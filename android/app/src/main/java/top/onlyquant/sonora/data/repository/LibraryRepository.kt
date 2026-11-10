package top.onlyquant.sonora.data.repository

import androidx.paging.Pager
import androidx.paging.PagingConfig
import androidx.paging.PagingData
import androidx.paging.map
import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.flow.emptyFlow
import kotlinx.coroutines.flow.flatMapLatest
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import timber.log.Timber
import top.onlyquant.sonora.data.api.SonoraApi
import top.onlyquant.sonora.data.api.toAlbum
import top.onlyquant.sonora.data.api.toArtist
import top.onlyquant.sonora.data.api.toPage
import top.onlyquant.sonora.data.api.toSong
import top.onlyquant.sonora.data.auth.SessionStore
import top.onlyquant.sonora.data.database.LibraryStateEntity
import top.onlyquant.sonora.data.database.MusicDao
import top.onlyquant.sonora.data.model.Album
import top.onlyquant.sonora.data.model.Artist
import top.onlyquant.sonora.data.model.Song

/**
 * The library on the device. Sonora: in the original, MusicRepository reads Room and the
 * Navidrome/Jellyfin SyncWorkers fill it from the servers. Here one repository does both: [sync]
 * brings the local copy up to date with `GET library/changes`, and the paging flows read it.
 */
@Singleton
class LibraryRepository @Inject constructor(
    private val api: SonoraApi,
    private val dao: MusicDao,
    private val sessions: SessionStore,
) {
    private val syncMutex = Mutex()
    private val _isSyncing = MutableStateFlow(false)
    val isSyncing: StateFlow<Boolean> = _isSyncing.asStateFlow()

    private val server: Flow<String?> = sessions.session.map { it?.server }.distinctUntilChanged()

    /**
     * Applies the changes since the last sync, page by page, each page with its cursor, so an
     * interrupted sync continues where it stopped. A different server address, or the server's
     * `reset`, starts over from an empty library. Concurrent calls wait for the running one and
     * then fetch what changed in between.
     */
    suspend fun sync() = syncMutex.withLock {
        val server = sessions.current()?.server ?: return@withLock
        _isSyncing.value = true
        try {
            var state = dao.getState()
            if (state == null || state.server != server) {
                state = LibraryStateEntity(server = server, cursor = 0)
                dao.clearLibrary(state)
            }
            var cursor = state.cursor
            var resets = 0
            while (true) {
                val changes = api.libraryChanges(since = cursor)
                if (changes.reset) {
                    // The server's log is older than our cursor (restored from a backup).
                    check(++resets <= 1) { "library/changes asked to reset twice" }
                    Timber.tag(TAG).w("Server asked for a reset at cursor %d", cursor)
                    cursor = 0
                    dao.clearLibrary(LibraryStateEntity(server = server, cursor = 0))
                    continue
                }
                dao.applyChanges(changes.toPage(server))
                cursor = changes.cursor
                if (!changes.more) break
            }
            Timber.tag(TAG).d("Library synced to cursor %d", cursor)
        } finally {
            _isSyncing.value = false
        }
    }

    @OptIn(ExperimentalCoroutinesApi::class)
    fun songsPaging(sortOrder: String): Flow<PagingData<Song>> = server.flatMapLatest { server ->
        if (server == null) emptyFlow()
        else Pager(PAGING_CONFIG) { dao.getSongsPaginated(sortOrder) }.flow
            .map { data -> data.map { it.toSong(server) } }
    }

    @OptIn(ExperimentalCoroutinesApi::class)
    fun albumsPaging(sortOrder: String, hideSingles: Boolean): Flow<PagingData<Album>> =
        server.flatMapLatest { server ->
            if (server == null) emptyFlow()
            else Pager(PAGING_CONFIG) { dao.getAlbumsPaginated(sortOrder, hideSingles) }.flow
                .map { data -> data.map { it.toAlbum(server) } }
        }

    @OptIn(ExperimentalCoroutinesApi::class)
    fun artistsPaging(sortOrder: String): Flow<PagingData<Artist>> = server.flatMapLatest { server ->
        if (server == null) emptyFlow()
        else Pager(PAGING_CONFIG) { dao.getArtistsPaginated(sortOrder) }.flow
            .map { data -> data.map { it.toArtist(server) } }
    }

    val songCount: Flow<Int> = dao.getSongCount()

    /** Every song in [sortOrder], for playing or shuffling the whole tab. */
    suspend fun allSongs(sortOrder: String): List<Song> {
        val server = sessions.current()?.server ?: return emptyList()
        return dao.getAllSongs(sortOrder).map { it.toSong(server) }
    }

    private companion object {
        const val TAG = "LibraryRepository"
        // The original's defaultLibraryPagingConfig (MusicRepositoryImpl).
        val PAGING_CONFIG = PagingConfig(pageSize = 50, enablePlaceholders = true, maxSize = 250)
    }
}
