@file:OptIn(ExperimentalMaterial3ExpressiveApi::class)

package top.onlyquant.sonora.presentation.screens

import top.onlyquant.sonora.presentation.navigation.navigateSafelyReplacing
import top.onlyquant.sonora.presentation.components.DetailHeroArtistChip
import top.onlyquant.sonora.presentation.components.DetailHeroFloatingLayer
import top.onlyquant.sonora.presentation.components.DetailHeroPinnedActionsHeight
import top.onlyquant.sonora.presentation.components.DetailSongsSummary
import top.onlyquant.sonora.presentation.components.HeroPlayState
import top.onlyquant.sonora.presentation.components.MediaDetailHero
import top.onlyquant.sonora.presentation.components.rememberDetailHeroDock
import top.onlyquant.sonora.presentation.components.rememberSongCollectionMenuActions
import top.onlyquant.sonora.presentation.screens.search.components.GenreTypography
import top.onlyquant.sonora.presentation.utils.getSegmentedItemShape

import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.Spring
import androidx.compose.animation.core.spring
import androidx.compose.animation.core.animateDpAsState
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.asPaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBars
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBars
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.rounded.CloudDownload
import androidx.compose.material.icons.rounded.CloudDone
import androidx.compose.material3.ContainedLoadingIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExperimentalMaterial3ExpressiveApi
import androidx.compose.material3.FilledIconButton
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButtonDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.FilledTonalButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clipToBounds
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.compositeOver
import androidx.compose.ui.graphics.lerp as lerpColor
import androidx.compose.ui.input.nestedscroll.NestedScrollConnection
import androidx.compose.ui.input.nestedscroll.NestedScrollSource
import androidx.compose.ui.input.nestedscroll.nestedScroll
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextGeometricTransform
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.IntOffset
import androidx.compose.ui.unit.Velocity
import androidx.compose.ui.unit.dp
import top.onlyquant.sonora.ui.theme.LocalPixelPlayerDarkTheme
import top.onlyquant.sonora.ui.theme.RoundedSans
import top.onlyquant.sonora.ui.theme.LocalShowScrollbar
import top.onlyquant.sonora.ui.theme.PixelPlayerStatusBarStyle
import androidx.compose.ui.unit.sp
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import androidx.media3.common.util.UnstableApi
import androidx.navigation.NavController
import coil.compose.AsyncImagePainter
import coil.size.Size
import top.onlyquant.sonora.R
import top.onlyquant.sonora.data.model.Album
import top.onlyquant.sonora.data.model.Song
import top.onlyquant.sonora.presentation.components.CollapsibleCommonTopBar
import top.onlyquant.sonora.presentation.components.ExpressiveScrollBar
import top.onlyquant.sonora.presentation.components.MiniPlayerHeight
import top.onlyquant.sonora.presentation.components.SmartImage
import top.onlyquant.sonora.presentation.components.SongInfoBottomSheet
import top.onlyquant.sonora.presentation.components.resolveNavBarOccupiedHeight
import top.onlyquant.sonora.presentation.navigation.Screen
import top.onlyquant.sonora.presentation.viewmodel.AlbumDetailViewModel
import top.onlyquant.sonora.presentation.viewmodel.PlayerViewModel
import top.onlyquant.sonora.utils.formatSongCount
import kotlinx.coroutines.launch
import kotlin.math.roundToInt
import androidx.compose.ui.res.stringResource


