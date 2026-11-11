package top.onlyquant.sonora.presentation.viewmodel

import android.content.Context
import androidx.compose.runtime.Immutable
import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import top.onlyquant.sonora.R
import top.onlyquant.sonora.data.model.Artist
import top.onlyquant.sonora.data.model.Song
import top.onlyquant.sonora.data.preferences.UserPreferencesRepository
import top.onlyquant.sonora.data.repository.LibraryRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.catch
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.launchIn
import kotlinx.coroutines.flow.onEach
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.flow.updateAndGet
import kotlinx.coroutines.launch
import timber.log.Timber
import javax.inject.Inject

/**
 * Holds the full UI state for ArtistDetailScreen.
 *
 * [effectiveImageUrl] is the resolved image to display (custom takes priority over Deezer).
 * It is updated after artist data loads and again whenever the user changes the custom image.
 *
 * [albumsCollapsedByDefault] is the "start with albums collapsed" preference, read once per load
 * and set in the same update as [albumSections], so the first frame never shows them expanded.
 */
data class ArtistDetailUiState(
    val artist: Artist? = null,
    val songs: List<Song> = emptyList(),
    val albumSections: List<ArtistAlbumSection> = emptyList(),
    val albumsCollapsedByDefault: Boolean = false,
    val effectiveImageUrl: String? = null,
    val isLoading: Boolean = false,
    val error: String? = null
)

@Immutable
data class ArtistAlbumSection(
    val albumId: Long,
    val title: String,
    val year: Int?,
    val albumArtUriString: String?,
    val songs: List<Song>,
    /** The genre most of the album's tracks share, or null when none is tagged. */
    val genre: String? = null
)

/**
 * Sonora: the artist and their songs come from the library on the device ([LibraryRepository]).
 * The artist's image is the server's; the original looks one up on Deezer
 * (ArtistImageRepository) and lets the user pick a custom one, which Sonora leaves out (Eric,
 * 2026-11-11).
 */
