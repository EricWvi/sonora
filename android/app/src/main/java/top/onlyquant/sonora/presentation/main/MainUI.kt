package top.onlyquant.sonora.presentation.main

import android.os.Build
import androidx.compose.animation.core.LinearOutSlowInEasing
import androidx.compose.animation.core.animateDpAsState
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutVertically
import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.asPaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBars
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.NavigationBarDefaults
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Surface
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clipToBounds
import androidx.compose.ui.graphics.RectangleShape
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.hapticfeedback.HapticFeedback
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.platform.LocalHapticFeedback
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.lerp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavHostController
import androidx.navigation.compose.currentBackStackEntryAsState
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.flow.map
import racra.compose.smooth_corner_rect_library.AbsoluteSmoothCornerShape
import top.onlyquant.sonora.MainNavItems
import top.onlyquant.sonora.data.preferences.NavBarStyle
import top.onlyquant.sonora.presentation.components.DismissUndoBar
import top.onlyquant.sonora.presentation.components.MiniPlayerBottomSpacer
import top.onlyquant.sonora.presentation.components.MiniPlayerHeight
import top.onlyquant.sonora.presentation.components.PlayerInternalNavigationBar
import top.onlyquant.sonora.presentation.components.UnifiedPlayerSheetV2
import top.onlyquant.sonora.presentation.components.calculatePlayerSheetCollapsedTargetY
import top.onlyquant.sonora.presentation.components.resolveNavBarOccupiedHeight
import top.onlyquant.sonora.presentation.components.resolveNavBarSurfaceHeight
import top.onlyquant.sonora.presentation.components.sanitizeNavigationBarBottomInset
import top.onlyquant.sonora.presentation.utils.AppHapticsConfig
import top.onlyquant.sonora.presentation.utils.LocalAppHapticsConfig
import top.onlyquant.sonora.presentation.viewmodel.PlayerViewModel
import top.onlyquant.sonora.ui.theme.LocalShowScrollbar

private data class DismissUndoBarSlice(
    val isVisible: Boolean = false,
    val durationMillis: Long = 4000L
)

/**
 * Reduced from MainActivity.MainUI in PixelPlayerOSS v0.4.0-alpha.53 (GPL-3.0): the bottom
 * navigation bar, the scrim and the player sheet over the screens, and the undo bar for a dismissed
 * queue, with the original's layout values.
 *
 * Left out for now: the navigation rail on wide screens (Sonora targets phones; PKB110), the
 * sidebar drawer (opened from the home screen, which comes later), hiding the bar on detail and
 * settings routes and while searching (those screens come later), and SmartMix. The nav bar
 * settings read their defaults (style DEFAULT, not compact, 32 dp corners, smooth corners,
 * haptics on) until the settings screen is copied.
 */
