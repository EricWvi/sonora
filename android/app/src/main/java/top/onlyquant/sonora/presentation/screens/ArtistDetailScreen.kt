@file:OptIn(ExperimentalMaterial3ExpressiveApi::class)

package top.onlyquant.sonora.presentation.screens

import top.onlyquant.sonora.presentation.components.DetailHeroFloatingLayer
import top.onlyquant.sonora.presentation.components.DetailHeroPinnedActionsHeight
import top.onlyquant.sonora.presentation.components.DetailSongsSummary
import top.onlyquant.sonora.presentation.components.HeroPlayState
import top.onlyquant.sonora.presentation.components.MediaDetailHero
import top.onlyquant.sonora.presentation.components.rememberDetailHeroDock
import top.onlyquant.sonora.presentation.components.rememberSongCollectionMenuActions
import top.onlyquant.sonora.presentation.utils.getSegmentedItemShape
import top.onlyquant.sonora.presentation.utils.SegmentedItemShapeDefaults
import androidx.lifecycle.compose.collectAsStateWithLifecycle

import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.Spring
import androidx.compose.animation.core.animateDpAsState
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.spring
import androidx.compose.animation.core.tween
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.GraphicEq
import androidx.compose.material.icons.rounded.Album
import androidx.compose.material.icons.rounded.ExpandMore
import androidx.compose.material.icons.rounded.Headphones
import androidx.compose.material.icons.rounded.Mic
import androidx.compose.material.icons.rounded.MusicNote
import androidx.compose.material.icons.rounded.PlayArrow
import androidx.compose.material.icons.rounded.SurroundSound
import androidx.compose.material.icons.rounded.UnfoldLess
import androidx.compose.material.icons.rounded.UnfoldMore
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.clipToBounds
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.RectangleShape
import androidx.compose.ui.graphics.compositeOver
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.graphics.lerp as lerpColor
import androidx.compose.ui.input.nestedscroll.NestedScrollConnection
import androidx.compose.ui.input.nestedscroll.NestedScrollSource
import androidx.compose.ui.input.nestedscroll.nestedScroll
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.Velocity
import androidx.compose.ui.unit.dp
import top.onlyquant.sonora.ui.theme.LocalPixelPlayerDarkTheme
import top.onlyquant.sonora.ui.theme.RoundedSans
import top.onlyquant.sonora.ui.theme.LocalShowScrollbar
import top.onlyquant.sonora.ui.theme.PixelPlayerStatusBarStyle
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import androidx.media3.common.util.UnstableApi
import androidx.navigation.NavController
import top.onlyquant.sonora.data.model.Artist
import top.onlyquant.sonora.data.model.Song
import top.onlyquant.sonora.presentation.components.CollapsibleCommonTopBar
import top.onlyquant.sonora.presentation.components.ExpressiveScrollBar
import top.onlyquant.sonora.presentation.components.MiniPlayerHeight
import top.onlyquant.sonora.presentation.components.SmartImageCompactListTargetSize
import top.onlyquant.sonora.presentation.components.SmartImage
import top.onlyquant.sonora.presentation.components.SongInfoBottomSheet
import top.onlyquant.sonora.presentation.components.resolveNavBarOccupiedHeight
import top.onlyquant.sonora.presentation.viewmodel.ArtistDetailViewModel
import top.onlyquant.sonora.presentation.viewmodel.ArtistAlbumSection
import top.onlyquant.sonora.presentation.viewmodel.PlayerViewModel
import top.onlyquant.sonora.presentation.viewmodel.collapseKey
import top.onlyquant.sonora.presentation.viewmodel.isAlbumSectionExpanded
import top.onlyquant.sonora.presentation.viewmodel.isAnyAlbumSectionExpanded
import top.onlyquant.sonora.utils.formatSongCount
import kotlinx.coroutines.launch
import top.onlyquant.sonora.presentation.components.subcomps.EnhancedSongListItem
import kotlin.math.roundToInt
import androidx.compose.ui.unit.IntOffset
import kotlinx.coroutines.delay
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.style.TextGeometricTransform
import androidx.compose.ui.unit.sp
import coil.size.Size
import racra.compose.smooth_corner_rect_library.AbsoluteSmoothCornerShape
import androidx.compose.ui.res.stringResource
import top.onlyquant.sonora.R


