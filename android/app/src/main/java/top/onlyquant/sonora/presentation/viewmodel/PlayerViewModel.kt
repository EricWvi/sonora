package top.onlyquant.sonora.presentation.viewmodel

import android.content.ComponentName
import android.content.Context
import androidx.compose.animation.core.Animatable
import androidx.compose.runtime.snapshotFlow
import androidx.core.content.ContextCompat
import androidx.core.net.toUri
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import androidx.media3.common.C
import androidx.media3.common.MediaItem
import androidx.media3.common.MediaMetadata
import androidx.media3.common.PlaybackException
import androidx.media3.common.Player
import androidx.media3.common.Timeline
import androidx.media3.common.Tracks
import androidx.media3.session.MediaController
import androidx.media3.session.SessionToken
import com.google.common.util.concurrent.ListenableFuture
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import kotlinx.collections.immutable.ImmutableList
import kotlinx.collections.immutable.PersistentList
import kotlinx.collections.immutable.persistentListOf
import kotlinx.collections.immutable.toImmutableList
import kotlinx.collections.immutable.toPersistentList
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.channels.BufferOverflow
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.flowOf
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import kotlinx.coroutines.withTimeoutOrNull
import kotlinx.coroutines.yield
import timber.log.Timber
import top.onlyquant.sonora.R
import top.onlyquant.sonora.data.model.Album
import top.onlyquant.sonora.data.model.Artist
import top.onlyquant.sonora.data.model.LibraryTabId
import top.onlyquant.sonora.data.model.SortOption
import top.onlyquant.sonora.data.repository.LibraryRepository
import top.onlyquant.sonora.presentation.selection.flattenDistinctGroups
import top.onlyquant.sonora.data.model.Song
import top.onlyquant.sonora.data.preferences.AlbumArtQuality
import top.onlyquant.sonora.data.preferences.CarouselStyle
import top.onlyquant.sonora.data.preferences.FullPlayerLoadingTweaks
import top.onlyquant.sonora.data.preferences.NavBarStyle
import top.onlyquant.sonora.data.preferences.ThemePreference
import top.onlyquant.sonora.data.preferences.ThemePreferencesRepository
import top.onlyquant.sonora.data.preferences.UserPreferencesRepository
import top.onlyquant.sonora.data.preferences.restartsAtPosition
import top.onlyquant.sonora.data.service.MusicService
import top.onlyquant.sonora.data.service.player.DualPlayerEngine
import top.onlyquant.sonora.di.DispatcherProvider
import top.onlyquant.sonora.presentation.navigation.PlayerSheetBackCoordinator
import top.onlyquant.sonora.utils.MediaItemBuilder
import top.onlyquant.sonora.data.api.ApiException
import top.onlyquant.sonora.data.api.FmState
import top.onlyquant.sonora.data.api.SonoraApi
import top.onlyquant.sonora.data.auth.SessionStore
import top.onlyquant.sonora.data.fm.FM_QUEUE_TAG
import top.onlyquant.sonora.data.fm.albumEnded
import top.onlyquant.sonora.data.fm.songs
import top.onlyquant.sonora.data.fm.songsToAppend
import top.onlyquant.sonora.data.fm.startSong
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import top.onlyquant.sonora.utils.QueueUtils

private fun List<Song>.toPlaybackQueue(): ImmutableList<Song> = when (this) {
    is PersistentList<Song> -> this
    is ImmutableList<Song> -> this
    else -> this.toPersistentList()
}

private fun ImmutableList<Song>.asPersistentPlaybackQueue(): PersistentList<Song> =
    this as? PersistentList<Song> ?: this.toPersistentList()

private fun ImmutableList<Song>.moveSong(fromIndex: Int, toIndex: Int): ImmutableList<Song> {
    if (fromIndex == toIndex || fromIndex !in indices || toIndex !in indices) return this
    val movedSong = this[fromIndex]
    return asPersistentPlaybackQueue()
        .removingAt(fromIndex)
        .addingAt(toIndex, movedSong)
}

private fun moveQueueIndex(index: Int, fromIndex: Int, toIndex: Int): Int {
    if (index == C.INDEX_UNSET || fromIndex == toIndex) return index
    return when {
        index == fromIndex -> toIndex
        fromIndex < toIndex && index in (fromIndex + 1)..toIndex -> index - 1
        toIndex < fromIndex && index in toIndex until fromIndex -> index + 1
        else -> index
    }
}

private const val SHUFFLE_ALL_SONG_LIMIT = 10_000

data class LocateScrollRequest(
    val tab: LibraryTabId,
    val index: Int
)

private data class PreparedPlaybackQueueSegments(
    val beforeCurrent: List<MediaItem>,
    val afterCurrent: List<MediaItem>,
    val currentIndex: Int
)

private data class QueueTimelineSignature(
    val count: Int,
    val orderHash: Long,
    val firstMediaId: String?,
    val lastMediaId: String?
)

/** Format details of the playing track, read from the selected audio track. */
data class PlaybackAudioMetadata(
    val mediaId: String? = null,
    val mimeType: String? = null,
    val bitrate: Int? = null,
    val sampleRate: Int? = null,
    val channelCount: Int? = null,
    val bitDepth: Int? = null
)

/**
 * Reduced from PlayerViewModel in PixelPlayerOSS v0.4.0-alpha.53 (GPL-3.0; 5,191 lines there):
 * what the player sheet, the mini player and the queue use — playback, the queue, shuffle and
 * repeat, the sleep timer, the album-art color scheme and the sheet state. The method bodies follow
 * the original.
 *
 * The library tabs (songs, albums, artists) read the local library through [LibraryRepository]
 * and [LibraryStateHolder], as the original reads MusicRepository.
 *
 * Left out until the screens that need them are copied: folders, search, daily mixes, stats,
 * lyrics (step 5), likes (step 5), playlists (step 5), metadata editing, deletion, external media,
 * the cast and cloud sources, app shortcuts and queue persistence.
 */
