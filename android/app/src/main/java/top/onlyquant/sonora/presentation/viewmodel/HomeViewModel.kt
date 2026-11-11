package top.onlyquant.sonora.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.collections.immutable.ImmutableList
import kotlinx.collections.immutable.persistentListOf
import kotlinx.collections.immutable.toImmutableList
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import timber.log.Timber
import top.onlyquant.sonora.data.api.ApiException
import top.onlyquant.sonora.data.api.SonoraApi
import top.onlyquant.sonora.data.auth.SessionStore
import top.onlyquant.sonora.data.fm.FM_HOME_ALBUMS
import top.onlyquant.sonora.data.model.Song
import top.onlyquant.sonora.data.api.toSong
import top.onlyquant.sonora.presentation.model.RecentlyPlayedSongUiModel

/**
 * What the home screen shows.
 *
 * @property fmAlbums one song per album the radio plays next, the current album first, for the
 *     album art collage (it draws songs' covers); null until the first load
 * @property recentlyPlayed `GET history`, newest first
 */
data class HomeUiState(
    val fmAlbums: ImmutableList<Song>? = null,
    val recentlyPlayed: ImmutableList<RecentlyPlayedSongUiModel> = persistentListOf(),
)

/**
 * Sonora: the home screen's data. The original's HomeScreen reads its daily mix, "your mix" and
 * on-device playback history from PlayerViewModel and StatsViewModel; Sonora's home shows the
 * Roaming FM's next albums (`GET fm?count=5`, Eric 2026-11-11) and the server's history.
 */
@HiltViewModel
class HomeViewModel @Inject constructor(
    private val api: SonoraApi,
    private val sessions: SessionStore,
) : ViewModel() {
    private val _uiState = MutableStateFlow(HomeUiState())
    val uiState: StateFlow<HomeUiState> = _uiState.asStateFlow()

    private var refreshJob: Job? = null

    /** Reloads both sections; called when home is shown and when the radio moves on. */
    fun refresh() {
        refreshJob?.cancel()
        refreshJob = viewModelScope.launch {
            val server = sessions.current()?.server ?: return@launch
            launch {
                try {
                    val state = api.fm(FM_HOME_ALBUMS)
                    val albums = state.albums.mapNotNull { album ->
                        album.tracks.firstOrNull()?.toSong(server)?.copy(albumId = album.id)
                    }
                    _uiState.update { it.copy(fmAlbums = albums.toImmutableList()) }
                } catch (e: ApiException) {
                    Timber.tag(TAG).w(e, "Loading the radio failed")
                    _uiState.update { it.copy(fmAlbums = it.fmAlbums ?: persistentListOf()) }
                }
            }
            launch {
                try {
                    val songs = api.history()
                    _uiState.update { state ->
                        state.copy(recentlyPlayed = songs.map { RecentlyPlayedSongUiModel(it, 0L) }.toImmutableList())
                    }
                } catch (e: ApiException) {
                    Timber.tag(TAG).w(e, "Loading the history failed")
                }
            }
        }
    }

    private companion object {
        const val TAG = "HomeViewModel"
    }
}
