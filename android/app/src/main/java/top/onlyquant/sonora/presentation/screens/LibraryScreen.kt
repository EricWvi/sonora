@file:OptIn(ExperimentalMaterial3ExpressiveApi::class)

package top.onlyquant.sonora.presentation.screens

import top.onlyquant.sonora.presentation.navigation.navigateSafely
import top.onlyquant.sonora.presentation.navigation.navigateSafelyReplacing

import android.os.Trace
import top.onlyquant.sonora.utils.traceSection
import android.text.format.Formatter
import androidx.compose.animation.ExperimentalAnimationApi
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.animateContentSize
import androidx.compose.animation.core.FastOutSlowInEasing
import androidx.compose.animation.core.animateDpAsState
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.tween
import androidx.compose.animation.core.updateTransition
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.wrapContentHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListItemInfo
import androidx.compose.ui.zIndex
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.GridItemSpan
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.lazy.grid.itemsIndexed
import androidx.compose.foundation.lazy.grid.rememberLazyGridState
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.rounded.ViewList
import androidx.compose.material.icons.filled.Album
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material.icons.filled.FavoriteBorder
import androidx.compose.material.icons.rounded.Bookmark
import androidx.compose.material.icons.rounded.MoreVert
import androidx.compose.material.icons.rounded.MoreHoriz
import androidx.compose.material.icons.rounded.SelectAll
import androidx.compose.material.icons.rounded.Deselect
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.FilledTonalButton
import androidx.compose.material3.LinearWavyProgressIndicator
import androidx.compose.material.icons.rounded.ViewModule
import top.onlyquant.sonora.presentation.components.ToggleSegmentButton
import androidx.compose.material3.LoadingIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButtonDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.PrimaryScrollableTabRow
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Surface
import androidx.compose.material3.TabRowDefaults
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.material3.ExperimentalMaterial3ExpressiveApi
import androidx.compose.material3.FilledTonalButton
import top.onlyquant.sonora.ui.theme.ShapeCache
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.repeatOnLifecycle
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.TransformOrigin
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.hapticfeedback.HapticFeedbackType
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.platform.LocalHapticFeedback
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.AnnotatedString
import androidx.compose.ui.text.ExperimentalTextApi
import androidx.compose.ui.text.PlatformTextStyle
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.Font
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontVariation
import androidx.compose.ui.text.rememberTextMeasurer
import androidx.compose.ui.text.style.LineHeightStyle
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import top.onlyquant.sonora.ui.theme.LocalPixelPlayerDarkTheme
import androidx.compose.ui.unit.sp
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import androidx.media3.common.util.UnstableApi
import androidx.navigation.NavController
import androidx.paging.compose.collectAsLazyPagingItems
import coil.compose.AsyncImage
import coil.compose.AsyncImagePainter
import coil.imageLoader
import coil.request.ImageRequest
import coil.size.Size
import top.onlyquant.sonora.R
import top.onlyquant.sonora.presentation.components.ShimmerBox
import top.onlyquant.sonora.data.model.Album
import top.onlyquant.sonora.data.model.Artist
import top.onlyquant.sonora.data.model.Song
import top.onlyquant.sonora.data.model.SortOption
import top.onlyquant.sonora.data.model.StorageFilter
import top.onlyquant.sonora.presentation.components.MiniPlayerHeight
import top.onlyquant.sonora.presentation.components.SmartImage
import top.onlyquant.sonora.presentation.components.resolveMainScreenBottomGradientHeight
import top.onlyquant.sonora.presentation.components.resolveNavBarOccupiedHeight
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.navigationBars
import androidx.compose.foundation.layout.asPaddingValues
import androidx.compose.ui.res.stringResource
import top.onlyquant.sonora.presentation.components.PlayerSheetAwareBackHandler
import top.onlyquant.sonora.presentation.components.SongInfoBottomSheet
import top.onlyquant.sonora.presentation.components.subcomps.LibraryActionRow
import top.onlyquant.sonora.presentation.navigation.Screen
import top.onlyquant.sonora.presentation.components.MultiSelectionBottomSheet
import top.onlyquant.sonora.presentation.components.subcomps.SelectionActionRow
import top.onlyquant.sonora.presentation.components.subcomps.SelectionCountPill
import top.onlyquant.sonora.presentation.viewmodel.ColorSchemePair
import top.onlyquant.sonora.presentation.viewmodel.PlayerUiState
import top.onlyquant.sonora.presentation.viewmodel.PlayerViewModel
import top.onlyquant.sonora.presentation.viewmodel.sharedPlayerViewModel
import top.onlyquant.sonora.data.model.LibraryTabId
import top.onlyquant.sonora.data.model.toLibraryTabIdOrNull
import top.onlyquant.sonora.presentation.screens.search.components.GenreTypography
import top.onlyquant.sonora.presentation.utils.SegmentedItemShapeDefaults
import top.onlyquant.sonora.presentation.utils.clipCardCornersFallback
import top.onlyquant.sonora.presentation.utils.getSegmentedItemShape
import top.onlyquant.sonora.presentation.utils.seamlessSelectionScale
import top.onlyquant.sonora.presentation.viewmodel.LibraryViewModel
import top.onlyquant.sonora.presentation.selection.appendDistinctSelection
import top.onlyquant.sonora.presentation.selection.dragToSelect
import top.onlyquant.sonora.presentation.selection.rememberDragToSelectState
import top.onlyquant.sonora.presentation.selection.selectionIndex
import top.onlyquant.sonora.presentation.selection.toggleOrderedSelection
import top.onlyquant.sonora.utils.formatSongCount
import androidx.paging.compose.collectAsLazyPagingItems
import android.content.Intent
import android.net.Uri
import android.os.Build
import android.os.Environment
import android.provider.Settings
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.annotation.RequiresApi
import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInHorizontally
import androidx.compose.animation.slideOutHorizontally
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.basicMarquee
import androidx.compose.foundation.gestures.detectHorizontalDragGestures
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.interaction.collectIsPressedAsState
import androidx.compose.foundation.layout.IntrinsicSize
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.heightIn
import androidx.compose.material.icons.rounded.Edit
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.TextButton
import kotlinx.collections.immutable.ImmutableList
import kotlinx.collections.immutable.persistentListOf
import kotlinx.collections.immutable.toImmutableList
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.launch
import kotlinx.coroutines.CancellationException
import racra.compose.smooth_corner_rect_library.AbsoluteSmoothCornerShape
import androidx.compose.material3.FilledIconButton
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.material3.pulltorefresh.PullToRefreshDefaults
import androidx.compose.material3.pulltorefresh.rememberPullToRefreshState
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.window.Dialog
import androidx.compose.ui.window.DialogProperties
import top.onlyquant.sonora.presentation.components.AutoScrollingTextOnDemand
import top.onlyquant.sonora.presentation.components.subcomps.PlayingEqIcon
import top.onlyquant.sonora.ui.theme.RoundedSans
import java.util.Locale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalLocale
import android.widget.Toast
import androidx.compose.foundation.layout.wrapContentWidth
import androidx.compose.ui.focus.focusModifier
import kotlinx.coroutines.flow.first
import androidx.paging.compose.collectAsLazyPagingItems
import androidx.paging.compose.itemContentType
import androidx.paging.LoadState
import top.onlyquant.sonora.presentation.components.ExpressiveScrollBar
import top.onlyquant.sonora.presentation.components.animatedScrollbarInset
import top.onlyquant.sonora.ui.theme.LocalShowScrollbar
import top.onlyquant.sonora.presentation.components.LibrarySortBottomSheet
import top.onlyquant.sonora.presentation.library.KeepUnscrolledListAtTop
import timber.log.Timber
import java.io.File
import kotlin.math.abs
import top.onlyquant.sonora.presentation.components.rememberModalSheetState

val ListExtraBottomGap = 30.dp
val PlayerSheetCollapsedCornerRadius = 32.dp

/** Tabs whose header shows the quick source filter (all / cloud / on-device). */
private fun LibraryTabId.hasStorageFilterButton(): Boolean = when (this) {
    LibraryTabId.SONGS, LibraryTabId.ALBUMS, LibraryTabId.ARTISTS -> true
    LibraryTabId.PLAYLISTS -> false
}

private const val PULL_REFRESH_MIN_VISIBLE_MS = 900L
private const val PULL_REFRESH_MAX_VISIBLE_MS = 1_500L
private const val INLINE_SYNC_MIN_VISIBLE_MS = 600L

private data class LibraryScreenPlayerProjection(
    val currentStorageFilter: StorageFilter = StorageFilter.ALL,
    val currentSongSortOption: SortOption = SortOption.SongTitleAZ,
    val currentAlbumSortOption: SortOption = SortOption.AlbumTitleAZ,
    val currentArtistSortOption: SortOption = SortOption.ArtistNameAZ,
    val isAlbumsListView: Boolean = false,
    val isLoadingLibraryCategories: Boolean = true,
    val isSyncingLibrary: Boolean = false,
    val isLoadingInitialSongs: Boolean = true
)

