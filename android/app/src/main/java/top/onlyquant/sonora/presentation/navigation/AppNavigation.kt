package top.onlyquant.sonora.presentation.navigation

import androidx.annotation.OptIn
import androidx.compose.animation.EnterTransition
import androidx.compose.animation.ExitTransition
import androidx.compose.animation.core.CubicBezierEasing
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInHorizontally
import androidx.compose.animation.slideOutHorizontally
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.runtime.Composable
import androidx.compose.ui.unit.IntOffset
import androidx.media3.common.util.UnstableApi
import androidx.navigation.NavHostController
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.navArgument
import top.onlyquant.sonora.presentation.components.ScreenWrapper
import top.onlyquant.sonora.presentation.screens.AlbumDetailScreen
import top.onlyquant.sonora.presentation.screens.ArtistDetailScreen
import top.onlyquant.sonora.presentation.screens.HomeScreen
import top.onlyquant.sonora.presentation.screens.LibraryScreen
import top.onlyquant.sonora.presentation.screens.SearchScreen
import top.onlyquant.sonora.presentation.viewmodel.PlayerViewModel

/**
 * The navigation graph, from AppNavigation in PixelPlayerOSS v0.4.0-alpha.53 (GPL-3.0). Sonora:
 * only the destinations that exist so far, with the original's transitions: the three main tabs,
 * the album page and the artist page. The original starts on the launch tab from the settings;
 * Sonora always starts on Home, and there is no sidebar drawer (Eric, 2026-11-11), so Home has no
 * onOpenSidebar.
 */
@OptIn(UnstableApi::class)
@Composable
fun AppNavigation(
    playerViewModel: PlayerViewModel,
    navController: NavHostController,
    paddingValues: PaddingValues,
    onSignOut: () -> Unit,
) {
    NavHost(
        navController = navController,
        startDestination = Screen.Home.route
    ) {
        composable(
            Screen.Home.route,
            enterTransition = {
                mainRootEnterTransition(
                    fromRoute = initialState.destination.route,
                    toRoute = targetState.destination.route,
                    fallback = enterTransition()
                )
            },
            exitTransition = {
                mainRootExitTransition(
                    fromRoute = initialState.destination.route,
                    toRoute = targetState.destination.route,
                    fallback = exitTransition()
                )
            },
            popEnterTransition = {
                mainRootEnterTransition(
                    fromRoute = initialState.destination.route,
                    toRoute = targetState.destination.route,
                    fallback = popEnterTransition()
                )
            },
            popExitTransition = {
                mainRootExitTransition(
                    fromRoute = initialState.destination.route,
                    toRoute = targetState.destination.route,
                    fallback = popExitTransition()
                )
            },
        ) {
            ScreenWrapper(navController = navController, playerViewModel = playerViewModel) {
                HomeScreen(
                    navController = navController,
                    paddingValuesParent = paddingValues,
                    playerViewModel = playerViewModel,
                    onSignOut = onSignOut
                )
            }
        }
        composable(
            Screen.Search.route,
            enterTransition = {
                mainRootEnterTransition(
                    fromRoute = initialState.destination.route,
                    toRoute = targetState.destination.route,
                    fallback = enterTransition()
                )
            },
            exitTransition = {
                mainRootExitTransition(
                    fromRoute = initialState.destination.route,
                    toRoute = targetState.destination.route,
                    fallback = exitTransition()
                )
            },
            popEnterTransition = {
                mainRootEnterTransition(
                    fromRoute = initialState.destination.route,
                    toRoute = targetState.destination.route,
                    fallback = popEnterTransition()
                )
            },
            popExitTransition = {
                mainRootExitTransition(
                    fromRoute = initialState.destination.route,
                    toRoute = targetState.destination.route,
                    fallback = popExitTransition()
                )
            },
        ) {
            ScreenWrapper(navController = navController, playerViewModel = playerViewModel) {
                SearchScreen(
                    paddingValues = paddingValues,
                    playerViewModel = playerViewModel,
                    navController = navController,
                    onSignOut = onSignOut
                )
            }
        }
        composable(
            Screen.Library.route,
            enterTransition = {
                mainRootEnterTransition(
                    fromRoute = initialState.destination.route,
                    toRoute = targetState.destination.route,
                    fallback = enterTransition()
                )
            },
            exitTransition = {
                mainRootExitTransition(
                    fromRoute = initialState.destination.route,
                    toRoute = targetState.destination.route,
                    fallback = exitTransition()
                )
            },
            popEnterTransition = {
                mainRootEnterTransition(
                    fromRoute = initialState.destination.route,
                    toRoute = targetState.destination.route,
                    fallback = popEnterTransition()
                )
            },
            popExitTransition = {
                mainRootExitTransition(
                    fromRoute = initialState.destination.route,
                    toRoute = targetState.destination.route,
                    fallback = popExitTransition()
                )
            },
        ) {
            ScreenWrapper(navController = navController, playerViewModel = playerViewModel) {
                LibraryScreen(navController = navController, playerViewModel = playerViewModel)
            }
        }
        composable(
            route = Screen.AlbumDetail.route,
            arguments = listOf(navArgument("albumId") { type = NavType.StringType }),
            enterTransition = { enterTransition() },
            exitTransition = { exitTransition() },
            popEnterTransition = { popEnterTransition() },
            popExitTransition = { popExitTransition() },
        ) { backStackEntry ->
            val albumId = backStackEntry.arguments?.getString("albumId")
            if (albumId != null) {
                ScreenWrapper(navController = navController, playerViewModel = playerViewModel) {
                    AlbumDetailScreen(
                        albumId = albumId,
                        navController = navController,
                        playerViewModel = playerViewModel
                    )
                }
            }
        }
        composable(
            route = Screen.ArtistDetail.route,
            arguments = listOf(navArgument("artistId") { type = NavType.StringType }),
            enterTransition = { enterTransition() },
            exitTransition = { exitTransition() },
            popEnterTransition = { popEnterTransition() },
            popExitTransition = { popExitTransition() },
        ) { backStackEntry ->
            val artistId = backStackEntry.arguments?.getString("artistId")
            if (artistId != null) {
                ScreenWrapper(navController = navController, playerViewModel = playerViewModel) {
                    ArtistDetailScreen(
                        artistId = artistId,
                        navController = navController,
                        playerViewModel = playerViewModel
                    )
                }
            }
        }
    }
}

