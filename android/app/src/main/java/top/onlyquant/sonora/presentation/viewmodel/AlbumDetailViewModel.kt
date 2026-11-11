package top.onlyquant.sonora.presentation.viewmodel

import android.content.Context
import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import top.onlyquant.sonora.data.model.Album
import top.onlyquant.sonora.data.model.Song
import top.onlyquant.sonora.data.repository.LibraryRepository
import top.onlyquant.sonora.R
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.catch
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class AlbumDetailUiState(
    val album: Album? = null,
    val songs: List<Song> = emptyList(),
    val isLoading: Boolean = false,
    val error: String? = null
)

/**
 * Sonora: the album and its songs come from the library on the device ([LibraryRepository]),
 * which sync keeps up to date. The original also tracks the album's offline downloads
 * (CloudOfflineRepository); caching comes with step 7, so nothing is downloaded yet.
 */
@HiltViewModel
class AlbumDetailViewModel @Inject constructor(
    @ApplicationContext private val context: Context,
    private val libraryRepository: LibraryRepository,
    savedStateHandle: SavedStateHandle
) : ViewModel() {

    private val _uiState = MutableStateFlow(AlbumDetailUiState())
    val uiState: StateFlow<AlbumDetailUiState> = _uiState.asStateFlow()
    // Sonora: the original's completed offline downloads; none until step 7.
    val completedOfflineUris: StateFlow<Set<String>> = MutableStateFlow<Set<String>>(emptySet()).asStateFlow()

    private var loadedAlbumId: Long? = null

    init {
        val albumIdString: String? = savedStateHandle.get("albumId")
        if (albumIdString != null) {
            val albumId = albumIdString.toLongOrNull()
            if (albumId != null) {
                loadedAlbumId = albumId
                loadAlbumData(albumId)
            } else {
                _uiState.update { it.copy(error = context.getString(R.string.invalid_album_id), isLoading = false) }
            }
        } else {
            _uiState.update { it.copy(error = context.getString(R.string.album_id_not_found), isLoading = false) }
        }
    }

    private fun loadAlbumData(id: Long) {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            try {
                val albumDetailsFlow = libraryRepository.album(id)
                val albumSongsFlow = libraryRepository.songsForAlbumFlow(id)

                combine(albumDetailsFlow, albumSongsFlow) { album, songs ->
                    if (album != null) {
                        AlbumDetailUiState(
                            album = album,
                            // Sonora: already in the original's order (disc, track, title) from
                            // MusicDao, with titles by the server's sortKey instead of lowercase.
                            songs = songs,
                            isLoading = false
                        )
                    } else {
                        AlbumDetailUiState(
                            error = context.getString(R.string.album_not_found),
                            isLoading = false
                        )
                    }
                }
                    .catch { e ->
                        emit(
                            AlbumDetailUiState(
                                error = context.getString(R.string.error_loading_album, e.localizedMessage ?: ""),
                                isLoading = false
                            )
                        )
                    }
                    .collect { newState ->
                        _uiState.value = newState
                    }

            } catch (e: Exception) {
                _uiState.update {
                    it.copy(
                        error = context.getString(R.string.error_loading_album, e.localizedMessage ?: ""),
                        isLoading = false
                    )
                }
            }
        }
    }

    /** Re-attempts loading the album after a failure (wired to the error-state retry button). */
    fun retry() {
        loadedAlbumId?.let { loadAlbumData(it) }
    }

    fun update(songs: List<Song>) {
        _uiState.update {
            it.copy(
                isLoading = false,
                songs = songs
            )
        }
    }

    // Sonora: the original enqueues the album's cloud songs for offline download. The button
    // stays and does nothing until caching comes with step 7 (Eric, 2026-11-11).
    @Suppress("UNUSED_PARAMETER")
    fun downloadAlbum(songs: List<Song>) {
    }

    // Sonora: the original removes the album's offline downloads; see downloadAlbum.
    @Suppress("UNUSED_PARAMETER")
    fun removeAlbumDownloads(songs: List<Song>) {
    }
}