private fun PlayerUiState.toLibraryScreenProjection(): LibraryScreenPlayerProjection =
    LibraryScreenPlayerProjection(
        currentStorageFilter = currentStorageFilter,
        currentSongSortOption = currentSongSortOption,
        currentAlbumSortOption = currentAlbumSortOption,
        currentArtistSortOption = currentArtistSortOption,
        isAlbumsListView = isAlbumsListView,
        isLoadingLibraryCategories = isLoadingLibraryCategories,
        isSyncingLibrary = isSyncingLibrary,
        isLoadingInitialSongs = isLoadingInitialSongs
    )

/**
 * From LibraryScreen in PixelPlayerOSS v0.4.0-alpha.53 (GPL-3.0), with the songs, albums, artists
 * and playlists tabs (Eric, 2026-11-11). Sonora:
 * - No folders, liked or genres tabs, and no compact "pill" navigation (a setting whose default
 *   is the tab row, which Sonora has no screen for yet).
 * - The tab row has no "edit tabs" button and the tabs keep their default order (Eric,
 *   2026-11-11).
 * - The header has no audio bookmarks button (Sonora has no bookmarks); the settings button does
 *   nothing until there is a settings screen.
 * - The sort sheet has no "cloud only" switch: every song is on the server.
 * - The storage filter button is shown, but does nothing (Eric, 2026-11-11).
 * - The song menu keeps play, like, add to queue, play next and add to playlist; the selection
 *   sheet keeps add to queue, play next, add to playlist and like. Like and playlists come with
 *   step 5, so those do nothing yet.
 * - The playlists tab is empty until step 5 brings the playlists (the original's PlaylistContainer).
 * - Pull to refresh and the sync indicators follow [LibraryViewModel.sync], the original's
 *   SyncManager.
 */
