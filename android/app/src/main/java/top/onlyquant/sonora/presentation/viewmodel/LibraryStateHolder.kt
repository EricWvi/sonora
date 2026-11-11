package top.onlyquant.sonora.presentation.viewmodel

import androidx.paging.PagingData
import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.collections.immutable.ImmutableList
import kotlinx.collections.immutable.persistentListOf
import kotlinx.collections.immutable.toImmutableList
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.conflate
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.flatMapLatest
import kotlinx.coroutines.flow.flowOn
import kotlinx.coroutines.launch
import timber.log.Timber
import top.onlyquant.sonora.data.model.Album
import top.onlyquant.sonora.data.model.Artist
import top.onlyquant.sonora.data.model.Song
import top.onlyquant.sonora.data.model.SortOption
import top.onlyquant.sonora.data.preferences.UserPreferencesRepository
import top.onlyquant.sonora.data.repository.LibraryRepository

/**
 * Reduced from LibraryStateHolder in PixelPlayerOSS v0.4.0-alpha.53 (GPL-3.0): the songs, albums
 * and artists tabs' sort options, paging flows and lists. Sonora: the library comes from
 * [LibraryRepository], which sorts in SQL by the server's sort keys, so the lists are re-read in
 * the new order instead of sorted in memory (the original's sortAlbumsList and sortArtistsList),
 * and the albums leave out singles, as the web player's albums page does (Eric, 2026-11-11).
 * There is no storage filter (every song is on the server), no folders and no genres; the original
 * also keeps every song in memory (allSongs), which Sonora reads from Room when it needs them.
 */
