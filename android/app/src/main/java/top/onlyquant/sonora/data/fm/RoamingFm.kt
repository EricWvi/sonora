package top.onlyquant.sonora.data.fm

import top.onlyquant.sonora.data.api.FmState
import top.onlyquant.sonora.data.api.toSong
import top.onlyquant.sonora.data.model.Song

/*
 * Roaming FM (漫游电台), as the Web player plays it (frontend/src/core/player/fm.ts and the FM part
 * of PlayerController): the server lines up whole albums (docs/api.md#漫游-fm), the songs go into
 * the ordinary queue tagged with [FM_QUEUE_TAG], and the progress lives on the server. Leaving an
 * album's last song for the next one counts the album as heard (`finish`); every song played saves
 * the position, and a song of a later album passes over the ones before it. Not in PixelPlayerOSS.
 */

/**
 * The queue tag (MediaItemBuilder's playlist ID) of the radio's songs. Songs the user inserts
 * ("play next") carry no tag, as in the Web player, where only the radio's entries have an album.
 */
const val FM_QUEUE_TAG = "sonora_fm"

/**
 * How many albums the home screen shows: its album art collage takes up to six songs, but the
 * pattern it uses (the original's default, COSMIC_SWIRL) has five places.
 */
const val FM_HOME_ALBUMS = 5

/** The songs of all the state's albums, in playing order. */
fun FmState.songs(server: String): List<Song> = albums.flatMap { album ->
    album.tracks.map { it.toSong(server).copy(albumId = album.id) }
}

/** The song to resume at: the server's position, or the current album's first song. */
fun FmState.startSong(songs: List<Song>): Song? =
    songs.firstOrNull { it.id == trackId?.toString() } ?: songs.firstOrNull()

/**
 * The songs to add after the queue: albums already queued after the current entry are left out,
 * and the current album starts at the song to resume at (Web: fmTracksToAppend).
 *
 * @param queuedAlbums the albums already in the queue after the current entry, and the current
 *     entry's album unless it was just finished (a small library can line it up again at once)
 */
fun FmState.songsToAppend(server: String, queuedAlbums: Set<Long>): List<Song> {
    val all = songs(server)
    val head = albums.firstOrNull()?.id
    val resumeAt = trackId?.toString()?.let { id -> all.indexOfFirst { it.id == id } } ?: -1
    return all.filterIndexed { index, song ->
        song.albumId !in queuedAlbums && (song.albumId != head || resumeAt < 0 || index >= resumeAt)
    }
}

/**
 * Whether moving on from a radio song of [albumId] leaves its album behind: no later entry of the
 * queue belongs to it (Web: fmBeforeNext).
 *
 * @param laterAlbums the albums of the radio entries after the one being left, in queue order
 */
fun albumEnded(albumId: Long, laterAlbums: List<Long>): Boolean = albumId !in laterAlbums