@RequiresApi(Build.VERSION_CODES.R)
@OptIn(ExperimentalMaterial3Api::class, ExperimentalMaterial3ExpressiveApi::class)
@androidx.annotation.OptIn(UnstableApi::class)
@Composable
fun LibraryScreen(
    navController: NavController,
    playerViewModel: PlayerViewModel = sharedPlayerViewModel(),
    libraryViewModel: LibraryViewModel = hiltViewModel()
) {
    val context = LocalContext.current
    val haptic = LocalHapticFeedback.current
    val lastTabIndex by playerViewModel.lastLibraryTabIndexFlow.collectAsStateWithLifecycle()
    val favoriteIds by playerViewModel.favoriteSongIds.collectAsStateWithLifecycle()
    val scope = rememberCoroutineScope()
    var isRefreshing by remember { mutableStateOf(false) }
    // Sonora: one sync both fetches and applies the changes, so both flags follow it.
    val isFetchingChanges by libraryViewModel.isSyncing.collectAsStateWithLifecycle()
    val isSyncing by libraryViewModel.isSyncing.collectAsStateWithLifecycle()

    var showSongInfoBottomSheet by remember { mutableStateOf(false) }
    var selectedSongForInfo by remember { mutableStateOf<Song?>(null) }
    val tabTitles by playerViewModel.libraryTabsFlow.collectAsStateWithLifecycle()
    val currentTabId by playerViewModel.currentLibraryTabId.collectAsStateWithLifecycle()
    val tabCount = tabTitles.size.coerceAtLeast(1)
    val normalizedLastTabIndex = positiveMod(lastTabIndex, tabCount)
    val pagerState = rememberPagerState(initialPage = normalizedLastTabIndex) { tabCount }
    val currentTabIndex by remember(pagerState, tabTitles) {
        derivedStateOf {
            resolveTabIndex(
                page = pagerState.currentPage,
                tabCount = tabTitles.size
            )
        }
    }
    // The pager's tab, available in this frame. currentTabId only catches up after
    // onLibraryTabSelected runs, so header actions, selection and back read this one instead.
    val visibleCurrentTabId = resolveVisibleLibraryTabId(tabTitles, currentTabIndex, fallback = currentTabId)
    val isSortSheetVisible by playerViewModel.isSortingSheetVisible.collectAsStateWithLifecycle()

    val multiSelectionState = playerViewModel.multiSelectionStateHolder
    val selectedSongs by multiSelectionState.selectedSongs.collectAsStateWithLifecycle()
    val isSelectionMode by multiSelectionState.isSelectionMode.collectAsStateWithLifecycle()
    val selectedSongIds by multiSelectionState.selectedSongIds.collectAsStateWithLifecycle()
    var showMultiSelectionSheet by remember { mutableStateOf(false) }
    var selectedAlbums by remember { mutableStateOf<ImmutableList<Album>>(persistentListOf()) }
    val selectedAlbumIds = remember(selectedAlbums) { selectedAlbums.map { it.id }.toSet() }
    val isAlbumSelectionMode = selectedAlbums.isNotEmpty()
    var selectedArtists by remember { mutableStateOf<ImmutableList<Artist>>(persistentListOf()) }
    val selectedArtistIds = remember(selectedArtists) { selectedArtists.map { it.id }.toSet() }
    val isArtistSelectionMode = selectedArtists.isNotEmpty()
    val latestSelectedAlbums by rememberUpdatedState(selectedAlbums)
    val latestSelectedArtists by rememberUpdatedState(selectedArtists)

    var songsShowLocateButton by remember { mutableStateOf(false) }
    var songsLocateAction by remember { mutableStateOf<(() -> Unit)?>(null) }

    val onSongLongPress: (Song) -> Unit = remember(multiSelectionState, haptic) {
        { song ->
            haptic.performHapticFeedback(HapticFeedbackType.LongPress)
            multiSelectionState.toggleSelection(song)
        }
    }

    val onSongSelectionToggle: (Song) -> Unit = remember(multiSelectionState) {
        { song -> multiSelectionState.toggleSelection(song) }
    }

    // Drag-to-select reads the holder directly: the collected selectedSongIds lag a frame behind
    // the long press that starts the drag.
    val currentSelectedSongIds: () -> Set<String> = remember(multiSelectionState) {
        { multiSelectionState.selectedSongIds.value }
    }

    val onSongDragSelectionChange: (List<Song>, List<Song>) -> Unit = remember(multiSelectionState) {
        { select, deselect -> multiSelectionState.applyDragSelection(select, deselect) }
    }

    val toggleAlbumSelection: (Album) -> Unit = remember(selectedAlbums, multiSelectionState) {
        { album ->
            // A changed category selection invalidates any songs resolved for an earlier snapshot.
            multiSelectionState.clearSelection()
            showMultiSelectionSheet = false
            selectedAlbums = toggleOrderedSelection(selectedAlbums, album, Album::id).toImmutableList()
        }
    }

    val onAlbumLongPress: (Album) -> Unit = remember(toggleAlbumSelection, haptic) {
        { album ->
            haptic.performHapticFeedback(HapticFeedbackType.LongPress)
            toggleAlbumSelection(album)
        }
    }

    val onAlbumSelectionToggle: (Album) -> Unit = remember(toggleAlbumSelection) {
        { album -> toggleAlbumSelection(album) }
    }

    val getAlbumSelectionIndex: (Long) -> Int? = remember(selectedAlbums) {
        { albumId -> selectionIndex(selectedAlbums, albumId, Album::id) }
    }

    val toggleArtistSelection: (Artist) -> Unit = remember(selectedArtists, multiSelectionState) {
        { artist ->
            // A changed category selection invalidates any songs resolved for an earlier snapshot.
            multiSelectionState.clearSelection()
            showMultiSelectionSheet = false
            selectedArtists = toggleOrderedSelection(selectedArtists, artist, Artist::id).toImmutableList()
        }
    }

    val onArtistLongPress: (Artist) -> Unit = remember(toggleArtistSelection, haptic) {
        { artist ->
            haptic.performHapticFeedback(HapticFeedbackType.LongPress)
            toggleArtistSelection(artist)
        }
    }

    val onArtistSelectionToggle: (Artist) -> Unit = remember(toggleArtistSelection) {
        { artist -> toggleArtistSelection(artist) }
    }

    val getArtistSelectionIndex: (Long) -> Int? = remember(selectedArtists) {
        { artistId -> selectionIndex(selectedArtists, artistId, Artist::id) }
    }

    val presentResolvedCategorySongs: (List<Song>) -> Unit = remember(
        multiSelectionState,
        playerViewModel,
        context
    ) {
        { songs ->
            if (songs.isEmpty()) {
                playerViewModel.sendToast(context.getString(R.string.no_valid_songs))
            } else {
                multiSelectionState.replaceSelection(songs)
                showMultiSelectionSheet = true
            }
        }
    }

    val openSelectedAlbumActions: () -> Unit = remember(
        selectedAlbums,
        playerViewModel,
        scope,
        presentResolvedCategorySongs
    ) {
        {
            val selectionSnapshot = selectedAlbums.toList()
            val snapshotIds = selectionSnapshot.map(Album::id)
            scope.launch {
                runCatching {
                    playerViewModel.resolveAlbumSongsForBatchActions(selectionSnapshot)
                }.onSuccess { songs ->
                    if (latestSelectedAlbums.map(Album::id) == snapshotIds) {
                        presentResolvedCategorySongs(songs)
                    }
                }
                    .onFailure { error ->
                        if (error is CancellationException) throw error
                        Timber.tag("LibrarySelection").e(error, "Unable to resolve selected albums")
                        playerViewModel.sendToast(context.getString(R.string.no_valid_songs))
                    }
            }
        }
    }

    val openSelectedArtistActions: () -> Unit = remember(
        selectedArtists,
        playerViewModel,
        scope,
        presentResolvedCategorySongs
    ) {
        {
            val selectionSnapshot = selectedArtists.toList()
            val snapshotIds = selectionSnapshot.map(Artist::id)
            scope.launch {
                runCatching {
                    playerViewModel.resolveArtistSongsForBatchActions(selectionSnapshot)
                }.onSuccess { songs ->
                    if (latestSelectedArtists.map(Artist::id) == snapshotIds) {
                        presentResolvedCategorySongs(songs)
                    }
                }
                    .onFailure { error ->
                        if (error is CancellationException) throw error
                        Timber.tag("LibrarySelection").e(error, "Unable to resolve selected artists")
                        playerViewModel.sendToast(context.getString(R.string.no_valid_songs))
                    }
            }
        }
    }

    // Sonora: the original's playerViewModel.selectSongForInfo keeps the song in the view model
    // for its song info sheet; the reduced menu keeps it here.
    val stableOnMoreOptionsClick: (Song) -> Unit = remember {
        { song ->
            selectedSongForInfo = song
            showSongInfoBottomSheet = true
        }
    }
    var isMinDelayActive by remember { mutableStateOf(false) }
    var refreshGeneration by remember { mutableStateOf(0) }

    val onRefresh: () -> Unit = remember(scope, libraryViewModel) {
        {
            val currentRefreshGeneration = refreshGeneration + 1
            refreshGeneration = currentRefreshGeneration
            isMinDelayActive = true
            isRefreshing = true
            libraryViewModel.sync()
            scope.launch {
                kotlinx.coroutines.delay(PULL_REFRESH_MIN_VISIBLE_MS)
                if (currentRefreshGeneration != refreshGeneration) return@launch
                isMinDelayActive = false
                val stillFetching = libraryViewModel.isSyncing.first()
                if (!stillFetching) {
                    isRefreshing = false
                    return@launch
                }

                val remainingVisibleMs =
                    (PULL_REFRESH_MAX_VISIBLE_MS - PULL_REFRESH_MIN_VISIBLE_MS)
                        .coerceAtLeast(0L)
                if (remainingVisibleMs > 0L) {
                    kotlinx.coroutines.delay(remainingVisibleMs)
                }
                if (currentRefreshGeneration != refreshGeneration) return@launch
                isRefreshing = false
            }
        }
    }

    LaunchedEffect(isFetchingChanges) {
        if (!isFetchingChanges && !isMinDelayActive) {
            isRefreshing = false
        }
    }

    var inlineSyncVisible by remember { mutableStateOf(false) }
    var inlineSyncShownAt by remember { mutableStateOf<Long?>(null) }
    LaunchedEffect(isSyncing, isRefreshing) {
        if (isSyncing && !isRefreshing) {
            if (!inlineSyncVisible) {
                inlineSyncShownAt = System.currentTimeMillis()
                inlineSyncVisible = true
            }
        } else if (isRefreshing) {
            inlineSyncVisible = false
            inlineSyncShownAt = null
        } else if (inlineSyncVisible) {
            val shownAt = inlineSyncShownAt
            val elapsed = if (shownAt != null) {
                System.currentTimeMillis() - shownAt
            } else {
                INLINE_SYNC_MIN_VISIBLE_MS
            }
            val remaining = INLINE_SYNC_MIN_VISIBLE_MS - elapsed
            if (remaining > 0) {
                kotlinx.coroutines.delay(remaining)
            }
            inlineSyncVisible = false
            inlineSyncShownAt = null
        }
    }

    val hasSelectionInCurrentTab = when (visibleCurrentTabId) {
        LibraryTabId.ALBUMS -> isAlbumSelectionMode
        LibraryTabId.SONGS -> isSelectionMode
        LibraryTabId.ARTISTS -> isArtistSelectionMode
        // Sonora: playlist selection comes with the playlists in step 5.
        LibraryTabId.PLAYLISTS -> false
    }

    // The coordinator restores this callback as soon as the sheet stops handling back, even if
    // LibraryScreen has not recomposed yet. It also preserves player priority after late registration.
    PlayerSheetAwareBackHandler(
        coordinator = playerViewModel.playerSheetBackCoordinator,
        enabled = hasSelectionInCurrentTab
    ) {
        when (visibleCurrentTabId) {
            LibraryTabId.ALBUMS -> {
                selectedAlbums = persistentListOf()
                multiSelectionState.clearSelection()
                showMultiSelectionSheet = false
            }

            LibraryTabId.ARTISTS -> {
                selectedArtists = persistentListOf()
                multiSelectionState.clearSelection()
                showMultiSelectionSheet = false
            }

            LibraryTabId.SONGS -> {
                multiSelectionState.clearSelection()
                showMultiSelectionSheet = false
            }

            LibraryTabId.PLAYLISTS -> Unit
        }
    }

    LaunchedEffect(Unit) {
        traceSection("LibraryScreen.InitialTabLoad") {
            playerViewModel.onLibraryTabSelected(normalizedLastTabIndex)
        }
    }

    // Also keyed on the tab itself: editing the tab list can put a different tab at the same index.
    LaunchedEffect(currentTabIndex, tabTitles.getOrNull(currentTabIndex)) {
        traceSection("LibraryScreen.PageChangeTabLoad") {
            playerViewModel.onLibraryTabSelected(currentTabIndex)
        }

        multiSelectionState.clearSelection()
        selectedAlbums = persistentListOf()
        selectedArtists = persistentListOf()
        showMultiSelectionSheet = false
    }

    val fabState by remember { derivedStateOf { currentTabIndex } }
    val transition = updateTransition(
        targetState = fabState,
        label = "Action Button Icon Transition"
    )

    val systemNavBarInset = WindowInsets.navigationBars.asPaddingValues().calculateBottomPadding()
    // Sonora: as in MainUI, the navigation bar is never compact until there are settings for it.
    val navBarCompactMode = false
    val bottomBarHeightDp = resolveNavBarOccupiedHeight(systemNavBarInset, navBarCompactMode)
    val bottomGradientHeight = resolveMainScreenBottomGradientHeight(navBarCompactMode)

    val dm = LocalPixelPlayerDarkTheme.current

    val iconRotation by transition.animateFloat(
        label = "Action Button Icon Rotation",
        transitionSpec = {
            tween(durationMillis = 300, easing = FastOutSlowInEasing)
        }
    ) { page ->
        when (tabTitles.getOrNull(page)?.toLibraryTabIdOrNull()) {
            LibraryTabId.PLAYLISTS -> 0f
            else -> 360f
        }
    }

    val primaryContainer = MaterialTheme.colorScheme.primaryContainer
    val onPrimaryContainer = MaterialTheme.colorScheme.onPrimaryContainer
    val gradientColorsDark = remember(primaryContainer) {
        listOf(primaryContainer.copy(alpha = 0.5f), Color.Transparent).toImmutableList()
    }
    val gradientColorsLight = remember(onPrimaryContainer) {
        listOf(onPrimaryContainer.copy(alpha = 0.2f), Color.Transparent).toImmutableList()
    }

    val gradientColors = if (dm) gradientColorsDark else gradientColorsLight

    val gradientBrush = remember(gradientColors) {
        Brush.verticalGradient(colors = gradientColors)
    }

    val headerContainerColor = MaterialTheme.colorScheme.primaryContainer.copy(alpha = 0.4f)

    Scaffold(
        modifier = Modifier.background(brush = gradientBrush),
        topBar = {
            Column(
                modifier = Modifier.background(headerContainerColor)
            ) {
                TopAppBar(
                    title = {
                        Text(
                            modifier = Modifier.padding(start = 8.dp),
                            text = stringResource(R.string.presentation_batch_d_library_title),
                            fontFamily = RoundedSans,
                            fontWeight = FontWeight.ExtraBold,
                            color = MaterialTheme.colorScheme.primary,
                            fontSize = 40.sp,
                            letterSpacing = 1.sp
                        )
                    },
                    actions = {
                        // Sonora: no audio bookmarks button (Sonora has no bookmarks, Eric 2026-11-11).
                        FilledIconButton(
                            modifier = Modifier.padding(end = 14.dp),
                            colors = IconButtonDefaults.filledIconButtonColors(
                                containerColor = MaterialTheme.colorScheme.primaryContainer,
                                contentColor = MaterialTheme.colorScheme.onPrimaryContainer
                            ),
                            onClick = {
                                // Sonora: opens Screen.Settings in the original; nothing yet,
                                // there is no settings screen (Eric, 2026-11-11).
                            }
                        ) {
                            Icon(
                                painter = painterResource(R.drawable.rounded_settings_24),
                                contentDescription = stringResource(R.string.presentation_batch_d_open_settings_cd)
                            )
                        }
                    },
                    colors = TopAppBarDefaults.topAppBarColors(
                        containerColor = Color.Transparent,
                        scrolledContainerColor = Color.Transparent
                    )
                )
                val showTabIndicator = false
                PrimaryScrollableTabRow(
                    selectedTabIndex = currentTabIndex,
                    containerColor = Color.Transparent,
                    edgePadding = 12.dp,
                    indicator = {
                        if (showTabIndicator) {
                            TabRowDefaults.PrimaryIndicator(
                                modifier = Modifier.tabIndicatorOffset(selectedTabIndex = currentTabIndex),
                                height = 3.dp,
                                color = MaterialTheme.colorScheme.primary
                            )
                        }
                    },
                    divider = {}
                ) {
                    tabTitles.forEachIndexed { index, rawId ->
                        val tabId = rawId.toLibraryTabIdOrNull() ?: LibraryTabId.SONGS
                        TabAnimation(
                            index = index,
                            title = tabId.storageKey,
                            selectedIndex = currentTabIndex,
                            onClick = {
                                scope.launch {
                                    pagerState.animateScrollToPage(
                                        targetPageForTabIndex(
                                            targetTabIndex = index,
                                            tabCount = tabTitles.size
                                        )
                                    )
                                }
                            }
                        ) {
                            Text(
                                text = stringResource(tabId.titleRes).uppercase(LocalLocale.current.platformLocale),
                                style = MaterialTheme.typography.labelLarge,
                                fontWeight = if (currentTabIndex == index) FontWeight.Bold else FontWeight.Medium
                            )
                        }
                    }
                    // Sonora: no "edit tabs" button after the tabs (Eric, 2026-11-11).
                }
            }
        }
    ) { innerScaffoldPadding ->
        // Starts from the current state rather than defaults: coming back to Library with a
        // non-default sort or the albums list view would otherwise see a sort change and a
        // grid/list switch on the first frame and lose the saved scroll position.
        val playerUiState by remember(playerViewModel) {
            playerViewModel.playerUiState
                .map { uiState -> uiState.toLibraryScreenProjection() }
                .distinctUntilChanged()
        }.collectAsStateWithLifecycle(
            initialValue = remember(playerViewModel) {
                playerViewModel.playerUiState.value.toLibraryScreenProjection()
            }
        )
        val isLibraryContentEmpty by remember(playerViewModel) {
            combine(
                playerViewModel.songCountFlow,
                playerViewModel.albumsFlow,
                playerViewModel.artistsFlow
            ) { songCount, albums, artists ->
                songCount == 0 && albums.isEmpty() && artists.isEmpty()
            }.distinctUntilChanged()
        }.collectAsStateWithLifecycle(initialValue = true)

        Box(
            modifier = Modifier
                .padding(top = innerScaffoldPadding.calculateTopPadding())
                .fillMaxSize()
        ) {
            Column(
                modifier = Modifier
                    .background(color = headerContainerColor)
                    .fillMaxSize()
            ) {
                Surface(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(horizontal = 0.dp, vertical = 0.dp),
                    color = MaterialTheme.colorScheme.surface,
                    shape = AbsoluteSmoothCornerShape(
                        cornerRadiusTL = 34.dp,
                        smoothnessAsPercentBL = 60,
                        cornerRadiusBL = 0.dp,
                        smoothnessAsPercentBR = 60,
                        cornerRadiusBR = 0.dp,
                        smoothnessAsPercentTR = 60,
                        cornerRadiusTR = 34.dp,
                        smoothnessAsPercentTL = 60
                    )
                ) {
                    Column(Modifier.fillMaxSize()) {
                        val availableSortOptions by playerViewModel.availableSortOptions.collectAsStateWithLifecycle()
                        val sanitizedSortOptions = remember(availableSortOptions, currentTabId) {
                            val cleaned = availableSortOptions.filterIsInstance<SortOption>()
                            val ensured = if (cleaned.any { option ->
                                    option.storageKey == currentTabId.defaultSort.storageKey
                                }
                            ) {
                                cleaned
                            } else {
                                buildList {
                                    add(currentTabId.defaultSort)
                                    addAll(cleaned)
                                }
                            }

                            val distinctByKey = ensured.distinctBy { it.storageKey }
                            distinctByKey.ifEmpty { listOf(currentTabId.defaultSort) }.toImmutableList()
                        }

                        val isLibraryLoading by libraryViewModel.isLoadingLibrary.collectAsStateWithLifecycle()
                        val hasCurrentSong by remember(playerViewModel) {
                            playerViewModel.stablePlayerState
                                .map { state -> state.currentSong != null && state.currentSong != Song.emptySong() }
                                .distinctUntilChanged()
                        }.collectAsStateWithLifecycle(initialValue = false)
                        val isShuffleEnabled by remember(playerViewModel) {
                            playerViewModel.stablePlayerState
                                .map { it.isShuffleEnabled }
                                .distinctUntilChanged()
                        }.collectAsStateWithLifecycle(initialValue = false)

                        val currentSelectedSortOption: SortOption? = when (currentTabId) {
                            LibraryTabId.SONGS -> playerUiState.currentSongSortOption
                            LibraryTabId.ALBUMS -> playerUiState.currentAlbumSortOption
                            LibraryTabId.ARTISTS -> playerUiState.currentArtistSortOption
                            // Sonora: the playlists' sort comes with them in step 5.
                            LibraryTabId.PLAYLISTS -> null
                        }

                        val locateAction = when (visibleCurrentTabId) {
                            LibraryTabId.SONGS -> songsLocateAction
                            else -> null
                        }
                        // A tab that reports the button without registering an action would show a dead button.
                        val showLocateButton = locateAction != null && when (visibleCurrentTabId) {
                            LibraryTabId.SONGS -> songsShowLocateButton
                            else -> false
                        }

                        val onSortOptionChanged: (SortOption) -> Unit = remember(playerViewModel, currentTabId) {
                            { option ->
                                when (currentTabId) {
                                    LibraryTabId.SONGS -> playerViewModel.sortSongs(option)
                                    LibraryTabId.ALBUMS -> playerViewModel.sortAlbums(option)
                                    LibraryTabId.ARTISTS -> playerViewModel.sortArtists(option)
                                    // Sonora: the playlists' sort comes with them in step 5.
                                    LibraryTabId.PLAYLISTS -> Unit
                                }
                            }
                        }

                        AnimatedContent(
                            targetState = hasSelectionInCurrentTab,
                            label = "ActionRowModeSwitch",
                            transitionSpec = {
                                (slideInHorizontally { -it } + fadeIn()) togetherWith
                                        (slideOutHorizontally { it } + fadeOut())
                            },
                            modifier = Modifier
                                .padding(
                                    top = 6.dp,
                                    start = 10.dp,
                                    end = 10.dp
                                )
                                .heightIn(min = 56.dp)
                        ) { inSelectionMode ->
                            if (inSelectionMode) {
                                if (visibleCurrentTabId == LibraryTabId.ALBUMS && isAlbumSelectionMode) {
                                    SelectionActionRow(
                                        selectedCount = selectedAlbums.size,
                                        onSelectAll = {
                                            multiSelectionState.clearSelection()
                                            showMultiSelectionSheet = false
                                            selectedAlbums = appendDistinctSelection(
                                                current = selectedAlbums,
                                                candidates = playerViewModel.albumsFlow.value,
                                                keyOf = Album::id
                                            ).toImmutableList()
                                        },
                                        onDeselect = {
                                            selectedAlbums = persistentListOf()
                                            multiSelectionState.clearSelection()
                                            showMultiSelectionSheet = false
                                        },
                                        onOptionsClick = openSelectedAlbumActions
                                    )
                                } else if (visibleCurrentTabId == LibraryTabId.ARTISTS && isArtistSelectionMode) {
                                    SelectionActionRow(
                                        selectedCount = selectedArtists.size,
                                        onSelectAll = {
                                            multiSelectionState.clearSelection()
                                            showMultiSelectionSheet = false
                                            selectedArtists = appendDistinctSelection(
                                                current = selectedArtists,
                                                candidates = playerViewModel.artistsFlow.value,
                                                keyOf = Artist::id
                                            ).toImmutableList()
                                        },
                                        onDeselect = {
                                            selectedArtists = persistentListOf()
                                            multiSelectionState.clearSelection()
                                            showMultiSelectionSheet = false
                                        },
                                        onOptionsClick = openSelectedArtistActions
                                    )
                                } else {
                                    SelectionActionRow(
                                        selectedCount = selectedSongs.size,
                                        onSelectAll = {
                                            when (visibleCurrentTabId) {
                                                LibraryTabId.SONGS -> {
                                                    scope.launch {
                                                        val songsToSelect =
                                                            playerViewModel.getSongsForCurrentLibrarySelection()
                                                        multiSelectionState.selectAll(songsToSelect)
                                                    }
                                                }
                                                else -> Unit
                                            }
                                        },
                                        onDeselect = { multiSelectionState.clearSelection() },
                                        onOptionsClick = { showMultiSelectionSheet = true }
                                    )
                                }
                            } else {
                                LibraryActionRow(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(end = 4.dp),
                                    onMainActionClick = {
                                        when (visibleCurrentTabId) {
                                            // Sonora: creating playlists comes with step 5.
                                            LibraryTabId.PLAYLISTS -> Unit
                                            LibraryTabId.ALBUMS -> playerViewModel.playRandomAlbum()
                                            LibraryTabId.ARTISTS -> playerViewModel.shuffleRandomArtist()
                                            else -> playerViewModel.shuffleAllSongs()
                                        }
                                    },
                                    iconRotation = iconRotation,
                                    showSortButton = sanitizedSortOptions.isNotEmpty(),
                                    showLocateButton = showLocateButton,
                                    onSortClick = { playerViewModel.showSortingSheet() },
                                    onLocateClick = { locateAction?.invoke() },
                                    isPlaylistTab = visibleCurrentTabId == LibraryTabId.PLAYLISTS,
                                    // Sonora: no M3U import next to "new" on the playlists tab.
                                    showImportButton = false,
                                    isShuffleEnabled = isShuffleEnabled,
                                    showStorageFilterButton = visibleCurrentTabId.hasStorageFilterButton(),
                                    currentStorageFilter = playerUiState.currentStorageFilter,
                                    onStorageFilterClick = { playerViewModel.toggleStorageFilter() }
                                )
                            }
                        }

                        LibraryInlineSyncIndicator(
                            visible = inlineSyncVisible && !isLibraryContentEmpty
                        )

                        if (isSortSheetVisible && sanitizedSortOptions.isNotEmpty()) {
                            val currentSelectionKey = currentSelectedSortOption?.storageKey
                            val selectedOptionForSheet = sanitizedSortOptions.firstOrNull { option ->
                                option.storageKey == currentSelectionKey
                            }
                                ?: sanitizedSortOptions.firstOrNull { option ->
                                    option.storageKey == currentTabId.defaultSort.storageKey
                                }
                                ?: sanitizedSortOptions.first()


                            val isAlbumTab = currentTabId == LibraryTabId.ALBUMS

                            LibrarySortBottomSheet(
                                title = stringResource(R.string.presentation_batch_d_sort_by),
                                options = sanitizedSortOptions,
                                selectedOption = selectedOptionForSheet,
                                onDismiss = { playerViewModel.hideSortingSheet() },
                                onOptionSelected = { option ->
                                    onSortOptionChanged(option)
                                    playerViewModel.hideSortingSheet()
                                },
                                onDirectionToggle = { option ->
                                    onSortOptionChanged(option)
                                },
                                viewSectionTitle = stringResource(R.string.presentation_batch_d_view_section_view),
                                viewToggleContent = if (isAlbumTab) {
                                    {
                                        Row(
                                            modifier = Modifier.fillMaxWidth().height(48.dp),
                                            horizontalArrangement = Arrangement.spacedBy(8.dp)
                                        ) {
                                            val isList = playerUiState.isAlbumsListView

                                            ToggleSegmentButton(
                                                modifier = Modifier.weight(1f),
                                                active = !isList,
                                                activeColor = MaterialTheme.colorScheme.primary,
                                                inactiveColor = MaterialTheme.colorScheme.surfaceVariant,
                                                activeContentColor = MaterialTheme.colorScheme.onPrimary,
                                                inactiveContentColor = MaterialTheme.colorScheme.onSurfaceVariant,
                                                activeCornerRadius = 32.dp,
                                                onClick = { playerViewModel.setAlbumsListView(false) },
                                                text = stringResource(R.string.presentation_batch_d_view_grid),
                                                imageVector = Icons.Rounded.ViewModule
                                            )

                                            ToggleSegmentButton(
                                                modifier = Modifier.weight(1f),
                                                active = isList,
                                                activeColor = MaterialTheme.colorScheme.primary,
                                                inactiveColor = MaterialTheme.colorScheme.surfaceVariant,
                                                activeContentColor = MaterialTheme.colorScheme.onPrimary,
                                                inactiveContentColor = MaterialTheme.colorScheme.onSurfaceVariant,
                                                activeCornerRadius = 32.dp,
                                                onClick = { playerViewModel.setAlbumsListView(true) },
                                                text = stringResource(R.string.presentation_batch_d_view_list),
                                                imageVector = Icons.AutoMirrored.Rounded.ViewList
                                            )
                                        }
                                    }
                                } else null
                                // Sonora: no "cloud only" switch under the options, every song is
                                // on the server (Eric, 2026-11-11); the playlists' layout switch
                                // comes with them in step 5.
                            )
                        }

                        Box(modifier = Modifier.fillMaxSize()) {
                            HorizontalPager(
                                state = pagerState,
                                modifier = Modifier
                                    .fillMaxSize()
                                    .padding(top = 8.dp),
                                pageSpacing = 0.dp,
                                beyondViewportPageCount = 1,
                                key = { it }
                            ) { page ->
                                val tabIndex = resolveTabIndex(
                                    page = page,
                                    tabCount = tabTitles.size
                                )
                                when (tabTitles.getOrNull(tabIndex)?.toLibraryTabIdOrNull()) {
                                    LibraryTabId.SONGS -> {
                                        // Each tab collects its own paging flow, so pages the pager has disposed stop
                                        // collecting; the flows are cachedIn the view model, so a page that comes back
                                        // starts from the cached items.
                                        val allSongsLazyPagingItems = libraryViewModel.songsPagingFlow.collectAsLazyPagingItems()
                                        LibrarySongsTab(
                                            songs = allSongsLazyPagingItems,
                                            isLoading = isLibraryLoading,
                                            playerViewModel = playerViewModel,
                                            bottomBarHeight = bottomBarHeightDp,
                                            onMoreOptionsClick = stableOnMoreOptionsClick,
                                            isRefreshing = isRefreshing,
                                            onRefresh = {
                                                onRefresh()
                                                allSongsLazyPagingItems.refresh()
                                            },
                                            isSelectionMode = isSelectionMode,
                                            selectedSongIds = selectedSongIds,
                                            onSongLongPress = onSongLongPress,
                                            onSongSelectionToggle = onSongSelectionToggle,
                                            getSelectionIndex = playerViewModel.multiSelectionStateHolder::getSelectionIndex,
                                            currentSelectedSongIds = currentSelectedSongIds,
                                            onSongDragSelectionChange = onSongDragSelectionChange,
                                            onLocateCurrentSongVisibilityChanged = { songsShowLocateButton = it },
                                            onRegisterLocateCurrentSongAction = { songsLocateAction = it },
                                            sortOption = playerUiState.currentSongSortOption,
                                            storageFilter = playerUiState.currentStorageFilter,
                                            hasCurrentSong = hasCurrentSong
                                        )
                                    }
                                    LibraryTabId.ALBUMS -> {
                                        val albumsLazyPagingItems = libraryViewModel.albumsPagingFlow.collectAsLazyPagingItems()
                                        val isLoading = playerUiState.isLoadingLibraryCategories

                                        val stableOnAlbumClick: (Long) -> Unit = remember(navController) {
                                            { albumId: Long ->
                                                navController.navigateSafelyReplacing(
                                                    route = Screen.AlbumDetail.createRoute(albumId),
                                                    patternToPop = Screen.AlbumDetail.route
                                                )
                                            }
                                        }
                                        LibraryAlbumsTab(
                                            albums = albumsLazyPagingItems,
                                            isLoading = isLoading,
                                            playerViewModel = playerViewModel,
                                            bottomBarHeight = bottomBarHeightDp,
                                            isListView = playerUiState.isAlbumsListView,
                                            currentAlbumSortOption = playerUiState.currentAlbumSortOption,
                                            onAlbumClick = stableOnAlbumClick,
                                            isRefreshing = isRefreshing,
                                            onRefresh = onRefresh,
                                            isSelectionMode = isAlbumSelectionMode,
                                            selectedAlbumIds = selectedAlbumIds,
                                            onAlbumLongPress = onAlbumLongPress,
                                            onAlbumSelectionToggle = onAlbumSelectionToggle,
                                            getSelectionIndex = getAlbumSelectionIndex,
                                            storageFilter = playerUiState.currentStorageFilter
                                        )
                                    }

                                    LibraryTabId.ARTISTS -> {
                                        val artistsLazyPagingItems = libraryViewModel.artistsPagingFlow.collectAsLazyPagingItems()
                                        val isLoading = playerUiState.isLoadingLibraryCategories

                                        LibraryArtistsTab(
                                            artists = artistsLazyPagingItems,
                                            isLoading = isLoading,
                                            playerViewModel = playerViewModel,
                                            bottomBarHeight = bottomBarHeightDp,
                                            currentArtistSortOption = playerUiState.currentArtistSortOption,
                                            onArtistClick = { artistId ->
                                                navController.navigateSafelyReplacing(
                                                    route = Screen.ArtistDetail.createRoute(artistId),
                                                    patternToPop = Screen.ArtistDetail.route
                                                )
                                            },
                                            isRefreshing = isRefreshing,
                                            onRefresh = onRefresh,
                                            isSelectionMode = isArtistSelectionMode,
                                            selectedArtistIds = selectedArtistIds,
                                            onArtistLongPress = onArtistLongPress,
                                            onArtistSelectionToggle = onArtistSelectionToggle,
                                            getSelectionIndex = getArtistSelectionIndex,
                                            storageFilter = playerUiState.currentStorageFilter
                                        )
                                    }

                                    // Sonora: the original's LibraryPlaylistsTab comes with the
                                    // playlists in step 5; until then the page is empty.
                                    LibraryTabId.PLAYLISTS -> Box(Modifier.fillMaxSize())

                                    null -> Unit
                                }
                            }

                            val selectionCount = when (visibleCurrentTabId) {
                                LibraryTabId.ALBUMS -> selectedAlbums.size.takeIf { isAlbumSelectionMode } ?: 0
                                LibraryTabId.ARTISTS -> selectedArtists.size.takeIf { isArtistSelectionMode } ?: 0
                                LibraryTabId.SONGS -> selectedSongs.size
                                LibraryTabId.PLAYLISTS -> 0
                            }
                            SelectionCountPill(
                                selectedCount = selectionCount,
                                modifier = Modifier
                                    .align(Alignment.TopCenter)
                                    .zIndex(1f)
                            )
                        }
                    }
                }
                if (
                    isLibraryContentEmpty &&
                    (
                            playerUiState.isSyncingLibrary ||
                                    playerUiState.isLoadingInitialSongs ||
                                    playerUiState.isLoadingLibraryCategories
                            )
                ) {
                    LibrarySyncOverlay()
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
    }

    if (showSongInfoBottomSheet && selectedSongForInfo != null) {
        val currentSong = selectedSongForInfo
        val isFavorite = remember(currentSong?.id, favoriteIds) { derivedStateOf { currentSong?.let {
            favoriteIds.contains(
                it.id)
        } } }.value ?: false

        if (currentSong != null) {
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
                    playerViewModel.sendToast(context.getString(R.string.toast_added_to_queue))
                },
                onAddNextToQueue = {
                    playerViewModel.addSongNextToQueue(currentSong)
                    showSongInfoBottomSheet = false
                    playerViewModel.sendToast(context.getString(R.string.toast_playing_next))
                },
                onAddToPlayList = {
                    // Sonora: the playlist sheet comes with the playlists in step 5.
                }
            )
        }
    }

    val clearResolvedCategorySelection: () -> Unit = {
        selectedAlbums = persistentListOf()
        selectedArtists = persistentListOf()
    }

    if (showMultiSelectionSheet && selectedSongs.isNotEmpty()) {
        MultiSelectionBottomSheet(
            selectedSongs = selectedSongs,
            favoriteSongIds = favoriteIds,
            onDismiss = { showMultiSelectionSheet = false },
            onPlayAll = {
                playerViewModel.playSelectedSongs(selectedSongs)
                clearResolvedCategorySelection()
                showMultiSelectionSheet = false
            },
            onAddToQueue = {
                playerViewModel.addSelectedToQueue(selectedSongs)
                clearResolvedCategorySelection()
                showMultiSelectionSheet = false
            },
            onPlayNext = {
                playerViewModel.addSelectedAsNext(selectedSongs)
                clearResolvedCategorySelection()
                showMultiSelectionSheet = false
            },
            onAddToPlaylist = {
                // Sonora: the playlist sheet comes with the playlists in step 5.
            },
            onToggleLikeAll = { shouldLike ->
                if (shouldLike) {
                    playerViewModel.likeSelectedSongs(selectedSongs)
                } else {
                    playerViewModel.unlikeSelectedSongs(selectedSongs)
                }
                clearResolvedCategorySelection()
                showMultiSelectionSheet = false
            }
        )
    }
}

/**
 * Slim, non-intrusive indicator for sync work that should not keep the list pulled
 * down: automatic startup syncs, background maintenance, and manual refreshes after
 * the short pull-to-refresh confirmation window. It sits just below
 * [LibraryActionRow] and collapses to zero height when not active.
 *
 * Distinct from [LibrarySyncOverlay], which is reserved for initial empty-library
 * loads. The parent screen also gates this indicator off while the pull spinner is
 * visible, so the two feedback channels do not compete.
 */
@Composable
private fun LibraryInlineSyncIndicator(
    visible: Boolean
) {
    AnimatedVisibility(
        visible = visible,
        enter = androidx.compose.animation.expandVertically(
            expandFrom = Alignment.Top,
            animationSpec = tween(durationMillis = 220, easing = FastOutSlowInEasing)
        ) + androidx.compose.animation.fadeIn(animationSpec = tween(180)),
        exit = androidx.compose.animation.shrinkVertically(
            shrinkTowards = Alignment.Top,
            animationSpec = tween(durationMillis = 200, easing = FastOutSlowInEasing)
        ) + androidx.compose.animation.fadeOut(animationSpec = tween(160))
    ) {
        // Sonora: the original names the phase of its media scan (SyncProgress); Sonora's sync
        // has one phase, which the original labels "sync in progress".
        val phaseLabel = stringResource(R.string.sync_in_progress)

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp, vertical = 6.dp)
        ) {
            Text(
                text = phaseLabel,
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis
            )
            Spacer(modifier = Modifier.height(4.dp))
            LinearWavyProgressIndicator(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(4.dp),
                color = MaterialTheme.colorScheme.primary,
                trackColor = MaterialTheme.colorScheme.surfaceContainerHighest
            )
        }
    }
}