@androidx.annotation.OptIn(UnstableApi::class)
@OptIn(ExperimentalMaterial3Api::class, ExperimentalFoundationApi::class, ExperimentalMaterial3ExpressiveApi::class)
@Composable
fun ArtistDetailScreen(
    artistId: String,
    navController: NavController,
    playerViewModel: PlayerViewModel,
    viewModel: ArtistDetailViewModel = hiltViewModel(),
    // Sonora: no PlaylistViewModel, playlists come with step 5.
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    // Kept as State and read only by the album list and the expand-all button, so a header tap
    // does not recompose the whole screen.
    val albumExpansionOverrides = viewModel.albumExpansionOverrides.collectAsStateWithLifecycle()
    val stablePlayerState by playerViewModel.stablePlayerState.collectAsStateWithLifecycle()
    
    var isTransitionFinished by remember { mutableStateOf(false) }
    LaunchedEffect(Unit) {
        delay(300)
        isTransitionFinished = true
    }

    val lazyListState = rememberLazyListState()
    val favoriteIds by playerViewModel.favoriteSongIds.collectAsStateWithLifecycle()
    // Sonora: as in MainUI, the navigation bar is never compact until there are settings for it.
    val navBarCompactMode = false
    var showSongInfoBottomSheet by remember { mutableStateOf(false) }
    // Sonora: the original's playerViewModel.selectSongForInfo keeps the song in the view model
    // for its song info sheet; the reduced menu keeps it here, as in LibraryScreen.
    var selectedSongForInfo by remember { mutableStateOf<Song?>(null) }
    val systemNavBarInset = WindowInsets.navigationBars.asPaddingValues().calculateBottomPadding()
    val bottomBarHeightDp = resolveNavBarOccupiedHeight(systemNavBarInset, navBarCompactMode)
    val configuration = LocalConfiguration.current
    val density = LocalDensity.current
    val coroutineScope = rememberCoroutineScope()
    val isDarkTheme = LocalPixelPlayerDarkTheme.current
    val baseColorScheme = MaterialTheme.colorScheme

    val artistColorSchemePair by viewModel.artistColorScheme.collectAsStateWithLifecycle()
    val artistColorScheme = remember(artistColorSchemePair, isDarkTheme) {
        artistColorSchemePair?.let { pair -> if (isDarkTheme) pair.dark else pair.light }
            ?: baseColorScheme
    }

    // Sonora: no image picker, the artist's image is the server's (Eric, 2026-11-11).

    LaunchedEffect(Unit) {
        playerViewModel.collapsePlayerSheet()
    }

    val statusBarHeight = WindowInsets.statusBars.asPaddingValues().calculateTopPadding()
    val minTopBarHeight = 64.dp + statusBarHeight
    val maxTopBarHeight = 300.dp

    val minTopBarHeightPx = with(density) { minTopBarHeight.toPx() }
    val maxTopBarHeightPx = with(density) { maxTopBarHeight.toPx() }
    val headerImageRequestSize = remember(
        configuration.screenWidthDp,
        density.density,
        maxTopBarHeightPx
    ) {
        Size(
            width = with(density) { configuration.screenWidthDp.dp.roundToPx() },
            height = maxTopBarHeightPx.roundToInt()
        )
    }

    val topBarHeight = remember { Animatable(maxTopBarHeightPx) }

    // Hiding the system bars drops the status-bar inset to zero, which shrinks
    // minTopBarHeightPx. Pull the animated height back into range when the inset returns,
    // otherwise the collapsed bar stays too short.
    LaunchedEffect(minTopBarHeightPx, maxTopBarHeightPx) {
        val clamped = topBarHeight.value.coerceIn(minTopBarHeightPx, maxTopBarHeightPx)
        if (clamped != topBarHeight.value) topBarHeight.snapTo(clamped)
    }
    val collapseFraction by remember(minTopBarHeightPx, maxTopBarHeightPx) {
        derivedStateOf {
            1f - ((topBarHeight.value - minTopBarHeightPx) / (maxTopBarHeightPx - minTopBarHeightPx)).coerceIn(0f, 1f)
        }
    }

    // Direction of the latest scroll delta. A gesture that ends while heading back to the top
    // (a fling landing there included) must leave the header fully expanded.
    var lastScrollTowardsTop by remember { mutableStateOf(true) }
    val nestedScrollConnection = remember(minTopBarHeightPx, maxTopBarHeightPx) {
        object : NestedScrollConnection {
            override fun onPreScroll(available: Offset, source: NestedScrollSource): Offset {
                val delta = available.y
                val isScrollingDown = delta < 0
                if (delta != 0f) lastScrollTowardsTop = !isScrollingDown

                if (!isScrollingDown && (lazyListState.firstVisibleItemIndex > 0 || lazyListState.firstVisibleItemScrollOffset > 0)) {
                    return Offset.Zero
                }

                val previousHeight = topBarHeight.value
                val newHeight = (previousHeight + delta).coerceIn(minTopBarHeightPx, maxTopBarHeightPx)
                val consumed = newHeight - previousHeight

                if (consumed.roundToInt() != 0) {
                    coroutineScope.launch {
                        topBarHeight.snapTo(newHeight)
                    }
                }

                val canConsumeScroll = !(isScrollingDown && newHeight == minTopBarHeightPx)
                return if (canConsumeScroll) Offset(0f, consumed) else Offset.Zero
            }

            override suspend fun onPostFling(consumed: Velocity, available: Velocity): Velocity {
                return super.onPostFling(consumed, available)
            }
        }
    }

    LaunchedEffect(lazyListState.isScrollInProgress) {
        if (!lazyListState.isScrollInProgress) {
            val canExpand = lazyListState.firstVisibleItemIndex == 0 && lazyListState.firstVisibleItemScrollOffset == 0
            val shouldExpand = lastScrollTowardsTop ||
                topBarHeight.value > (minTopBarHeightPx + maxTopBarHeightPx) / 2

            val targetValue = if (shouldExpand && canExpand) {
                maxTopBarHeightPx
            } else {
                minTopBarHeightPx
            }

            if (topBarHeight.value != targetValue) {
                coroutineScope.launch {
                    topBarHeight.animateTo(targetValue, spring(stiffness = Spring.StiffnessMedium))
                }
            }
        }
    }

    MaterialTheme(
        colorScheme = artistColorScheme,
        typography = MaterialTheme.typography,
        shapes = MaterialTheme.shapes
    ) {
    Surface(
        modifier = Modifier.fillMaxSize(),
        color = MaterialTheme.colorScheme.surface
    ) {
        Box(modifier = Modifier.nestedScroll(nestedScrollConnection)) {
            when {
                uiState.isLoading && uiState.artist == null -> {
                    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        ContainedLoadingIndicator()
                    }
                }
                uiState.error != null && uiState.artist == null -> {
                    Box(
                        modifier = Modifier.fillMaxSize().padding(16.dp),
                        contentAlignment = Alignment.Center
                    ) {
                        Column(
                            horizontalAlignment = Alignment.CenterHorizontally,
                            verticalArrangement = Arrangement.spacedBy(12.dp)
                        ) {
                            Text(
                                text = uiState.error!!,
                                color = MaterialTheme.colorScheme.error,
                                style = MaterialTheme.typography.bodyLarge
                            )
                            FilledTonalButton(onClick = { viewModel.retry() }) {
                                Text(stringResource(R.string.library_retry))
                            }
                        }
                    }
                }
                uiState.artist != null -> {
                    val artist = uiState.artist!!
                    val songs = uiState.songs
                    val artistDurationMs = remember(songs) { songs.sumOf { it.duration } }
                    val currentTopBarHeightDp = with(density) { topBarHeight.value.toDp() }

                    val albumSections = uiState.albumSections
                    val heroOverlapPx = with(density) { ArtistHeroHeaderOverlap.toPx() }
                    val listTopPadding = minTopBarHeight + 8.dp
                    val listTopPaddingPx = with(density) { listTopPadding.toPx() }
                    // The list rides down with the header, overlapping its artwork by the hero overlap.
                    val listOffsetPx: () -> Int = {
                        (topBarHeight.value - minTopBarHeightPx - heroOverlapPx * (1f - collapseFraction))
                            .roundToInt()
                            .coerceAtLeast(0)
                    }
                    val heroDock = rememberDetailHeroDock(
                        heroTopPx = {
                            if (lazyListState.firstVisibleItemIndex == 0) {
                                listOffsetPx() + listTopPaddingPx - lazyListState.firstVisibleItemScrollOffset
                            } else {
                                null
                            }
                        },
                        // Docking starts once the header is collapsed and the list itself scrolls, so a
                        // list too short to scroll never leaves the title half-way.
                        dockStartHeroTopPx = listTopPaddingPx,
                        heroStartPadding = 16.dp,
                        collapsedBarBottom = minTopBarHeight,
                        hasBarSubtitle = true,
                        // Sonora: no image menu on the right of the bar (the original keeps
                        // 88dp for it), so the title ends 24dp from the edge as on the album page.
                        collapsedTitleEndPadding = 24.dp
                    )
                    val artistMenuActions = rememberSongCollectionMenuActions(
                        // Sonora: the playlist sheet comes with the playlists in step 5.
                        onAddToPlaylist = {},
                        onPlayNext = { if (songs.isNotEmpty()) playerViewModel.addSelectedAsNext(songs) },
                        onAddToQueue = { if (songs.isNotEmpty()) playerViewModel.addSelectedToQueue(songs) }
                    )
                    val songIdsInScope = remember(songs) { songs.mapTo(HashSet()) { it.id } }
                    val currentSongId = stablePlayerState.currentSong?.id
                    val heroPlayState = when {
                        currentSongId == null || currentSongId !in songIdsInScope -> HeroPlayState.Idle
                        stablePlayerState.isPlaying -> HeroPlayState.Playing
                        else -> HeroPlayState.Paused
                    }
                    val onHeroShuffle: () -> Unit = {
                        if (songs.isNotEmpty()) {
                            playerViewModel.playSongsShuffled(songs, artist.name, startAtZero = true)
                        }
                    }
                    val onHeroPlay: () -> Unit = {
                        if (heroPlayState == HeroPlayState.Idle) {
                            songs.firstOrNull()?.let { first -> playerViewModel.showAndPlaySong(first, songs) }
                        } else {
                            playerViewModel.playPause()
                        }
                    }
                    val barScrimFadePx = with(density) { 24.dp.toPx() }
                    val barContainerAlpha by remember(barScrimFadePx) {
                        derivedStateOf {
                            if (lazyListState.firstVisibleItemIndex > 0) 1f
                            else (lazyListState.firstVisibleItemScrollOffset / barScrimFadePx).coerceIn(0f, 1f)
                        }
                    }
                    val visibleSections = remember(albumSections) { albumSections.filter { it.songs.isNotEmpty() } }
                    val visibleSectionKeys = remember(visibleSections) { visibleSections.map { it.collapseKey() } }
                    val albumCount = visibleSections.size
                    val albumCountLabel = if (albumCount == 1) {
                        stringResource(R.string.detail_hero_album_count_one)
                    } else {
                        stringResource(R.string.detail_hero_album_count_many, albumCount)
                    }
                    val albumsCollapsedByDefault = uiState.albumsCollapsedByDefault
                    val isSectionExpanded: (String) -> Boolean = { sectionKey ->
                        isAlbumSectionExpanded(
                            key = sectionKey,
                            overrides = albumExpansionOverrides.value,
                            collapsedByDefault = albumsCollapsedByDefault,
                            sectionCount = albumCount
                        )
                    }
                    // Derived and read only by the summary row: this scope recomposes on every
                    // frame of the header collapse, which should not re-check every album.
                    val anySectionExpanded = remember(
                        albumExpansionOverrides,
                        visibleSectionKeys,
                        albumsCollapsedByDefault
                    ) {
                        derivedStateOf {
                            isAnyAlbumSectionExpanded(
                                sectionKeys = visibleSectionKeys,
                                overrides = albumExpansionOverrides.value,
                                collapsedByDefault = albumsCollapsedByDefault
                            )
                        }
                    }

                    val isScrollbarEnabled = LocalShowScrollbar.current
                    val showScrollBar by remember(isScrollbarEnabled) {
                        derivedStateOf {
                            isScrollbarEnabled &&
                            collapseFraction > 0.95f &&
                                (lazyListState.canScrollForward || lazyListState.canScrollBackward)
                        }
                    }
                    val scrollbarInset by animateDpAsState(
                        targetValue = if (showScrollBar) 8.dp else 0.dp,
                        label = "artistScrollbarInset"
                    )

                    // Artwork behind the list so the hero can ride up over its lower part.
                    ArtistHeaderBackdrop(
                        artistName = artist.name,
                        effectiveImageUrl = uiState.effectiveImageUrl,
                        collapseFraction = collapseFraction,
                        headerHeight = currentTopBarHeightDp,
                        headerImageRequestSize = headerImageRequestSize
                    )

                    LazyColumn(
                        state = lazyListState,
                        modifier = Modifier
                            .fillMaxSize()
                            .offset { IntOffset(0, listOffsetPx()) },
                        contentPadding = PaddingValues(
                            top = listTopPadding,
                            start = 16.dp,
                            // Fixed, so the hero never shifts; only album cards make room for
                            // the scrollbar (scrollbarInset).
                            end = 16.dp,
                            bottom = MiniPlayerHeight + WindowInsets.navigationBars.asPaddingValues().calculateBottomPadding() + 8.dp
                        )
                    ) {
                        item(key = "artist_hero", contentType = "artist_hero") {
                            MediaDetailHero(
                                title = artist.name,
                                // Song count and total length live in the summary above the list.
                                metaLine = albumCountLabel.takeIf { albumCount > 0 },
                                onShuffle = onHeroShuffle,
                                onPlay = onHeroPlay,
                                playState = heroPlayState,
                                menuActions = artistMenuActions,
                                enabled = songs.isNotEmpty(),
                                modifier = Modifier.padding(bottom = 8.dp),
                                dock = heroDock
                            )
                        }
                        if (songs.isNotEmpty()) {
                            item(key = "artist_songs_summary", contentType = "detail_songs_summary") {
                                Row(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(start = 8.dp, bottom = 12.dp)
                                        .padding(end = scrollbarInset),
                                    verticalAlignment = Alignment.CenterVertically
                                ) {
                                    DetailSongsSummary(
                                        songCount = songs.size,
                                        durationMs = artistDurationMs,
                                        modifier = Modifier.weight(1f)
                                    )
                                    if (albumCount > 1) {
                                        val anyExpanded = anySectionExpanded.value
                                        ArtistAlbumsExpandAllButton(
                                            anyExpanded = anyExpanded,
                                            onClick = { viewModel.setAllAlbumSectionsExpanded(!anyExpanded) }
                                        )
                                    }
                                }
                            }
                        }
                        visibleSections.forEachIndexed { index, section ->
                            val sectionKey = visibleSectionKeys[index]
                            val isFirstSection = index == 0
                            val isLastSection = index == visibleSections.lastIndex
                            val isExpanded = isSectionExpanded(sectionKey)
                            val sectionSongs = if (isTransitionFinished) section.songs else section.songs.take(5)

                            item(
                                key = "${sectionKey}_header",
                                contentType = "artist_section_header"
                            ) {
                                CollapsibleAlbumSectionHeader(
                                    modifier = Modifier.padding(end = scrollbarInset),
                                    section = section,
                                    isExpanded = isExpanded,
                                    isFirstSection = isFirstSection,
                                    isLastSection = isLastSection,
                                    onToggleExpanded = {
                                        viewModel.setAlbumSectionExpanded(sectionKey, !isExpanded)
                                    },
                                    onPlayAlbum = {
                                        section.songs.firstOrNull()?.let { firstSong ->
                                            playerViewModel.showAndPlaySong(firstSong, section.songs)
                                        }
                                    }
                                )
                            }

                            if (isExpanded) {
                                item(
                                    key = "${sectionKey}_song_group_spacer",
                                    contentType = "artist_section_spacer"
                                ) {
                                    Box(
                                        modifier = Modifier
                                            .animateItem(
                                                fadeInSpec = tween(durationMillis = 160),
                                                fadeOutSpec = tween(durationMillis = 120),
                                                placementSpec = tween(durationMillis = 180)
                                            )
                                            .padding(end = scrollbarInset)
                                            .fillMaxWidth()
                                            .height(10.dp)
                                            .background(MaterialTheme.colorScheme.surfaceContainerLow.copy(alpha = 0.5f))
                                    )
                                }

                                itemsIndexed(
                                    items = sectionSongs,
                                    key = { songIndex, song -> "${sectionKey}_song_${song.id}_$songIndex" },
                                    contentType = { _, _ -> "artist_section_song" }
                                ) { songIndex, song ->
                                    ArtistAlbumSectionSongItem(
                                        modifier = Modifier
                                            .animateItem(
                                                fadeInSpec = tween(durationMillis = 180),
                                                fadeOutSpec = tween(durationMillis = 120),
                                                placementSpec = tween(durationMillis = 200)
                                            )
                                            .padding(end = scrollbarInset),
                                        song = song,
                                        songIndex = songIndex,
                                        songCount = section.songs.size,
                                        isLastSection = isLastSection,
                                        isCurrentSong = stablePlayerState.currentSong?.id == song.id,
                                        isPlaying = stablePlayerState.isPlaying,
                                        onSongClick = {
                                            playerViewModel.showAndPlaySong(song, section.songs)
                                        },
                                        onMoreOptionsClick = {
                                            selectedSongForInfo = song
                                            showSongInfoBottomSheet = true
                                        }
                                    )
                                }
                            }

                            item(
                                key = "${sectionKey}_footer",
                                contentType = "artist_section_footer"
                            ) {
                                // Album cards sit 4dp apart, as one segmented group.
                                Spacer(modifier = Modifier.height(if (isLastSection) 24.dp else 4.dp))
                            }
                        }
                    }

                    if (showScrollBar) {
                        ExpressiveScrollBar(
                            listState = lazyListState,
                            modifier = Modifier
                                .align(Alignment.CenterEnd)
                                .padding(
                                    top = minTopBarHeight + DetailHeroPinnedActionsHeight + 4.dp,
                                    bottom = MiniPlayerHeight + systemNavBarInset + 8.dp
                                )
                        )
                    }

                    // Sonora: the original keeps an older top bar (CustomCollapsingTopBar) behind
                    // UseSharedCollapsibleTopBarProbe, which is always true; it is not copied.
                    SharedArtistTopBarProbe(
                        artist = artist,
                        songsCount = songs.size,
                        collapseFraction = collapseFraction,
                        titleAlpha = heroDock.barTitleAlpha,
                        subtitleAlpha = heroDock.barSubtitleAlpha,
                        containerAlpha = maxOf(barContainerAlpha, heroDock.barContainerAlpha),
                        headerHeight = currentTopBarHeightDp,
                        onBackPressed = { navController.popBackStack() }
                    )

                    DetailHeroFloatingLayer(
                        title = artist.name,
                        dock = heroDock,
                        listState = lazyListState,
                        onShuffle = onHeroShuffle,
                        onPlay = onHeroPlay,
                        menuActions = artistMenuActions,
                        playState = heroPlayState,
                        enabled = songs.isNotEmpty(),
                        metaLine = albumCountLabel.takeIf { albumCount > 0 }
                    )
                }
            }
        }
    }

    // Sonora: the original's playlist sheet for all of the artist's songs comes with step 5.

    if (showSongInfoBottomSheet && selectedSongForInfo != null) {
        val currentSong = selectedSongForInfo
        val isFavorite = remember(currentSong?.id, favoriteIds) {
            currentSong?.let { favoriteIds.contains(it.id) } ?: false
        }

        if (currentSong != null) {
            // Sonora: the reduced song menu (Eric, 2026-11-11), as in LibraryScreen. The original
            // also passes delete, navigation to the album, artists and genre, and metadata editing.
            SongInfoBottomSheet(
                song = currentSong,
                isFavorite = isFavorite,
                onToggleFavorite = {
                    playerViewModel.toggleFavoriteSpecificSong(currentSong)
                },
                onDismiss = { showSongInfoBottomSheet = false },
                onPlaySong = {
                    playerViewModel.showAndPlaySong(currentSong)
                    showSongInfoBottomSheet = false
                },
                onAddToQueue = {
                    playerViewModel.addSongToQueue(currentSong)
                    showSongInfoBottomSheet = false
                },
                onAddNextToQueue = {
                    playerViewModel.addSongNextToQueue(currentSong)
                    showSongInfoBottomSheet = false
                },
                onAddToPlayList = {
                    // Sonora: the playlist sheet comes with the playlists in step 5.
                }
            )
        }
    }
    }
}

