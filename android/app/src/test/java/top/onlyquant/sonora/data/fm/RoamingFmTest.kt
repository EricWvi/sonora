package top.onlyquant.sonora.data.fm

import com.google.common.truth.Truth.assertThat
import org.junit.jupiter.api.Test
import top.onlyquant.sonora.data.api.FmAlbum
import top.onlyquant.sonora.data.api.FmState
import top.onlyquant.sonora.data.api.Track

class RoamingFmTest {
    private fun album(id: Long, vararg tracks: Long) = FmAlbum(id, tracks.map { Track(id = it, title = "t$it") })

    private val state = FmState(albums = listOf(album(1, 11, 12, 13), album(2, 21, 22)), trackId = 12)

    @Test
    fun `songs carry their album and play from the saved position`() {
        val songs = state.songs("http://s")
        assertThat(songs.map { it.id }).containsExactly("11", "12", "13", "21", "22").inOrder()
        assertThat(songs.map { it.albumId }).containsExactly(1L, 1L, 1L, 2L, 2L).inOrder()
        assertThat(state.startSong(songs)?.id).isEqualTo("12")
        assertThat(state.copy(trackId = null).startSong(songs)?.id).isEqualTo("11")
        assertThat(state.copy(trackId = 99).startSong(songs)?.id).isEqualTo("11")
    }

    @Test
    fun `only albums not yet queued are appended, the current one from the saved position`() {
        // Album 1 is playing and 2 is queued: nothing new.
        assertThat(state.songsToAppend("http://s", setOf(1, 2))).isEmpty()
        // Album 1 just finished and the server moved on to 2 and 3; 2 is already queued.
        val next = FmState(albums = listOf(album(2, 21, 22), album(3, 31)), trackId = null)
        assertThat(next.songsToAppend("http://s", setOf(2)).map { it.id }).containsExactly("31")
        // Nothing queued: the current album from the saved position, then the next.
        assertThat(state.songsToAppend("http://s", emptySet()).map { it.id })
            .containsExactly("12", "13", "21", "22").inOrder()
    }

    @Test
    fun `an album ends when no later radio entry belongs to it`() {
        assertThat(albumEnded(1, listOf(1, 2, 2))).isFalse()
        assertThat(albumEnded(1, listOf(2, 2))).isTrue()
        assertThat(albumEnded(1, emptyList())).isTrue()
    }
}
