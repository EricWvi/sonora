package top.onlyquant.sonora.data.service

import android.app.PendingIntent
import android.content.Intent
import android.os.Bundle
import androidx.media3.common.MediaItem
import androidx.media3.common.MediaMetadata
import androidx.media3.common.Player
import androidx.media3.session.DefaultMediaNotificationProvider
import androidx.media3.session.MediaSession
import androidx.media3.session.MediaSessionService
import androidx.media3.session.SessionCommand
import androidx.media3.session.SessionCommands
import androidx.media3.session.SessionError
import androidx.media3.session.SessionResult
import com.google.common.util.concurrent.Futures
import com.google.common.util.concurrent.ListenableFuture
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.launch
import kotlin.math.abs
import top.onlyquant.sonora.data.preferences.UserPreferencesRepository
import top.onlyquant.sonora.MainActivity
import top.onlyquant.sonora.R
import top.onlyquant.sonora.data.service.player.DualPlayerEngine

/**
 * Reduced from MusicService in PixelPlayerOSS v0.4.0-alpha.53 (GPL-3.0). Keeps the session wiring
 * around DualPlayerEngine, the sleep timer (its custom session commands and player events) and
 * ReplayGain, which applies the server's loudness gain;
 * widgets, the equalizer, stats, queue persistence, the other custom session commands and the
 * custom notification layout are left out, and come back only when needed.
 */
@AndroidEntryPoint
class MusicService : MediaSessionService() {
    @Inject lateinit var engine: DualPlayerEngine
    @Inject lateinit var sleepTimerSession: SleepTimerSession
    @Inject lateinit var userPreferencesRepository: UserPreferencesRepository

    private var mediaSession: MediaSession? = null
    private val serviceScope = CoroutineScope(SupervisorJob() + Dispatchers.Main)
    private var userPlaybackSpeed = 1f

    private val sleepTimerController by lazy {
        SleepTimerController(
            engine = engine,
            session = sleepTimerSession,
            scope = serviceScope,
            alarmScheduler = AlarmManagerSleepTimerScheduler(this),
            sessionPlayer = { mediaSession?.player ?: engine.masterPlayer },
        )
    }

    private val replayGainProcessor by lazy {
        ReplayGainProcessor(
            engine = engine,
            currentSessionMediaItem = { mediaSession?.player?.currentMediaItem },
        )
    }

    /** The player [playerListener] is on; the engine swaps players for crossfades. */
    private var listenedPlayer: Player? = null

    private val playerSwapListener: (Player) -> Unit = { newPlayer ->
        mediaSession?.player = newPlayer
        listenedPlayer?.removeListener(playerListener)
        newPlayer.addListener(playerListener)
        listenedPlayer = newPlayer
        replayGainProcessor.prepareForTransition(newPlayer)
        applyPlaybackSpeed(newPlayer)
    }

    /*
     * Sonora: the original also publishes the incoming crossfade player to the session here;
     * that comes with TransitionController, which is not copied yet (nothing starts a crossfade).
     */
    private val transitionDisplayPlayerListener: (Player) -> Unit = { displayPlayer ->
        replayGainProcessor.prepareForTransition(displayPlayer)
    }

    private val transitionFinishedListener: () -> Unit = {
        replayGainProcessor.onTransitionFinished()
    }

    private val playerListener = object : Player.Listener {
        override fun onVolumeChanged(volume: Float) {
            replayGainProcessor.onPlayerVolumeChanged(volume)
        }

        override fun onIsPlayingChanged(isPlaying: Boolean) {
            if (isPlaying) {
                replayGainProcessor.reapplyLastAppliedVolume(mediaSession?.player ?: engine.masterPlayer)
            }
        }

        override fun onPlaybackStateChanged(playbackState: Int) {
            if (playbackState == Player.STATE_ENDED) sleepTimerController.onPlaybackEnded()
        }

        override fun onPositionDiscontinuity(
            oldPosition: Player.PositionInfo,
            newPosition: Player.PositionInfo,
            reason: Int,
        ) {
            if (reason == Player.DISCONTINUITY_REASON_AUTO_TRANSITION ||
                reason == Player.DISCONTINUITY_REASON_SEEK
            ) {
                val oldMediaId = oldPosition.mediaItem?.mediaId
                val newMediaId = newPosition.mediaItem?.mediaId
                if (oldMediaId != null && oldMediaId == newMediaId) {
                    replayGainProcessor.reapplyLastAppliedVolume(engine.masterPlayer)
                } else {
                    replayGainProcessor.apply(mediaSession?.player?.currentMediaItem)
                }
            }
        }

        override fun onMediaItemTransition(mediaItem: MediaItem?, reason: Int) {
            sleepTimerController.onMediaItemTransition(mediaItem, reason)
            replayGainProcessor.apply(mediaSession?.player?.currentMediaItem)
        }

        override fun onMediaMetadataChanged(mediaMetadata: MediaMetadata) {
            val activePlayer = mediaSession?.player ?: engine.masterPlayer
            replayGainProcessor.onMediaMetadataChanged(activePlayer.currentMediaItem)
        }

        override fun onRepeatModeChanged(repeatMode: Int) {
            sleepTimerController.onRepeatModeChanged(repeatMode)
        }
    }

