package top.onlyquant.sonora.presentation.screens

import top.onlyquant.sonora.presentation.navigation.navigateSafely
import top.onlyquant.sonora.presentation.navigation.navigateSafelyReplacing

import android.content.Intent
import androidx.activity.compose.ReportDrawnWhen
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.DragInteraction
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.rounded.MusicNote
import androidx.compose.material.icons.rounded.Refresh
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExperimentalMaterial3ExpressiveApi
import androidx.compose.material3.FilledTonalButton
import androidx.compose.material3.Icon
import androidx.compose.material3.LargeExtendedFloatingActionButton
import androidx.compose.material3.LoadingIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.derivedStateOf
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.ExperimentalTextApi
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.Font
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontVariation
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import androidx.media3.common.util.UnstableApi
import androidx.navigation.NavController
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.lifecycle.compose.currentStateAsState
import top.onlyquant.sonora.R
import top.onlyquant.sonora.data.model.Song
import top.onlyquant.sonora.data.preferences.CollagePattern
import top.onlyquant.sonora.presentation.components.AlbumArtCollage
import top.onlyquant.sonora.presentation.components.HomeGradientTopBar
import top.onlyquant.sonora.presentation.components.MiniPlayerHeight
import top.onlyquant.sonora.presentation.components.RecentlyPlayedSection
import top.onlyquant.sonora.presentation.components.RecentlyPlayedSectionMinSongsToShow
import top.onlyquant.sonora.presentation.components.SmartImage
import top.onlyquant.sonora.presentation.components.resolveMainScreenBottomGradientHeight
import top.onlyquant.sonora.presentation.components.subcomps.PlayingEqIcon
import top.onlyquant.sonora.presentation.navigation.Screen
import top.onlyquant.sonora.presentation.viewmodel.PlayerViewModel
import top.onlyquant.sonora.presentation.viewmodel.HomeViewModel
import top.onlyquant.sonora.presentation.components.SongInfoBottomSheet
import androidx.lifecycle.repeatOnLifecycle
import androidx.compose.material.icons.rounded.Radio
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.TextButton
import top.onlyquant.sonora.presentation.viewmodel.sharedPlayerViewModel
import top.onlyquant.sonora.ui.theme.ExpTitleTypography
import kotlinx.collections.immutable.persistentListOf
import kotlinx.collections.immutable.toImmutableList
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.flow.filterIsInstance
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.launch
import top.onlyquant.sonora.ui.theme.ShapeCache
import racra.compose.smooth_corner_rect_library.AbsoluteSmoothCornerShape
import androidx.compose.ui.res.stringResource

private const val HomeLoadingPlaceholderMinDurationMillis = 1200L
private const val HomeScrollRestoreGraceMillis = 1_500L

/**
 * The home screen, from HomeScreen in PixelPlayerOSS v0.4.0-alpha.53 (GPL-3.0). Sonora (Eric,
 * 2026-11-11): the header and the album art collage are the Roaming FM (漫游电台): the title is the
 * radio's, its button starts the radio, and the collage shows the five albums it plays next
 * (`GET fm?count=5`, one per place of the collage), a tap opening the album. The original's header and collage are "your mix",
 * built on the device. Left out: the daily mix card, the yearly recap card and the listening stats
 * card (no on-device statistics), and the sheets of the buttons the top bar no longer has. Recently
 * played comes from the server's history (`GET history`).
 */