private val SectionOuterCorner = 24.dp
private val SectionInnerCorner = 8.dp

/** "Collapse all" while any album shows its songs, "Expand all" once every album is collapsed. */
@Composable
private fun ArtistAlbumsExpandAllButton(
    anyExpanded: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    val height = ButtonDefaults.ExtraSmallContainerHeight
    TextButton(
        onClick = onClick,
        modifier = modifier.heightIn(min = height),
        contentPadding = ButtonDefaults.contentPaddingFor(height, hasStartIcon = true)
    ) {
        Icon(
            imageVector = if (anyExpanded) Icons.Rounded.UnfoldLess else Icons.Rounded.UnfoldMore,
            contentDescription = null,
            modifier = Modifier.size(ButtonDefaults.iconSizeFor(height))
        )
        Spacer(modifier = Modifier.width(ButtonDefaults.iconSpacingFor(height)))
        Text(
            text = stringResource(
                if (anyExpanded) R.string.artist_page_collapse_all_albums else R.string.artist_page_expand_all_albums
            ),
            style = ButtonDefaults.textStyleFor(height),
            maxLines = 1
        )
    }
}

@Composable
private fun CollapsibleAlbumSectionHeader(
    section: ArtistAlbumSection,
    isExpanded: Boolean,
    isFirstSection: Boolean,
    isLastSection: Boolean,
    onToggleExpanded: () -> Unit,
    modifier: Modifier = Modifier,
    onPlayAlbum: () -> Unit
) {
    val expandIconRotation by animateFloatAsState(
        targetValue = if (isExpanded) 180f else 0f,
        animationSpec = tween(durationMillis = 260),
        label = "ArtistAlbumExpandRotation"
    )
    val songCountText = formatSongCount(section.songs.size)
    val subtitle = remember(section.year, section.genre, songCountText) {
        // The genre goes last so a long one is what gets ellipsized, not the song count.
        listOfNotNull(
            section.year?.takeIf { it > 0 }?.toString(),
            songCountText,
            section.genre
        ).joinToString(" • ")
    }

    val containerColor = MaterialTheme.colorScheme.surfaceContainerLow.copy(alpha = 0.5f)
    // Album cards form one segmented group: only the group's outer corners stay round.
    val top = if (isFirstSection) SectionOuterCorner else SectionInnerCorner
    val bottom = when {
        isExpanded -> 0.dp
        isLastSection -> SectionOuterCorner
        else -> SectionInnerCorner
    }
    val shape = remember(top, bottom) {
        AbsoluteSmoothCornerShape(
            cornerRadiusTR = top, smoothnessAsPercentTR = 60,
            cornerRadiusTL = top, smoothnessAsPercentTL = 60,
            cornerRadiusBR = bottom, smoothnessAsPercentBR = 60,
            cornerRadiusBL = bottom, smoothnessAsPercentBL = 60
        )
    }

    Box(
        modifier = modifier
            .fillMaxWidth()
            .background(containerColor, shape)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .clip(shape)
                .clickable(onClick = onToggleExpanded)
                .padding(horizontal = 14.dp, vertical = 14.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            SmartImage(
                model = section.albumArtUriString,
                contentDescription = section.title,
                targetSize = SmartImageCompactListTargetSize,
                modifier = Modifier
                    .size(52.dp)
                    .clip(RoundedCornerShape(10.dp))
            )
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = section.title,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold,
                    color = MaterialTheme.colorScheme.onSurface,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = subtitle,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
            }
            FilledTonalIconButton(
                onClick = onPlayAlbum,
                colors = IconButtonDefaults.filledTonalIconButtonColors(
                    containerColor = MaterialTheme.colorScheme.tertiaryContainer,
                    contentColor = MaterialTheme.colorScheme.onTertiaryContainer
                )
            ) {
                Icon(Icons.Rounded.PlayArrow, contentDescription = stringResource(R.string.presentation_batch_d_cd_play_title, section.title))
            }
            Icon(
                imageVector = Icons.Rounded.ExpandMore,
                contentDescription = if (isExpanded) {
                    stringResource(R.string.presentation_batch_d_cd_collapse_title, section.title)
                } else {
                    stringResource(R.string.presentation_batch_d_cd_expand_title, section.title)
                },
                tint = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.graphicsLayer {
                    rotationZ = expandIconRotation
                }
            )
        }
    }
}