    private val sessionCallback = object : MediaSession.Callback {
        override fun onConnect(
            session: MediaSession,
            controller: MediaSession.ControllerInfo,
        ): MediaSession.ConnectionResult {
            val defaultResult = super.onConnect(session, controller)
            val commands = SessionCommands.Builder()
                .addSessionCommands(defaultResult.availableSessionCommands.commands)
            if (isPrivilegedController(controller)) {
                SLEEP_TIMER_COMMANDS.forEach { commands.add(SessionCommand(it, Bundle.EMPTY)) }
            }
            return MediaSession.ConnectionResult.accept(commands.build(), defaultResult.availablePlayerCommands)
        }

        override fun onCustomCommand(
            session: MediaSession,
            controller: MediaSession.ControllerInfo,
            customCommand: SessionCommand,
            args: Bundle,
        ): ListenableFuture<SessionResult> {
            if (!isPrivilegedController(controller)) {
                return Futures.immediateFuture(SessionResult(SessionError.ERROR_PERMISSION_DENIED))
            }
            when (customCommand.customAction) {
                MusicNotificationProvider.CUSTOM_COMMAND_COUNTED_PLAY ->
                    sleepTimerController.startCountedPlay(args.getInt("count", 1))
                MusicNotificationProvider.CUSTOM_COMMAND_CANCEL_COUNTED_PLAY ->
                    sleepTimerController.stopCountedPlay()
                MusicNotificationProvider.CUSTOM_COMMAND_SET_SLEEP_TIMER_DURATION ->
                    sleepTimerController.setDurationTimer(
                        args.getInt(MusicNotificationProvider.EXTRA_SLEEP_TIMER_MINUTES, 0),
                    )
                MusicNotificationProvider.CUSTOM_COMMAND_SET_SLEEP_TIMER_END_OF_TRACK ->
                    sleepTimerController.setEndOfTrackTimer(
                        args.getBoolean(MusicNotificationProvider.EXTRA_END_OF_TRACK_ENABLED, true),
                    )
                MusicNotificationProvider.CUSTOM_COMMAND_CANCEL_SLEEP_TIMER ->
                    sleepTimerController.cancelTimer()
            }
            return Futures.immediateFuture(SessionResult(SessionResult.RESULT_SUCCESS))
        }
    }

    override fun onCreate() {
        super.onCreate()
        engine.initialize()
        replayGainProcessor.captureUserVolume(engine.masterPlayer.volume)
        engine.addPlayerSwapListener(playerSwapListener)
        engine.addTransitionDisplayPlayerListener(transitionDisplayPlayerListener)
        engine.addTransitionFinishedListener(transitionFinishedListener)
        engine.masterPlayer.addListener(playerListener)
        listenedPlayer = engine.masterPlayer
        mediaSession = MediaSession.Builder(this, engine.masterPlayer)
            .setCallback(sessionCallback)
            // As in the original: notification artwork through Coil, so covers carry the device token.
            .setBitmapLoader(CoilBitmapLoader(this, serviceScope))
            .setSessionActivity(
                PendingIntent.getActivity(
                    this,
                    0,
                    Intent(this, MainActivity::class.java),
                    PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT,
                ),
            )
            .build()
        sleepTimerController.attach()
        // As in the original: the speed set in the queue sheet, re-applied on every player swap.
        serviceScope.launch {
            userPreferencesRepository.playbackSpeedFlow.collect { speed ->
                userPlaybackSpeed = speed
                applyPlaybackSpeed(mediaSession?.player ?: engine.masterPlayer)
            }
        }
        // As in the original: the switch takes effect on the current song at once.
        serviceScope.launch {
            userPreferencesRepository.replayGainEnabledFlow.collect { enabled ->
                replayGainProcessor.setEnabled(enabled)
                replayGainProcessor.apply(mediaSession?.player?.currentMediaItem)
            }
        }
        // As in the original: the monochrome player glyph in the status bar.
        setMediaNotificationProvider(
            DefaultMediaNotificationProvider.Builder(this).build().also {
                it.setSmallIcon(R.drawable.monochrome_player)
            },
        )
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        if (intent?.action == ACTION_SLEEP_TIMER_EXPIRED) sleepTimerController.onDurationExpired()
        return super.onStartCommand(intent, flags, startId)
    }

    override fun onGetSession(controllerInfo: MediaSession.ControllerInfo): MediaSession? = mediaSession

    override fun onDestroy() {
        sleepTimerController.release()
        listenedPlayer?.removeListener(playerListener)
        listenedPlayer = null
        engine.removePlayerSwapListener(playerSwapListener)
        engine.removeTransitionDisplayPlayerListener(transitionDisplayPlayerListener)
        engine.removeTransitionFinishedListener(transitionFinishedListener)
        mediaSession?.release()
        mediaSession = null
        engine.release()
        serviceScope.cancel()
        super.onDestroy()
    }

    private fun applyPlaybackSpeed(player: Player) {
        if (abs(player.playbackParameters.speed - userPlaybackSpeed) > 0.001f) {
            player.setPlaybackSpeed(userPlaybackSpeed)
        }
    }

    /** This app, and trusted controllers such as System UI. */
    private fun isPrivilegedController(controller: MediaSession.ControllerInfo): Boolean =
        controller.packageName == packageName || controller.isTrusted

    companion object {
        const val ACTION_SLEEP_TIMER_EXPIRED = "top.onlyquant.sonora.ACTION_SLEEP_TIMER_EXPIRED"

        private val SLEEP_TIMER_COMMANDS = listOf(
            MusicNotificationProvider.CUSTOM_COMMAND_COUNTED_PLAY,
            MusicNotificationProvider.CUSTOM_COMMAND_CANCEL_COUNTED_PLAY,
            MusicNotificationProvider.CUSTOM_COMMAND_SET_SLEEP_TIMER_DURATION,
            MusicNotificationProvider.CUSTOM_COMMAND_SET_SLEEP_TIMER_END_OF_TRACK,
            MusicNotificationProvider.CUSTOM_COMMAND_CANCEL_SLEEP_TIMER,
        )
    }
}
