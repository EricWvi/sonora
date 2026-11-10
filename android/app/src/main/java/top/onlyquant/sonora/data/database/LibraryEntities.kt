package top.onlyquant.sonora.data.database

import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.Index
import androidx.room.PrimaryKey
import kotlinx.serialization.Serializable
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.json.Json
import top.onlyquant.sonora.data.model.ArtistRef

/*
 * Sonora: the original's SongEntity, AlbumEntity, ArtistEntity and SongArtistCrossRef describe
 * local files and Navidrome/Jellyfin copies (paths, MediaStore dates, user edits, source types).
 * Here they hold the Sonora library as `GET library/changes` sends it (docs/api.md#曲库变更); the
 * table and column names follow the original where the meaning is the same.
 *
 * There are no foreign keys: a page of changes can bring a song before its album or artist.
 * Covers are kept as their names on the server, since their URLs depend on the server address.
 * Names sort by the server's sortKey (pinyin and numeric order, as the Web player); an artist's
 * or album's sortKey lives only in its own table and is joined in when sorting songs by it.
 */

@Entity(
    tableName = "songs",
    indices = [
        Index(value = ["album_id"]),
        Index(value = ["artist_id"]),
        Index(value = ["sort_key"]),
        Index(value = ["duration"]),
    ]
)
data class SongEntity(
    @PrimaryKey val id: Long,
    @ColumnInfo(name = "title") val title: String,
    /** The first artist, as the original's primary artist. */
    @ColumnInfo(name = "artist_name") val artistName: String,
    @ColumnInfo(name = "artist_id") val artistId: Long,
    @ColumnInfo(name = "artists_json") val artistsJson: String,
    @ColumnInfo(name = "album_name") val albumName: String,
    @ColumnInfo(name = "album_id") val albumId: Long,
    @ColumnInfo(name = "album_artist") val albumArtist: String?,
    @ColumnInfo(name = "cover_name") val coverName: String?,
    @ColumnInfo(name = "duration") val duration: Long,
    @ColumnInfo(name = "track_number") val trackNumber: Int,
    @ColumnInfo(name = "disc_number") val discNumber: Int,
    @ColumnInfo(name = "year") val year: Int,
    @ColumnInfo(name = "format") val format: String,
    @ColumnInfo(name = "has_lyrics") val hasLyrics: Boolean,
    @ColumnInfo(name = "sort_key") val sortKey: String,
)

@Entity(
    tableName = "albums",
    indices = [Index(value = ["sort_key"]), Index(value = ["artist_id"])]
)
data class AlbumEntity(
    @PrimaryKey val id: Long,
    @ColumnInfo(name = "title") val title: String,
    @ColumnInfo(name = "artist_name") val artistName: String,
    @ColumnInfo(name = "artist_id") val artistId: Long,
    @ColumnInfo(name = "artists_json") val artistsJson: String,
    @ColumnInfo(name = "year") val year: Int,
    @ColumnInfo(name = "cover_name") val coverName: String?,
    @ColumnInfo(name = "description") val description: String,
    @ColumnInfo(name = "song_count") val songCount: Int,
    @ColumnInfo(name = "duration") val duration: Long,
    /** Marked as a single on the server (`hideSingles` and `onlySingles`). */
    @ColumnInfo(name = "single") val single: Boolean,
    @ColumnInfo(name = "sort_key") val sortKey: String,
)

@Entity(tableName = "artists", indices = [Index(value = ["sort_key"])])
data class ArtistEntity(
    @PrimaryKey val id: Long,
    @ColumnInfo(name = "name") val name: String,
    @ColumnInfo(name = "image_name") val imageName: String?,
    @ColumnInfo(name = "description") val description: String,
    @ColumnInfo(name = "track_count") val trackCount: Int,
    @ColumnInfo(name = "album_count") val albumCount: Int,
    @ColumnInfo(name = "sort_key") val sortKey: String,
)

/** As the original's junction table: every artist of a song, the first one primary. */
@Entity(
    tableName = "song_artist_cross_ref",
    primaryKeys = ["song_id", "artist_id"],
    indices = [Index(value = ["artist_id"])]
)
data class SongArtistCrossRef(
    @ColumnInfo(name = "song_id") val songId: Long,
    @ColumnInfo(name = "artist_id") val artistId: Long,
    @ColumnInfo(name = "is_primary") val isPrimary: Boolean,
)

/** Not in the original: where the last sync stopped, and for which server. One row. */
@Entity(tableName = "library_state")
data class LibraryStateEntity(
    @PrimaryKey val id: Int = 0,
    @ColumnInfo(name = "server") val server: String,
    @ColumnInfo(name = "cursor") val cursor: Long,
)

/*
 * The original's serializeArtistRefs and parseArtistsJson (SongEntity.kt) write the same JSON with
 * org.json; kotlinx.serialization here, so that the conversions run in unit tests, where
 * android.jar's org.json only returns defaults.
 */
@Serializable
private data class StoredArtistRef(val id: Long, val name: String, val primary: Boolean = false)

private val artistsJsonFormat = Json { ignoreUnknownKeys = true }

fun serializeArtistRefs(artists: List<ArtistRef>): String =
    artistsJsonFormat.encodeToString(
        ListSerializer(StoredArtistRef.serializer()),
        artists.map { StoredArtistRef(it.id, it.name, it.isPrimary) }
    )

fun parseArtistsJson(json: String?): List<ArtistRef> {
    if (json.isNullOrBlank()) return emptyList()
    return try {
        artistsJsonFormat.decodeFromString(ListSerializer(StoredArtistRef.serializer()), json)
            .map { ArtistRef(id = it.id, name = it.name, isPrimary = it.primary) }
    } catch (_: Exception) {
        emptyList()
    }
}
