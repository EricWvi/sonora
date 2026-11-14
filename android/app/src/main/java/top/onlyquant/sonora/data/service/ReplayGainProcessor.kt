package top.onlyquant.sonora.data.service

import androidx.media3.common.MediaItem
import androidx.media3.common.Player
import top.onlyquant.sonora.data.service.player.DualPlayerEngine
import top.onlyquant.sonora.utils.MediaItemBuilder
import timber.log.Timber
import kotlin.math.abs
import kotlin.math.pow

/**
 * The player volume for a song's loudness normalization gain, or null when no normalization
 * applies and the user-selected volume stays: the switch is off, the server has not measured
 * the song ([gainDb] null), or the song is online (online songs are not normalized,
 * docs/api.md#音量均衡). The gain is 10^(gainDb/20), capped at 1 because a player's volume
 * cannot go above it; only a few songs need a boost, by a few dB at most.
 */
internal fun normalizationVolume(enabled: Boolean, gainDb: Double?, online: Boolean): Float? {
    if (!enabled || online || gainDb == null || gainDb.isNaN()) return null
    return 10.0.pow(gainDb / 20.0).toFloat().coerceIn(0f, 1f)
}

/**
 * Copied from ReplayGainProcessor in PixelPlayerOSS v0.4.0-alpha.53 (GPL-3.0).
 *
 * Owns all ReplayGain volume-normalization state and logic, extracted from
 * [MusicService].
 *
 * The processor encapsulates the RG enable flag, the "expected vs. user-selected" volume
 * bookkeeping that keeps a programmatic RG volume change from being mistaken for a user gesture,
 * and the crossfade-aware application of the computed gain. It reaches the players only
 * through the injected [DualPlayerEngine]; the service supplies the current
 * session media item via [currentSessionMediaItem].
 *
 * Sonora: the original reads ReplayGain tags from the local file on an IO thread
 * (ReplayGainManager, with a cache, prefetching and stale-result checks) and has an album-gain
 * switch. Here the gain is the server's per-song `gainDb`, already on the media item
 * ([MediaItemBuilder.EXTERNAL_EXTRA_GAIN_DB]), so it is applied at once; the server does not
 * tell album and track gain apart, so there is no album-gain switch.
 *
 * All public methods are main-thread affine, matching the original in-service behaviour.
 */
