package top.onlyquant.sonora.presentation.viewmodel

import android.content.Context
import android.graphics.Bitmap
import android.graphics.Canvas
import android.util.LruCache
import androidx.compose.material3.ColorScheme
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.toArgb
import androidx.core.graphics.createBitmap
import coil.imageLoader
import coil.request.CachePolicy
import coil.request.ImageRequest
import coil.size.Size
import top.onlyquant.sonora.data.preferences.AlbumArtColorAccuracy
import top.onlyquant.sonora.data.preferences.AlbumArtPaletteStyle
import top.onlyquant.sonora.utils.LocalArtworkUri
import top.onlyquant.sonora.utils.traceAsyncSection
import top.onlyquant.sonora.ui.theme.clearExtractedColorCache
import top.onlyquant.sonora.ui.theme.extractSeedColor
import top.onlyquant.sonora.ui.theme.generateColorSchemeFromSeed
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import kotlinx.coroutines.withContext
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Sonora: the original also stores schemes in Room (AlbumArtThemeDao); until the app has a
 * database (the library step of phase 8) they are kept only in the in-memory LRU cache below.
 *
 * Efficient color scheme processor for album art.
 * 
 * Optimizations:
 * - In-memory LRU cache to avoid disk reads for recently accessed schemes
 * - Mutex-protected processing to avoid duplicate work
 * - Batched bitmap operations on IO dispatcher
 * - Reduced bitmap size (128x128) for faster processing
 * 
 * Extracted from PlayerViewModel to improve modularity.
 */
