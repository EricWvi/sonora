package top.onlyquant.sonora

import androidx.annotation.DrawableRes
import androidx.annotation.StringRes
import kotlinx.collections.immutable.ImmutableList
import kotlinx.collections.immutable.persistentListOf
import top.onlyquant.sonora.presentation.navigation.Screen

// From MainActivity.kt of PixelPlayerOSS v0.4.0-alpha.53 (GPL-3.0), where these sit next to the
// activity; split out so the copied navigation bar and rail resolve them in the same package.

data class BottomNavItem(
    /** Shown as the tab label and content description, so it follows the app language. */
    @StringRes val labelResId: Int,
    @DrawableRes val iconResId: Int,
    @DrawableRes val selectedIconResId: Int? = null,
    val screen: Screen
)

/** Top-level destinations shared by the bottom navigation bar and the navigation rail. */
internal val MainNavItems: ImmutableList<BottomNavItem> = persistentListOf(
    BottomNavItem(R.string.tab_home, R.drawable.rounded_home_24, R.drawable.home_24_rounded_filled, Screen.Home),
    BottomNavItem(R.string.search, R.drawable.rounded_search_24, R.drawable.rounded_search_24, Screen.Search),
    BottomNavItem(R.string.tab_library, R.drawable.rounded_library_music_24, R.drawable.round_library_music_24, Screen.Library)
)