@Singleton
class LibraryStateHolder @Inject constructor(
    private val libraryRepository: LibraryRepository,
    private val userPreferencesRepository: UserPreferencesRepository
) {

    private val _albums = MutableStateFlow<ImmutableList<Album>>(persistentListOf())
    val albums = _albums.asStateFlow()

    private val _artists = MutableStateFlow<ImmutableList<Artist>>(persistentListOf())
    val artists = _artists.asStateFlow()

    private val _isLoadingLibrary = MutableStateFlow(false)
    val isLoadingLibrary = _isLoadingLibrary.asStateFlow()

    private val _isLoadingCategories = MutableStateFlow(false)
    val isLoadingCategories = _isLoadingCategories.asStateFlow()

    private val _currentSongSortOption = MutableStateFlow<SortOption>(SortOption.SongDefaultOrder)
    val currentSongSortOption = _currentSongSortOption.asStateFlow()

    @OptIn(ExperimentalCoroutinesApi::class)
    val songsPagingFlow: Flow<PagingData<Song>> =
        _currentSongSortOption.flatMapLatest { sortOption ->
            libraryRepository.songsPaging(sortOption.storageKey)
        }
        .flowOn(Dispatchers.IO)

    private val _currentAlbumSortOption = MutableStateFlow<SortOption>(SortOption.AlbumTitleAZ)
    val currentAlbumSortOption = _currentAlbumSortOption.asStateFlow()

    private val _currentArtistSortOption = MutableStateFlow<SortOption>(SortOption.ArtistNameAZ)
    val currentArtistSortOption = _currentArtistSortOption.asStateFlow()

    @OptIn(ExperimentalCoroutinesApi::class)
    val albumsPagingFlow: Flow<PagingData<Album>> =
        _currentAlbumSortOption.flatMapLatest { sortOption ->
            libraryRepository.albumsPaging(sortOption.storageKey, hideSingles = HIDE_SINGLES)
        }
        .flowOn(Dispatchers.IO)

    @OptIn(ExperimentalCoroutinesApi::class)
    val artistsPagingFlow: Flow<PagingData<Artist>> =
        _currentArtistSortOption.flatMapLatest { sortOption ->
            libraryRepository.artistsPaging(sortOption.storageKey)
        }
        .flowOn(Dispatchers.IO)

    private var scope: CoroutineScope? = null
    private val owners = OwnerStack<CoroutineScope> { it }

    fun initialize(scope: CoroutineScope) {
        owners.push(scope)
        this.scope = scope
        scope.launch {
            val songSortKey = userPreferencesRepository.songsSortOptionFlow.first()
            _currentSongSortOption.value = SortOption.SONGS.find { it.storageKey == songSortKey } ?: SortOption.SongDefaultOrder

            val albumSortKey = userPreferencesRepository.albumsSortOptionFlow.first()
            _currentAlbumSortOption.value = SortOption.ALBUMS.find { it.storageKey == albumSortKey } ?: SortOption.AlbumTitleAZ

            val artistSortKey = userPreferencesRepository.artistsSortOptionFlow.first()
            _currentArtistSortOption.value = SortOption.ARTISTS.find { it.storageKey == artistSortKey } ?: SortOption.ArtistNameAZ
        }
    }

    fun onCleared(owner: CoroutineScope) {
        // A stale PlayerViewModel clearing leaves the newer binding alone; the current one
        // clearing hands the holder back to the previous live PlayerViewModel. Going scope-less
        // instead would make sorting and library loading silently no-op until process restart.
        owners.remove(owner) { next -> scope = next }
        // The library collectors die with the scope they were launched in, whichever binding
        // that was, so move them to the scope that is current now.
        if (observerScope === owner) {
            cancelLibraryObservers()
            if (scope != null) startObservingLibraryData()
        }
    }

    private var songsJob: Job? = null
    private var albumsJob: Job? = null
    private var artistsJob: Job? = null
    /** The scope the library collectors were launched in, or null while they are stopped. */
    private var observerScope: CoroutineScope? = null

    @OptIn(ExperimentalCoroutinesApi::class)
    fun startObservingLibraryData() {
        if (
            songsJob?.isActive == true &&
            albumsJob?.isActive == true &&
            artistsJob?.isActive == true
        ) {
            return
        }

        Timber.tag("LibraryStateHolder").d("startObservingLibraryData called.")
        observerScope = scope

        // Sonora: the songs tab pages from Room, so this only tells the tab when the first read
        // is done; the original collects every song here.
        songsJob = scope?.launch {
            _isLoadingLibrary.value = true
            libraryRepository.songCount.conflate().collect {
                _isLoadingLibrary.value = false
            }
        }

        albumsJob = scope?.launch {
            _isLoadingCategories.value = true
            _currentAlbumSortOption.flatMapLatest { sortOption ->
                libraryRepository.albums(sortOption.storageKey, hideSingles = HIDE_SINGLES)
            }.conflate().collect { albums ->
                _albums.value = albums.toImmutableList()
                _isLoadingCategories.value = false
            }
        }

        artistsJob = scope?.launch {
            _isLoadingCategories.value = true
            _currentArtistSortOption.flatMapLatest { sortOption ->
                libraryRepository.artists(sortOption.storageKey)
            }.conflate().collect { artists ->
                _artists.value = artists.toImmutableList()
                _isLoadingCategories.value = false
            }
        }
    }

    private fun cancelLibraryObservers() {
        songsJob?.cancel()
        albumsJob?.cancel()
        artistsJob?.cancel()
        songsJob = null
        albumsJob = null
        artistsJob = null
        observerScope = null
    }

    fun sortSongs(sortOption: SortOption, persist: Boolean = true) {
        scope?.launch {
            if (persist && _currentSongSortOption.value.storageKey == sortOption.storageKey) {
                return@launch
            }
            if (persist) {
                userPreferencesRepository.setSongsSortOption(sortOption.storageKey)
            }
            _currentSongSortOption.value = sortOption
        }
    }

    fun sortAlbums(sortOption: SortOption, persist: Boolean = true) {
        scope?.launch {
            if (persist && _currentAlbumSortOption.value.storageKey == sortOption.storageKey) {
                return@launch
            }
            if (persist) {
                userPreferencesRepository.setAlbumsSortOption(sortOption.storageKey)
            }
            _currentAlbumSortOption.value = sortOption
        }
    }

    fun sortArtists(sortOption: SortOption, persist: Boolean = true) {
        scope?.launch {
            if (persist && _currentArtistSortOption.value.storageKey == sortOption.storageKey) {
                return@launch
            }
            if (persist) {
                userPreferencesRepository.setArtistsSortOption(sortOption.storageKey)
            }
            _currentArtistSortOption.value = sortOption
        }
    }

    private companion object {
        /** The albums tab hides albums marked as singles, as the web player does. */
        const val HIDE_SINGLES = true
    }
}