@Composable
fun MainUI(
    playerViewModel: PlayerViewModel,
    navController: NavHostController,
    content: @Composable (androidx.compose.foundation.layout.PaddingValues) -> Unit,
) {
    val navBackStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = navBackStackEntry?.destination?.route

    val navBarStyle = NavBarStyle.DEFAULT
    val navBarCompactMode = false
    val navBarCornerRadius = 32
    val useSmoothCorners = true
    val isMiniPlayerDismissing by playerViewModel.isMiniPlayerDismissing.collectAsStateWithLifecycle()
    val appHapticsConfig = remember { AppHapticsConfig(enabled = true) }
    val scopedHapticFeedback: HapticFeedback = LocalHapticFeedback.current

    val systemNavBarInset = sanitizeNavigationBarBottomInset(
        WindowInsets.navigationBars.asPaddingValues().calculateBottomPadding()
    )

    val horizontalPadding = if (navBarStyle == NavBarStyle.DEFAULT) {
        if (systemNavBarInset > 30.dp) 14.dp else systemNavBarInset
    } else {
        0.dp
    }
    val animatedBottomBarPadding by animateDpAsState(
        targetValue = if (navBarStyle == NavBarStyle.FULL_WIDTH) 0.dp else systemNavBarInset,
        animationSpec = tween(400),
        label = "BottomBarPadding"
    )
    val bottomBarPadding = animatedBottomBarPadding
    val navBarHeight = resolveNavBarSurfaceHeight(navBarStyle, systemNavBarInset, navBarCompactMode)
    val navBarOccupiedHeight = resolveNavBarOccupiedHeight(systemNavBarInset, navBarCompactMode)
    val shouldHideNavigationBar = false
    val navBarVisibilityProgressState = animateFloatAsState(
        targetValue = if (shouldHideNavigationBar) 0f else 1f,
        animationSpec = tween(
            durationMillis = 220,
            easing = LinearOutSlowInEasing
        ),
        label = "NavBarVisibilityProgress"
    )
    val navBarVisibilityProgress by navBarVisibilityProgressState
    val visibleNavBarOccupiedHeight by remember(navBarOccupiedHeight, navBarVisibilityProgress) {
        derivedStateOf { navBarOccupiedHeight * navBarVisibilityProgress }
    }
    val miniPlayerBottomMargin by remember(systemNavBarInset, visibleNavBarOccupiedHeight) {
        derivedStateOf {
            if (visibleNavBarOccupiedHeight > systemNavBarInset) {
                visibleNavBarOccupiedHeight
            } else {
                systemNavBarInset
            }
        }
    }
    val shouldRenderNavigationBar by remember(shouldHideNavigationBar, navBarVisibilityProgress) {
        derivedStateOf {
            !shouldHideNavigationBar || navBarVisibilityProgress > 0.01f
        }
    }
    val isNavBarEffectivelyHidden by remember(shouldHideNavigationBar, navBarVisibilityProgress) {
        derivedStateOf {
            shouldHideNavigationBar && navBarVisibilityProgress <= 0.01f
        }
    }

    CompositionLocalProvider(
        LocalAppHapticsConfig provides appHapticsConfig,
        LocalHapticFeedback provides scopedHapticFeedback,
        LocalShowScrollbar provides true
    ) {
        Scaffold(
            modifier = Modifier.fillMaxSize(),
            bottomBar = {
                if (shouldRenderNavigationBar) {
                    val currentSongId by remember {
                        playerViewModel.stablePlayerState
                            .map { it.currentSong?.id }
                            .distinctUntilChanged()
                    }.collectAsStateWithLifecycle(initialValue = null)
                    val showPlayerContentArea = currentSongId != null
                    val navBarElevation = 3.dp

                    val animatedNavBarCornerRadius = animateDpAsState(
                        targetValue = navBarCornerRadius.dp,
                        animationSpec = tween(400),
                        label = "NavBarCornerRadius"
                    )

                    val animatedDefaultTopCornerRadius = animateDpAsState(
                        targetValue = if (showPlayerContentArea && !isMiniPlayerDismissing) 10.dp else navBarCornerRadius.dp,
                        animationSpec = tween(400),
                        label = "NavBarDefaultTopCornerRadius"
                    )

                    var componentHeightPx by remember { mutableStateOf(0) }
                    val density = LocalDensity.current
                    val shadowOverflowPx = remember(navBarElevation, density) {
                        with(density) { (navBarElevation * 8).toPx() }
                    }
                    val bottomBarPaddingPx = remember(bottomBarPadding, density) {
                        with(density) { bottomBarPadding.toPx() }
                    }
                    val navBarElevationPx = remember(navBarElevation, density) {
                        with(density) { navBarElevation.toPx() }
                    }
                    val navBarShapeCache = remember { NavBarShapeCache() }

                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(navBarOccupiedHeight)
                            .clipToBounds()
                    ) {
                        Surface(
                            modifier = Modifier
                                .align(Alignment.BottomCenter)
                                .fillMaxWidth()
                                .padding(bottom = bottomBarPadding)
                                .onSizeChanged { componentHeightPx = it.height }
                                .graphicsLayer {
                                    val expansionHide = if (showPlayerContentArea) {
                                        playerViewModel.playerContentExpansionFraction.value.coerceIn(0f, 1f)
                                    } else {
                                        0f
                                    }
                                    val routeHide = (1f - navBarVisibilityProgressState.value).coerceIn(0f, 1f)
                                    val hideFraction = maxOf(expansionHide, routeHide)
                                    translationY = (componentHeightPx + shadowOverflowPx + bottomBarPaddingPx) * hideFraction
                                    alpha = 1f
                                }
                                .height(navBarHeight)
                                .padding(horizontal = horizontalPadding)
                                .graphicsLayer {
                                    val fraction = playerViewModel.playerContentExpansionFraction.value
                                    val topDp = when {
                                        navBarStyle == NavBarStyle.DEFAULT -> animatedDefaultTopCornerRadius.value
                                        navBarStyle == NavBarStyle.FULL_WIDTH -> lerp(navBarCornerRadius.dp, 26.dp, fraction)
                                        showPlayerContentArea -> if (fraction < 0.2f) {
                                            lerp(navBarCornerRadius.dp, 26.dp, (fraction / 0.2f).coerceIn(0f, 1f))
                                        } else {
                                            26.dp
                                        }
                                        else -> navBarCornerRadius.dp
                                    }
                                    val bottomDp = when (navBarStyle) {
                                        NavBarStyle.FULL_WIDTH -> 0.dp
                                        else -> animatedNavBarCornerRadius.value
                                    }
                                    shape = navBarShapeCache.get(this, topDp.toPx(), bottomDp.toPx(), useSmoothCorners)
                                    clip = true
                                    shadowElevation = navBarElevationPx
                                },
                            color = NavigationBarDefaults.containerColor
                        ) {
                            PlayerInternalNavigationBar(
                                navController = navController,
                                navItems = MainNavItems,
                                currentRoute = currentRoute,
                                navBarStyle = navBarStyle,
                                compactMode = navBarCompactMode,
                                bottomBarPadding = bottomBarPadding,
                                modifier = Modifier.fillMaxSize()
                            )
                        }
                    }
                }
            }
        ) { innerPadding ->
            BoxWithConstraints(modifier = Modifier.fillMaxSize()) {
                val density = LocalDensity.current
                val containerHeight = this.maxHeight
                val screenHeightPx = remember(containerHeight, density) {
                    with(density) { containerHeight.toPx() }
                }

                val showPlayerContentInitially by remember {
                    playerViewModel.stablePlayerState
                        .map { it.currentSong?.id != null }
                        .distinctUntilChanged()
                }.collectAsStateWithLifecycle(initialValue = false)
                val shouldHideMiniPlayer = false

                val miniPlayerH = with(density) { MiniPlayerHeight.toPx() }
                val totalSheetHeightWhenContentCollapsedPx = if (showPlayerContentInitially && !shouldHideMiniPlayer) miniPlayerH else 0f

                val bottomMargin = miniPlayerBottomMargin

                val spacerPx = with(density) { MiniPlayerBottomSpacer.toPx() }
                val bottomMarginPx = with(density) { bottomMargin.toPx() }
                val sheetCollapsedTargetY = calculatePlayerSheetCollapsedTargetY(
                    containerHeightPx = screenHeightPx,
                    collapsedContentHeightPx = totalSheetHeightWhenContentCollapsedPx,
                    bottomMarginPx = bottomMarginPx,
                    bottomSpacerPx = spacerPx
                )

                Box(modifier = Modifier.fillMaxSize()) {
                    content(innerPadding)
                }

                val isExpandedOrExpanding by remember {
                    derivedStateOf {
                        playerViewModel.playerContentExpansionFraction.value > 0.01f
                    }
                }
                androidx.compose.animation.AnimatedVisibility(
                    visible = isExpandedOrExpanding,
                    enter = fadeIn(animationSpec = tween(durationMillis = 350)),
                    exit = fadeOut(animationSpec = tween(durationMillis = 350)),
                    modifier = Modifier.fillMaxSize()
                ) {
                    Box(
                        modifier = Modifier
                            .fillMaxSize()
                            .background(
                                MaterialTheme.colorScheme.surfaceContainerLowest.copy(
                                    alpha = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) 0.35f else 0.6f
                                )
                            )
                            .pointerInput(Unit) {
                                detectTapGestures {
                                    playerViewModel.collapsePlayerSheet()
                                }
                            }
                    )
                }

                UnifiedPlayerSheetV2(
                    playerViewModel = playerViewModel,
                    sheetCollapsedTargetY = sheetCollapsedTargetY,
                    collapsedStateHorizontalPadding = horizontalPadding,
                    hideMiniPlayer = shouldHideMiniPlayer,
                    containerHeight = containerHeight,
                    navController = navController,
                    isNavBarHidden = isNavBarEffectivelyHidden
                )

                val dismissUndoBarSlice by remember {
                    playerViewModel.playerUiState
                        .map { state ->
                            DismissUndoBarSlice(
                                isVisible = state.showDismissUndoBar,
                                durationMillis = state.undoBarVisibleDuration
                            )
                        }
                        .distinctUntilChanged()
                }.collectAsStateWithLifecycle(initialValue = DismissUndoBarSlice())
                val onUndoDismissPlaylist = remember(playerViewModel) {
                    { playerViewModel.undoDismissPlaylist() }
                }
                val onCloseDismissUndoBar = remember(playerViewModel) {
                    { playerViewModel.hideDismissUndoBar() }
                }

                androidx.compose.animation.AnimatedVisibility(
                    visible = dismissUndoBarSlice.isVisible,
                    enter = slideInVertically(initialOffsetY = { it }) + fadeIn(),
                    exit = slideOutVertically(targetOffsetY = { -it }) + fadeOut(),
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .padding(bottom = innerPadding.calculateBottomPadding() + MiniPlayerBottomSpacer)
                        .padding(horizontal = horizontalPadding)
                ) {
                    DismissUndoBar(
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(MiniPlayerHeight)
                            .padding(horizontal = 14.dp),
                        onUndo = onUndoDismissPlaylist,
                        onClose = onCloseDismissUndoBar,
                        durationMillis = dismissUndoBarSlice.durationMillis
                    )
                }
            }
        }
    }
}