@Composable
private fun ArtistAlbumSectionSongItem(
    modifier: Modifier = Modifier,
    song: Song,
    songIndex: Int,
    songCount: Int,
    isLastSection: Boolean,
    isCurrentSong: Boolean,
    isPlaying: Boolean,
    onSongClick: () -> Unit,
    onMoreOptionsClick: () -> Unit
) {
    val isLastSong = songIndex == songCount - 1

    val songItemShape = getSegmentedItemShape(
        index = songIndex,
        count = songCount,
        outerRadius = SegmentedItemShapeDefaults.CardOuterCornerRadius
    )

    val containerShape = remember(isLastSong, isLastSection) {
        if (isLastSong) {
            val bottom = if (isLastSection) SectionOuterCorner else SectionInnerCorner
            AbsoluteSmoothCornerShape(
                cornerRadiusTR = 0.dp, smoothnessAsPercentTR = 0,
                cornerRadiusTL = 0.dp, smoothnessAsPercentTL = 0,
                cornerRadiusBR = bottom, smoothnessAsPercentBR = 60,
                cornerRadiusBL = bottom, smoothnessAsPercentBL = 60
            )
        } else {
            RectangleShape
        }
    }

    Box(
        modifier = modifier
            .fillMaxWidth()
            .background(MaterialTheme.colorScheme.surfaceContainerLow.copy(alpha = 0.5f), containerShape)
            .padding(horizontal = 8.dp)
    ) {
        Column {
            if (songIndex > 0) {
                Spacer(modifier = Modifier.height(4.dp))
            }

            EnhancedSongListItem(
                modifier = Modifier.fillMaxWidth(),
                song = song,
                isCurrentSong = isCurrentSong,
                isPlaying = isPlaying,
                showAlbumArt = false,
                customShape = songItemShape,
                onMoreOptionsClick = { onMoreOptionsClick() },
                onClick = onSongClick
            )

            if (isLastSong) {
                Spacer(modifier = Modifier.height(8.dp))
            }
        }
    }
}