@androidx.annotation.OptIn(UnstableApi::class)
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HomeScreen(
    navController: NavController,
    paddingValuesParent: PaddingValues,
    playerViewModel: PlayerViewModel = sharedPlayerViewModel(),
    homeViewModel: HomeViewModel = hiltViewModel(),
    onSignOut: () -> Unit = {},
) {
    val context = LocalContext.current
    val homeUiState by homeViewModel.uiState.collectAsStateWithLifecycle()
    val lifecycleOwner = LocalLifecycleOwner.current

    val fmAlbums = homeUiState.fmAlbums
    val yourMixSongs = fmAlbums ?: persistentListOf()
    var homePlaceholderRefreshGeneration by rememberSaveable { mutableIntStateOf(0) }
    var hasHomeLoadingMinimumElapsed by rememberSaveable(homePlaceholderRefreshGeneration) {
        mutableStateOf(false)
    }

    LaunchedEffect(homePlaceholderRefreshGeneration, yourMixSongs.isEmpty()) {
        if (yourMixSongs.isEmpty()) {
            hasHomeLoadingMinimumElapsed = false
            delay(HomeLoadingPlaceholderMinDurationMillis)
            hasHomeLoadingMinimumElapsed = true
        } else {
            hasHomeLoadingMinimumElapsed = true
        }
    }

    // Sonora: also waits for the first answer from the server.
    val shouldShowYourMixLoadingPlaceholder =
        yourMixSongs.isEmpty() && (fmAlbums == null || !hasHomeLoadingMinimumElapsed)

    // Sonora: the original keeps the pills in place while Home is on screen; the server's history
    // is reloaded when Home is shown, so it simply follows it.
    val recentlyPlayedSongs = homeUiState.recentlyPlayed

    val recentlyPlayedQueue = remember(recentlyPlayedSongs) {
        recentlyPlayedSongs.map { it.song }.toImmutableList()
    }

    ReportDrawnWhen {
        yourMixSongs.isNotEmpty() || hasHomeLoadingMinimumElapsed
    }

    val recentlyPlayedQueueName = stringResource(R.string.presentation_batch_b_queue_recently_played)

    // Seed with the live value: starting from null made the bottom padding grow one frame
    // after returning to Home, once the scroll position was already restored, which left the
    // last section under the mini player and nav bar. The seed is read once inside remember
    // so the StateFlow is not read during composition.
    val initialCurrentSong = remember(playerViewModel) {
        playerViewModel.stablePlayerState.value.currentSong
    }
    val currentSong by remember(playerViewModel.stablePlayerState) {
        playerViewModel.stablePlayerState.map { it.currentSong }
    }.collectAsStateWithLifecycle(initialValue = initialCurrentSong)

    // Sonora: the button is highlighted while the radio plays (the original: while shuffle is on).
    val isFmPlaying by playerViewModel.isRoamingFmPlaying.collectAsStateWithLifecycle()

    // Sonora: reload when Home is shown, and when the radio moves on to another album.
    val fmAlbumId = currentSong?.albumId.takeIf { isFmPlaying }
    LaunchedEffect(lifecycleOwner, fmAlbumId, homePlaceholderRefreshGeneration) {
        lifecycleOwner.lifecycle.repeatOnLifecycle(Lifecycle.State.RESUMED) {
            homeViewModel.refresh()
        }
    }

    val bottomPadding = if (currentSong != null) MiniPlayerHeight else 0.dp
    // Sonora: as in MainUI, the navigation bar is never compact until there are settings for it.
    val navBarCompactMode = false
    val bottomGradientHeight = resolveMainScreenBottomGradientHeight(navBarCompactMode)

    var recentSongForOptions by remember { mutableStateOf<Song?>(null) }
    var showSignOutDialog by remember { mutableStateOf(false) }

    val listState = rememberSaveable(saver = LazyListState.Saver) { LazyListState() }
    val density = LocalDensity.current
    val scrollThresholdPx = remember(density) { with(density) { 180.dp.toPx() } }
    val isScrolledPastThreshold = remember {
        derivedStateOf { listState.firstVisibleItemIndex > 0 || listState.firstVisibleItemScrollOffset > scrollThresholdPx }
    }

    var savedScrollIndex by rememberSaveable { mutableIntStateOf(0) }
    var savedScrollOffset by rememberSaveable { mutableIntStateOf(0) }
    var needsScrollRestore by rememberSaveable { mutableStateOf(false) }
    // ON_PAUSE fires while Home is still composed (during its exit transition, or before the
    // window hides), so restoring right away would only re-apply the current position and save
    // the flag as false. Wait until Home is resumed again; the grace window starts there too.
    val lifecycleState = lifecycleOwner.lifecycle.currentStateAsState()
    val isHomeResumed by remember(lifecycleState) {
        derivedStateOf { lifecycleState.value.isAtLeast(Lifecycle.State.RESUMED) }
    }

    DisposableEffect(lifecycleOwner, listState) {
        val observer = LifecycleEventObserver { _, event ->
            if (event == Lifecycle.Event.ON_PAUSE) {
                savedScrollIndex = listState.firstVisibleItemIndex
                savedScrollOffset = listState.firstVisibleItemScrollOffset
                needsScrollRestore = true
            }
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose { lifecycleOwner.lifecycle.removeObserver(observer) }
    }

    // Sections can re-emit empty values for a moment when returning, so the list is briefly
    // shorter than when we left it. Keep re-applying the saved position as content arrives
    // until it is actually reached, the user drags the list, or a short grace window ends.
    LaunchedEffect(
        needsScrollRestore,
        isHomeResumed,
        yourMixSongs.isNotEmpty(),
        recentlyPlayedSongs.size,
        bottomPadding
    ) {
        if (!needsScrollRestore || !isHomeResumed) return@LaunchedEffect
        val totalItems = listState.layoutInfo.totalItemsCount
        if (totalItems == 0) return@LaunchedEffect
        val targetIndex = savedScrollIndex.coerceIn(0, (totalItems - 1).coerceAtLeast(0))
        listState.scrollToItem(targetIndex, savedScrollOffset)
        if (listState.firstVisibleItemIndex == savedScrollIndex &&
            listState.firstVisibleItemScrollOffset == savedScrollOffset
        ) {
            needsScrollRestore = false
        }
    }

    LaunchedEffect(needsScrollRestore, isHomeResumed) {
        if (!needsScrollRestore || !isHomeResumed) return@LaunchedEffect
        launch {
            listState.interactionSource.interactions
                .filterIsInstance<DragInteraction.Start>()
                .first()
            needsScrollRestore = false
        }
        delay(HomeScrollRestoreGraceMillis)
        needsScrollRestore = false
    }

    Box(
        modifier = Modifier.fillMaxSize()
    ) {
        Scaffold(
            modifier = Modifier.fillMaxSize(),
            topBar = {
                HomeGradientTopBar(
                    onNavigationIconClick = {
                        // Sonora: opens Screen.Settings in the original. Until the settings
                        // screen is copied it offers signing out, the only setting so far (Eric,
                        // 2026-11-11: a temporary way to sign out).
                        showSignOutDialog = true
                    },
                    isScrolled = isScrolledPastThreshold.value
                )
            }
        ) { innerPadding ->
            LazyColumn(
                state = listState,
                modifier = Modifier
                    .fillMaxSize()
                    .background(MaterialTheme.colorScheme.background),
                contentPadding = PaddingValues(
                    top = innerPadding.calculateTopPadding(),
                    bottom = paddingValuesParent.calculateBottomPadding()
                            + 38.dp + bottomPadding
                ),
                verticalArrangement = Arrangement.spacedBy(24.dp)
            ) {
                if (yourMixSongs.isEmpty()) {
                    item(
                        key = "your_mix_placeholder",
                        contentType = "your_mix_placeholder"
                    ) {
                        if (shouldShowYourMixLoadingPlaceholder) {
                            YourMixLoadingPlaceholder()
                        } else {
                            YourMixEmptyPlaceholder(
                                onRefresh = {
                                    homePlaceholderRefreshGeneration++
                                }
                            )
                        }
                    }
                } else {
                    item(
                        key = "your_mix_header",
                        contentType = "your_mix_header"
                    ) {
                        YourMixHeader(
                            song = stringResource(R.string.home_roaming_subtitle),
                            isShuffleEnabled = isFmPlaying,
                            onPlayShuffled = { playerViewModel.playRoamingFm() }
                        )
                    }
                }

                if (yourMixSongs.isNotEmpty()) {
                    item(
                        key = "album_art_collage",
                        contentType = "album_art_collage"
                    ) {
                        // Sonora: the original's pattern setting (default, no auto-rotate).
                        AlbumArtCollage(
                            modifier = Modifier.fillMaxWidth(),
                            songs = yourMixSongs,
                            padding = 14.dp,
                            height = 400.dp,
                            pattern = CollagePattern.default,
                            onSongClick = { song ->
                                navController.navigateSafely(Screen.AlbumDetail.createRoute(song.albumId))
                            }
                        )
                    }
                }

                // Sonora: no yearly recap card and no daily mix card (Eric, 2026-11-11).

                if (recentlyPlayedSongs.size >= RecentlyPlayedSectionMinSongsToShow) {
                    item(
                        key = "recently_played_section",
                        contentType = "recently_played_section"
                    ) {
                        RecentlyPlayedSection(
                            songs = recentlyPlayedSongs,
                            onSongClick = { song ->
                                if (recentlyPlayedQueue.isNotEmpty()) {
                                    playerViewModel.playSongs(
                                        songsToPlay = recentlyPlayedQueue,
                                        startSong = song,
                                        queueName = recentlyPlayedQueueName
                                    )
                                }
                            },
                            onSongLongClick = { song ->
                                recentSongForOptions = song
                            },
                            onOpenAllClick = {
                                // Sonora: the recently played screen comes with step 5.
                            },
                            themeStateHolder = playerViewModel.themeStateHolder,
                            currentSongId = currentSong?.id,
                            contentPadding = PaddingValues(start = 8.dp, end = 24.dp)
                        )
                    }
                }

                // Sonora: no listening stats card (no on-device statistics).
            }
        }
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .align(Alignment.BottomCenter)
                .height(bottomGradientHeight)
                .background(
                    brush = Brush.verticalGradient(
                        colorStops = arrayOf(
                            0.0f to Color.Transparent,
                            0.2f to Color.Transparent,
                            0.8f to MaterialTheme.colorScheme.surfaceContainerLowest,
                            1.0f to MaterialTheme.colorScheme.surfaceContainerLowest
                        )
                    )
                )
        ) {

        }
    }
    if (showSignOutDialog) {
        AlertDialog(
            onDismissRequest = { showSignOutDialog = false },
            title = { Text(stringResource(R.string.sign_out)) },
            text = { Text(stringResource(R.string.sign_out_message)) },
            confirmButton = {
                TextButton(onClick = {
                    showSignOutDialog = false
                    onSignOut()
                }) { Text(stringResource(R.string.sign_out)) }
            },
            dismissButton = {
                TextButton(onClick = { showSignOutDialog = false }) { Text(stringResource(R.string.cancel)) }
            }
        )
    }
    // Sonora: the original's RecentlyPlayedSongOptionsSheet with the reduced song menu (Eric,
    // 2026-11-11), as in LibraryScreen.
    recentSongForOptions?.let { song ->
        val favoriteIds by playerViewModel.favoriteSongIds.collectAsStateWithLifecycle()
        SongInfoBottomSheet(
            song = song,
            isFavorite = favoriteIds.contains(song.id),
            onToggleFavorite = { playerViewModel.toggleFavoriteSpecificSong(song) },
            onDismiss = { recentSongForOptions = null },
            onPlaySong = {
                if (recentlyPlayedQueue.any { it.id == song.id }) {
                    playerViewModel.playSongs(
                        songsToPlay = recentlyPlayedQueue,
                        startSong = song,
                        queueName = recentlyPlayedQueueName
                    )
                } else {
                    playerViewModel.showAndPlaySong(song)
                }
                recentSongForOptions = null
            },
            onAddToQueue = {
                playerViewModel.addSongToQueue(song)
                playerViewModel.sendToast(context.getString(R.string.toast_added_to_queue))
                recentSongForOptions = null
            },
            onAddNextToQueue = {
                playerViewModel.addSongNextToQueue(song)
                playerViewModel.sendToast(context.getString(R.string.toast_playing_next))
                recentSongForOptions = null
            },
            onAddToPlayList = {
                // Sonora: the playlist sheet comes with the playlists in step 5.
            }
        )
    }
}