/**
 * Returns a cached Shape instance for a quantized (top, bottom) radius pair.
 * Because the instance identity is stable while the radii don't move past a
 * sub-pixel threshold, the graphics layer reuses its cached Outline between
 * frames and only re-clips when the radius actually changes.
 */
private class NavBarShapeCache {
    private var lastTopPx: Float = Float.NaN
    private var lastBottomPx: Float = Float.NaN
    private var lastSmooth: Boolean = true
    private var cached: androidx.compose.ui.graphics.Shape = RectangleShape

    fun get(
        density: androidx.compose.ui.unit.Density,
        topPx: Float,
        bottomPx: Float,
        smooth: Boolean
    ): androidx.compose.ui.graphics.Shape {
        if (smooth == lastSmooth &&
            !lastTopPx.isNaN() &&
            kotlin.math.abs(topPx - lastTopPx) < 0.5f &&
            kotlin.math.abs(bottomPx - lastBottomPx) < 0.5f
        ) {
            return cached
        }
        lastTopPx = topPx
        lastBottomPx = bottomPx
        lastSmooth = smooth
        cached = with(density) {
            DynamicSmoothCornerShape(
                useSmoothCorners = smooth,
                topRadius = topPx.toDp(),
                bottomRadius = bottomPx.toDp()
            )
        }
        return cached
    }
}