/**
 * Isolated sync/loading overlay for the first load of an empty library. Sonora: without the
 * original's progress bar (SyncProgressBar), which counts scanned files; library/changes pages
 * have no total.
 */
@Composable
private fun LibrarySyncOverlay() {
    Surface(
        modifier = Modifier.fillMaxSize(),
        color = MaterialTheme.colorScheme.scrim.copy(alpha = 0.5f)
    ) {
        Box(contentAlignment = Alignment.Center) {
            Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                modifier = Modifier.padding(32.dp)
            ) {
                LoadingIndicator(modifier = Modifier.size(64.dp))
                Spacer(modifier = Modifier.height(16.dp))
                Text(
                    text = stringResource(R.string.syncing_library),
                    style = MaterialTheme.typography.titleMedium,
                    color = MaterialTheme.colorScheme.onSurface
                )
            }
        }
    }
}

private fun positiveMod(value: Int, mod: Int): Int {
    if (mod <= 0) return 0
    return ((value % mod) + mod) % mod
}

// Sonora: without the compact navigation's endless pager (infinitePagerInitialPage), so pages
// are tab indices.
private fun resolveTabIndex(page: Int, tabCount: Int): Int {
    if (tabCount <= 0) return 0
    return page.coerceIn(0, tabCount - 1)
}

