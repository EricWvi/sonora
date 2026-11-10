package top.onlyquant.sonora.data.preferences

import javax.inject.Qualifier

/*
 * The constants below are copied from data/preferences/UserPreferencesRepository.kt of
 * PixelPlayerOSS v0.4.0-alpha.53 (GPL-3.0). The repository itself is not: the player reads its
 * defaults until the settings screen is copied.
 */

/**
 * The DataStore for settings ("settings", as in the original). The login session is kept apart,
 * in the "session" store.
 */
@Qualifier
@Retention(AnnotationRetention.BINARY)
annotation class SettingsDataStore

object ThemePreference {
    const val DEFAULT = "default"
    const val DYNAMIC = "dynamic"
    const val ALBUM_ART = "album_art"
    const val GLOBAL = "global"
}

object AppThemeMode {
    const val FOLLOW_SYSTEM = "follow_system"
    const val LIGHT = "light"
    const val DARK = "dark"
}

const val MIN_NAV_BAR_CORNER_RADIUS = 0
const val MAX_NAV_BAR_CORNER_RADIUS = 60

internal fun sanitizeNavBarCornerRadius(radius: Int): Int =
        radius.coerceIn(MIN_NAV_BAR_CORNER_RADIUS, MAX_NAV_BAR_CORNER_RADIUS)

/**
 * Album art quality settings for developer options.
 * Controls maximum resolution for album artwork in player view.
 * Thumbnails in lists always use low resolution for performance.
 *
 * @property maxSize Maximum size in pixels (0 = original size)
 * @property label Human-readable label for UI
 */
enum class AlbumArtQuality(val maxSize: Int, val label: String) {
    LOW(256, "Low (256px) - Better performance"),
    MEDIUM(512, "Medium (512px) - Balanced"),
    HIGH(800, "High (800px) - Best quality"),
    ORIGINAL(0, "Original - Maximum quality")
}