/**
 * Fixed-radius corner shape. Swaps AbsoluteSmoothCornerShape for a plain
 * RoundedCornerShape when smooth corners are disabled in settings. The radius
 * values are identical in both branches, so the animated radius behavior is
 * unchanged regardless of which delegate is active. The resulting Outline is
 * cached per (size, layoutDirection) so repeated draws are cheap.
 */
private class DynamicSmoothCornerShape(
    private val useSmoothCorners: Boolean,
    private val topRadius: androidx.compose.ui.unit.Dp,
    private val bottomRadius: androidx.compose.ui.unit.Dp
) : androidx.compose.ui.graphics.Shape {

    private var cachedSize: androidx.compose.ui.geometry.Size =
        androidx.compose.ui.geometry.Size.Unspecified
    private var cachedLayoutDirection: androidx.compose.ui.unit.LayoutDirection? = null
    private var cachedOutline: androidx.compose.ui.graphics.Outline? = null

    override fun createOutline(
        size: androidx.compose.ui.geometry.Size,
        layoutDirection: androidx.compose.ui.unit.LayoutDirection,
        density: androidx.compose.ui.unit.Density
    ): androidx.compose.ui.graphics.Outline {
        cachedOutline?.let {
            if (cachedSize == size && cachedLayoutDirection == layoutDirection) return it
        }

        val delegate: androidx.compose.ui.graphics.Shape = if (useSmoothCorners) {
            AbsoluteSmoothCornerShape(
                cornerRadiusTL = topRadius,
                smoothnessAsPercentTL = 60,
                cornerRadiusTR = topRadius,
                smoothnessAsPercentTR = 60,
                cornerRadiusBL = bottomRadius,
                smoothnessAsPercentBL = 60,
                cornerRadiusBR = bottomRadius,
                smoothnessAsPercentBR = 60
            )
        } else {
            RoundedCornerShape(
                topStart = topRadius,
                topEnd = topRadius,
                bottomEnd = bottomRadius,
                bottomStart = bottomRadius
            )
        }

        return delegate.createOutline(size, layoutDirection, density).also {
            cachedSize = size
            cachedLayoutDirection = layoutDirection
            cachedOutline = it
        }
    }
}
