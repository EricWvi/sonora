package top.onlyquant.sonora.data.database

import androidx.paging.PagingSource
import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import androidx.room.Transaction
import kotlinx.coroutines.flow.Flow

/*
 * Sonora: the library queries of the original's MusicDao (getSongsPaginated, getAlbumsPaginated,
 * getArtistsPaginated), without its directory and storage filters, which have no meaning for a
 * library that only comes from the server. The sort keys are the original's SortOption storage
 * keys. Orders follow the server's lists (docs/api.md, 排序) instead of the original's NOCASE
 * code-point order: names by sortKey; ties in the list's own order (songs newest first, albums and
 * artists by ID); "Date Added" by ID, which the server assigns in upload order and never reuses;
 * the default order is the server's default (songs newest first, albums by title).
 */
@Dao
interface MusicDao {

    @Query(
        """
        SELECT songs.* FROM songs
        LEFT JOIN artists AS ar ON ar.id = songs.artist_id
        LEFT JOIN albums AS al ON al.id = songs.album_id
        ORDER BY
            CASE WHEN :sortOrder = 'song_title_az' THEN songs.sort_key END ASC,
            CASE WHEN :sortOrder = 'song_title_za' THEN songs.sort_key END DESC,
            CASE WHEN :sortOrder = 'song_artist' THEN COALESCE(ar.sort_key, '') END ASC,
            CASE WHEN :sortOrder = 'song_artist_desc' THEN COALESCE(ar.sort_key, '') END DESC,
            CASE WHEN :sortOrder = 'song_album' THEN COALESCE(al.sort_key, '') END ASC,
            CASE WHEN :sortOrder = 'song_album_desc' THEN COALESCE(al.sort_key, '') END DESC,
            CASE WHEN :sortOrder = 'song_date_added_asc' THEN songs.id END ASC,
            CASE WHEN :sortOrder = 'song_duration' THEN songs.duration END DESC,
            CASE WHEN :sortOrder = 'song_duration_asc' THEN songs.duration END ASC,
            songs.id DESC
        """
    )
    fun getSongsPaginated(sortOrder: String): PagingSource<Int, SongEntity>

    /** Every song in [sortOrder], for playing or shuffling the whole tab. */
    @Query(
        """
        SELECT songs.* FROM songs
        LEFT JOIN artists AS ar ON ar.id = songs.artist_id
        LEFT JOIN albums AS al ON al.id = songs.album_id
        ORDER BY
            CASE WHEN :sortOrder = 'song_title_az' THEN songs.sort_key END ASC,
            CASE WHEN :sortOrder = 'song_title_za' THEN songs.sort_key END DESC,
            CASE WHEN :sortOrder = 'song_artist' THEN COALESCE(ar.sort_key, '') END ASC,
            CASE WHEN :sortOrder = 'song_artist_desc' THEN COALESCE(ar.sort_key, '') END DESC,
            CASE WHEN :sortOrder = 'song_album' THEN COALESCE(al.sort_key, '') END ASC,
            CASE WHEN :sortOrder = 'song_album_desc' THEN COALESCE(al.sort_key, '') END DESC,
            CASE WHEN :sortOrder = 'song_date_added_asc' THEN songs.id END ASC,
            CASE WHEN :sortOrder = 'song_duration' THEN songs.duration END DESC,
            CASE WHEN :sortOrder = 'song_duration_asc' THEN songs.duration END ASC,
            songs.id DESC
        """
    )
    suspend fun getAllSongs(sortOrder: String): List<SongEntity>

