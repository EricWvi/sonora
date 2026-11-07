package top.onlyquant.sonora.data.service

import android.app.PendingIntent
import android.content.Intent
import androidx.media3.common.Player
import androidx.media3.session.DefaultMediaNotificationProvider
import androidx.media3.session.MediaSession
import androidx.media3.session.MediaSessionService
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject
import top.onlyquant.sonora.MainActivity
import top.onlyquant.sonora.R
import top.onlyquant.sonora.data.service.player.DualPlayerEngine

/**
 * Reduced from MusicService in PixelPlayerOSS v0.4.0-alpha.53 (GPL-3.0). Keeps the session wiring
 * around DualPlayerEngine; widgets, the equalizer, stats, queue persistence, sleep timer, custom
 * session commands and the custom notification layout are left out, and come back only when needed.
 */
@AndroidEntryPoint
class MusicService : MediaSessionService() {
    @Inject lateinit var engine: DualPlayerEngine

    private var mediaSession: MediaSession? = null

    private val playerSwapListener: (Player) -> Unit = { newPlayer ->
        mediaSession?.player = newPlayer
    }

    override fun onCreate() {
        super.onCreate()
        engine.initialize()
        engine.addPlayerSwapListener(playerSwapListener)
        mediaSession = MediaSession.Builder(this, engine.masterPlayer)
            .setSessionActivity(
                PendingIntent.getActivity(
                    this,
                    0,
                    Intent(this, MainActivity::class.java),
                    PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT,
                ),
            )
            .build()
        // As in the original: the monochrome player glyph in the status bar.
        setMediaNotificationProvider(
            DefaultMediaNotificationProvider.Builder(this).build().also {
                it.setSmallIcon(R.drawable.monochrome_player)
            },
        )
    }

    override fun onGetSession(controllerInfo: MediaSession.ControllerInfo): MediaSession? = mediaSession

    override fun onDestroy() {
        engine.removePlayerSwapListener(playerSwapListener)
        mediaSession?.release()
        mediaSession = null
        engine.release()
        super.onDestroy()
    }
}