@OptIn(ExperimentalMaterial3ExpressiveApi::class)
@Composable
private fun YourMixLoadingPlaceholder() {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .height(256.dp)
            .padding(16.dp),
        contentAlignment = Alignment.Center
    ) {
        LoadingIndicator(
            modifier = Modifier.size(128.dp),
            color = MaterialTheme.colorScheme.primary
        )
    }
}

@Composable
private fun YourMixEmptyPlaceholder(
    onRefresh: () -> Unit
) {
    val colors = MaterialTheme.colorScheme

    Box(
        modifier = Modifier
            .fillMaxWidth()
            .heightIn(min = 256.dp)
            .padding(horizontal = 24.dp, vertical = 16.dp),
        contentAlignment = Alignment.Center
    ) {
        Column(
            modifier = Modifier.fillMaxWidth(),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            Surface(
                modifier = Modifier.size(76.dp),
                shape = ShapeCache.expressiveHero,
                color = colors.secondaryContainer,
                contentColor = colors.onSecondaryContainer
            ) {
                Box(contentAlignment = Alignment.Center) {
                    Icon(
                        imageVector = Icons.Rounded.MusicNote,
                        contentDescription = null,
                        modifier = Modifier.size(34.dp)
                    )
                }
            }

            Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(6.dp)
            ) {
                Text(
                    text = stringResource(R.string.home_empty_placeholder_title),
                    style = MaterialTheme.typography.titleLarge,
                    color = colors.onSurface,
                    textAlign = TextAlign.Center
                )
                Text(
                    // Sonora: the Web player's words when the radio has no album; the original's
                    // speaks of scanning and server sources.
                    text = stringResource(R.string.fm_no_albums),
                    style = MaterialTheme.typography.bodyMedium,
                    color = colors.onSurfaceVariant,
                    textAlign = TextAlign.Center,
                    modifier = Modifier.fillMaxWidth()
                )
            }

            FilledTonalButton(
                onClick = onRefresh,
                shape = AbsoluteSmoothCornerShape(
                    cornerRadiusTL = 22.dp,
                    smoothnessAsPercentTR = 60,
                    cornerRadiusBR = 22.dp,
                    smoothnessAsPercentTL = 60,
                    cornerRadiusBL = 22.dp,
                    smoothnessAsPercentBR = 60,
                    cornerRadiusTR = 22.dp,
                    smoothnessAsPercentBL = 60,
                )
            ) {
                Icon(
                    imageVector = Icons.Rounded.Refresh,
                    contentDescription = null,
                    modifier = Modifier.size(20.dp)
                )
                Spacer(modifier = Modifier.width(8.dp))
                Text(text = stringResource(R.string.home_empty_placeholder_refresh))
            }
        }
    }
}

