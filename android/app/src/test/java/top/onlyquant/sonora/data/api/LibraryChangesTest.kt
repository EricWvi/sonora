package top.onlyquant.sonora.data.api

import com.google.common.truth.Truth.assertThat
import kotlinx.serialization.json.Json
import org.junit.jupiter.api.Test
import top.onlyquant.sonora.data.database.SongArtistCrossRef
import top.onlyquant.sonora.data.model.ArtistRef

class LibraryChangesTest {
    private val json = Json { ignoreUnknownKeys = true }

    // Shape as returned by the dev server (docs/api.md#曲库变更).
    private val body = """
        {"cursor":602,"more":true,"reset":false,
         "tracks":[{"id":75,"title":"爱的四季","artists":[{"id":58,"name":"李健"},{"id":9,"name":"王菲"}],
           "album":{"id":50,"title":"音乐傲骨","artists":[{"id":58,"name":"李健"}]},
           "trackNo":3,"discNo":0,"year":2009,"durationMs":272560,"format":"mp3",
           "cover":{"name":"de21.webp","thumbhash":"Zg"},"hasLyrics":true,"onlineId":null,"sortKey":"9600"}],
         "albums":[{"id":86,"title":"三生三世","artists":[{"id":19,"name":"张杰"}],"year":2017,
           "cover":null,"description":"","trackCount":1,"durationMs":257634,"onlineId":null,
           "single":true,"sortKey":"9664"}],
         "artists":[{"id":1,"name":"221小伙伴","image":null,"description":"","trackCount":1,
           "albumCount":1,"onlineId":null,"sortKey":"14f0"}],
         "deleted":{"tracks":[3],"albums":[],"artists":[4,5]}}
    """

    @Test
    fun `a page of changes becomes rows with its cursor`() {
        val page = json.decodeFromString(LibraryChanges.serializer(), body).toPage("http://s")

        assertThat(page.state.cursor).isEqualTo(602)
        assertThat(page.state.server).isEqualTo("http://s")
        assertThat(page.deletedSongs).containsExactly(3L)
        assertThat(page.deletedArtists).containsExactly(4L, 5L)
        assertThat(page.crossRefs).containsExactly(
            SongArtistCrossRef(75, 58, isPrimary = true),
            SongArtistCrossRef(75, 9, isPrimary = false),
        ).inOrder()

        val song = page.songs.single()
        assertThat(song.artistName).isEqualTo("李健")
        assertThat(song.albumId).isEqualTo(50)
        assertThat(song.coverName).isEqualTo("de21.webp")
        assertThat(song.sortKey).isEqualTo("9600")
        assertThat(page.albums.single().single).isTrue()
        assertThat(page.albums.single().coverName).isNull()
        assertThat(page.artists.single().sortKey).isEqualTo("14f0")
    }

    @Test
    fun `a stored song reads back as the original's Song`() {
        val page = json.decodeFromString(LibraryChanges.serializer(), body).toPage("http://s")
        val song = page.songs.single().toSong("http://s")

        assertThat(song.id).isEqualTo("75")
        assertThat(song.artists).containsExactly(
            ArtistRef(58, "李健", isPrimary = true),
            ArtistRef(9, "王菲", isPrimary = false),
        ).inOrder()
        assertThat(song.contentUriString).isEqualTo("http://s/api/v1/tracks/75/audio")
        assertThat(song.albumArtUriString).isEqualTo("http://s/api/v1/covers/de21.webp?size=1024")
        assertThat(song.discNumber).isNull()
        assertThat(song.trackNumber).isEqualTo(3)
        assertThat(song.mimeType).isEqualTo("audio/mpeg")
    }

    @Test
    fun `a song without album or artists`() {
        val track = ChangedTrack(id = 1, title = "x")
        val song = track.toEntity().toSong("http://s")
        assertThat(song.albumId).isEqualTo(-1)
        assertThat(song.artistId).isEqualTo(-1)
        assertThat(song.artists).isEmpty()
        assertThat(song.albumArtUriString).isNull()
        assertThat(song.gainDb).isNull()
    }

    @Test
    fun `the loudness gain survives the local copy`() {
        val changes = json.decodeFromString(
            LibraryChanges.serializer(),
            """{"cursor":1,"tracks":[{"id":7,"title":"Red Flag","gainDb":-0.58}]}""",
        )
        val entity = changes.toPage("http://s").songs.single()
        assertThat(entity.gainDb).isEqualTo(-0.58)
        assertThat(entity.toSong("http://s").gainDb).isEqualTo(-0.58)
    }

    @Test
    fun `a track from the player API keeps its loudness gain`() {
        val track = json.decodeFromString(
            Track.serializer(),
            """{"id":7,"title":"Three Laps Is A Lifetime","gainDb":-3.04}""",
        )
        assertThat(track.toSong("http://s").gainDb).isEqualTo(-3.04)
        val unmeasured = json.decodeFromString(Track.serializer(), """{"id":8,"title":"x","gainDb":null}""")
        assertThat(unmeasured.toSong("http://s").gainDb).isNull()
    }
}
