package top.onlyquant.sonora.data.api

import kotlinx.serialization.Serializable
import top.onlyquant.sonora.data.database.AlbumEntity
import top.onlyquant.sonora.data.database.ArtistEntity
import top.onlyquant.sonora.data.database.LibraryPage
import top.onlyquant.sonora.data.database.LibraryStateEntity
import top.onlyquant.sonora.data.database.SongArtistCrossRef
import top.onlyquant.sonora.data.database.SongEntity
import top.onlyquant.sonora.data.database.parseArtistsJson
import top.onlyquant.sonora.data.database.serializeArtistRefs
import top.onlyquant.sonora.data.model.Album
import top.onlyquant.sonora.data.model.Artist
import top.onlyquant.sonora.data.model.ArtistRef
import top.onlyquant.sonora.data.model.Song

/** `GET library/changes` (docs/api.md#曲库变更). Only the fields the app keeps. */
@Serializable
data class LibraryChanges(
    val cursor: Long,
    val more: Boolean = false,
    val reset: Boolean = false,
    val tracks: List<ChangedTrack> = emptyList(),
    val albums: List<ChangedAlbum> = emptyList(),
    val artists: List<ChangedArtist> = emptyList(),
    val deleted: Deleted = Deleted(),
) {
    @Serializable
    data class Deleted(
        val tracks: List<Long> = emptyList(),
        val albums: List<Long> = emptyList(),
        val artists: List<Long> = emptyList(),
    )
}

@Serializable
data class ChangedTrack(
    val id: Long,
    val title: String,
    val artists: List<Ref> = emptyList(),
    val album: AlbumRef? = null,
    val trackNo: Int = 0,
    val discNo: Int = 0,
    val year: Int = 0,
    val durationMs: Long = 0,
    val format: String = "",
    val cover: Cover? = null,
    val hasLyrics: Boolean = false,
    val sortKey: String = "",
)

@Serializable
data class ChangedAlbum(
    val id: Long,
    val title: String,
    val artists: List<Ref> = emptyList(),
    val year: Int = 0,
    val cover: Cover? = null,
    val description: String = "",
    val trackCount: Int = 0,
    val durationMs: Long = 0,
    val single: Boolean = false,
    val sortKey: String = "",
)

@Serializable
data class ChangedArtist(
    val id: Long,
    val name: String,
    val image: Cover? = null,
    val description: String = "",
    val trackCount: Int = 0,
    val albumCount: Int = 0,
    val sortKey: String = "",
)

private fun List<Ref>.toArtistRefs(): List<ArtistRef> =
    mapIndexed { i, a -> ArtistRef(id = a.id, name = a.name, isPrimary = i == 0) }

/** The rows for one page of changes from [server], ending at the page's cursor. */
fun LibraryChanges.toPage(server: String): LibraryPage = LibraryPage(
    songs = tracks.map { it.toEntity() },
    albums = albums.map { it.toEntity() },
    artists = artists.map { it.toEntity() },
    crossRefs = tracks.flatMap { t ->
        t.artists.distinctBy { it.id }.mapIndexed { i, a -> SongArtistCrossRef(t.id, a.id, isPrimary = i == 0) }
    },
    deletedSongs = deleted.tracks,
    deletedAlbums = deleted.albums,
    deletedArtists = deleted.artists,
    state = LibraryStateEntity(server = server, cursor = cursor),
)

internal fun ChangedTrack.toEntity() = SongEntity(
    id = id,
    title = title,
    artistName = artists.firstOrNull()?.name.orEmpty(),
    artistId = artists.firstOrNull()?.id ?: -1L,
    artistsJson = serializeArtistRefs(artists.toArtistRefs()),
    albumName = album?.title.orEmpty(),
    albumId = album?.id ?: -1L,
    albumArtist = album?.artists?.takeIf { it.isNotEmpty() }?.joinToString(", ") { it.name },
    coverName = cover?.name,
    duration = durationMs,
    trackNumber = trackNo,
    discNumber = discNo,
    year = year,
    format = format,
    hasLyrics = hasLyrics,
    sortKey = sortKey,
)

internal fun ChangedAlbum.toEntity() = AlbumEntity(
    id = id,
    title = title,
    artistName = artists.joinToString(", ") { it.name },
    artistId = artists.firstOrNull()?.id ?: -1L,
    artistsJson = serializeArtistRefs(artists.toArtistRefs()),
    year = year,
    coverName = cover?.name,
    description = description,
    songCount = trackCount,
    duration = durationMs,
    single = single,
    sortKey = sortKey,
)

internal fun ChangedArtist.toEntity() = ArtistEntity(
    id = id,
    name = name,
    imageName = image?.name,
    description = description,
    trackCount = trackCount,
    albumCount = albumCount,
    sortKey = sortKey,
)

/** As [Track.toSong], from the local copy. */
fun SongEntity.toSong(server: String): Song = Song(
    id = id.toString(),
    title = title,
    artist = artistName,
    artistId = artistId,
    artists = parseArtistsJson(artistsJson),
    album = albumName,
    albumId = albumId,
    albumArtist = albumArtist,
    path = "",
    contentUriString = audioUrl(server, id),
    albumArtUriString = coverName?.let { coverUrl(server, Cover(it)) },
    duration = duration,
    trackNumber = trackNumber,
    discNumber = discNumber.takeIf { it > 0 },
    year = year,
    dateAdded = id,
    mimeType = formatMimeType(format),
    bitrate = null,
    sampleRate = null,
)

/**
 * The original's Album. `dateAdded` is the album ID, which the server assigns in upload order
 * (the player API does not expose upload times).
 */
fun AlbumEntity.toAlbum(server: String): Album = Album(
    id = id,
    title = title,
    artist = artistName,
    year = year,
    dateAdded = id,
    albumArtUriString = coverName?.let { coverUrl(server, Cover(it)) },
    songCount = songCount,
    albumArtist = artistName.takeIf { it.isNotBlank() },
)

fun ArtistEntity.toArtist(server: String): Artist = Artist(
    id = id,
    name = name,
    songCount = trackCount,
    imageUrl = imageName?.let { coverUrl(server, Cover(it)) },
)