@Composable
private fun ArtistHeaderBackdrop(
    artistName: String,
    effectiveImageUrl: String?,
    collapseFraction: Float,
    headerHeight: Dp,
    headerImageRequestSize: Size,
) {
    val surfaceColor = MaterialTheme.colorScheme.surface
    val expandedContentAlpha = 1f - (collapseFraction * 2f).coerceIn(0f, 1f)
    val displayUrl = effectiveImageUrl?.takeIf { it.isNotBlank() }
    val headerOverlayBrush = remember(surfaceColor, expandedContentAlpha) {
        Brush.verticalGradient(
            colors = listOf(
                Color.Transparent,
                surfaceColor.copy(alpha = 0.24f * expandedContentAlpha),
                surfaceColor.copy(alpha = 0.84f * expandedContentAlpha),
                surfaceColor
            )
        )
    }
    if (expandedContentAlpha <= 0.01f) return
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .height(headerHeight)
            .clipToBounds()
    ) {
        if (!displayUrl.isNullOrEmpty()) {
            SmartImage(
                model = displayUrl,
                contentDescription = artistName,
                contentScale = ContentScale.Crop,
                targetSize = headerImageRequestSize,
                allowHardware = true,
                crossfadeDurationMillis = 0,
                alpha = expandedContentAlpha,
                modifier = Modifier.fillMaxSize()
            )
        } else {
            MusicIconPattern(
                modifier = Modifier
                    .fillMaxSize()
                    .graphicsLayer { alpha = expandedContentAlpha }
            )
        }
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(headerOverlayBrush)
        )
    }
}

