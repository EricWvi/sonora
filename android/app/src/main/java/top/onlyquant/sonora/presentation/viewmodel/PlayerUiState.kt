package top.onlyquant.sonora.presentation.viewmodel

import androidx.compose.runtime.Immutable
import kotlinx.collections.immutable.ImmutableList
import kotlinx.collections.immutable.persistentListOf
import top.onlyquant.sonora.data.model.Song
import top.onlyquant.sonora.data.model.SortOption
import top.onlyquant.sonora.data.model.StorageFilter

/**
 * Reduced from PlayerUiState in PixelPlayerOSS v0.4.0-alpha.53 (GPL-3.0): the player sheet, the
 * queue and their undo bars, and the songs, albums and artists tabs' sort and view. The folder,
 * liked, playlist and search fields come with those screens; there is no storage filter state
 * (every song is on the server; the filter button is shown but does nothing, Eric 2026-11-11).
 */
@Immutable
data class PlayerUiState(
    val currentPlaybackQueue: ImmutableList<Song> = persistentListOf(),
    val currentQueueSourceName: String = "All Songs",
    val isLoadingInitialSongs: Boolean = true,
    val showDismissUndoBar: Boolean = false,
    val dismissedSong: Song? = null,
    val dismissedQueue: ImmutableList<Song> = persistentListOf(),
    val dismissedQueueName: String = "",
    val dismissedPosition: Long = 0L,
    val undoBarVisibleDuration: Long = 4000L,
    val preparingSongId: String? = null,
    val showQueueItemUndoBar: Boolean = false,
    val lastRemovedQueueSong: Song? = null,
    val lastRemovedQueueIndex: Int = -1,
    val currentStorageFilter: StorageFilter = StorageFilter.ALL,
    val currentSongSortOption: SortOption = SortOption.SongTitleAZ,
    val currentAlbumSortOption: SortOption = SortOption.AlbumTitleAZ,
    val currentArtistSortOption: SortOption = SortOption.ArtistNameAZ,
    val isAlbumsListView: Boolean = false,
    val isLoadingLibraryCategories: Boolean = true,
    val isSyncingLibrary: Boolean = false
)