    @Query(
        """
        SELECT albums.* FROM albums
        LEFT JOIN artists AS ar ON ar.id = albums.artist_id
        WHERE :hideSingles = 0 OR albums.single = 0
        ORDER BY
            CASE WHEN :sortOrder = 'album_title_za' THEN albums.sort_key END DESC,
            CASE WHEN :sortOrder = 'album_artist' THEN COALESCE(ar.sort_key, '') END ASC,
            CASE WHEN :sortOrder = 'album_artist_desc' THEN COALESCE(ar.sort_key, '') END DESC,
            CASE WHEN :sortOrder = 'album_release_year' THEN albums.year END DESC,
            CASE WHEN :sortOrder = 'album_release_year_asc' THEN albums.year END ASC,
            CASE WHEN :sortOrder = 'album_date_added' THEN albums.id END DESC,
            CASE WHEN :sortOrder = 'album_size_asc' THEN albums.song_count END ASC,
            CASE WHEN :sortOrder = 'album_size_desc' THEN albums.song_count END DESC,
            albums.sort_key ASC,
            albums.id ASC
        """
    )
    fun getAlbumsPaginated(sortOrder: String, hideSingles: Boolean): PagingSource<Int, AlbumEntity>

    @Query(
        """
        SELECT * FROM artists
        ORDER BY
            CASE WHEN :sortOrder = 'artist_name_za' THEN sort_key END DESC,
            CASE WHEN :sortOrder = 'artist_num_songs_desc' THEN track_count END DESC,
            CASE WHEN :sortOrder = 'artist_num_songs_asc' THEN track_count END ASC,
            sort_key ASC,
            id ASC
        """
    )
    fun getArtistsPaginated(sortOrder: String): PagingSource<Int, ArtistEntity>

    /** The albums tab as a list, for selecting all and the random album (the original's getAlbums). */
    @Query(
        """
        SELECT albums.* FROM albums
        LEFT JOIN artists AS ar ON ar.id = albums.artist_id
        WHERE :hideSingles = 0 OR albums.single = 0
        ORDER BY
            CASE WHEN :sortOrder = 'album_title_za' THEN albums.sort_key END DESC,
            CASE WHEN :sortOrder = 'album_artist' THEN COALESCE(ar.sort_key, '') END ASC,
            CASE WHEN :sortOrder = 'album_artist_desc' THEN COALESCE(ar.sort_key, '') END DESC,
            CASE WHEN :sortOrder = 'album_release_year' THEN albums.year END DESC,
            CASE WHEN :sortOrder = 'album_release_year_asc' THEN albums.year END ASC,
            CASE WHEN :sortOrder = 'album_date_added' THEN albums.id END DESC,
            CASE WHEN :sortOrder = 'album_size_asc' THEN albums.song_count END ASC,
            CASE WHEN :sortOrder = 'album_size_desc' THEN albums.song_count END DESC,
            albums.sort_key ASC,
            albums.id ASC
        """
    )
    fun getAlbums(sortOrder: String, hideSingles: Boolean): Flow<List<AlbumEntity>>

    /** The artists tab as a list (the original's getArtists). */
    @Query(
        """
        SELECT * FROM artists
        ORDER BY
            CASE WHEN :sortOrder = 'artist_name_za' THEN sort_key END DESC,
            CASE WHEN :sortOrder = 'artist_num_songs_desc' THEN track_count END DESC,
            CASE WHEN :sortOrder = 'artist_num_songs_asc' THEN track_count END ASC,
            sort_key ASC,
            id ASC
        """
    )
    fun getArtists(sortOrder: String): Flow<List<ArtistEntity>>

    /** As the original: disc, then numbered tracks, then title (here by sortKey). */
    @Query(
        """
        SELECT * FROM songs
        WHERE album_id = :albumId
        ORDER BY
            CASE WHEN disc_number <= 0 THEN 1 ELSE disc_number END ASC,
            CASE WHEN track_number > 0 THEN 0 ELSE 1 END ASC,
            CASE WHEN track_number > 0 THEN track_number END ASC,
            sort_key ASC,
            id ASC
        """
    )
    fun getSongsByAlbumId(albumId: Long): Flow<List<SongEntity>>

    /** As the original, through the cross references; by title (here by sortKey). */
    @Query(
        """
        SELECT songs.* FROM songs
        INNER JOIN song_artist_cross_ref ON songs.id = song_artist_cross_ref.song_id
        WHERE song_artist_cross_ref.artist_id = :artistId
        ORDER BY songs.sort_key ASC, songs.id ASC
        """
    )
    fun getSongsForArtist(artistId: Long): Flow<List<SongEntity>>

