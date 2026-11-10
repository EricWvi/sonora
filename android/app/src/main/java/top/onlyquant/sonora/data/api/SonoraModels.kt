package top.onlyquant.sonora.data.api

import kotlinx.serialization.Serializable
import top.onlyquant.sonora.data.model.ArtistRef
import top.onlyquant.sonora.data.model.Song

/** Shapes of the player API, as in docs/api.md ("数据形态"). Only the fields the app reads. */
@Serializable
data class Ref(val id: Long, val name: String)

@Serializable
data class Cover(val name: String, val thumbhash: String = "")

@Serializable
data class AlbumRef(val id: Long, val title: String, val artists: List<Ref> = emptyList())

@Serializable
data class Track(
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
)

@Serializable
data class Page<T>(val items: List<T>, val total: Int)

/** Where the app reads a track's audio; the device token is added by [AuthHeaderInterceptor]. */
fun audioUrl(server: String, trackId: Long): String = "$server/api/v1/tracks/$trackId/audio"

/**
 * A cover at the size the full player shows (1024, as the Web player's large cover). Smaller
 * places decode it down; the 100 and 300 sizes come with the library screens.
 */
fun coverUrl(server: String, cover: Cover): String = "$server/api/v1/covers/${cover.name}?size=1024"

/**
 * The original's [Song] for a Sonora track. The original keys songs by MediaStore or server IDs;
 * here the ID is the track ID, the content URI is the audio endpoint, and there is no file path.
 */
fun Track.toSong(server: String): Song {
    val refs = artists.mapIndexed { i, a -> ArtistRef(id = a.id, name = a.name, isPrimary = i == 0) }
    return Song(
        id = id.toString(),
        title = title,
        artist = artists.firstOrNull()?.name.orEmpty(),
        artistId = artists.firstOrNull()?.id ?: -1L,
        artists = refs,
        album = album?.title.orEmpty(),
        albumId = album?.id ?: -1L,
        albumArtist = album?.artists?.takeIf { it.isNotEmpty() }?.joinToString(", ") { it.name },
        path = "",
        contentUriString = audioUrl(server, id),
        albumArtUriString = cover?.let { coverUrl(server, it) },
        duration = durationMs,
        trackNumber = trackNo,
        discNumber = discNo.takeIf { it > 0 },
        year = year,
        mimeType = formatMimeType(format),
        bitrate = null,
        sampleRate = null,
    )
}

/** MIME types of the formats Sonora ingests (docs/api.md, 415 `unsupported_format`). */
internal fun formatMimeType(format: String): String? = when (format.lowercase()) {
    "mp3" -> "audio/mpeg"
    "flac" -> "audio/flac"
    "aac", "m4a" -> "audio/mp4"
    "vorbis", "ogg" -> "audio/ogg"
    "opus" -> "audio/opus"
    else -> null
}