class ReplayGainProcessor(
    private val engine: DualPlayerEngine,
    private val currentSessionMediaItem: () -> MediaItem?,
) {
    private companion object {
        private const val TAG = "ReplayGain"
    }

    private var enabled = false

    // Volume the user last chose by hand; restored whenever RG is disabled.
    private var userSelectedVolume = 1f
    // Volume we just wrote programmatically — used to ignore the echoed onVolumeChanged.
    private var expectedVolume: Float? = null
    // RG volume computed mid-crossfade, applied once the transition finishes.
    private var pendingVolume: Float? = null
    // Last successfully applied RG volume, re-applied on resume, same-track seeks and
    // queue edits.
    private var lastAppliedVolume: Float? = null
    // MediaId for which lastAppliedVolume was computed.
    private var lastMediaId: String? = null

    fun setEnabled(value: Boolean) {
        enabled = value
    }

    /** Seeds the user-selected volume from the player's current volume on startup. */
    fun captureUserVolume(volume: Float) {
        userSelectedVolume = volume.coerceIn(0f, 1f)
    }

    /**
     * Mirrors [Player.Listener.onVolumeChanged]: distinguishes a programmatic RG
     * volume change (which we ignore) from a genuine user gesture (which updates
     * [userSelectedVolume]).
     */
    fun onPlayerVolumeChanged(volume: Float) {
        if (engine.isTransitionRunning()) return
        val expected = expectedVolume
        if (expected != null && abs(expected - volume) < 0.001f) {
            expectedVolume = null
            return
        }
        expectedVolume = null
        userSelectedVolume = volume.coerceIn(0f, 1f)
    }

    private fun setPlayerVolume(player: Player, volume: Float) {
        val clampedVolume = volume.coerceIn(0f, 1f)
        expectedVolume = clampedVolume
        player.volume = clampedVolume
    }

    /**
     * Re-applies the last computed RG volume immediately unless a
     * crossfade is running, preventing a brief full-volume spike on resume,
     * same-track seeks, and queue edits.
     */
    fun reapplyLastAppliedVolume(player: Player) {
        if (engine.isTransitionRunning()) return
        lastAppliedVolume?.let { setPlayerVolume(player, it) }
    }

    /**
     * Computes ReplayGain for the incoming crossfade track and seeds
     * [DualPlayerEngine.incomingTrackReplayGainVolume] so the fade loop ends at the correct
     * volume, then runs [apply], which (since the transition is running) stores the result as a
     * pending volume.
     */
    fun prepareForTransition(player: Player) {
        val incomingItem = player.currentMediaItem
        if (enabled && incomingItem != null) {
            engine.incomingTrackReplayGainVolume = targetVolumeFor(incomingItem)
        }
        apply(incomingItem)
    }

    /** Applies ReplayGain volume normalization to [mediaItem] on the master player. */
    fun apply(mediaItem: MediaItem?) {
        if (mediaItem == null) {
            return
        }

        if (!enabled) {
            pendingVolume = null
            // Sonora: the original keeps these, so resuming after turning RG off puts the old
            // gain back (reapplyLastAppliedVolume); clearing them keeps the user volume.
            lastAppliedVolume = null
            lastMediaId = null
            if (!engine.isTransitionRunning()) {
                setPlayerVolume(engine.masterPlayer, userSelectedVolume)
            }
            return
        }

        val mediaId = mediaItem.mediaId
        val volume = targetVolumeFor(mediaItem)

        if (engine.isTransitionRunning()) {
            // Store for application after transition completes.
            // Also pass to engine so the crossfade loop ends at the correct RG
            // volume instead of hard-coding 1f, preventing the audible jump.
            pendingVolume = volume
            engine.incomingTrackReplayGainVolume = volume
            Timber.tag(TAG).d("ReplayGain: Stored pending volume=%.4f for %s (transition running)",
                volume, mediaItem.mediaMetadata.title
            )
        } else {
            pendingVolume = null
            engine.incomingTrackReplayGainVolume = null
            lastAppliedVolume = volume
            lastMediaId = mediaId
            setPlayerVolume(engine.masterPlayer, volume)
            Timber.tag(TAG).d("ReplayGain: Applied volume=%.4f for %s",
                volume, mediaItem.mediaMetadata.title
            )
        }
    }

    /**
     * The volume [mediaItem] plays at with RG on: its normalization gain, or the user-selected
     * volume when it has none (the original falls back the same way for tracks without a file).
     */
    private fun targetVolumeFor(mediaItem: MediaItem): Float {
        val extras = mediaItem.mediaMetadata.extras
        val gainDb = extras?.takeIf { it.containsKey(MediaItemBuilder.EXTERNAL_EXTRA_GAIN_DB) }
            ?.getDouble(MediaItemBuilder.EXTERNAL_EXTRA_GAIN_DB)
        // Sonora: the Android app has no online songs yet; pass true for them once it does.
        return normalizationVolume(enabled, gainDb, online = false) ?: userSelectedVolume
    }

    /**
     * Applies the volume that was held back while a crossfade was running, or
     * triggers a fresh computation if none was pending.
     */
    fun onTransitionFinished() {
        val player = engine.masterPlayer
        val pending = pendingVolume
        pendingVolume = null

        if (!enabled) {
            setPlayerVolume(player, userSelectedVolume)
            Timber.tag(TAG).d("ReplayGain: Transition finished, RG disabled — restored userSelectedVolume=%.2f", userSelectedVolume)
            return
        }

        if (pending != null) {
            // The crossfade loop ramps to this value; apply it now as the stable post-fade volume.
            // Also update lastAppliedVolume so any subsequent onPositionDiscontinuity
            // (REASON_AUTO_TRANSITION fires right after crossfade ends) uses this value
            // immediately.
            lastAppliedVolume = pending
            // Sonora: also note the track, so a metadata change right after does not redo it.
            lastMediaId = currentSessionMediaItem()?.mediaId
            setPlayerVolume(player, pending)
            Timber.tag(TAG).d("ReplayGain: Transition finished, applied pending volume=%.4f", pending)
        } else {
            // No pending volume was computed during transition, trigger full computation
            apply(currentSessionMediaItem())
            Timber.tag(TAG).d("ReplayGain: Transition finished, no pending volume — triggering full recomputation")
        }
    }

    /**
     * Recomputes RG only when the track actually changed; otherwise re-applies the
     * last known volume. [onMediaMetadataChanged] also fires on queue edits without a
     * track change.
     */
    fun onMediaMetadataChanged(currentItem: MediaItem?) {
        val currentMediaId = currentItem?.mediaId ?: return
        if (currentMediaId != lastMediaId) {
            apply(currentItem)
        } else {
            reapplyLastAppliedVolume(engine.masterPlayer)
        }
    }
}
