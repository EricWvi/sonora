package top.onlyquant.sonora.utils

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update

/**
 * Reduced from AlbumArtUtils in PixelPlayerOSS v0.4.0-alpha.53 (GPL-3.0): only the artwork cache
 * version the player observes. The rest extracts and caches artwork embedded in local files, and
 * Sonora covers always come from the server (`covers/{name}`).
 */
object AlbumArtUtils {
    private val _artworkCacheVersion = MutableStateFlow(0L)
    val artworkCacheVersion = _artworkCacheVersion.asStateFlow()

    /** Publish only after file and Coil caches are cleared, so mounted images can reload safely. */
    internal fun notifyArtworkCacheInvalidated() {
        _artworkCacheVersion.update { it + 1L }
    }
}