/** How far the hero rides up over the bottom of the expanded artist header. */
private val ArtistHeroHeaderOverlap = 120.dp

@Composable
private fun SharedArtistTopBarProbe(
    artist: Artist,
    songsCount: Int,
    collapseFraction: Float,
    titleAlpha: Float,
    subtitleAlpha: Float,
    containerAlpha: Float,
    headerHeight: Dp,
    onBackPressed: () -> Unit,
) {
    val surfaceColor = MaterialTheme.colorScheme.surface
    val statusBarColor =
        if (LocalPixelPlayerDarkTheme.current) Color.Black.copy(alpha = 0.6f)
        else Color.White.copy(alpha = 0.4f)
    val solidAlpha = (collapseFraction * 2f).coerceIn(0f, 1f)
    val statusBarBrush = remember(statusBarColor) {
        Brush.verticalGradient(colors = listOf(statusBarColor, Color.Transparent))
    }
    val expandedStatusBarFallback = remember(statusBarColor, surfaceColor) {
        statusBarColor.compositeOver(surfaceColor)
    }
    val fallbackStatusBarColor = remember(expandedStatusBarFallback, surfaceColor, solidAlpha) {
        lerpColor(expandedStatusBarFallback, surfaceColor, solidAlpha)
    }

    PixelPlayerStatusBarStyle(color = fallbackStatusBarColor)

    Box(
        modifier = Modifier
            .fillMaxWidth()
            .height(headerHeight)
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(80.dp)
                .background(statusBarBrush)
                .align(Alignment.TopCenter)
        )

        CollapsibleCommonTopBar(
            title = artist.name,
            subtitle = formatSongCount(songsCount),
            collapseFraction = collapseFraction,
            headerHeight = headerHeight,
            onBackClick = onBackPressed,
            // Transparent until something scrolls under the bar; the hero overlaps the
            // header, so a collapse-driven fill would dim it mid-gesture.
            containerColor = surfaceColor.copy(alpha = containerAlpha * solidAlpha),
            collapsedTitleStartPadding = 68.dp,
            expandedTitleStartPadding = 24.dp,
            // Sonora: 24dp as on the album page; the original keeps 88dp for the image menu.
            collapsedTitleEndPadding = 24.dp,
            expandedTitleEndPadding = 24.dp,
            containerHeightRange = 112.dp to 56.dp,
            titleStyle = MaterialTheme.typography.headlineMedium.copy(
                fontFamily = RoundedSans,
                fontWeight = FontWeight.SemiBold,
                textGeometricTransform = TextGeometricTransform(scaleX = 1.08f)
            ),
            titleScaleRange = 1f to 1f,
            titleFontSizeRange = 30.sp to 18.sp,
            maxLines = 1,
            collapsedSubtitleMaxLines = 1,
            expandedSubtitleMaxLines = 1,
            // The hero title docks into this slot itself (DetailHeroFloatingLayer); the bar's
            // own title only shows for titles too long to dock.
            contentColor = MaterialTheme.colorScheme.onSurface.copy(alpha = titleAlpha),
            subtitleColor = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = subtitleAlpha),
            fadeSubtitleOnCollapse = false,
            syncStatusBarWithContainer = false,
            // Sonora: no edit button with the "change photo / reset to default" menu here,
            // the artist's image is the server's (Eric, 2026-11-11).
        )
    }
}