@HiltViewModel
class PlayerViewModel @Inject constructor(
    @param:ApplicationContext private val context: Context,
    themePreferencesRepository: ThemePreferencesRepository,
    private val userPreferencesRepository: UserPreferencesRepository,
    private val dualPlayerEngine: DualPlayerEngine,
    private val queueStateHolder: QueueStateHolder,
    private val queueUndoStateHolder: QueueUndoStateHolder,
    private val playlistDismissUndoStateHolder: PlaylistDismissUndoStateHolder,
    private val playbackStateHolder: PlaybackStateHolder,
    private val sleepTimerStateHolder: SleepTimerStateHolder,
    val themeStateHolder: ThemeStateHolder,
    private val libraryRepository: LibraryRepository,
    private val libraryStateHolder: LibraryStateHolder,
    private val libraryTabsStateHolder: LibraryTabsStateHolder,
    val multiSelectionStateHolder: MultiSelectionStateHolder,
    private val dispatchers: DispatcherProvider,
    // Sonora: Roaming FM (see the section of that name below).
    private val api: SonoraApi,
    private val sessions: SessionStore,
) : ViewModel() {

    private val _playerUiState = MutableStateFlow(PlayerUiState())
    val playerUiState: StateFlow<PlayerUiState> = _playerUiState.asStateFlow()

    val queueFlow: StateFlow<ImmutableList<Song>> = _playerUiState
        .map { it.currentPlaybackQueue }
        .distinctUntilChanged()
        .stateIn(
            scope = viewModelScope,
            started = SharingStarted.WhileSubscribed(5000),
            initialValue = persistentListOf()
        )

    val stablePlayerState: StateFlow<StablePlayerState> = playbackStateHolder.stablePlayerState

    /**
     * High-frequency playback position should not force global UI recomposition.
     * Keep a dedicated position flow for real-time UI elements (seek bars, lyrics timing).
     */
    val currentPlaybackPosition: StateFlow<Long> = playbackStateHolder.currentPosition

    private val _sheetState = MutableStateFlow(PlayerSheetState.COLLAPSED)
    val sheetState: StateFlow<PlayerSheetState> = _sheetState.asStateFlow()
    private val _isSheetVisible = MutableStateFlow(false)
    val isSheetVisible: StateFlow<Boolean> = _isSheetVisible.asStateFlow()
    private val _predictiveBackCollapseFraction = MutableStateFlow(0f)
    val predictiveBackCollapseFraction: StateFlow<Float> = _predictiveBackCollapseFraction.asStateFlow()
    private val _predictiveBackSwipeEdge = MutableStateFlow<Int?>(null)
    val predictiveBackSwipeEdge: StateFlow<Int?> = _predictiveBackSwipeEdge.asStateFlow()
    private val _isQueueSheetVisible = MutableStateFlow(false)
    val isQueueSheetVisible: StateFlow<Boolean> = _isQueueSheetVisible.asStateFlow()
    internal val playerSheetBackCoordinator = PlayerSheetBackCoordinator()

    val playerContentExpansionFraction = Animatable(0f)

    private val _isMiniPlayerDismissing = MutableStateFlow(false)
    val isMiniPlayerDismissing: StateFlow<Boolean> = _isMiniPlayerDismissing.asStateFlow()

    fun setMiniPlayerDismissing(dismissing: Boolean) {
        _isMiniPlayerDismissing.value = dismissing
    }

    val activePlayerColorSchemePair: StateFlow<ColorSchemePair?> = themeStateHolder.activePlayerColorSchemePair
    val currentThemedAlbumArtUri: StateFlow<String?> = themeStateHolder.currentAlbumArtUri

    val playerThemePreference: StateFlow<String> = themePreferencesRepository.playerThemePreferenceFlow
        .stateIn(
            scope = viewModelScope,
            started = SharingStarted.WhileSubscribed(5000),
            initialValue = ThemePreference.ALBUM_ART
        )

    private val _isImmersiveTemporarilyDisabled = MutableStateFlow(false)

    fun setImmersiveTemporarilyDisabled(disabled: Boolean) {
        _isImmersiveTemporarilyDisabled.value = disabled
    }

    // Sonora: in the original these two are in SettingsViewModel (the queue sheet reads them from
    // there); they move back when the settings screen is copied.
    val playbackSpeed: StateFlow<Float> = userPreferencesRepository.playbackSpeedFlow
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), 1.0f)

    fun setPlaybackSpeed(speed: Float) {
        viewModelScope.launch {
            userPreferencesRepository.setPlaybackSpeed(speed)
        }
    }

    val showQueueHistory: StateFlow<Boolean> = userPreferencesRepository.showQueueHistoryFlow
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), false)

    val isEndOfTrackTimerActive: StateFlow<Boolean> = sleepTimerStateHolder.isEndOfTrackTimerActive
    val activeTimerValueDisplay: StateFlow<String?> = sleepTimerStateHolder.activeTimerValueDisplay
    val activeTimerDurationMinutes: StateFlow<Int?> = sleepTimerStateHolder.activeTimerDurationMinutes
    val playCount: StateFlow<Float> = sleepTimerStateHolder.playCount

    private var bufferingDebounceJob: Job? = null

    private val _toastEvents = MutableSharedFlow<String>(
        extraBufferCapacity = 1,
        onBufferOverflow = BufferOverflow.DROP_OLDEST
    )
    val toastEvents = _toastEvents.asSharedFlow()

    private val _albumNavigationRequests = MutableSharedFlow<Long>(extraBufferCapacity = 1)
    val albumNavigationRequests = _albumNavigationRequests.asSharedFlow()
    private val _artistNavigationRequests = MutableSharedFlow<Long>(extraBufferCapacity = 1)
    val artistNavigationRequests = _artistNavigationRequests.asSharedFlow()

    private var albumNavigationJob: Job? = null
    private var artistNavigationJob: Job? = null
    private var fullQueuePlaybackJob: Job? = null
    private var fullQueuePlaybackToken: Long = 0L
    private var directPlaybackJob: Job? = null
    private var directPlaybackToken: Long = 0L
    private var pendingQueueSegmentsJob: Job? = null
    private var transitionSchedulerJob: Job? = null

    private val _playbackAudioMetadata = MutableStateFlow(PlaybackAudioMetadata())

    /** Sonora: likes come with step 5, so the player shows every song as not liked. */
    val isCurrentSongFavorite: StateFlow<Boolean> = MutableStateFlow(false).asStateFlow()

    /**
     * Sonora: the original reads the artists of the current song from the local library; until
     * the library is copied they come from the song itself.
     */
    private val currentSongArtists: StateFlow<ImmutableList<Artist>> = stablePlayerState
        .map { state ->
            state.currentSong?.artists.orEmpty()
                .map { Artist(id = it.id, name = it.name, songCount = 0) }
                .toPersistentList()
        }
        .distinctUntilChanged()
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), persistentListOf())

    data class FullPlayerSlice(
        val currentSongArtists: ImmutableList<Artist> = persistentListOf(),
        val lyricsSyncOffset: Int = 0,
        val albumArtQuality: AlbumArtQuality = AlbumArtQuality.MEDIUM,
        val audioMetadata: PlaybackAudioMetadata = PlaybackAudioMetadata(),
        val showPlayerFileInfo: Boolean = true,
        val immersiveLyricsEnabled: Boolean = false,
        val immersiveLyricsTimeout: Long = 4000L,
        val isImmersiveTemporarilyDisabled: Boolean = false,
        val isBluetoothEnabled: Boolean = false,
        val bluetoothName: String? = null
    )

    // Sonora: the settings in these slices keep the original's defaults until the settings
    // screen is copied.
    val fullPlayerSlice: StateFlow<FullPlayerSlice> = combine(
        currentSongArtists,
        _playbackAudioMetadata,
        _isImmersiveTemporarilyDisabled
    ) { artists, audioMeta, immersiveDisabled ->
        FullPlayerSlice(
            currentSongArtists = artists,
            audioMetadata = audioMeta,
            isImmersiveTemporarilyDisabled = immersiveDisabled
        )
    }
        .distinctUntilChanged()
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), FullPlayerSlice())

    data class PlayerConfigSlice(
        val navBarCornerRadius: Int = 32,
        val navBarStyle: String = NavBarStyle.DEFAULT,
        val carouselStyle: String = CarouselStyle.NO_PEEK,
        val fullPlayerLoadingTweaks: FullPlayerLoadingTweaks = FullPlayerLoadingTweaks(),
        val tapBackgroundClosesPlayer: Boolean = false,
        val useSmoothCorners: Boolean = true,
        val playerThemePreference: String = ThemePreference.ALBUM_ART,
        val miniPlayerSwipeToSkip: Boolean = false
    )

    val playerConfigSlice: StateFlow<PlayerConfigSlice> = playerThemePreference
        .map { PlayerConfigSlice(playerThemePreference = it) }
        .distinctUntilChanged()
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), PlayerConfigSlice())

    /** Sonora: songs are known from the queue only, see the class comment. */
    fun observeSong(songId: String?): Flow<Song?> {
        if (songId == null) return flowOf(null)
        return _playerUiState
            .map { state -> state.currentPlaybackQueue.firstOrNull { it.id == songId } }
            .distinctUntilChanged()
    }

    private val sessionToken = SessionToken(context, ComponentName(context, MusicService::class.java))
    private val mediaControllerFuture: ListenableFuture<MediaController> =
        MediaController.Builder(context, sessionToken).buildAsync()
    private var mediaController: MediaController? = null
    private var mediaControllerPlaybackListener: Player.Listener? = null
    private var pendingPlaybackAction: (() -> Unit)? = null

    init {
        playbackStateHolder.initialize(coroutineScope = viewModelScope)
        themeStateHolder.initialize(viewModelScope)
        libraryStateHolder.initialize(viewModelScope)
        observeLibraryState()

        mediaControllerFuture.addListener({
            try {
                mediaController = mediaControllerFuture.get()
                playbackStateHolder.setMediaController(mediaController)
                setupMediaControllerListeners()
                pendingPlaybackAction?.invoke()
                pendingPlaybackAction = null
            } catch (e: Exception) {
                _playerUiState.update { it.copy(isLoadingInitialSongs = false) }
                Timber.tag("PlayerViewModel").e(e, "Error setting up MediaController")
            }
        }, ContextCompat.getMainExecutor(context))

        sleepTimerStateHolder.initialize(
            scope = viewModelScope,
            toastEmitter = { msg -> _toastEvents.emit(msg) },
            mediaControllerProvider = { mediaController },
            songTitleResolver = { songId ->
                _playerUiState.value.currentPlaybackQueue.firstOrNull { it.id == songId }?.title
            }
        )

        playlistDismissUndoStateHolder.observeUndoStateAgainstPlayback(
            scope = viewModelScope,
            currentSongIdFlow = stablePlayerState.map { it.currentSong?.id },
            getUiState = { _playerUiState.value },
            onHideDismissUndoBar = { hideDismissUndoBar() }
        )
    }

    // ------------------------
    // Library (the original's library section, with LibraryRepository for MusicRepository)
    // ------------------------

    private val _scrollToIndexEvent = MutableSharedFlow<LocateScrollRequest>(extraBufferCapacity = 1)
    val scrollToIndexEvent = _scrollToIndexEvent.asSharedFlow()

    /** Sonora: likes come with step 5; until then no song is liked. */
    val favoriteSongIds: StateFlow<Set<String>> = MutableStateFlow<Set<String>>(emptySet()).asStateFlow()

    val lastLibraryTabIndexFlow: StateFlow<Int> =
        userPreferencesRepository.lastLibraryTabIndexFlow.stateIn(
            scope = viewModelScope,
            started = SharingStarted.WhileSubscribed(5000),
            initialValue = 0
        )

    private val defaultLibraryTabs: ImmutableList<String> =
        LibraryTabId.entries.map { it.storageKey }.toImmutableList()

    /**
     * Sonora: always the default order. The original reads the order saved by its "reorder tabs"
     * sheet, which Sonora leaves out (Eric, 2026-11-11).
     */
    val libraryTabsFlow: StateFlow<ImmutableList<String>> =
        MutableStateFlow(defaultLibraryTabs).asStateFlow()

    private val _loadedTabs = MutableStateFlow(emptySet<String>())

    private val _currentLibraryTabId = MutableStateFlow(LibraryTabId.SONGS)
    val currentLibraryTabId: StateFlow<LibraryTabId> = _currentLibraryTabId.asStateFlow()

    private val _isSortingSheetVisible = MutableStateFlow(false)
    val isSortingSheetVisible: StateFlow<Boolean> = _isSortingSheetVisible.asStateFlow()

    val availableSortOptions: StateFlow<List<SortOption>> =
        currentLibraryTabId.map { tabId ->
            when (tabId) {
                LibraryTabId.SONGS -> SortOption.SONGS
                LibraryTabId.ALBUMS -> SortOption.ALBUMS
                LibraryTabId.ARTISTS -> SortOption.ARTISTS
                LibraryTabId.PLAYLISTS -> SortOption.PLAYLISTS
            }
        }.stateIn(
            scope = viewModelScope,
            started = SharingStarted.WhileSubscribed(5000),
            initialValue = SortOption.SONGS
        )

    val songCountFlow: StateFlow<Int> = libraryRepository.songCount
        .stateIn(
            scope = viewModelScope,
            started = SharingStarted.WhileSubscribed(5000),
            initialValue = 0
        )

    val albumsFlow: StateFlow<ImmutableList<Album>> = libraryStateHolder.albums
    val artistsFlow: StateFlow<ImmutableList<Artist>> = libraryStateHolder.artists

    private fun observeLibraryState() {
        viewModelScope.launch {
            userPreferencesRepository.isAlbumsListViewFlow.collect { albumsList ->
                _playerUiState.update { it.copy(isAlbumsListView = albumsList) }
            }
        }
        viewModelScope.launch {
            // Sonora: the original's SyncManager.isSyncing; here the library sync.
            libraryRepository.isSyncing.collect { isSyncing ->
                _playerUiState.update { it.copy(isSyncingLibrary = isSyncing) }
            }
        }
        viewModelScope.launch {
            combine(
                libraryStateHolder.isLoadingLibrary,
                libraryStateHolder.isLoadingCategories,
            ) { loadingLibrary, loadingCategories ->
                loadingLibrary to loadingCategories
            }.collect { (loadingLibrary, loadingCategories) ->
                _playerUiState.update {
                    it.copy(
                        isLoadingInitialSongs = loadingLibrary,
                        isLoadingLibraryCategories = loadingCategories,
                    )
                }
            }
        }
        viewModelScope.launch {
            combine(
                libraryStateHolder.currentSongSortOption,
                libraryStateHolder.currentAlbumSortOption,
                libraryStateHolder.currentArtistSortOption,
            ) { songSort, albumSort, artistSort ->
                Triple(songSort, albumSort, artistSort)
            }.collect { (songSort, albumSort, artistSort) ->
                _playerUiState.update {
                    it.copy(
                        currentSongSortOption = songSort,
                        currentAlbumSortOption = albumSort,
                        currentArtistSortOption = artistSort,
                    )
                }
            }
        }
    }

    /** Finds the current song in the songs tab's list and emits its index on [scrollToIndexEvent]. */
    fun requestLocateCurrentSong(tab: LibraryTabId) {
        val currentSong = stablePlayerState.value.currentSong ?: return

        viewModelScope.launch {
            try {
                val state = playerUiState.value
                // Sonora: only the songs tab locates; the original also locates in liked.
                val sortedIds = libraryRepository.allSongs(state.currentSongSortOption.storageKey)
                    .map { it.id }
                val index = sortedIds.indexOf(currentSong.id)

                if (index != -1) {
                    _scrollToIndexEvent.emit(LocateScrollRequest(tab = tab, index = index))
                } else {
                    sendToast(context.getString(R.string.player_song_not_found_in_list))
                }
            } catch (e: Exception) {
                if (e is CancellationException) throw e
                Timber.e(e, "Failed to locate current song")
                sendToast(context.getString(R.string.player_could_not_locate_song))
            }
        }
    }

    fun showAndPlaySongFromLibrary(
        song: Song,
        queueName: String = context.getString(R.string.tab_library),
        isVoluntaryPlay: Boolean = true
    ) {
        launchLatestFullQueuePlayback(
            song = song,
            queueName = queueName,
            isVoluntaryPlay = isVoluntaryPlay,
            failureMessage = "Failed to build full library queue for songId=%s"
        ) {
            libraryRepository.allSongs(playerUiState.value.currentSongSortOption.storageKey)
        }
    }

    suspend fun getSongsForCurrentLibrarySelection(): List<Song> =
        libraryRepository.allSongs(playerUiState.value.currentSongSortOption.storageKey)

    /**
     * Sonora: the provider returns the songs in order, read from Room; the original's returns
     * their IDs, resolved through its in-memory library (resolvePlaybackQueueFromSortedIds).
     */
    private fun launchLatestFullQueuePlayback(
        song: Song,
        queueName: String,
        isVoluntaryPlay: Boolean,
        failureMessage: String,
        queueProvider: suspend () -> List<Song>
    ) {
        cancelPendingFullQueuePlayback()
        cancelPendingDirectPlayback()
        val requestToken = fullQueuePlaybackToken

        fullQueuePlaybackJob = viewModelScope.launch {
            try {
                val fullQueue = queueProvider()
                throwIfFullQueuePlaybackRequestIsStale(requestToken)

                showAndPlaySong(
                    song = song,
                    contextSongs = fullQueue.ifEmpty { listOf(song) },
                    queueName = queueName,
                    isVoluntaryPlay = isVoluntaryPlay,
                    cancelPendingQueueBuild = false
                )
            } catch (cancelled: CancellationException) {
                throw cancelled
            } catch (error: Exception) {
                if (requestToken != fullQueuePlaybackToken) {
                    return@launch
                }

                Timber.e(error, failureMessage, song.id)
                showAndPlaySong(
                    song = song,
                    contextSongs = listOf(song),
                    queueName = queueName,
                    isVoluntaryPlay = isVoluntaryPlay,
                    cancelPendingQueueBuild = false
                )
            }
        }
    }

    private fun cancelPendingFullQueuePlayback() {
        fullQueuePlaybackToken += 1L
        fullQueuePlaybackJob?.cancel()
        fullQueuePlaybackJob = null
    }

    private fun throwIfFullQueuePlaybackRequestIsStale(requestToken: Long) {
        if (requestToken != fullQueuePlaybackToken) {
            throw CancellationException("Stale full-queue playback request")
        }
    }

    private fun cancelPendingDirectPlayback() {
        cancelPendingDirectPlaybackBuild()
        pendingQueueSegmentsJob?.cancel()
        pendingQueueSegmentsJob = null
    }

    fun shuffleAllSongs(queueName: String = context.getString(R.string.queue_name_all_songs_shuffled)) {
        Timber.tag("ShuffleDebug").d("shuffleAllSongs called.")

        viewModelScope.launch {
            val randomSongs = libraryRepository.randomSongs(limit = SHUFFLE_ALL_SONG_LIMIT)
            if (randomSongs.isNotEmpty()) {
                playSongsShuffled(randomSongs, queueName, startAtZero = true)
            }
        }
    }

    /** Only the album pick is random; its tracks play in album order. */
    fun playRandomAlbum() {
        libraryStateHolder.albums.value.randomOrNull()?.let(::playAlbum)
    }

    fun shuffleRandomArtist() {
        viewModelScope.launch {
            val allArtists = libraryStateHolder.artists.value
            if (allArtists.isNotEmpty()) {
                val randomArtist = allArtists.random()
                val artistSongs = libraryRepository.songsForArtist(randomArtist.id)
                if (artistSongs.isNotEmpty()) {
                    playSongsShuffled(artistSongs, randomArtist.name, startAtZero = true)
                }
            }
        }
    }

    /** Sonora: the repository already orders the album by disc and track, as the original sorts here. */
    fun playAlbum(album: Album) {
        Timber.tag("ShuffleDebug").d("playAlbum called for album: ${album.title}")
        viewModelScope.launch {
            try {
                val songsList: List<Song> = withContext(dispatchers.io) {
                    libraryRepository.songsForAlbum(album.id)
                }

                if (songsList.isNotEmpty()) {
                    playSongs(songsList, songsList.first(), album.title, null)
                    _isSheetVisible.value = true
                } else {
                    Timber.tag("PlayerViewModel").w("Album '${album.title}' has no playable songs.")
                }
            } catch (e: Exception) {
                if (e is CancellationException) throw e
                Timber.tag("PlayerViewModel").e(e, "Error playing album ${album.title}")
            }
        }
    }

    fun playSongsShuffled(
        songsToPlay: List<Song>,
        queueName: String = "None",
        playlistId: String? = null,
        startAtZero: Boolean = false
    ) {
        cancelPendingFullQueuePlayback()
        val requestToken = beginDirectPlaybackRequest()
        directPlaybackJob = viewModelScope.launch {
            val result = queueStateHolder.prepareShuffledQueueSuspending(
                songsToPlay,
                startAtZero,
                playlistId
            )
            throwIfDirectPlaybackRequestIsStale(requestToken)
            if (result == null) {
                sendToast(context.getString(R.string.player_no_songs_to_shuffle))
                return@launch
            }

            val (shuffledQueue, startSong) = result
            transitionSchedulerJob?.cancel()

            playbackStateHolder.updateStablePlayerState { it.copy(isShuffleEnabled = true) }
            launch { userPreferencesRepository.setShuffleOn(true) }

            internalPlaySongs(shuffledQueue, startSong, queueName, playlistId)
            if (requestToken == directPlaybackToken) {
                directPlaybackJob = null
            }
        }
    }

    fun playSelectedSongs(songs: List<Song>) {
        if (songs.isEmpty()) return
        playSongs(songs, songs.first(), context.getString(R.string.queue_name_selected_songs))
        multiSelectionStateHolder.clearSelection()
    }

    fun addSelectedToQueue(songs: List<Song>) {
        songs.forEach { addSongToQueue(it) }
        viewModelScope.launch {
            val n = songs.size
            _toastEvents.emit(
                context.resources.getQuantityString(R.plurals.n_songs_added_to_queue, n, n),
            )
        }
        multiSelectionStateHolder.clearSelection()
    }

    fun addSelectedAsNext(songs: List<Song>) {
        songs.reversed().forEach { addSongNextToQueue(it) }
        viewModelScope.launch {
            val n = songs.size
            _toastEvents.emit(
                context.resources.getQuantityString(R.plurals.n_songs_will_play_next, n, n),
            )
        }
        multiSelectionStateHolder.clearSelection()
    }

    /** Sonora: likes come with step 5; until then liking the selection does nothing. */
    @Suppress("UNUSED_PARAMETER")
    fun likeSelectedSongs(songs: List<Song>) = Unit

    /** Sonora: likes come with step 5; until then unliking the selection does nothing. */
    @Suppress("UNUSED_PARAMETER")
    fun unlikeSelectedSongs(songs: List<Song>) = Unit

    suspend fun resolveAlbumSongsForBatchActions(albums: List<Album>): List<Song> {
        if (albums.isEmpty()) return emptyList()
        val visibleSongIds = currentLibrarySelectionSongIds()
        return withContext(dispatchers.io) {
            flattenDistinctGroups(
                groups = albums.map { album ->
                    libraryRepository.songsForAlbum(album.id)
                        .filter { it.id in visibleSongIds }
                },
                keyOf = Song::id
            )
        }
    }

    suspend fun resolveArtistSongsForBatchActions(artists: List<Artist>): List<Song> {
        if (artists.isEmpty()) return emptyList()
        val visibleSongIds = currentLibrarySelectionSongIds()
        return withContext(dispatchers.io) {
            flattenDistinctGroups(
                groups = artists.map { artist ->
                    libraryRepository.songsForArtist(artist.id)
                        .filter { it.id in visibleSongIds }
                },
                keyOf = Song::id
            )
        }
    }

    private suspend fun currentLibrarySelectionSongIds(): Set<String> =
        libraryRepository.allSongs(playerUiState.value.currentSongSortOption.storageKey)
            .mapTo(hashSetOf()) { it.id }

    fun sortSongs(sortOption: SortOption, persist: Boolean = true) {
        libraryStateHolder.sortSongs(sortOption, persist)
    }

    fun sortAlbums(sortOption: SortOption, persist: Boolean = true) {
        libraryStateHolder.sortAlbums(sortOption, persist)
    }

    fun sortArtists(sortOption: SortOption, persist: Boolean = true) {
        libraryStateHolder.sortArtists(sortOption, persist)
    }

    fun setAlbumsListView(isList: Boolean) {
        viewModelScope.launch {
            userPreferencesRepository.setAlbumsListView(isList)
        }
    }

    /** Sonora: the button is shown as in the original but does nothing (Eric, 2026-11-11). */
    fun toggleStorageFilter() = Unit

    fun showSortingSheet() {
        libraryTabsStateHolder.showSortingSheet(_isSortingSheetVisible)
    }

    fun hideSortingSheet() {
        libraryTabsStateHolder.hideSortingSheet(_isSortingSheetVisible)
    }

    fun onLibraryTabSelected(tabIndex: Int) {
        libraryTabsStateHolder.onLibraryTabSelected(
            tabIndex = tabIndex,
            libraryTabs = libraryTabsFlow.value,
            loadedTabs = _loadedTabs,
            currentLibraryTabId = _currentLibraryTabId,
            saveLastTabIndex = { index -> userPreferencesRepository.saveLastLibraryTabIndex(index) },
            scope = viewModelScope,
            loadSongs = { libraryStateHolder.startObservingLibraryData() },
            loadAlbums = { libraryStateHolder.startObservingLibraryData() },
            loadArtists = { libraryStateHolder.startObservingLibraryData() }
        )
    }

    fun sendToast(message: String) {
        viewModelScope.launch {
            _toastEvents.emit(message)
        }
    }

    fun updatePredictiveBackCollapseFraction(fraction: Float) {
        _predictiveBackCollapseFraction.value = fraction.coerceIn(0f, 1f)
    }

    fun updatePredictiveBackSwipeEdge(edge: Int?) {
        _predictiveBackSwipeEdge.value = edge
    }

    fun resetPredictiveBackState() {
        _predictiveBackCollapseFraction.value = 0f
        _predictiveBackSwipeEdge.value = null
    }

    fun updateQueueSheetVisibility(visible: Boolean) {
        _isQueueSheetVisible.value = visible
    }

    fun showAndPlaySong(
        song: Song,
        contextSongs: List<Song>,
        queueName: String = "Current Context",
        isVoluntaryPlay: Boolean = true,
        cancelPendingQueueBuild: Boolean = true,
        playlistId: String? = null,
        indexInQueue: Int? = null
    ) {
        val playbackContext =
            if (contextSongs.any { it.id == song.id }) contextSongs else listOf(song)
        val controller = mediaController
        val currentQueue = _playerUiState.value.currentPlaybackQueue
        val songIndexInQueue = indexInQueue ?: currentQueue.indexOfFirst { it.id == song.id }
        val queueMatchesContext = currentQueue.matchesSongOrder(playbackContext)
        val reusableTargetIndex = if (
            controller != null &&
            controller.isConnected &&
            !dualPlayerEngine.isTransitionRunning() &&
            songIndexInQueue != -1 &&
            queueMatchesContext
        ) {
            controller.resolveReusablePlaybackTargetIndex(
                songIndexInQueue = songIndexInQueue,
                songId = song.id,
                isExplicitQueueTarget = indexInQueue != null
            )
        } else {
            null
        }

        if (controller != null && reusableTargetIndex != null) {
            cancelPendingDirectPlaybackBuild()
            playLoadedControllerItem(controller, reusableTargetIndex)
        } else {
            playSongs(playbackContext, song, queueName, playlistId)
        }
        resetPredictiveBackState()
    }

    fun showAndPlaySong(song: Song) {
        showAndPlaySong(song, listOf(song), context.getString(R.string.tab_library))
    }

    private fun List<Song>.matchesSongOrder(contextSongs: List<Song>): Boolean {
        if (size != contextSongs.size) return false
        return indices.all { this[it].id == contextSongs[it].id }
    }

    private fun MediaController.resolveReusablePlaybackTargetIndex(
        songIndexInQueue: Int,
        songId: String,
        isExplicitQueueTarget: Boolean = false
    ): Int? {
        if (!isExplicitQueueTarget) {
            currentMediaItem?.takeIf { it.mediaId == songId }?.let {
                return currentMediaItemIndex.takeIf { index -> index != C.INDEX_UNSET } ?: 0
            }
        }

        if (songIndexInQueue !in 0 until mediaItemCount) return null

        val mediaIdAtTarget = runCatching { getMediaItemAt(songIndexInQueue).mediaId }.getOrNull()
        return songIndexInQueue.takeIf { mediaIdAtTarget == songId }
    }

    private fun playLoadedControllerItem(controller: MediaController, targetIndex: Int) {
        val shouldSeekToStart =
            controller.currentMediaItemIndex != targetIndex ||
                controller.playbackState == Player.STATE_ENDED

        if (shouldSeekToStart) {
            controller.seekTo(targetIndex, 0L)
        }
        if (controller.playbackState == Player.STATE_IDLE && controller.mediaItemCount > 0) {
            controller.prepare()
        }
        controller.play()
    }

    fun removeSongFromQueue(songId: String) {
        queueUndoStateHolder.removeSongFromQueue(
            scope = viewModelScope,
            mediaController = mediaController,
            songId = songId,
            getUiState = { _playerUiState.value },
            updateUiState = { mutation -> _playerUiState.update(mutation) }
        )
    }

    fun undoRemoveSongFromQueue() {
        queueUndoStateHolder.undoRemoveSongFromQueue(
            mediaController = mediaController,
            getUiState = { _playerUiState.value },
            updateUiState = { mutation -> _playerUiState.update(mutation) }
        )
    }

    fun hideQueueItemUndoBar() {
        queueUndoStateHolder.hideQueueItemUndoBar { mutation ->
            _playerUiState.update(mutation)
        }
    }

    fun reorderQueueItem(fromIndex: Int, toIndex: Int) {
        mediaController?.let { controller ->
            if (fromIndex >= 0 && fromIndex < controller.mediaItemCount &&
                toIndex >= 0 && toIndex < controller.mediaItemCount) {
                val currentIndexBeforeMove = controller.currentMediaItemIndex
                    .takeIf { it != C.INDEX_UNSET }
                    ?: playbackStateHolder.stablePlayerState.value.currentMediaItemIndex
                val updatedCurrentIndex = moveQueueIndex(currentIndexBeforeMove, fromIndex, toIndex)

                controller.moveMediaItem(fromIndex, toIndex)

                _playerUiState.update { state ->
                    val updatedQueue = state.currentPlaybackQueue.moveSong(fromIndex, toIndex)
                    if (updatedQueue === state.currentPlaybackQueue) {
                        state
                    } else {
                        state.copy(currentPlaybackQueue = updatedQueue)
                    }
                }

                playbackStateHolder.updateStablePlayerState { state ->
                    if (updatedCurrentIndex == C.INDEX_UNSET ||
                        state.currentMediaItemIndex == updatedCurrentIndex
                    ) {
                        state
                    } else {
                        state.copy(currentMediaItemIndex = updatedCurrentIndex)
                    }
                }
            }
        }
    }

    fun togglePlayerSheetState(resetPredictiveState: Boolean = true) {
        _sheetState.value = if (_sheetState.value == PlayerSheetState.COLLAPSED) {
            PlayerSheetState.EXPANDED
        } else {
            PlayerSheetState.COLLAPSED
        }
        if (resetPredictiveState) {
            resetPredictiveBackState()
        }
    }

    fun expandPlayerSheet(resetPredictiveState: Boolean = true) {
        _sheetState.value = PlayerSheetState.EXPANDED
        if (resetPredictiveState) {
            resetPredictiveBackState()
        }
    }

    fun collapsePlayerSheet(resetPredictiveState: Boolean = true) {
        _sheetState.value = PlayerSheetState.COLLAPSED
        if (resetPredictiveState) {
            resetPredictiveBackState()
        }
    }

    fun triggerAlbumNavigationFromPlayer(albumId: Long) {
        if (albumId == -1L) return

        val existingJob = albumNavigationJob
        if (existingJob != null && existingJob.isActive) return

        albumNavigationJob?.cancel()
        albumNavigationJob = viewModelScope.launch {
            collapsePlayerSheet()

            withTimeoutOrNull(900) {
                awaitSheetState(PlayerSheetState.COLLAPSED)
                awaitPlayerCollapse()
            }

            _albumNavigationRequests.emit(albumId)
        }
    }

    /**
     * Sonora: the original looks an artist without an ID up by name in the local library; every
     * Sonora artist reference has an ID, so there is nothing to resolve.
     */
    fun triggerArtistNavigationFromPlayer(artistId: Long) {
        if (artistId == 0L || artistId == -1L) return

        val existingJob = artistNavigationJob
        if (existingJob != null && existingJob.isActive) return

        artistNavigationJob?.cancel()
        artistNavigationJob = viewModelScope.launch {
            collapsePlayerSheet()

            withTimeoutOrNull(900) {
                awaitSheetState(PlayerSheetState.COLLAPSED)
                awaitPlayerCollapse()
            }

            _artistNavigationRequests.emit(artistId)
        }
    }

    suspend fun awaitSheetState(target: PlayerSheetState) {
        sheetState.first { it == target }
    }

    suspend fun awaitPlayerCollapse(threshold: Float = 0.1f, timeoutMillis: Long = 800L) {
        withTimeoutOrNull(timeoutMillis) {
            snapshotFlow { playerContentExpansionFraction.value }
                .first { it <= threshold }
        }
    }

    private fun resolveSongFromMediaItem(
        mediaItem: MediaItem,
        songsById: Map<String, Song>? = null
    ): Song? =
        songsById?.get(mediaItem.mediaId)
            ?: _playerUiState.value.currentPlaybackQueue.find { it.id == mediaItem.mediaId }
            ?: knownSongs[mediaItem.mediaId]

    /**
     * Songs this view model has queued, so that items keep resolving after the UI queue moves on
     * (the original falls back to the library). Cleared when a new queue starts.
     */
    private val knownSongs = mutableMapOf<String, Song>()

    private var lastQueueUpdateRequestId = 0L
    private var lastQueueSignature: QueueTimelineSignature? = null
    private var lastQueueUpdateJob: Job? = null

    private fun updateCurrentPlaybackQueueFromPlayer(playerCtrl: MediaController?) {
        val currentMediaController = playerCtrl ?: mediaController ?: return
        val requestId = ++lastQueueUpdateRequestId
        lastQueueUpdateJob?.cancel()
        lastQueueUpdateJob = viewModelScope.launch {
            delay(100)

            val isWindowed = dualPlayerEngine.isUsingWindowedQueue()
            val mediaItems = if (isWindowed) {
                dualPlayerEngine.getFullQueue()
            } else {
                val timeline = currentMediaController.currentTimeline
                val windowCount = timeline.windowCount
                val list = ArrayList<MediaItem>(windowCount)
                val window = Timeline.Window()
                for (i in 0 until windowCount) {
                    list.add(timeline.getWindow(i, window).mediaItem)
                    if (i % 500 == 0) yield()
                }
                list
            }

            val count = mediaItems.size
            if (count == 0) {
                if (requestId != lastQueueUpdateRequestId) return@launch
                val emptySignature = QueueTimelineSignature(
                    count = 0,
                    orderHash = 0L,
                    firstMediaId = null,
                    lastMediaId = null
                )
                if (lastQueueSignature != emptySignature) {
                    lastQueueSignature = emptySignature
                    _playerUiState.update { it.copy(currentPlaybackQueue = persistentListOf()) }
                }
                return@launch
            }

            var orderHash = 1125899906842597L
            var firstMediaId: String? = null
            var lastMediaId: String? = null

            for (i in 0 until count) {
                val mediaItem = mediaItems[i]
                val mediaId = mediaItem.mediaId
                if (i == 0) firstMediaId = mediaId
                if (i == count - 1) lastMediaId = mediaId
                orderHash = (orderHash * 31) + mediaId.hashCode()
                if (i % 500 == 0) yield()
            }

            val signature = QueueTimelineSignature(
                count = count,
                orderHash = orderHash,
                firstMediaId = firstMediaId,
                lastMediaId = lastMediaId
            )
            if (requestId != lastQueueUpdateRequestId) return@launch
            if (signature == lastQueueSignature) return@launch

            val songsById = knownSongs.toMap()

            val queue = withContext(dispatchers.default) {
                mediaItems.mapNotNull { mediaItem ->
                    resolveSongFromMediaItem(mediaItem, songsById)
                }
            }

            if (requestId != lastQueueUpdateRequestId) return@launch

            lastQueueSignature = signature
            _playerUiState.update { it.copy(currentPlaybackQueue = queue.toPlaybackQueue()) }
            if (queue.isNotEmpty()) {
                _isSheetVisible.value = true
            }
        }
    }

    private fun resetPlaybackAudioMetadata() {
        _playbackAudioMetadata.value = PlaybackAudioMetadata()
    }

    private fun preparePlaybackAudioMetadataForMedia(mediaId: String?) {
        // Queue-only changes (shuffle/unshuffle, reordering, repeat toggles) re-announce
        // the same media id while the same track keeps playing. No tracks-changed event
        // follows, so wiping here would blank the player file info until the next real
        // transition. Keep the already-resolved format info instead.
        if (mediaId != null && _playbackAudioMetadata.value.mediaId == mediaId) return
        _playbackAudioMetadata.value = PlaybackAudioMetadata(mediaId = mediaId)
    }

    private fun extractBitDepthFromPcmEncoding(pcmEncoding: Int): Int? {
        return when (pcmEncoding) {
            C.ENCODING_PCM_8BIT -> 8
            C.ENCODING_PCM_16BIT -> 16
            C.ENCODING_PCM_24BIT -> 24
            C.ENCODING_PCM_32BIT -> 32
            C.ENCODING_PCM_FLOAT -> 32
            else -> null
        }
    }

    /**
     * Sonora: the original also probes missing fields with MediaMetadataRetriever on the local
     * file; Sonora audio is a network stream, so only the selected track's format is read.
     */
    private fun refreshPlaybackAudioMetadata(player: Player, tracks: Tracks = player.currentTracks) {
        runCatching {
            val mediaId = player.currentMediaItem?.mediaId
            if (mediaId == null) {
                resetPlaybackAudioMetadata()
                return@runCatching
            }

            val selectedAudioFormat = tracks.groups
                .asSequence()
                .filter { it.type == C.TRACK_TYPE_AUDIO }
                .flatMap { group ->
                    (0 until group.length)
                        .asSequence()
                        .filter { index -> group.isTrackSelected(index) }
                        .map { index -> group.getTrackFormat(index) }
                }
                .firstOrNull()

            val current = _playbackAudioMetadata.value.takeIf { it.mediaId == mediaId }
            _playbackAudioMetadata.value = PlaybackAudioMetadata(
                mediaId = mediaId,
                mimeType = selectedAudioFormat?.sampleMimeType
                    ?: selectedAudioFormat?.containerMimeType
                    ?: current?.mimeType,
                bitrate = selectedAudioFormat?.bitrate?.takeIf { it > 0 }
                    ?: current?.bitrate,
                sampleRate = selectedAudioFormat?.sampleRate?.takeIf { it > 0 }
                    ?: current?.sampleRate,
                channelCount = selectedAudioFormat?.channelCount?.takeIf { it > 0 } ?: current?.channelCount,
                bitDepth = selectedAudioFormat?.pcmEncoding?.let(::extractBitDepthFromPcmEncoding) ?: current?.bitDepth
            )
        }.onFailure { throwable ->
            Timber.w(throwable, "Failed to refresh playback audio metadata")
        }
    }

    private fun syncPlaybackPositionFromPlayer(
        mediaId: String?,
        reportedPositionMs: Long
    ): Long {
        playbackStateHolder.syncCurrentPositionFromPlayer(mediaId, reportedPositionMs)
        return playbackStateHolder.currentPosition.value
    }

    private fun extractColorScheme(song: Song) {
        viewModelScope.launch {
            val uri = song.albumArtUriString?.toUri()
            val currentUri = playbackStateHolder.stablePlayerState.value.currentSong?.albumArtUriString
            themeStateHolder.extractAndGenerateColorScheme(uri, currentUri)
        }
    }

    private fun syncDisplayedMediaItemIfChanged(player: Player) {
        val mediaItem = player.currentMediaItem ?: return
        val currentSongId = playbackStateHolder.stablePlayerState.value.currentSong?.id
        val currentIndex = playbackStateHolder.stablePlayerState.value.currentMediaItemIndex
        val expectedIndex = if (dualPlayerEngine.isUsingWindowedQueue()) {
            dualPlayerEngine.getCurrentAbsoluteIndex()
        } else {
            player.currentMediaItemIndex
        }
        if (currentSongId == mediaItem.mediaId && currentIndex == expectedIndex) return

        playbackStateHolder.onPlaybackOccurrenceTransition(mediaItem.mediaId)
        preparePlaybackAudioMetadataForMedia(mediaItem.mediaId)
        transitionSchedulerJob?.cancel()

        val song = resolveSongFromMediaItem(mediaItem)
        val currentPosition = player.currentPosition.coerceAtLeast(0L)
        val resolvedDuration = if (song != null) {
            playbackStateHolder.resolveDurationForPlaybackState(
                reportedDurationMs = player.duration,
                songDurationHintMs = song.duration.coerceAtLeast(0L),
                currentPositionMs = currentPosition
            )
        } else {
            0L
        }

        playbackStateHolder.updateStablePlayerState {
            it.copy(
                currentSong = song,
                currentMediaItemIndex = expectedIndex,
                totalDuration = resolvedDuration,
                lyrics = null,
                isLoadingLyrics = false,
                isPlaying = player.isPlaying,
                playWhenReady = player.playWhenReady
            )
        }
        syncPlaybackPositionFromPlayer(mediaItem.mediaId, currentPosition)

        song?.let(::extractColorScheme)
    }

    private fun setupMediaControllerListeners() {
        val playerCtrl = mediaController ?: return
        playbackStateHolder.updateStablePlayerState {
            it.copy(
                isShuffleEnabled = it.isShuffleEnabled,
                repeatMode = playerCtrl.repeatMode,
                isPlaying = playerCtrl.isPlaying,
                playWhenReady = playerCtrl.playWhenReady
            )
        }
        preparePlaybackAudioMetadataForMedia(playerCtrl.currentMediaItem?.mediaId)
        refreshPlaybackAudioMetadata(playerCtrl)

        updateCurrentPlaybackQueueFromPlayer(playerCtrl)

        playerCtrl.currentMediaItem?.let { mediaItem ->
            playbackStateHolder.ensureCurrentPlaybackOccurrence(mediaItem.mediaId)
            val song = resolveSongFromMediaItem(mediaItem)

            if (song != null) {
                val initialPosition = playerCtrl.currentPosition.coerceAtLeast(0L)
                val resolvedDuration = playbackStateHolder.resolveDurationForPlaybackState(
                    reportedDurationMs = playerCtrl.duration,
                    songDurationHintMs = song.duration.coerceAtLeast(0L),
                    currentPositionMs = initialPosition
                )
                playbackStateHolder.updateStablePlayerState {
                    it.copy(
                        currentSong = song,
                        totalDuration = resolvedDuration
                    )
                }
                syncPlaybackPositionFromPlayer(mediaItem.mediaId, initialPosition)
                extractColorScheme(song)
                if (playerCtrl.isPlaying) {
                    _isSheetVisible.value = true
                    startProgressUpdates()
                }
            } else {
                playbackStateHolder.updateStablePlayerState {
                    it.copy(
                        currentSong = null,
                        isPlaying = false,
                        playWhenReady = false
                    )
                }
                playbackStateHolder.clearCurrentPositionHints()
                playbackStateHolder.setCurrentPosition(0L)
                resetPlaybackAudioMetadata()
            }
        }

        mediaControllerPlaybackListener?.let(playerCtrl::removeListener)
        mediaControllerPlaybackListener = object : Player.Listener {
            override fun onPlayerError(error: PlaybackException) {
                viewModelScope.launch {
                    _toastEvents.emit(context.getString(R.string.player_error_skipping))
                }
            }

            override fun onIsPlayingChanged(isPlaying: Boolean) {
                playbackStateHolder.updateStablePlayerState {
                    it.copy(
                        isPlaying = isPlaying,
                        playWhenReady = playerCtrl.playWhenReady
                    )
                }
                val shouldKeepSampling = playerCtrl.playWhenReady &&
                    playerCtrl.playbackState != Player.STATE_IDLE &&
                    playerCtrl.playbackState != Player.STATE_ENDED
                if (isPlaying || shouldKeepSampling) {
                    _isSheetVisible.value = true
                    if (isPlaying) {
                        clearPreparingSongIfMatching(playerCtrl.currentMediaItem?.mediaId)
                    }
                    startProgressUpdates()
                } else {
                    stopProgressUpdates()
                    val pausedPosition = playerCtrl.currentPosition.coerceAtLeast(0L)
                    syncPlaybackPositionFromPlayer(playerCtrl.currentMediaItem?.mediaId, pausedPosition)
                }
            }

            override fun onPlayWhenReadyChanged(playWhenReady: Boolean, reason: Int) {
                playbackStateHolder.updateStablePlayerState { it.copy(playWhenReady = playWhenReady) }
                if (
                    playWhenReady &&
                    playerCtrl.playbackState != Player.STATE_IDLE &&
                    playerCtrl.playbackState != Player.STATE_ENDED
                ) {
                    startProgressUpdates()
                }
            }

            override fun onMediaItemTransition(mediaItem: MediaItem?, reason: Int) {
                onFmTransition(mediaItem, reason, playerCtrl)
                playbackStateHolder.onPlaybackOccurrenceTransition(mediaItem?.mediaId)
                preparePlaybackAudioMetadataForMedia(mediaItem?.mediaId)
                transitionSchedulerJob?.cancel()
                transitionSchedulerJob = viewModelScope.launch {
                    mediaItem?.let { transitionedItem ->
                        val song = resolveSongFromMediaItem(transitionedItem)

                        val resolvedDuration = if (song != null) {
                            playbackStateHolder.resolveDurationForPlaybackState(
                                reportedDurationMs = playerCtrl.duration,
                                songDurationHintMs = song.duration.coerceAtLeast(0L),
                                currentPositionMs = playerCtrl.currentPosition.coerceAtLeast(0L)
                            )
                        } else {
                            0L
                        }
                        playbackStateHolder.updateStablePlayerState {
                            it.copy(
                                currentSong = song,
                                currentMediaItemIndex = if (dualPlayerEngine.isUsingWindowedQueue()) {
                                    dualPlayerEngine.getCurrentAbsoluteIndex()
                                } else {
                                    playerCtrl.currentMediaItemIndex
                                },
                                totalDuration = resolvedDuration,
                                lyrics = null,
                                isLoadingLyrics = false,
                                playWhenReady = playerCtrl.playWhenReady
                            )
                        }
                        syncPlaybackPositionFromPlayer(
                            transitionedItem.mediaId,
                            playerCtrl.currentPosition.coerceAtLeast(0L)
                        )

                        song?.let(::extractColorScheme)
                    } ?: run {
                        playbackStateHolder.updateStablePlayerState {
                            it.copy(
                                currentSong = null,
                                isPlaying = false,
                                playWhenReady = false,
                                lyrics = null,
                                isLoadingLyrics = false,
                                totalDuration = 0L
                            )
                        }
                        playbackStateHolder.clearCurrentPositionHints()
                        resetPlaybackAudioMetadata()
                    }
                }
            }

            override fun onPlaybackStateChanged(playbackState: Int) {
                // Sonora: the radio's queue ran out before the next album arrived.
                if (playbackState == Player.STATE_ENDED && isFmItem(playerCtrl.currentMediaItem)) {
                    fmAdvancePastEnd()
                }
                refreshPlaybackAudioMetadata(playerCtrl)
                syncDisplayedMediaItemIfChanged(playerCtrl)

                bufferingDebounceJob?.cancel()
                if (playbackState == Player.STATE_BUFFERING) {
                    bufferingDebounceJob = viewModelScope.launch {
                        delay(500)
                        playbackStateHolder.updateStablePlayerState { state ->
                            state.copy(isBuffering = true)
                        }
                    }
                } else {
                    playbackStateHolder.updateStablePlayerState { state ->
                        state.copy(isBuffering = false)
                    }
                }

                if (playbackState == Player.STATE_READY) {
                    clearPreparingSongIfMatching(playerCtrl.currentMediaItem?.mediaId)
                    val readyPosition = playerCtrl.currentPosition.coerceAtLeast(0L)
                    val songDurationHint = playbackStateHolder.stablePlayerState.value.currentSong?.duration ?: 0L
                    val resolvedDuration = playbackStateHolder.resolveDurationForPlaybackState(
                        reportedDurationMs = playerCtrl.duration,
                        songDurationHintMs = songDurationHint,
                        currentPositionMs = readyPosition
                    )
                    syncPlaybackPositionFromPlayer(playerCtrl.currentMediaItem?.mediaId, readyPosition)
                    playbackStateHolder.updateStablePlayerState { it.copy(totalDuration = resolvedDuration) }
                    startProgressUpdates()
                }
                if (playbackState == Player.STATE_IDLE && playerCtrl.mediaItemCount == 0) {
                    clearPreparingSongIfMatching()
                    playbackStateHolder.updateStablePlayerState {
                        it.copy(
                            currentSong = null,
                            isPlaying = false,
                            playWhenReady = false,
                            lyrics = null,
                            isLoadingLyrics = false,
                            totalDuration = 0L
                        )
                    }
                    playbackStateHolder.clearCurrentPositionHints()
                    playbackStateHolder.setCurrentPosition(0L)
                    resetPlaybackAudioMetadata()
                }
            }
            override fun onTracksChanged(tracks: Tracks) {
                refreshPlaybackAudioMetadata(playerCtrl, tracks)
            }
            override fun onMediaMetadataChanged(mediaMetadata: MediaMetadata) {
                syncDisplayedMediaItemIfChanged(playerCtrl)
            }
            override fun onShuffleModeEnabledChanged(shuffleModeEnabled: Boolean) {
                if (shuffleModeEnabled) {
                    playerCtrl.shuffleModeEnabled = false
                    if (!playbackStateHolder.stablePlayerState.value.isShuffleEnabled) {
                        toggleShuffle()
                    }
                }
            }
            override fun onRepeatModeChanged(repeatMode: Int) {
                playbackStateHolder.updateStablePlayerState { it.copy(repeatMode = repeatMode) }
                viewModelScope.launch { userPreferencesRepository.setRepeatMode(repeatMode) }
            }
            override fun onTimelineChanged(timeline: Timeline, reason: Int) {
                syncDisplayedMediaItemIfChanged(playerCtrl)
                if (dualPlayerEngine.isTransitionRunning()) return

                transitionSchedulerJob?.cancel()

                if (reason == Player.TIMELINE_CHANGE_REASON_PLAYLIST_CHANGED ||
                    reason == Player.TIMELINE_CHANGE_REASON_SOURCE_UPDATE) {
                    updateCurrentPlaybackQueueFromPlayer(mediaController)
                }
            }
        }
        playerCtrl.addListener(checkNotNull(mediaControllerPlaybackListener))
    }

    fun playSongs(songsToPlay: List<Song>, startSong: Song, queueName: String = "None", playlistId: String? = null) {
        val requestToken = beginDirectPlaybackRequest()
        directPlaybackJob = viewModelScope.launch {
            transitionSchedulerJob?.cancel()

            val validSongs = songsToPlay
            if (validSongs.isEmpty()) {
                _toastEvents.emit(context.getString(R.string.no_valid_songs))
                return@launch
            }

            val validStartSong =
                validSongs.firstOrNull { it.id == startSong.id } ?: validSongs.first()

            // Every entry is queued from playlistId, so shuffle off can put each tag back in place.
            queueStateHolder.setOriginalQueueOrder(validSongs, List(validSongs.size) { playlistId })

            val isPersistent = userPreferencesRepository.persistentShuffleEnabledFlow.first()
            throwIfDirectPlaybackRequestIsStale(requestToken)
            val isShuffleOn = playbackStateHolder.stablePlayerState.value.isShuffleEnabled

            if (!isPersistent) {
                playbackStateHolder.updateStablePlayerState { it.copy(isShuffleEnabled = false) }
            }

            val finalSongsToPlay = if (isPersistent && isShuffleOn) {
                withContext(dispatchers.default) {
                    QueueUtils.buildAnchoredShuffleQueueSuspending(
                        validSongs,
                        validSongs.indexOfFirst { it.id == validStartSong.id }.coerceAtLeast(0)
                    )
                }
            } else {
                validSongs
            }
            throwIfDirectPlaybackRequestIsStale(requestToken)

            internalPlaySongs(finalSongsToPlay, validStartSong, queueName, playlistId)
            if (requestToken == directPlaybackToken) {
                directPlaybackJob = null
            }
        }
    }

    private fun setPreparingSong(songId: String?) {
        _playerUiState.update { state ->
            if (state.preparingSongId == songId) state else state.copy(preparingSongId = songId)
        }
    }

    /** Sonora songs always stream, so the original's local-file check is gone. */
    private fun beginPreparingSong(song: Song) {
        setPreparingSong(song.id)
        viewModelScope.launch(dispatchers.io) {
            val albumArtUri = song.albumArtUriString
            if (albumArtUri.isNullOrBlank()) {
                themeStateHolder.extractAndGenerateColorScheme(
                    albumArtUriAsUri = null,
                    currentSongUriString = null,
                    isPreload = false
                )
            } else {
                themeStateHolder.extractAndGenerateColorScheme(
                    albumArtUriAsUri = albumArtUri.toUri(),
                    currentSongUriString = albumArtUri,
                    isPreload = false
                )
            }
        }
    }

    private fun clearPreparingSongIfMatching(mediaId: String? = null) {
        val preparingSongId = _playerUiState.value.preparingSongId ?: return
        if (mediaId == null || preparingSongId == mediaId) {
            setPreparingSong(null)
        }
    }

    private fun cancelPendingDirectPlaybackBuild() {
        directPlaybackToken += 1L
        directPlaybackJob?.cancel()
        directPlaybackJob = null
    }

    private fun beginDirectPlaybackRequest(): Long {
        directPlaybackToken += 1L
        directPlaybackJob?.cancel()
        directPlaybackJob = null
        pendingQueueSegmentsJob?.cancel()
        pendingQueueSegmentsJob = null
        return directPlaybackToken
    }

    private fun throwIfDirectPlaybackRequestIsStale(requestToken: Long) {
        if (requestToken != directPlaybackToken) {
            throw CancellationException("Stale direct playback request")
        }
    }

    private suspend fun preparePlaybackQueueSegments(
        songsToPlay: List<Song>,
        startSongId: String,
        playlistId: String?
    ): PreparedPlaybackQueueSegments = withContext(dispatchers.default) {
        val currentIndex = songsToPlay
            .indexOfFirst { it.id == startSongId }
            .takeIf { it >= 0 }
            ?: 0

        val beforeCurrent = List(currentIndex) { index ->
            buildPlaybackMediaItem(songsToPlay[index], playlistId)
        }
        val afterStartIndex = currentIndex + 1
        val afterCurrent = List((songsToPlay.size - afterStartIndex).coerceAtLeast(0)) { offset ->
            buildPlaybackMediaItem(songsToPlay[afterStartIndex + offset], playlistId)
        }

        PreparedPlaybackQueueSegments(
            beforeCurrent = beforeCurrent,
            afterCurrent = afterCurrent,
            currentIndex = currentIndex
        )
    }

    private suspend fun attachPreparedQueueSegmentsIfCurrent(
        player: Player,
        startSongId: String,
        preparedSegments: PreparedPlaybackQueueSegments
    ) {
        if (player.currentMediaItem?.mediaId != startSongId) return
        if (player.mediaItemCount != 1) return
        if (player.getMediaItemAt(0).mediaId != startSongId) return

        val batchSize = 200

        if (preparedSegments.beforeCurrent.isNotEmpty()) {
            var insertedCount = 0
            while (insertedCount < preparedSegments.beforeCurrent.size) {
                val end = (insertedCount + batchSize).coerceAtMost(preparedSegments.beforeCurrent.size)
                val batch = preparedSegments.beforeCurrent.subList(insertedCount, end)
                player.addMediaItems(insertedCount, batch)
                insertedCount = end
                yield()
            }
        }

        if (preparedSegments.afterCurrent.isNotEmpty()) {
            var insertedCount = 0
            while (insertedCount < preparedSegments.afterCurrent.size) {
                val end = (insertedCount + batchSize).coerceAtMost(preparedSegments.afterCurrent.size)
                val batch = preparedSegments.afterCurrent.subList(insertedCount, end)
                player.addMediaItems(preparedSegments.beforeCurrent.size + 1 + insertedCount, batch)
                insertedCount = end
                yield()
            }
        }

        playbackStateHolder.updateStablePlayerState {
            it.copy(currentMediaItemIndex = preparedSegments.currentIndex)
        }
    }

    private fun internalPlaySongs(songsToPlay: List<Song>, startSong: Song, queueName: String = "None", playlistId: String? = null) {
        if (songsToPlay.isEmpty()) {
            clearPreparingSongIfMatching()
            return
        }
        val effectiveStartSong = songsToPlay.firstOrNull { it.id == startSong.id } ?: songsToPlay.first()

        knownSongs.clear()
        songsToPlay.associateByTo(knownSongs) { it.id }

        beginPreparingSong(effectiveStartSong)
        _playerUiState.update {
            it.copy(
                currentPlaybackQueue = songsToPlay.toPlaybackQueue(),
                currentQueueSourceName = queueName
            )
        }
        playbackStateHolder.updateStablePlayerState {
            it.copy(
                currentSong = effectiveStartSong,
                currentMediaItemIndex = 0,
                isPlaying = true,
                playWhenReady = true,
                totalDuration = effectiveStartSong.duration.coerceAtLeast(0L)
            )
        }
        _isSheetVisible.value = true

        // Sonora: the original resolves cloud URIs here (DualPlayerEngine.resolveMediaItem); the
        // audio endpoint is played as is.
        val startMediaItem = buildPlaybackMediaItem(effectiveStartSong, playlistId)

        val playSongsAction = {
            dualPlayerEngine.cancelNext()
            val enginePlayer = dualPlayerEngine.masterPlayer

            enginePlayer.setMediaItem(startMediaItem, 0L)
            enginePlayer.prepare()
            enginePlayer.play()
            _playerUiState.update { it.copy(isLoadingInitialSongs = false) }

            if (songsToPlay.size > 1) {
                pendingQueueSegmentsJob?.cancel()
                pendingQueueSegmentsJob = viewModelScope.launch {
                    val preparedSegments = preparePlaybackQueueSegments(
                        songsToPlay = songsToPlay,
                        startSongId = effectiveStartSong.id,
                        playlistId = playlistId
                    )
                    withContext(Dispatchers.Main.immediate) {
                        attachPreparedQueueSegmentsIfCurrent(
                            player = dualPlayerEngine.masterPlayer,
                            startSongId = effectiveStartSong.id,
                            preparedSegments = preparedSegments
                        )
                    }
                }
            }
        }

        if (mediaController == null) {
            Timber.w("MediaController not available. Queuing playback action.")
            pendingPlaybackAction = playSongsAction
        } else {
            playSongsAction()
        }
    }

    private fun loadAndPlaySong(song: Song) {
        beginPreparingSong(song)
        playbackStateHolder.updateStablePlayerState {
            it.copy(
                currentSong = song,
                isPlaying = true,
                playWhenReady = true
            )
        }
        _isSheetVisible.value = true

        val controller = mediaController
        if (controller == null) {
            pendingPlaybackAction = {
                loadAndPlaySong(song)
            }
            return
        }

        knownSongs[song.id] = song
        val mediaItem = buildPlaybackMediaItem(song)
        if (controller.currentMediaItem?.mediaId == song.id) {
            if (!controller.isPlaying) controller.play()
        } else {
            controller.setMediaItem(mediaItem)
            controller.prepare()
            controller.play()
        }
    }

    fun toggleShuffle(currentSongOverride: Song? = null) {
        val currentQueue = _playerUiState.value.currentPlaybackQueue.toList()
        val currentSong = currentSongOverride
            ?: playbackStateHolder.stablePlayerState.value.currentSong
            ?: mediaController?.currentMediaItem?.let { resolveSongFromMediaItem(it) }
            ?: currentQueue.firstOrNull()

        playbackStateHolder.toggleShuffle(
            currentSongs = currentQueue,
            currentSong = currentSong,
            updateQueueCallback = { newQueue ->
                _playerUiState.update { it.copy(currentPlaybackQueue = newQueue.toPlaybackQueue()) }
            }
        )
    }

    // --- Roaming FM --------------------------------------------------------------------------
    //
    // Sonora: not in PixelPlayerOSS. As the Web player (PlayerController's 漫游 FM part, Eric
    // 2026-11-07): the server lines up whole albums, their songs go into the ordinary queue
    // tagged with FM_QUEUE_TAG, and the progress is kept on the server. See data/fm/RoamingFm.kt.

    /** FM requests reach the server in the order they are made: a finish before the next position. */
    private val fmMutex = Mutex()
    private var fmLastItem: MediaItem? = null
    private var fmLastIndex = C.INDEX_UNSET
    private val _isRoamingFmPlaying = MutableStateFlow(false)
    /** Whether the current song is the radio's (the home screen highlights its button). */
    val isRoamingFmPlaying: StateFlow<Boolean> = _isRoamingFmPlaying.asStateFlow()

    /** Plays the radio from the position the server keeps (the home screen's radio button). */
    fun playRoamingFm() {
        cancelPendingFullQueuePlayback()
        val requestToken = beginDirectPlaybackRequest()
        directPlaybackJob = viewModelScope.launch {
            val server = sessions.current()?.server ?: return@launch
            val state = try {
                fmMutex.withLock { api.fm() }
            } catch (e: ApiException) {
                Timber.tag(FM_TAG).w(e, "Loading the radio failed")
                sendToast(context.getString(R.string.fm_load_failed))
                return@launch
            }
            throwIfDirectPlaybackRequestIsStale(requestToken)
            val songs = state.songs(server)
            val start = state.startSong(songs) ?: run {
                sendToast(context.getString(R.string.fm_no_albums))
                return@launch
            }
            // The server decides the order: no shuffle, and no repeat (the Web player has neither
            // button in FM; here the buttons stay, so the modes are switched off).
            playbackStateHolder.updateStablePlayerState { it.copy(isShuffleEnabled = false) }
            mediaController?.repeatMode = Player.REPEAT_MODE_OFF
            queueStateHolder.setOriginalQueueOrder(songs, List(songs.size) { FM_QUEUE_TAG })
            internalPlaySongs(songs, start, context.getString(R.string.fm_queue_name), FM_QUEUE_TAG)
            if (requestToken == directPlaybackToken) {
                directPlaybackJob = null
            }
        }
    }

    private fun isFmItem(item: MediaItem?): Boolean =
        item != null && MediaItemBuilder.playlistIdOf(item) == FM_QUEUE_TAG

    /** The albums of the radio entries after [index], in queue order. */
    private fun fmAlbumsAfter(player: Player, index: Int): List<Long> =
        (index + 1 until player.mediaItemCount)
            .map { player.getMediaItemAt(it) }
            .filter(::isFmItem)
            .mapNotNull { knownSongs[it.mediaId]?.albumId }

    /**
     * On every song change: leaving an album's last song for the next song counts the album as
     * heard and appends what the server lines up next; a radio song saves the position (Web:
     * fmBeforeNext and playSong). A song of a later album passes the ones before it over, which
     * the server does by itself.
     */
    private fun onFmTransition(item: MediaItem?, reason: Int, player: Player) {
        val previous = fmLastItem
        val previousIndex = fmLastIndex
        fmLastItem = item
        fmLastIndex = player.currentMediaItemIndex
        _isRoamingFmPlaying.value = isFmItem(item)
        if (item == null) return
        // A queue change reports the same entry again (appending the next album does).
        if (previous?.mediaId == item.mediaId && previousIndex == player.currentMediaItemIndex) return
        val previousAlbum = previous?.takeIf(::isFmItem)?.let { knownSongs[it.mediaId]?.albumId }
        val isFm = isFmItem(item)
        if (previousAlbum == null && !isFm) return

        val index = player.currentMediaItemIndex
        val moving = reason == Player.MEDIA_ITEM_TRANSITION_REASON_AUTO ||
            reason == Player.MEDIA_ITEM_TRANSITION_REASON_SEEK
        val forward = moving && index == previousIndex + 1
        // The entry being left was at index - 1 when moving forward.
        val ended = forward && previousAlbum != null &&
            albumEnded(previousAlbum, fmAlbumsAfter(player, index - 1))
        // Starting the radio queues its songs in the background; only a move within the queue
        // can tell that nothing follows the current album.
        val currentAlbum = knownSongs[item.mediaId]?.albumId
        val needsMore = isFm && moving && fmAlbumsAfter(player, index).all { it == currentAlbum }
        val trackId = item.mediaId.toLongOrNull()?.takeIf { isFm }
        if (!ended && !needsMore && trackId == null) return

        viewModelScope.launch {
            fmMutex.withLock {
                try {
                    val state = when {
                        ended -> api.fmFinish(previousAlbum!!)
                        needsMore -> api.fm()
                        else -> null
                    }
                    if (trackId != null) api.fmPosition(trackId)
                    state?.let { fmAppend(it, finished = previousAlbum.takeIf { ended }) }
                    Unit
                } catch (e: ApiException) {
                    Timber.tag(FM_TAG).w(e, "Radio request failed")
                }
            }
        }
    }

    /**
     * The current radio song is the queue's last: count its album as heard if this is its last
     * song, wait for what the server lines up next and move on to it (Web: fmBeforeNext awaits the
     * next album when the queue has nothing after the current entry). Normally the next album is
     * queued long before; this happens when the server could not line one up in time, or the
     * library has a single album the radio can play.
     */
    private fun fmAdvancePastEnd() {
        val player = mediaController ?: return
        val current = player.currentMediaItem ?: return
        val albumId = knownSongs[current.mediaId]?.albumId
        viewModelScope.launch {
            val appended = fmMutex.withLock {
                try {
                    val state = if (albumId != null) api.fmFinish(albumId) else api.fm()
                    fmAppend(state, finished = albumId)
                } catch (e: ApiException) {
                    Timber.tag(FM_TAG).w(e, "Radio request failed")
                    0
                }
            }
            // The controller's timeline catches up with addMediaItems later; its commands run in
            // order, so the seek lands on the first appended song.
            if (appended > 0) {
                player.seekTo(player.currentMediaItemIndex + 1, 0L)
                player.play()
            } else {
                sendToast(context.getString(R.string.fm_no_next_album))
            }
        }
    }

    /**
     * Appends the albums the server lined up that the queue does not have yet (Web: fmAppend).
     * [finished] is the album just heard: a small library can line it up again at once.
     */
    /** Returns how many songs were appended. */
    private suspend fun fmAppend(state: FmState, finished: Long?): Int {
        val player = mediaController ?: return 0
        val server = sessions.current()?.server ?: return 0
        val current = player.currentMediaItem
        if (!isFmItem(current) && fmAlbumsAfter(player, player.currentMediaItemIndex).isEmpty()) return 0
        val queued = fmAlbumsAfter(player, player.currentMediaItemIndex).toMutableSet()
        current?.let { knownSongs[it.mediaId]?.albumId }?.takeIf { it != finished }?.let(queued::add)
        val songs = state.songsToAppend(server, queued)
        if (songs.isEmpty()) return 0
        songs.forEach { knownSongs[it.id] = it }
        player.addMediaItems(songs.map { buildPlaybackMediaItem(it, FM_QUEUE_TAG) })
        Timber.tag(FM_TAG).d("Appended %d songs of albums %s", songs.size, songs.map { it.albumId }.distinct())
        return songs.size
    }

    fun cycleRepeatMode() {
        playbackStateHolder.cycleRepeatMode()
    }

    /** Sonora: likes come with step 5 (`likes`, `likes/ids`); until then this does nothing. */
    fun toggleFavorite() {
    }

    /** Sonora: likes come with step 5 (`likes`, `likes/ids`); until then this does nothing. */
    fun toggleFavoriteSpecificSong(song: Song, removing: Boolean = false) {
    }

    fun addSongToQueue(song: Song) {
        mediaController?.let { controller ->
            knownSongs[song.id] = song
            controller.addMediaItem(buildPlaybackMediaItem(song))
        }
    }

    fun addSongNextToQueue(song: Song) {
        mediaController?.let { controller ->
            knownSongs[song.id] = song
            val mediaItem = buildPlaybackMediaItem(song)

            val insertionIndex = if (controller.currentMediaItemIndex != C.INDEX_UNSET) {
                (controller.currentMediaItemIndex + 1).coerceAtMost(controller.mediaItemCount)
            } else {
                controller.mediaItemCount
            }

            controller.addMediaItem(insertionIndex, mediaItem)
        }
    }

    private fun buildPlaybackMediaItem(song: Song, playlistId: String? = null): MediaItem =
        MediaItemBuilder.withPlaylistId(MediaItemBuilder.build(song), playlistId)

    /**
     * Signal from the player sheet whether the slider-bearing UI is currently
     * rendered. Drives the position-ticker's resolution (250 ms vs 1 s).
     */
    fun setSliderUiMounted(mounted: Boolean) {
        playbackStateHolder.setSliderUiMounted(mounted)
    }

    fun playPause() {
        val controller = mediaController
        if (controller == null || !controller.isConnected) {
            playbackStateHolder.playPause()
            return
        }

        if (controller.isPlaying) {
            controller.pause()
        } else {
            if (controller.currentMediaItem == null) {
                val currentQueue = _playerUiState.value.currentPlaybackQueue
                val currentSong = playbackStateHolder.stablePlayerState.value.currentSong
                when {
                    currentQueue.isNotEmpty() && currentSong != null -> {
                        transitionSchedulerJob?.cancel()
                        internalPlaySongs(
                            currentQueue.toList(),
                            currentSong,
                            _playerUiState.value.currentQueueSourceName
                        )
                    }
                    currentSong != null -> loadAndPlaySong(currentSong)
                    else -> controller.play()
                }
            } else {
                if (controller.playbackState == Player.STATE_IDLE && controller.mediaItemCount > 0) {
                    controller.prepare()
                }
                controller.play()
            }
        }
    }

    fun seekTo(position: Long) {
        playbackStateHolder.seekTo(position)
    }

    fun nextSong() {
        // Sonora: at the end of a radio queue, wait for the next album (Roaming FM below).
        val controller = mediaController
        if (controller != null && isFmItem(controller.currentMediaItem) && !controller.hasNextMediaItem()) {
            fmAdvancePastEnd()
            return
        }
        playbackStateHolder.nextSong()
    }

    fun previousSong() {
        playbackStateHolder.previousSong()
    }

    /**
     * Whether [previousSong] at [positionMs] restarts the current track under the "Previous
     * button" preference, ignoring queue bounds. Read at event time, so it adds no UI state.
     */
    fun previousButtonRestartsAt(positionMs: Long): Boolean =
        playbackStateHolder.currentPreviousButtonBehavior().restartsAtPosition(positionMs)

    /** Steps back to the previous track even past the restart threshold of [previousSong]. */
    fun forcePreviousSong() {
        playbackStateHolder.forcePreviousSong()
    }

    private fun startProgressUpdates() {
        playbackStateHolder.startProgressUpdates()
    }

    private fun stopProgressUpdates() {
        playbackStateHolder.stopProgressUpdates()
    }

    fun clearQueueExceptCurrent() {
        mediaController?.let { controller ->
            val currentSongIndex = controller.currentMediaItemIndex
            if (currentSongIndex == C.INDEX_UNSET) return@let
            val indicesToRemove = (0 until controller.mediaItemCount)
                .filter { it != currentSongIndex }
                .sortedDescending()

            for (index in indicesToRemove) {
                controller.removeMediaItem(index)
            }
        }
    }

    override fun onCleared() {
        // Captured before super.onCleared() so the owner checks below compare against the
        // exact scope this instance bound the singleton holders to.
        val ownerScope = viewModelScope
        val controllerToRelease = mediaController
        mediaControllerPlaybackListener?.let { listener ->
            controllerToRelease?.removeListener(listener)
            mediaControllerPlaybackListener = null
        }
        playbackStateHolder.clearMediaController(controllerToRelease)
        controllerToRelease?.release()
        mediaController = null
        mediaControllerFuture.cancel(true)
        super.onCleared()
        // The holder moves progress updates to the PlayerViewModel left bound, or stops them when
        // none is; stopping them here would freeze the seek bar of the surviving instance.
        playbackStateHolder.onCleared(ownerScope)
        libraryStateHolder.onCleared(ownerScope)
        themeStateHolder.onCleared(ownerScope)
        sleepTimerStateHolder.onCleared(ownerScope)
        queueUndoStateHolder.onCleared()
        playlistDismissUndoStateHolder.onCleared()
    }

    fun setSleepTimer(durationMinutes: Int) {
        sleepTimerStateHolder.setSleepTimer(durationMinutes)
    }

    fun playCounted(count: Int) {
        sleepTimerStateHolder.playCounted(count)
    }

    fun cancelCountedPlay() {
        sleepTimerStateHolder.cancelCountedPlay()
    }

    fun setEndOfTrackTimer(enable: Boolean) {
        val currentSongId = stablePlayerState.value.currentSong?.id
        sleepTimerStateHolder.setEndOfTrackTimer(enable, currentSongId)
    }

    fun cancelSleepTimer(overrideToastMessage: String? = null, suppressDefaultToast: Boolean = false) {
        sleepTimerStateHolder.cancelSleepTimer(overrideToastMessage, suppressDefaultToast)
    }

    fun dismissPlaylistAndShowUndo() {
        setMiniPlayerDismissing(false)
        playlistDismissUndoStateHolder.dismissPlaylistAndShowUndo(
            scope = viewModelScope,
            currentSong = playbackStateHolder.stablePlayerState.value.currentSong,
            queue = _playerUiState.value.currentPlaybackQueue,
            queueName = _playerUiState.value.currentQueueSourceName,
            position = playbackStateHolder.currentPosition.value,
            getUiState = { _playerUiState.value },
            updateUiState = { mutation -> _playerUiState.update(mutation) },
            disconnectRemoteIfNeeded = {},
            clearPlayback = {
                mediaController?.stop()
                mediaController?.clearMediaItems()
            },
            clearStablePlaybackState = {
                playbackStateHolder.updateStablePlayerState {
                    it.copy(
                        currentSong = null,
                        isPlaying = false,
                        playWhenReady = false,
                        totalDuration = 0L
                    )
                }
            },
            setCurrentPosition = { playbackStateHolder.setCurrentPosition(it) },
            setSheetVisible = { _isSheetVisible.value = it }
        )
    }

    fun hideDismissUndoBar() {
        playlistDismissUndoStateHolder.hideDismissUndoBar { mutation ->
            _playerUiState.update(mutation)
        }
    }

    fun undoDismissPlaylist() {
        setMiniPlayerDismissing(false)
        playlistDismissUndoStateHolder.undoDismissPlaylist(
            scope = viewModelScope,
            getUiState = { _playerUiState.value },
            updateUiState = { mutation -> _playerUiState.update(mutation) },
            playSongs = { songs, startSong, queueName ->
                playSongs(songs, startSong, queueName)
            },
            seekTo = { position -> mediaController?.seekTo(position) },
            setSheetVisible = { _isSheetVisible.value = it },
            setSheetCollapsed = { _sheetState.value = PlayerSheetState.COLLAPSED },
            emitToast = { message -> _toastEvents.emit(message) }
        )
    }
}

private const val FM_TAG = "RoamingFm"
