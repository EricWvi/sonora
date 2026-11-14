package top.onlyquant.sonora.data.repository

import com.google.common.truth.Truth.assertThat
import org.junit.jupiter.api.Test

class LocalSearchTest {
    private data class Row(val title: String, val others: List<String> = emptyList())

    private fun search(q: String, rows: List<Row>) =
        rankLocalMatches(q, rows) { listOf(listOf(it.title), it.others) }.map { it.title }

    @Test
    fun `equal, then prefix, then contains, each tier before the next`() {
        val rows = listOf(
            Row("Love Story"),
            Row("Other", others = listOf("love")),
            Row("My love"),
            Row("love"),
            Row("Nothing"),
        )
        assertThat(search("love", rows)).containsExactly("love", "Love Story", "My love", "Other").inOrder()
    }

    @Test
    fun `rows of the same rank keep the list order`() {
        val rows = listOf(Row("b 晴天"), Row("a 晴天"), Row("c 晴天"))
        assertThat(search("晴天", rows)).containsExactly("b 晴天", "a 晴天", "c 晴天").inOrder()
    }

    @Test
    fun `case is ignored for ASCII letters only, and the query is trimmed`() {
        assertThat(search("  RED ", listOf(Row("Red Flag")))).containsExactly("Red Flag")
        assertThat(search("Ä", listOf(Row("ä")))).isEmpty()
    }

    @Test
    fun `a blank query matches nothing`() {
        assertThat(search(" ", listOf(Row("x")))).isEmpty()
    }

    @Test
    fun `wildcards match literally`() {
        assertThat(search("%", listOf(Row("100%"), Row("abc")))).containsExactly("100%")
    }
}