@Composable
private fun MusicIconPattern(modifier: Modifier = Modifier) {
    val color1 = MaterialTheme.colorScheme.onPrimaryContainer.copy(alpha = 0.1f)
    val color2 = MaterialTheme.colorScheme.primary.copy(alpha = 0.15f)

    Box(modifier = modifier.background(MaterialTheme.colorScheme.primaryContainer)) {
        Icon(
            imageVector = Icons.Rounded.MusicNote,
            contentDescription = null, tint = color1,
            modifier = Modifier
                .align(Alignment.TopStart)
                .offset(x = 76.dp, y = 72.dp)
                .size(60.dp)
                .graphicsLayer { rotationZ = -12f }
        )
        Icon(
            imageVector = Icons.Default.GraphicEq,
            contentDescription = null, tint = color1,
            modifier = Modifier
                .align(Alignment.CenterStart)
                .offset(x = 4.dp, y = 78.dp)
                .size(50.dp)
                .graphicsLayer { rotationZ = 18f }
        )
        Icon(
            imageVector = Icons.Rounded.Album,
            contentDescription = null, tint = color2,
            modifier = Modifier
                .align(Alignment.CenterEnd)
                .offset(x = 4.dp, y = (-70).dp)
                .size(70.dp)
                .graphicsLayer { rotationZ = 12f }
        )
        Icon(
            imageVector = Icons.Rounded.Mic,
            contentDescription = null, tint = color1,
            modifier = Modifier
                .align(Alignment.BottomCenter)
                .offset(x = 14.dp, y = (-6).dp)
                .size(60.dp)
                .graphicsLayer { rotationZ = 14f }
        )
        Icon(
            imageVector = Icons.Rounded.SurroundSound,
            contentDescription = null, tint = color2,
            modifier = Modifier
                .align(Alignment.TopCenter)
                .offset(x = (-24).dp, y = 32.dp)
                .size(80.dp)
                .graphicsLayer { rotationZ = 10f }
        )
        Icon(
            imageVector = Icons.Rounded.MusicNote,
            contentDescription = null, tint = color1,
            modifier = Modifier
                .align(Alignment.BottomEnd)
                .offset(x = (-18).dp, y = (-138).dp)
                .size(45.dp)
                .graphicsLayer { rotationZ = -18f }
        )
        Icon(
            imageVector = Icons.Rounded.Headphones,
            contentDescription = null, tint = color2,
            modifier = Modifier
                .align(Alignment.Center)
                .offset(x = 72.dp, y = 34.dp)
                .size(45.dp)
                .graphicsLayer { rotationZ = -8f }
        )
    }
}