/** The tab stored at [index], or [fallback] when the index or the stored identifier is unknown. */
internal fun resolveVisibleLibraryTabId(
    tabs: List<String>,
    index: Int,
    fallback: LibraryTabId
): LibraryTabId = tabs.getOrNull(index)?.toLibraryTabIdOrNull() ?: fallback

private fun targetPageForTabIndex(
    targetTabIndex: Int,
    tabCount: Int
): Int {
    if (tabCount <= 0) return 0
    return positiveMod(targetTabIndex, tabCount)
}

@Composable
private fun LibraryTabId.displayTitle(): String = stringResource(titleRes)


@androidx.annotation.OptIn(UnstableApi::class)
@Composable
fun AlbumGridItemRedesigned(
    album: Album,
    albumColorSchemePairFlow: StateFlow<ColorSchemePair?>,
    onClick: () -> Unit,
    isLoading: Boolean = false,
    isSelectionMode: Boolean = false,
    isSelected: Boolean = false,
    selectionIndex: Int? = null,
    onLongPress: () -> Unit = {},
    onSelectionToggle: () -> Unit = {},
    customShape: androidx.compose.ui.graphics.Shape? = null
) {
    val albumColorSchemePair by albumColorSchemePairFlow.collectAsStateWithLifecycle()
    val systemIsDark = LocalPixelPlayerDarkTheme.current

    val currentMaterialColorScheme = MaterialTheme.colorScheme

    val itemDesignColorScheme = remember(albumColorSchemePair, systemIsDark, currentMaterialColorScheme) {
        albumColorSchemePair?.let { pair ->
            if (systemIsDark) pair.dark else pair.light
        } ?: currentMaterialColorScheme
    }

    val gradientBaseColor = itemDesignColorScheme.primaryContainer
    val onGradientColor = itemDesignColorScheme.onPrimaryContainer
    val cardShape = remember(customShape) { customShape ?: RoundedCornerShape(20.dp) }
    val selectionScale by animateFloatAsState(
        targetValue = if (isSelected) 0.985f else 1f,
        animationSpec = tween(durationMillis = 220),
        label = "albumGridSelectionScale"
    )
    val selectionBorderWidth by animateDpAsState(
        targetValue = if (isSelected) 2.dp else 0.dp,
        animationSpec = tween(durationMillis = 220),
        label = "albumGridSelectionBorder"
    )

    if (isLoading) {
        Card(
            modifier = Modifier.fillMaxWidth(),
            shape = cardShape,
            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.3f))
        ) {
            Column(
                modifier = Modifier.background(
                    color = MaterialTheme.colorScheme.primaryContainer,
                    shape = cardShape
                )
            ) {
                ShimmerBox(
                    modifier = Modifier
                        .aspectRatio(3f / 2f)
                        .fillMaxSize()
                )
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(84.dp)
                        .padding(12.dp),
                    verticalArrangement = Arrangement.Center
                ) {
                    ShimmerBox(
                        modifier = Modifier
                            .fillMaxWidth(0.8f)
                            .height(20.dp)
                            .clip(RoundedCornerShape(4.dp))
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    ShimmerBox(
                        modifier = Modifier
                            .fillMaxWidth(0.6f)
                            .height(16.dp)
                            .clip(RoundedCornerShape(4.dp))
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    ShimmerBox(
                        modifier = Modifier
                            .fillMaxWidth(0.4f)
                            .height(16.dp)
                            .clip(RoundedCornerShape(4.dp))
                    )
                }
            }
        }
    } else {
        Card(
            modifier = Modifier
                .fillMaxWidth()
                .seamlessSelectionScale(selectionScale)
                .then(
                    if (isSelected) {
                        Modifier.border(
                            width = selectionBorderWidth,
                            color = MaterialTheme.colorScheme.primary,
                            shape = cardShape
                        )
                    } else {
                        Modifier
                    }
                )
                .clipCardCornersFallback(cardShape)
                .combinedClickable(
                    onClick = {
                        if (isSelectionMode) {
                            onSelectionToggle()
                        } else {
                            onClick()
                        }
                    },
                    onLongClick = onLongPress
                ),
            shape = cardShape,
            colors = CardDefaults.cardColors(containerColor = itemDesignColorScheme.surfaceVariant.copy(alpha = 0.3f))
        ) {
            Box {
                Column(
                    modifier = Modifier.background(
                        color = gradientBaseColor,
                        shape = cardShape
                    )
                ) {
                    Box(contentAlignment = Alignment.BottomStart) {
                        var isLoadingImage by remember { mutableStateOf(true) }
                        SmartImage(
                            model = album.albumArtUriString,
                            contentDescription = stringResource(R.string.cd_album_art_for_title, album.title),
                            contentScale = ContentScale.Crop,
                            targetSize = Size(256, 256),
                            modifier = Modifier
                                .aspectRatio(3f / 2f)
                                .fillMaxSize(),
                            onState = { state ->
                                isLoadingImage = state is AsyncImagePainter.State.Loading
                            }
                        )
                        if (isLoadingImage) {
                            ShimmerBox(
                                modifier = Modifier
                                    .aspectRatio(3f / 2f)
                                    .fillMaxSize()
                            )
                        }
                        Box(
                            modifier = Modifier
                                .fillMaxSize()
                                .aspectRatio(3f / 2f)
                                .background(
                                    remember(gradientBaseColor) {
                                        Brush.verticalGradient(
                                            colors = listOf(
                                                Color.Transparent, gradientBaseColor
                                            )
                                        )
                                    })
                        )
                    }
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(84.dp)
                            .padding(12.dp),
                        verticalArrangement = Arrangement.Center
                    ) {
                        Text(
                            album.title,
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.Bold,
                            color = onGradientColor,
                            maxLines = 1,
                            overflow = TextOverflow.Ellipsis
                        )
                        Text(album.artist, style = MaterialTheme.typography.bodySmall, color = onGradientColor.copy(alpha = 0.85f), maxLines = 1, overflow = TextOverflow.Ellipsis)
                        Text(formatSongCount(album.songCount), style = MaterialTheme.typography.bodySmall, color = onGradientColor.copy(alpha = 0.7f), maxLines = 1, overflow = TextOverflow.Ellipsis)
                    }
                }

                if (isSelectionMode && isSelected) {
                    Box(
                        modifier = Modifier
                            .align(Alignment.TopEnd)
                            .padding(10.dp)
                            .size(28.dp)
                            .background(
                                color = MaterialTheme.colorScheme.primary,
                                shape = CircleShape
                            ),
                        contentAlignment = Alignment.Center
                    ) {
                        Text(
                            text = selectionIndex?.toString() ?: "✓",
                            color = MaterialTheme.colorScheme.onPrimary,
                            style = MaterialTheme.typography.labelMedium,
                            fontWeight = FontWeight.Bold
                        )
                    }
                }
            }
        }
    }
}