    @Query("SELECT * FROM albums WHERE id = :id")
    fun getAlbumById(id: Long): Flow<AlbumEntity?>

    @Query("SELECT * FROM artists WHERE id = :id")
    fun getArtistById(id: Long): Flow<ArtistEntity?>

    @Query("SELECT * FROM songs ORDER BY RANDOM() LIMIT :limit")
    suspend fun getRandomSongs(limit: Int): List<SongEntity>

    @Query("SELECT COUNT(*) FROM songs")
    fun getSongCount(): Flow<Int>

    @Query("SELECT * FROM songs WHERE id IN (:ids)")
    suspend fun getSongsByIds(ids: List<Long>): List<SongEntity>

    // ------------------------
    // Sync (not in the original's MusicDao, whose sync rewrites whole sources)
    // ------------------------

    @Query("SELECT * FROM library_state WHERE id = 0")
    suspend fun getState(): LibraryStateEntity?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun setState(state: LibraryStateEntity)

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun upsertSongs(songs: List<SongEntity>)

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun upsertAlbums(albums: List<AlbumEntity>)

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun upsertArtists(artists: List<ArtistEntity>)

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertCrossRefs(refs: List<SongArtistCrossRef>)

    @Query("DELETE FROM song_artist_cross_ref WHERE song_id IN (:songIds)")
    suspend fun deleteCrossRefsForSongs(songIds: List<Long>)

    @Query("DELETE FROM songs WHERE id IN (:ids)")
    suspend fun deleteSongs(ids: List<Long>)

    @Query("DELETE FROM albums WHERE id IN (:ids)")
    suspend fun deleteAlbums(ids: List<Long>)

    @Query("DELETE FROM artists WHERE id IN (:ids)")
    suspend fun deleteArtists(ids: List<Long>)

    @Query("DELETE FROM songs")
    suspend fun clearSongs()

    @Query("DELETE FROM albums")
    suspend fun clearAlbums()

    @Query("DELETE FROM artists")
    suspend fun clearArtists()

    @Query("DELETE FROM song_artist_cross_ref")
    suspend fun clearCrossRefs()

    /** One page of `GET library/changes`, applied together with its cursor. */
    @Transaction
    suspend fun applyChanges(page: LibraryPage) {
        if (page.deletedSongs.isNotEmpty() || page.songs.isNotEmpty()) {
            val touched = page.deletedSongs + page.songs.map { it.id }
            touched.chunked(SQL_VARIABLE_CHUNK).forEach { deleteCrossRefsForSongs(it) }
        }
        page.deletedSongs.chunked(SQL_VARIABLE_CHUNK).forEach { deleteSongs(it) }
        page.deletedAlbums.chunked(SQL_VARIABLE_CHUNK).forEach { deleteAlbums(it) }
        page.deletedArtists.chunked(SQL_VARIABLE_CHUNK).forEach { deleteArtists(it) }
        if (page.artists.isNotEmpty()) upsertArtists(page.artists)
        if (page.albums.isNotEmpty()) upsertAlbums(page.albums)
        if (page.songs.isNotEmpty()) upsertSongs(page.songs)
        if (page.crossRefs.isNotEmpty()) insertCrossRefs(page.crossRefs)
        setState(page.state)
    }

    /** Forgets the library, as when the server asks for a reset or the server address changed. */
    @Transaction
    suspend fun clearLibrary(state: LibraryStateEntity) {
        clearCrossRefs()
        clearSongs()
        clearAlbums()
        clearArtists()
        setState(state)
    }

    private companion object {
        const val SQL_VARIABLE_CHUNK = 900
    }
}

/** A page of changes as rows, with the state to store once it is applied. */
data class LibraryPage(
    val songs: List<SongEntity>,
    val albums: List<AlbumEntity>,
    val artists: List<ArtistEntity>,
    val crossRefs: List<SongArtistCrossRef>,
    val deletedSongs: List<Long>,
    val deletedAlbums: List<Long>,
    val deletedArtists: List<Long>,
    val state: LibraryStateEntity,
)