@HiltViewModel
class ArtistDetailViewModel @Inject constructor(
    @ApplicationContext private val context: Context,
    private val libraryRepository: LibraryRepository,
    private val userPreferencesRepository: UserPreferencesRepository,
    val themeStateHolder: ThemeStateHolder,
    savedStateHandle: SavedStateHandle
) : ViewModel() {

    private val _uiState = MutableStateFlow(ArtistDetailUiState())
    val uiState: StateFlow<ArtistDetailUiState> = _uiState.asStateFlow()

    /**
     * Pre-warmed color scheme for the current artist image.
     * This is populated synchronously (from the processor's LRU/DB cache) before [uiState]
     * marks [ArtistDetailUiState.isLoading] = false, so the screen has the correct palette
     * on its very first composition — no flash from system colors.
     *
     * Consumers should read this directly instead of calling [ThemeStateHolder.getAlbumColorSchemeFlow]
     * in order to avoid the initial-null-emission that causes the flash.
     */
    private val _artistColorScheme = MutableStateFlow<ColorSchemePair?>(null)
    val artistColorScheme: StateFlow<ColorSchemePair?> = _artistColorScheme.asStateFlow()

    /**
     * Album sections the user expanded or collapsed on this visit, by [collapseKey]. Sections not
     * listed follow [ArtistDetailUiState.albumsCollapsedByDefault] (see [isAlbumSectionExpanded]).
     * Held here rather than in the screen so the choice survives configuration changes and
     * coming back from another destination.
     */
    private val _albumExpansionOverrides = MutableStateFlow<Map<String, Boolean>>(emptyMap())
    val albumExpansionOverrides: StateFlow<Map<String, Boolean>> = _albumExpansionOverrides.asStateFlow()

    init {
        savedStateHandle.getStateFlow<String?>("artistId", null)
            .onEach { idString ->
                if (idString != null) {
                    val artistId = idString.toLongOrNull()
                    if (artistId != null) {
                        loadArtistData(artistId)
                    } else {
                        _uiState.update { it.copy(error = context.getString(R.string.invalid_artist_id), isLoading = false) }
                    }
                } else {
                    _uiState.update { it.copy(error = context.getString(R.string.artist_id_not_found), isLoading = false) }
                }
            }
            .launchIn(viewModelScope)
    }

    private var currentLoadJob: Job? = null
    private var loadedArtistId: Long? = null

    private fun loadArtistData(id: Long) {
        loadedArtistId = id
        currentLoadJob?.cancel()
        currentLoadJob = viewModelScope.launch {
            Timber.tag("ArtistDebug").d("loadArtistData: id=$id")
            _uiState.update { it.copy(isLoading = true, error = null) }
            try {
                // "Start with": a page that is already open keeps its layout if the setting changes.
                val albumsCollapsedByDefault =
                    userPreferencesRepository.artistAlbumsCollapsedByDefaultFlow.first()
                val artistDetailsFlow = libraryRepository.artist(id)
                val artistSongsFlow = libraryRepository.songsForArtistFlow(id)

                combine(artistDetailsFlow, artistSongsFlow) { artist, songs ->
                    Timber.tag("ArtistDebug").d("loadArtistData: id=$id found=${artist != null} songs=${songs.size}")
                    artist to songs
                }
                    .catch { e ->
                        _uiState.update {
                            it.copy(
                                error = context.getString(R.string.error_loading_artist, e.localizedMessage ?: ""),
                                isLoading = false
                            )
                        }
                    }
                    .collect { (artist, songs) ->
                        if (artist == null) {
                            _uiState.update {
                                it.copy(error = context.getString(R.string.could_not_find_artist), isLoading = false)
                            }
                            return@collect
                        }

                        val albumSections = buildAlbumSections(songs)
                        val orderedSongs = albumSections.flatMap { it.songs }

                        // Sonora: the server's image; the original resolves a custom or Deezer one.
                        val effectiveUrl = artist.effectiveImageUrl

                        val newScheme = if (!effectiveUrl.isNullOrBlank()) {
                            try {
                                themeStateHolder.getOrGenerateColorScheme(effectiveUrl)
                            } catch (e: Exception) {
                                Timber.tag("ArtistDebug").w("Color scheme pre-warm failed: ${e.message}")
                                null
                            }
                        } else null

                        _artistColorScheme.value = newScheme
                        pruneAlbumExpansionOverrides(albumSections)
                        _uiState.value = ArtistDetailUiState(
                            artist = artist,
                            songs = orderedSongs,
                            albumSections = albumSections,
                            albumsCollapsedByDefault = albumsCollapsedByDefault,
                            effectiveImageUrl = effectiveUrl,
                            isLoading = false
                        )
                    }

            } catch (e: Exception) {
                _uiState.update {
                    it.copy(
                        error = context.getString(R.string.error_loading_artist, e.localizedMessage ?: ""),
                        isLoading = false
                    )
                }
            }
        }
    }

    /** Re-attempts loading the artist after a failure (wired to the error-state retry button). */
    fun retry() {
        loadedArtistId?.let { loadArtistData(it) }
    }

    // Sonora: no setCustomImage / clearCustomImage, the artist's image is the server's.

    fun removeSongFromAlbumSection(songId: String) {
        val updatedState = _uiState.updateAndGet { currentState ->
            val updatedAlbumSections = currentState.albumSections.map { section ->
                val updatedSongs = section.songs.filterNot { it.id == songId }
                section.copy(songs = updatedSongs)
            }.filter { it.songs.isNotEmpty() }

            currentState.copy(
                albumSections = updatedAlbumSections,
                songs = currentState.songs.filterNot { it.id == songId }
            )
        }
        pruneAlbumExpansionOverrides(updatedState.albumSections)
    }

    /** Expands or collapses one album section, keyed by [collapseKey], for the rest of this visit. */
    fun setAlbumSectionExpanded(sectionKey: String, expanded: Boolean) {
        _albumExpansionOverrides.update { it + (sectionKey to expanded) }
    }

    /** Expands or collapses every album section of the artist ("Expand all" / "Collapse all"). */
    fun setAllAlbumSectionsExpanded(expanded: Boolean) {
        _albumExpansionOverrides.value = _uiState.value.albumSections
            .filter { it.songs.isNotEmpty() }
            .associate { it.collapseKey() to expanded }
    }

    /** Drops the choices made for albums the artist no longer has. */
    private fun pruneAlbumExpansionOverrides(sections: List<ArtistAlbumSection>) {
        val currentKeys = sections.mapTo(HashSet()) { it.collapseKey() }
        _albumExpansionOverrides.update { overrides -> overrides.filterKeys { it in currentKeys } }
    }
}

// Sonora: titles keep the order they come in (MusicDao, by the server's sortKey) instead of
// lowercase titles; the sort is stable.
private val songDisplayComparator = compareBy<Song> { it.discNumber ?: 1 }
    .thenBy { if (it.trackNumber > 0) it.trackNumber else Int.MAX_VALUE }

internal fun buildAlbumSections(songs: List<Song>): List<ArtistAlbumSection> {
    if (songs.isEmpty()) return emptyList()

    val sections = songs
        .groupBy { it.albumId to it.album }
        .map { (key, albumSongs) ->
            val sortedSongs = albumSongs.sortedWith(songDisplayComparator)
            val albumYear = albumSongs.mapNotNull { song -> song.year.takeIf { it > 0 } }.maxOrNull()
            val albumArtUri = albumSongs.firstNotNullOfOrNull { it.albumArtUriString }
            ArtistAlbumSection(
                albumId = key.first,
                title = (key.second.takeIf { it.isNotBlank() } ?: "Unknown Album"),
                year = albumYear,
                albumArtUriString = albumArtUri,
                songs = sortedSongs,
                // Sonora: the server has no genres (the original takes the dominant one).
                genre = null
            )
        }

    val (withYear, withoutYear) = sections.partition { it.year != null }
    val withYearSorted = withYear.sortedWith(
        compareByDescending<ArtistAlbumSection> { it.year ?: Int.MIN_VALUE }
            .thenBy { it.title.lowercase() }
    )
    val withoutYearSorted = withoutYear.sortedBy { it.title.lowercase() }

    return withYearSorted + withoutYearSorted
}