@androidx.annotation.OptIn(UnstableApi::class)
@Composable
fun ArtistListItem(
    artist: Artist,
    onClick: () -> Unit,
    isLoading: Boolean = false,
    isSelectionMode: Boolean = false,
    isSelected: Boolean = false,
    selectionIndex: Int? = null,
    onLongPress: () -> Unit = {},
    onSelectionToggle: () -> Unit = {},
    customShape: androidx.compose.ui.graphics.Shape? = null
) {
    val cardShape = remember(customShape) { customShape ?: RoundedCornerShape(18.dp) }
    val selectionScale by animateFloatAsState(
        targetValue = if (isSelected) 0.99f else 1f,
        animationSpec = tween(durationMillis = 200),
        label = "artistSelectionScale"
    )
    val selectionBorderWidth by animateDpAsState(
        targetValue = if (isSelected) 2.dp else 0.dp,
        animationSpec = tween(durationMillis = 200),
        label = "artistSelectionBorder"
    )

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .seamlessSelectionScale(selectionScale)
            .then(
                if (isSelected) {
                    Modifier.border(
                        width = selectionBorderWidth,
                        color = MaterialTheme.colorScheme.primary,
                        shape = cardShape
                    )
                } else {
                    Modifier
                }
            )
            .clipCardCornersFallback(cardShape)
            .combinedClickable(
                enabled = !isLoading,
                onClick = {
                    if (isSelectionMode) onSelectionToggle() else onClick()
                },
                onLongClick = onLongPress
            ),
        shape = cardShape,
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceContainerLow
        )
    ) {
        Box(modifier = Modifier.fillMaxWidth()) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(
                        start = 12.dp,
                        top = 12.dp,
                        end = if (isSelected) 54.dp else 12.dp,
                        bottom = 12.dp
                    ),
                verticalAlignment = Alignment.CenterVertically
            ) {
                if (isLoading) {
                    ShimmerBox(
                        modifier = Modifier
                            .size(48.dp)
                            .clip(CircleShape)
                    )
                    Spacer(modifier = Modifier.width(16.dp))
                    Column(modifier = Modifier.weight(1f)) {
                        ShimmerBox(
                            modifier = Modifier
                                .fillMaxWidth(0.6f)
                                .height(20.dp)
                                .clip(RoundedCornerShape(4.dp))
                        )
                        Spacer(modifier = Modifier.height(4.dp))
                        ShimmerBox(
                            modifier = Modifier
                                .fillMaxWidth(0.3f)
                                .height(16.dp)
                                .clip(RoundedCornerShape(4.dp))
                        )
                    }
                } else {
                    Box(
                        modifier = Modifier
                            .size(48.dp)
                            .clip(ShapeCache.expressiveAvatar)
                            .background(MaterialTheme.colorScheme.primaryContainer),
                        contentAlignment = Alignment.Center
                    ) {
                        if (!artist.effectiveImageUrl.isNullOrEmpty()) {
                            AsyncImage(
                                model = ImageRequest.Builder(LocalContext.current)
                                    .data(artist.effectiveImageUrl)
                                    .crossfade(true)
                                    .build(),
                                contentDescription = artist.name,
                                contentScale = ContentScale.Crop,
                                modifier = Modifier.fillMaxSize()
                            )
                        } else {
                            Icon(
                                painter = painterResource(R.drawable.rounded_artist_24),
                                contentDescription = stringResource(R.string.presentation_batch_d_cd_artist),
                                modifier = Modifier.padding(8.dp),
                                tint = MaterialTheme.colorScheme.onPrimaryContainer
                            )
                        }
                    }
                    Spacer(modifier = Modifier.width(16.dp))
                    Column(modifier = Modifier.weight(1f)) {
                        Text(artist.name, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                        Text(formatSongCount(artist.songCount), style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    }
                }
            }

            if (isSelectionMode && isSelected) {
                Box(
                    modifier = Modifier
                        .align(Alignment.CenterEnd)
                        .padding(end = 14.dp)
                        .size(28.dp)
                        .background(MaterialTheme.colorScheme.primary, CircleShape),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = selectionIndex?.toString() ?: "✓",
                        color = MaterialTheme.colorScheme.onPrimary,
                        style = MaterialTheme.typography.labelMedium,
                        fontWeight = FontWeight.Bold
                    )
                }
            }
        }
    }
}