private enum class MainRootDirection {
    FORWARD,
    BACKWARD
}

private const val BOTTOM_NAV_TRANSITION_DURATION = 380

private val BottomNavEasing = CubicBezierEasing(0.2f, 0f, 0f, 1f)

private val MAIN_ROOT_TRANSITION_SPEC =
    tween<IntOffset>(durationMillis = BOTTOM_NAV_TRANSITION_DURATION, easing = BottomNavEasing)

private val MAIN_ROOT_FADE_SPEC =
    tween<Float>(durationMillis = BOTTOM_NAV_TRANSITION_DURATION / 2, easing = BottomNavEasing)

private fun mainRootDirection(
    fromRoute: String?,
    toRoute: String?
): MainRootDirection? {
    val fromIndex = mainRootRouteIndex(fromRoute) ?: return null
    val toIndex = mainRootRouteIndex(toRoute) ?: return null
    if (fromIndex == toIndex) return null
    return if (toIndex > fromIndex) MainRootDirection.FORWARD else MainRootDirection.BACKWARD
}

private fun mainRootEnterTransition(
    fromRoute: String?,
    toRoute: String?,
    fallback: EnterTransition
): EnterTransition = when (mainRootDirection(fromRoute, toRoute)) {
    MainRootDirection.FORWARD -> {
        slideInHorizontally(
            animationSpec = MAIN_ROOT_TRANSITION_SPEC,
            initialOffsetX = { (it * 0.5f).toInt() }
        ) + fadeIn(animationSpec = MAIN_ROOT_FADE_SPEC)
    }
    MainRootDirection.BACKWARD -> {
        slideInHorizontally(
            animationSpec = MAIN_ROOT_TRANSITION_SPEC,
            initialOffsetX = { -(it * 0.5f).toInt() }
        ) + fadeIn(animationSpec = MAIN_ROOT_FADE_SPEC)
    }
    null -> fallback
}

private fun mainRootExitTransition(
    fromRoute: String?,
    toRoute: String?,
    fallback: ExitTransition
): ExitTransition = when (mainRootDirection(fromRoute, toRoute)) {
    MainRootDirection.FORWARD -> {
        slideOutHorizontally(
            animationSpec = MAIN_ROOT_TRANSITION_SPEC,
            targetOffsetX = { -(it * 0.5f).toInt() }
        ) + fadeOut(animationSpec = MAIN_ROOT_FADE_SPEC)
    }
    MainRootDirection.BACKWARD -> {
        slideOutHorizontally(
            animationSpec = MAIN_ROOT_TRANSITION_SPEC,
            targetOffsetX = { (it * 0.5f).toInt() }
        ) + fadeOut(animationSpec = MAIN_ROOT_FADE_SPEC)
    }
    null -> fallback
}