@Singleton
class ColorSchemeProcessor @Inject constructor(
    @ApplicationContext private val context: Context,
) {
    private val memoryCache = LruCache<String, ColorSchemePair>(20)
    private val processingMutex = Mutex()
    private val inProgressUris = mutableSetOf<String>()
    private val localArtworkCacheMutex = Mutex()
    @Volatile
    private var localArtworkGeneration = 0L

    /**
     * Channel for queuing color scheme requests.
     * Used by PlayerViewModel for background processing.
     * Capacity is bounded with DROP_OLDEST to prevent unbounded growth during rapid
     * track changes (e.g. fast seek through a large playlist).
     */
    val requestChannel = Channel<String>(capacity = 32, onBufferOverflow = kotlinx.coroutines.channels.BufferOverflow.DROP_OLDEST)

    /**
     * Gets or generates a color scheme for the given album art URI.
     * Checks memory cache first, then database, then generates new.
     * All heavy operations are performed on appropriate dispatchers.
     */
    /**
     * Gets or generates a color scheme for the given album art URI.
     * Checks memory cache first, then database, then generates new.
     * @param forceRefresh If true, bypasses caches and forces regeneration from source image.
     */
    suspend fun getOrGenerateColorScheme(
        albumArtUri: String,
        paletteStyle: AlbumArtPaletteStyle,
        colorAccuracyLevel: Int = AlbumArtColorAccuracy.DEFAULT,
        forceRefresh: Boolean = false
    ): ColorSchemePair? = traceAsyncSection("ColorSchemeProcessor.getOrGenerate") {
        val resolvedAccuracyLevel = AlbumArtColorAccuracy.clamp(colorAccuracyLevel)
        val generation = localArtworkGeneration
        if (!forceRefresh) {
            loadCachedColorScheme(
                albumArtUri = albumArtUri,
                paletteStyle = paletteStyle,
                colorAccuracyLevel = resolvedAccuracyLevel,
                expectedLocalGeneration = generation
            )?.let { schemePair ->
                return@traceAsyncSection schemePair
            }
        }

        generateAndCacheColorScheme(
            albumArtUri = albumArtUri,
            paletteStyle = paletteStyle,
            colorAccuracyLevel = resolvedAccuracyLevel,
            forceRefresh = forceRefresh,
            expectedLocalGeneration = generation
        )
    }

    suspend fun getPreviewColorScheme(
        albumArtUri: String,
        paletteStyle: AlbumArtPaletteStyle,
        colorAccuracyLevel: Int = AlbumArtColorAccuracy.DEFAULT
    ): ColorSchemePair? {
        val resolvedAccuracyLevel = AlbumArtColorAccuracy.clamp(colorAccuracyLevel)
        val generation = localArtworkGeneration
        return loadCachedColorScheme(
            albumArtUri = albumArtUri,
            paletteStyle = paletteStyle,
            colorAccuracyLevel = resolvedAccuracyLevel,
            expectedLocalGeneration = generation
        ) ?: generateAndCacheColorScheme(
            albumArtUri = albumArtUri,
            paletteStyle = paletteStyle,
            colorAccuracyLevel = resolvedAccuracyLevel,
            expectedLocalGeneration = generation
        )
    }

    /**
     * Generates a color scheme from the album art bitmap.
     * All processing done on Default dispatcher for CPU-bound work.
     */
    private suspend fun generateAndCacheColorScheme(
        albumArtUri: String,
        paletteStyle: AlbumArtPaletteStyle,
        colorAccuracyLevel: Int,
        forceRefresh: Boolean = false,
        expectedLocalGeneration: Long
    ): ColorSchemePair? = traceAsyncSection("ColorSchemeProcessor.generate") {
        try {
            val cacheKey = buildCacheKey(albumArtUri, paletteStyle, colorAccuracyLevel)
            val bitmap = withContext(Dispatchers.IO) {
                loadBitmapForColorExtraction(albumArtUri, forceRefresh)
            } ?: return@traceAsyncSection null

            val schemePair = try {
                withContext(Dispatchers.Default) {
                    val seed = extractSeedColor(
                        bitmap = bitmap,
                        config = top.onlyquant.sonora.ui.theme.ColorExtractionConfig(
                            accuracyLevel = colorAccuracyLevel
                        )
                    )
                    generateColorSchemeFromSeed(seedColor = seed, paletteStyle = paletteStyle)
                }
            } finally {
                bitmap.recycle()
            }

            suspend fun cacheScheme(): ColorSchemePair {
                memoryCache.put(cacheKey, schemePair)
                return schemePair
            }

            if (LocalArtworkUri.isLocalArtworkUri(albumArtUri)) {
                localArtworkCacheMutex.withLock {
                    if (expectedLocalGeneration != localArtworkGeneration) null else cacheScheme()
                }
            } else {
                cacheScheme()
            }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            null
        }
    }

    /**
     * Loads a small bitmap optimized for color extraction.
     */
    private suspend fun loadBitmapForColorExtraction(uri: String, skipCache: Boolean): Bitmap? {
        return try {
            val cachePolicy = if (skipCache) CachePolicy.DISABLED else CachePolicy.ENABLED
            val diskCachePolicy = if (LocalArtworkUri.isLocalArtworkUri(uri)) CachePolicy.DISABLED else cachePolicy
            
            val request = ImageRequest.Builder(context)
                .data(uri)
                .allowHardware(false)
                .size(Size(128, 128))
                .bitmapConfig(Bitmap.Config.ARGB_8888)
                .memoryCachePolicy(cachePolicy)
                .diskCachePolicy(diskCachePolicy)
                .build()
            
            val drawable = context.imageLoader.execute(request).drawable ?: return null
            
            createBitmap(
                drawable.intrinsicWidth.coerceAtLeast(1),
                drawable.intrinsicHeight.coerceAtLeast(1)
            ).also { bmp ->
                Canvas(bmp).let { canvas ->
                    drawable.setBounds(0, 0, canvas.width, canvas.height)
                    drawable.draw(canvas)
                }
            }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            null
        }
    }

    /**
     * Checks if a URI is currently being processed.
     * Used to avoid duplicate work.
     */
    suspend fun markProcessing(uri: String): Boolean {
        return processingMutex.withLock {
            if (inProgressUris.contains(uri)) {
                false
            } else {
                inProgressUris.add(uri)
                true
            }
        }
    }

    /**
     * Marks a URI as finished processing.
     */
    suspend fun markComplete(uri: String) {
        processingMutex.withLock {
            inProgressUris.remove(uri)
        }
    }

    /**
     * Clears the in-memory cache.
     * Call when memory is low or on configuration changes.
     */
    fun clearMemoryCache() {
        memoryCache.evictAll()
    }

    /**
     * Removes a specific URI from the cache.
     */
    fun evictFromCache(uri: String) {
        removeUriFromMemoryCache(uri)
    }

    /**
     * Invalidates the color scheme for a URI in both memory and database.
     */
    suspend fun invalidateScheme(uri: String) {
        clearExtractedColorCache()
        removeUriFromMemoryCache(uri)
    }

    /** Retire only local palettes; extraction resumes lazily for visible/current artwork. */
    suspend fun invalidateLocalArtworkSchemes() {
        localArtworkCacheMutex.withLock {
            localArtworkGeneration += 1
            memoryCache.snapshot().keys
                .filter(LocalArtworkUri::isLocalArtworkUri)
                .forEach { memoryCache.remove(it) }
            clearExtractedColorCache()
        }
    }

    private fun removeUriFromMemoryCache(uri: String) {
        val prefix = "$uri$CACHE_KEY_SEPARATOR"
        memoryCache.snapshot().keys
            .filter { key -> key == uri || key.startsWith(prefix) }
            .forEach { key -> memoryCache.remove(key) }
    }

    private fun buildCacheKey(
        uri: String,
        paletteStyle: AlbumArtPaletteStyle,
        colorAccuracyLevel: Int
    ): String {
        return "$uri$CACHE_KEY_SEPARATOR${paletteStyleCacheKey(paletteStyle, colorAccuracyLevel)}"
    }

    private fun paletteStyleCacheKey(
        paletteStyle: AlbumArtPaletteStyle,
        colorAccuracyLevel: Int
    ): String {
        return buildString {
            append(paletteStyle.storageKey)
            append(CACHE_KEY_SEPARATOR)
            append("accuracy_")
            append(AlbumArtColorAccuracy.clamp(colorAccuracyLevel))
            append(CACHE_KEY_SEPARATOR)
            append(CACHE_ALGORITHM_VERSION)
        }
    }

    private suspend fun loadCachedColorScheme(
        albumArtUri: String,
        paletteStyle: AlbumArtPaletteStyle,
        colorAccuracyLevel: Int,
        expectedLocalGeneration: Long
    ): ColorSchemePair? {
        val cacheKey = buildCacheKey(albumArtUri, paletteStyle, colorAccuracyLevel)

        suspend fun loadScheme(): ColorSchemePair? {
            return memoryCache.get(cacheKey)
        }

        return if (LocalArtworkUri.isLocalArtworkUri(albumArtUri)) {
            localArtworkCacheMutex.withLock {
                if (expectedLocalGeneration != localArtworkGeneration) null else loadScheme()
            }
        } else {
            loadScheme()
        }
    }

    companion object {
        private const val TAG = "ColorSchemeProcessor"
        private const val CACHE_KEY_SEPARATOR = "|"
        private const val CACHE_ALGORITHM_VERSION = "algo_v7"
    }
}