@androidx.annotation.OptIn(UnstableApi::class)
@Composable
fun AlbumListItem(
    album: Album,
    albumColorSchemePairFlow: StateFlow<ColorSchemePair?>,
    onClick: () -> Unit,
    isLoading: Boolean = false,
    isSelectionMode: Boolean = false,
    isSelected: Boolean = false,
    selectionIndex: Int? = null,
    onLongPress: () -> Unit = {},
    onSelectionToggle: () -> Unit = {},
    customShape: androidx.compose.ui.graphics.Shape? = null
) {
    val albumColorSchemePair by albumColorSchemePairFlow.collectAsStateWithLifecycle()
    val systemIsDark = LocalPixelPlayerDarkTheme.current
    val currentMaterialColorScheme = MaterialTheme.colorScheme

    val itemDesignColorScheme = remember(albumColorSchemePair, systemIsDark, currentMaterialColorScheme) {
        albumColorSchemePair?.let { pair ->
            if (systemIsDark) pair.dark else pair.light
        } ?: currentMaterialColorScheme
    }

    val gradientBaseColor = itemDesignColorScheme.primaryContainer
    val onGradientColor = itemDesignColorScheme.onPrimaryContainer
    val cardShape = remember(customShape) { customShape ?: RoundedCornerShape(16.dp) }
    val selectionScale by animateFloatAsState(
        targetValue = if (isSelected) 0.99f else 1f,
        animationSpec = tween(durationMillis = 200),
        label = "albumListSelectionScale"
    )
    val selectionBorderWidth by animateDpAsState(
        targetValue = if (isSelected) 2.dp else 0.dp,
        animationSpec = tween(durationMillis = 200),
        label = "albumListSelectionBorder"
    )

    if (isLoading) {
        Card(
            modifier = Modifier
                .fillMaxWidth()
                .height(80.dp)
                .clipCardCornersFallback(cardShape),
            shape = cardShape,
            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.3f))
        ) {
            Row(modifier = Modifier.fillMaxSize()) {
                ShimmerBox(
                    modifier = Modifier
                        .aspectRatio(1f)
                        .fillMaxHeight()
                )
                Column(
                    modifier = Modifier
                        .weight(1f)
                        .padding(12.dp),
                    verticalArrangement = Arrangement.Center
                ) {
                    ShimmerBox(
                        modifier = Modifier
                            .fillMaxWidth(0.6f)
                            .height(16.dp)
                            .clip(RoundedCornerShape(4.dp))
                    )
                    Spacer(modifier = Modifier.height(8.dp))
                    ShimmerBox(
                        modifier = Modifier
                            .fillMaxWidth(0.4f)
                            .height(14.dp)
                            .clip(RoundedCornerShape(4.dp))
                    )
                }
            }
        }
    } else {
        Card(
            modifier = Modifier
                .fillMaxWidth()
                .height(88.dp)
                .seamlessSelectionScale(selectionScale)
                .then(
                    if (isSelected) {
                        Modifier.border(
                            width = selectionBorderWidth,
                            color = MaterialTheme.colorScheme.primary,
                            shape = cardShape
                        )
                    } else {
                        Modifier
                    }
                )
                .clipCardCornersFallback(cardShape)
                .combinedClickable(
                    onClick = {
                        if (isSelectionMode) {
                            onSelectionToggle()
                        } else {
                            onClick()
                        }
                    },
                    onLongClick = onLongPress
                ),
            shape = cardShape,
            colors = CardDefaults.cardColors(containerColor = itemDesignColorScheme.surfaceVariant.copy(alpha = 0.3f))
        ) {
            Box(modifier = Modifier.fillMaxSize()) {
                Row(
                    modifier = Modifier.fillMaxSize()
                ) {
                    Box(
                        modifier = Modifier
                            .aspectRatio(1f)
                            .fillMaxHeight()
                    ) {
                        var isLoadingImage by remember { mutableStateOf(true) }
                        SmartImage(
                            model = album.albumArtUriString,
                            contentDescription = stringResource(R.string.cd_album_art_for_title, album.title),
                            contentScale = ContentScale.Crop,
                            targetSize = Size(256, 256),
                            modifier = Modifier.fillMaxSize(),
                            onState = { state ->
                                isLoadingImage = state is AsyncImagePainter.State.Loading
                            }
                        )
                        if (isLoadingImage) {
                            ShimmerBox(modifier = Modifier.fillMaxSize())
                        }

                        Box(
                            modifier = Modifier
                                .fillMaxSize()
                                .background(
                                    Brush.horizontalGradient(
                                        colors = listOf(
                                            Color.Transparent,
                                            gradientBaseColor
                                        )
                                    )
                                )
                        )
                    }

                    Box(
                        modifier = Modifier
                            .weight(1f)
                            .fillMaxHeight()
                            .background(gradientBaseColor)
                    ) {
                        Column(
                            modifier = Modifier
                                .fillMaxSize()
                                .padding(horizontal = 12.dp, vertical = 10.dp),
                            verticalArrangement = Arrangement.Center
                        ) {
                            val variableTextStyle = remember(album.id, album.title) {
                                GenreTypography.getGenreStyle(album.id.toString(), album.title)
                            }

                            Text(
                                album.title,
                                style = variableTextStyle.copy(fontSize = 22.sp),
                                color = onGradientColor,
                                maxLines = 1,
                                overflow = TextOverflow.Ellipsis
                            )
                            Spacer(
                                modifier = Modifier.height(4.dp)
                            )
                            Text(
                                album.artist,
                                style = MaterialTheme.typography.bodySmall,
                                color = onGradientColor.copy(alpha = 0.85f),
                                maxLines = 1,
                                overflow = TextOverflow.Ellipsis
                            )
                            Text(
                                formatSongCount(album.songCount),
                                style = MaterialTheme.typography.bodySmall,
                                color = onGradientColor.copy(alpha = 0.7f),
                                maxLines = 1,
                                overflow = TextOverflow.Ellipsis
                            )
                        }
                    }
                }

                if (isSelectionMode && isSelected) {
                    Box(
                        modifier = Modifier
                            .align(Alignment.TopEnd)
                            .padding(8.dp)
                            .size(24.dp)
                            .background(
                                color = MaterialTheme.colorScheme.primary,
                                shape = CircleShape
                            ),
                        contentAlignment = Alignment.Center
                    ) {
                        Text(
                            text = selectionIndex?.toString() ?: "✓",
                            style = MaterialTheme.typography.labelSmall,
                            color = MaterialTheme.colorScheme.onPrimary,
                            fontWeight = FontWeight.Bold
                        )
                    }
                }
            }
        }
    }
}