@OptIn(ExperimentalMaterial3ExpressiveApi::class)
@Composable
fun YourMixHeader(
    song: String,
    isShuffleEnabled: Boolean = false,
    onPlayShuffled: () -> Unit
) {
    val buttonCorners = 68.dp
    val colors = MaterialTheme.colorScheme

    val titleStyle = rememberYourMixTitleStyle()

    Box(
        modifier = Modifier
            .fillMaxWidth()
            .height(256.dp)
            .padding(16.dp)
    ) {
        Column(
            modifier = Modifier
                .align(Alignment.TopStart)
                .padding(top = 48.dp, start = 12.dp)
        ) {
            Text(
                // Sonora: 漫游电台 (Eric, 2026-11-11); the original's 为你推荐.
                text = stringResource(R.string.home_roaming_title),
                style = titleStyle,
                color = MaterialTheme.colorScheme.onSurface,
                modifier = Modifier
            )

            Text(
                text = song,
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.7f),
                modifier = Modifier.padding(start = 8.dp)
            )
        }
        LargeExtendedFloatingActionButton(
            modifier = Modifier
                .align(Alignment.BottomEnd)
                .padding(end = 12.dp),
            onClick = onPlayShuffled,
            containerColor = if (isShuffleEnabled) colors.primary else colors.tertiaryContainer,
            contentColor = if (isShuffleEnabled) colors.onPrimary else colors.onTertiaryContainer,
            shape = AbsoluteSmoothCornerShape(
                cornerRadiusTL = buttonCorners,
                smoothnessAsPercentTR = 60,
                cornerRadiusBR = buttonCorners,
                smoothnessAsPercentTL = 60,
                cornerRadiusBL = buttonCorners,
                smoothnessAsPercentBR = 60,
                cornerRadiusTR = buttonCorners,
                smoothnessAsPercentBL = 60,
            )
        ) {
            // Sonora: a radio icon for the Roaming FM (Eric, 2026-11-11); the original shuffles.
            Icon(
                imageVector = Icons.Rounded.Radio,
                contentDescription = stringResource(R.string.cd_play_roaming_fm),
                modifier = Modifier.size(36.dp)
            )
        }
    }
}


// Sonora: SongListItemFavs and its wrapper are not used by this screen.

@OptIn(ExperimentalTextApi::class)
@Composable
private fun rememberYourMixTitleStyle(): TextStyle {
    return remember {
        TextStyle(
            fontFamily = FontFamily(
                Font(
                    resId = R.font.gflex_variable,
                    variationSettings = FontVariation.Settings(
                        FontVariation.weight(636),
                        FontVariation.width(152f),
                        FontVariation.Setting("ROND", 50f),
                        FontVariation.Setting("XTRA", 520f),
                        FontVariation.Setting("YOPQ", 90f),
                        FontVariation.Setting("YTLC", 505f)
                    )
                )
            ),
            fontWeight = FontWeight(760),
            fontSize = 64.sp,
            lineHeight = 62.sp
        )
    }
}