@androidx.annotation.OptIn(UnstableApi::class)
@OptIn(ExperimentalMaterial3Api::class, ExperimentalMaterial3ExpressiveApi::class)
@Composable
fun AlbumDetailScreen(
    albumId: String,
    navController: NavController,
    playerViewModel: PlayerViewModel,
    viewModel: AlbumDetailViewModel = hiltViewModel(),
    // Sonora: no PlaylistViewModel, playlists come with step 5.
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val stablePlayerState by playerViewModel.stablePlayerState.collectAsStateWithLifecycle()
    val isMiniPlayerVisible by remember {
        derivedStateOf { stablePlayerState.currentSong != null }
    }
    val favoriteIds by playerViewModel.favoriteSongIds.collectAsStateWithLifecycle()
    // Sonora: as in MainUI, the navigation bar is never compact until there are settings for it.
    val navBarCompactMode = false
    val completedOfflineUris by viewModel.completedOfflineUris.collectAsStateWithLifecycle()
    val allArtists by playerViewModel.artistsFlow.collectAsStateWithLifecycle()

    var showSongInfoBottomSheet by remember { mutableStateOf(false) }
    // Sonora: the original's playerViewModel.selectSongForInfo keeps the song in the view model
    // for its song info sheet; the reduced menu keeps it here, as in LibraryScreen.
    var selectedSongForInfo by remember { mutableStateOf<Song?>(null) }
    val systemNavBarInset = WindowInsets.navigationBars.asPaddingValues().calculateBottomPadding()
    val bottomBarHeightDp = resolveNavBarOccupiedHeight(systemNavBarInset, navBarCompactMode)
    val isDarkTheme = LocalPixelPlayerDarkTheme.current
    val baseColorScheme = MaterialTheme.colorScheme
    val albumArtUri = uiState.album?.albumArtUriString?.takeIf { it.isNotBlank() }
    val albumColorSchemeFlow = remember(albumArtUri) {
        albumArtUri?.let { playerViewModel.themeStateHolder.getAlbumColorSchemeFlow(it, eager = false) }
    }
    val albumColorSchemePair = albumColorSchemeFlow?.collectAsStateWithLifecycle()?.value
    val albumColorScheme = remember(albumColorSchemePair, isDarkTheme, baseColorScheme) {
        albumColorSchemePair?.let { pair -> if (isDarkTheme) pair.dark else pair.light }
            ?: baseColorScheme
    }
    var headerArtworkLoaded by remember(albumArtUri) { mutableStateOf(albumArtUri == null) }
    var themeRequestIssued by remember(albumArtUri) { mutableStateOf(albumArtUri == null) }
    LaunchedEffect(albumArtUri, headerArtworkLoaded, themeRequestIssued) {
        if (!themeRequestIssued && headerArtworkLoaded && albumArtUri != null) {
            themeRequestIssued = true
            playerViewModel.themeStateHolder.ensureAlbumColorScheme(albumArtUri)
        }
    }
    val configuration = LocalConfiguration.current
    val density = LocalDensity.current
    val coroutineScope = rememberCoroutineScope()

    MaterialTheme(
        colorScheme = albumColorScheme,
        typography = MaterialTheme.typography,
        shapes = MaterialTheme.shapes
    ) {

        val fabBottomPadding by animateDpAsState(
            targetValue = if (isMiniPlayerVisible) MiniPlayerHeight + 16.dp else 16.dp,
            label = "fabPadding"
        )

        when {
            uiState.isLoading && uiState.album == null -> {
                Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    ContainedLoadingIndicator()
                }
            }

            uiState.error != null && uiState.album == null -> {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(16.dp),
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

            uiState.album != null -> {
                val album = uiState.album!!
                val songs = uiState.songs
                // Sonora: every song streams from the server, so all of them are download
                // candidates (the original's CloudOfflineRepository.downloadCandidates).
                val cloudSongs = songs
                val allCloudSongsDownloaded = remember(cloudSongs, completedOfflineUris) {
                    cloudSongs.isNotEmpty() && cloudSongs.all { it.contentUriString in completedOfflineUris }
                }
                val songsByDisc = remember(songs) {
                    songs.groupBy { it.discNumber ?: 1 }
                }
                val lazyListState = rememberLazyListState()
                val albumArtist = remember(songs, allArtists) {
                    val artistId = songs.firstOrNull()?.artistId
                    allArtists.firstOrNull { it.id == artistId }
                }
                // Sonora: the server has no genres; the original shows the one most tracks share.
                val albumGenre: String? = null
                val albumDurationMs = remember(songs) { songs.sumOf { it.duration } }
                val songIdsInScope = remember(songs) { songs.mapTo(HashSet()) { it.id } }
                val currentSongId = stablePlayerState.currentSong?.id
                val heroPlayState = when {
                    currentSongId == null || currentSongId !in songIdsInScope -> HeroPlayState.Idle
                    stablePlayerState.isPlaying -> HeroPlayState.Playing
                    else -> HeroPlayState.Paused
                }
                // The bar only needs a background once content actually slides under it.
                val barScrimFadePx = with(density) { 24.dp.toPx() }
                val barContainerAlpha by remember(barScrimFadePx) {
                    derivedStateOf {
                        if (lazyListState.firstVisibleItemIndex > 0) 1f
                        else (lazyListState.firstVisibleItemScrollOffset / barScrimFadePx).coerceIn(0f, 1f)
                    }
                }
                val onHeroShuffle: () -> Unit = {
                    if (songs.isNotEmpty()) {
                        playerViewModel.playSongsShuffled(songs, album.title, startAtZero = true)
                    }
                }
                val onHeroPlay: () -> Unit = {
                    if (heroPlayState == HeroPlayState.Idle) {
                        songs.firstOrNull()?.let { first -> playerViewModel.showAndPlaySong(first, songs) }
                    } else {
                        playerViewModel.playPause()
                    }
                }
                val albumMenuActions = rememberSongCollectionMenuActions(
                    // Sonora: the playlist sheet comes with the playlists in step 5.
                    onAddToPlaylist = {},
                    onPlayNext = { if (songs.isNotEmpty()) playerViewModel.addSelectedAsNext(songs) },
                    onAddToQueue = { if (songs.isNotEmpty()) playerViewModel.addSelectedToQueue(songs) }
                )

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
                // minTopBarHeightPx. Pull the animated height back into range when the inset
                // returns, otherwise the collapsed bar stays too short.
                LaunchedEffect(minTopBarHeightPx, maxTopBarHeightPx) {
                    val clamped = topBarHeight.value.coerceIn(minTopBarHeightPx, maxTopBarHeightPx)
                    if (clamped != topBarHeight.value) topBarHeight.snapTo(clamped)
                }
                val collapseFraction by remember(minTopBarHeightPx, maxTopBarHeightPx) {
                    derivedStateOf {
                        1f - ((topBarHeight.value - minTopBarHeightPx) / (maxTopBarHeightPx - minTopBarHeightPx)).coerceIn(
                            0f,
                            1f
                        )
                    }
                }

                val heroOverlapPx = with(density) { HeroHeaderOverlap.toPx() }
                val listTopPadding = minTopBarHeight + 8.dp
                val listTopPaddingPx = with(density) { listTopPadding.toPx() }
                // The list rides down with the header, overlapping its artwork by the hero overlap.
                val listOffsetPx: () -> Int = {
                    (topBarHeight.value - minTopBarHeightPx - heroOverlapPx * (1f - collapseFraction))
                        .roundToInt()
                        .coerceAtLeast(0)
                }
                // Same per-album variable-font title as the album cards in the Library.
                val heroTitleStyle = remember(album.id, album.title) {
                    GenreTypography.getGenreStyle(album.id.toString(), album.title)
                }
                val heroMetaLine = listOfNotNull(
                    albumGenre,
                    album.year.takeIf { it > 0 }?.toString()
                ).joinToString(" · ").ifBlank { null }
                // Drawn by the floating layer, with a same-sized placeholder in the list.
                val heroChip: @Composable () -> Unit = {
                    val artistName = album.albumArtist?.takeIf { it.isNotBlank() } ?: album.artist
                    DetailHeroArtistChip(
                        name = artistName,
                        imageUrl = albumArtist?.effectiveImageUrl,
                        onClick = albumArtist?.let { artist ->
                            {
                                navController.navigateSafelyReplacing(
                                    route = Screen.ArtistDetail.createRoute(artist.id),
                                    patternToPop = Screen.ArtistDetail.route
                                )
                            }
                        }
                    )
                }
                val showDownloadAction = cloudSongs.isNotEmpty()
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
                    // The bar keeps the download action on the right when the album has cloud songs.
                    collapsedTitleEndPadding = if (showDownloadAction) 88.dp else 24.dp,
                    titleStyle = heroTitleStyle
                )

                // Direction of the latest scroll delta. A gesture that ends while heading back to
                // the top (a fling landing there included) must leave the header fully expanded.
                var lastScrollTowardsTop by remember { mutableStateOf(true) }
                val nestedScrollConnection = remember(minTopBarHeightPx, maxTopBarHeightPx) {
                    object : NestedScrollConnection {
                        override fun onPreScroll(
                            available: Offset,
                            source: NestedScrollSource
                        ): Offset {
                            val delta = available.y
                            val isScrollingDown = delta < 0
                            if (delta != 0f) lastScrollTowardsTop = !isScrollingDown

                            if (!isScrollingDown && (lazyListState.firstVisibleItemIndex > 0 || lazyListState.firstVisibleItemScrollOffset > 0)) {
                                return Offset.Zero
                            }

                            val previousHeight = topBarHeight.value
                            val newHeight =
                                (previousHeight + delta).coerceIn(minTopBarHeightPx, maxTopBarHeightPx)
                            val consumed = newHeight - previousHeight

                            if (consumed.roundToInt() != 0) {
                                coroutineScope.launch {
                                    topBarHeight.snapTo(newHeight)
                                }
                            }

                            val canConsumeScroll = !(isScrollingDown && newHeight == minTopBarHeightPx)
                            return if (canConsumeScroll) Offset(0f, consumed) else Offset.Zero
                        }

                        override suspend fun onPostFling(
                            consumed: Velocity,
                            available: Velocity
                        ): Velocity {
                            return super.onPostFling(consumed, available)
                        }
                    }
                }

                LaunchedEffect(lazyListState.isScrollInProgress) {
                    if (!lazyListState.isScrollInProgress) {
                        val canExpand =
                            lazyListState.firstVisibleItemIndex == 0 && lazyListState.firstVisibleItemScrollOffset == 0
                        val shouldExpand = lastScrollTowardsTop ||
                            topBarHeight.value > (minTopBarHeightPx + maxTopBarHeightPx) / 2

                        val targetValue = if (shouldExpand && canExpand) {
                            maxTopBarHeightPx
                        } else {
                            minTopBarHeightPx
                        }

                        if (topBarHeight.value != targetValue) {
                            coroutineScope.launch {
                                topBarHeight.animateTo(
                                    targetValue,
                                    spring(stiffness = Spring.StiffnessMedium)
                                )
                            }
                        }
                    }
                }

                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(
                            color = MaterialTheme.colorScheme.surface
                        )
                        .nestedScroll(nestedScrollConnection)
                ) {
                    val currentTopBarHeightDp = with(density) { topBarHeight.value.toDp() }
                    val showScrollBar =
                        LocalShowScrollbar.current &&
                        collapseFraction > 0.95f &&
                            (lazyListState.canScrollForward || lazyListState.canScrollBackward)
                    val scrollbarInset by animateDpAsState(
                        targetValue = if (showScrollBar) 8.dp else 0.dp,
                        label = "albumScrollbarInset"
                    )
                    // Sonora: the button stays and does nothing until caching comes with step 7
                    // (Eric, 2026-11-11); see AlbumDetailViewModel.downloadAlbum.
                    val onDownloadClick: () -> Unit = {
                        if (allCloudSongsDownloaded) viewModel.removeAlbumDownloads(cloudSongs)
                        else viewModel.downloadAlbum(cloudSongs)
                    }

                    // Artwork + gradient sit behind the list so the hero can ride up over the
                    // lower part of the image; only the bar (back button, collapsed title,
                    // download action) is drawn above the list.
                    AlbumHeaderBackdrop(
                        album = album,
                        collapseFraction = collapseFraction,
                        headerHeight = currentTopBarHeightDp,
                        headerImageRequestSize = headerImageRequestSize,
                        onHeaderArtworkState = { state ->
                            if (state is AsyncImagePainter.State.Success) {
                                headerArtworkLoaded = true
                            }
                        }
                    )

                    LazyColumn(
                        state = lazyListState,
                        modifier = Modifier
                            .fillMaxSize()
                            .offset { IntOffset(0, listOffsetPx()) },
                        contentPadding = PaddingValues(
                            top = listTopPadding,
                            start = 16.dp,
                            // Fixed, so the hero never shifts; only song rows make room for
                            // the scrollbar (scrollbarInset).
                            end = 16.dp,
                            bottom = fabBottomPadding + 80.dp
                        ),
                        verticalArrangement = Arrangement.spacedBy(4.dp)
                    ) {
                        item(key = "album_hero", contentType = "album_hero") {
                            MediaDetailHero(
                                title = album.title,
                                metaLine = heroMetaLine,
                                chip = heroChip,
                                onShuffle = onHeroShuffle,
                                onPlay = onHeroPlay,
                                playState = heroPlayState,
                                menuActions = albumMenuActions,
                                enabled = songs.isNotEmpty(),
                                dock = heroDock,
                                titleStyle = heroTitleStyle
                            )
                        }
                        if (songs.isNotEmpty()) {
                            item(key = "album_songs_summary", contentType = "detail_songs_summary") {
                                DetailSongsSummary(
                                    songCount = songs.size,
                                    durationMs = albumDurationMs,
                                    modifier = Modifier.padding(start = 8.dp, top = 4.dp, bottom = 8.dp)
                                )
                            }
                        }
                        songsByDisc.forEach { (discNumber, discSongs) ->
                            if (songsByDisc.size > 1) {
                                item(key = "disc_header_$discNumber") {
                                    Text(
                                        text = stringResource(R.string.disc_number_header, discNumber),
                                        style = MaterialTheme.typography.labelLarge,
                                        color = MaterialTheme.colorScheme.primary,
                                        fontWeight = FontWeight.Bold,
                                        modifier = Modifier
                                            .padding(top = 16.dp, bottom = 8.dp, start = 8.dp)
                                    )
                                }
                            }
                            itemsIndexed(
                                items = discSongs,
                                key = { _, song -> "album_song_${song.id}" },
                                contentType = { _, _ -> "album_song" }
                            ) { index, song ->
                                LibraryPlaybackAwareSongItem(
                                    modifier = Modifier.padding(end = scrollbarInset),
                                    song = song,
                                    playerViewModel = playerViewModel,
                                    customShape = getSegmentedItemShape(index, discSongs.size),
                                    showAlbumArt = false,
                                    onMoreOptionsClick = {
                                        selectedSongForInfo = song
                                        showSongInfoBottomSheet = true
                                    },
                                    onClick = { playerViewModel.showAndPlaySong(song, songs) }
                                )
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
                                    bottom = fabBottomPadding + 80.dp
                                )
                        )
                    }

                    // Sonora: the original keeps an older top bar (CollapsingAlbumTopBar) behind
                    // UseSharedCollapsibleTopBarProbe, which is always true; it is not copied.
                    SharedAlbumTopBarProbe(
                        album = album,
                        songsCount = songs.size,
                        collapseFraction = collapseFraction,
                        titleAlpha = heroDock.barTitleAlpha,
                        subtitleAlpha = heroDock.barSubtitleAlpha,
                        containerAlpha = maxOf(barContainerAlpha, heroDock.barContainerAlpha),
                        headerHeight = currentTopBarHeightDp,
                        onBackPressed = { navController.popBackStack() },
                        showDownloadAction = showDownloadAction,
                        allDownloaded = allCloudSongsDownloaded,
                        onDownloadClick = onDownloadClick
                    )

                    DetailHeroFloatingLayer(
                        title = album.title,
                        dock = heroDock,
                        listState = lazyListState,
                        onShuffle = onHeroShuffle,
                        onPlay = onHeroPlay,
                        menuActions = albumMenuActions,
                        playState = heroPlayState,
                        enabled = songs.isNotEmpty(),
                        titleStyle = heroTitleStyle,
                        metaLine = heroMetaLine,
                        chip = heroChip
                    )
                }

                // Sonora: the original's playlist sheet for the whole album comes with step 5.
            }
        }
        if (showSongInfoBottomSheet && selectedSongForInfo != null) {
            val currentSong = selectedSongForInfo
            val isFavorite = remember(currentSong?.id, favoriteIds) {
                derivedStateOf { currentSong?.let { favoriteIds.contains(it.id) } }
            }.value ?: false

            if (currentSong != null) {
                // Sonora: the reduced song menu (Eric, 2026-11-11), as in LibraryScreen. The
                // original also passes delete, navigation to the album, artists and genre, and
                // metadata editing.
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

@Composable
private fun AlbumHeaderBackdrop(
    album: Album,
    collapseFraction: Float,
    headerHeight: Dp,
    headerImageRequestSize: Size,
    onHeaderArtworkState: ((AsyncImagePainter.State) -> Unit)? = null,
) {
    val surfaceColor = MaterialTheme.colorScheme.surface
    val expandedContentAlpha = 1f - (collapseFraction * 2f).coerceIn(0f, 1f)
    val headerOverlayBrush = remember(surfaceColor, expandedContentAlpha) {
        Brush.verticalGradient(
            colors = listOf(
                Color.Transparent,
                surfaceColor.copy(alpha = 0.22f * expandedContentAlpha),
                surfaceColor.copy(alpha = 0.82f * expandedContentAlpha),
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
        SmartImage(
            model = album.albumArtUriString,
            contentDescription = stringResource(R.string.album_cover_for, album.title),
            contentScale = ContentScale.Crop,
            targetSize = headerImageRequestSize,
            allowHardware = true,
            crossfadeDurationMillis = 0,
            alpha = expandedContentAlpha,
            onState = onHeaderArtworkState,
            modifier = Modifier.fillMaxSize()
        )
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(headerOverlayBrush)
        )
    }
}

/** How far the hero rides up over the bottom of the expanded artwork header. */
private val HeroHeaderOverlap = 120.dp

@Composable
private fun SharedAlbumTopBarProbe(
    album: Album,
    songsCount: Int,
    collapseFraction: Float,
    titleAlpha: Float,
    subtitleAlpha: Float,
    containerAlpha: Float,
    headerHeight: Dp,
    onBackPressed: () -> Unit,
    showDownloadAction: Boolean,
    allDownloaded: Boolean,
    onDownloadClick: () -> Unit,
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

        val titleEndPadding = if (showDownloadAction) 88.dp else 24.dp
        CollapsibleCommonTopBar(
            title = album.title,
            subtitle = "${album.artist} • ${formatSongCount(songsCount)}",
            collapseFraction = collapseFraction,
            headerHeight = headerHeight,
            onBackClick = onBackPressed,
            // Transparent until something scrolls under the bar; the hero overlaps the
            // header, so a collapse-driven fill would dim it mid-gesture.
            containerColor = surfaceColor.copy(alpha = containerAlpha * solidAlpha),
            collapsedTitleStartPadding = 68.dp,
            expandedTitleStartPadding = 24.dp,
            collapsedTitleEndPadding = titleEndPadding,
            expandedTitleEndPadding = titleEndPadding,
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
            // own title only shows for titles too long to dock, and the subtitle settles in
            // under the docked title.
            contentColor = MaterialTheme.colorScheme.onSurface.copy(alpha = titleAlpha),
            subtitleColor = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = subtitleAlpha),
            fadeSubtitleOnCollapse = false,
            syncStatusBarWithContainer = false,
            actions = {
                if (showDownloadAction) {
                    Box(modifier = Modifier.padding(end = 12.dp, top = 4.dp)) {
                        FilledIconButton(
                            onClick = onDownloadClick,
                            colors = IconButtonDefaults.filledIconButtonColors(
                                containerColor = MaterialTheme.colorScheme.surfaceContainerLow
                            )
                        ) {
                            Icon(
                                imageVector = if (allDownloaded) Icons.Rounded.CloudDone else Icons.Rounded.CloudDownload,
                                contentDescription = stringResource(
                                    if (allDownloaded) R.string.cloud_album_remove_downloads else R.string.cloud_album_download
                                )
                            )
                        }
                    }
                }
            }
        )
    }
}
