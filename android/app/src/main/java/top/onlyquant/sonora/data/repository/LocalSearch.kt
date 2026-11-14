package top.onlyquant.sonora.data.repository

/*
 * Not in the original: the search on the device, for when the server cannot be reached. It
 * matches and ranks as the server's `q` does (backend/internal/store/search.go, docs/api.md
 * 通用约定): a row matches when any of its fields contains the query, ASCII letters ignoring
 * case; rows rank by the first tier of fields whose text equals the query, then starts with it,
 * then contains it; rows in the same rank keep the list's own order.
 */

/** [items] that match [query], most relevant first; [tiers] gives each item's fields by importance. */
internal fun <T> rankLocalMatches(query: String, items: List<T>, tiers: (T) -> List<List<String>>): List<T> {
    val q = foldAscii(query.trim())
    if (q.isEmpty()) return emptyList()
    return items
        .mapIndexedNotNull { index, item ->
            val rank = relevance(q, tiers(item)) ?: return@mapIndexedNotNull null
            Triple(rank, index, item)
        }
        .sortedWith(compareBy({ it.first }, { it.second }))
        .map { it.third }
}

/** As the server's CASE: per tier, equal, then prefix, then contains; null when nothing matches. */
private fun relevance(q: String, tiers: List<List<String>>): Int? {
    tiers.forEachIndexed { tier, fields ->
        val texts = fields.map(::foldAscii)
        when {
            texts.any { it == q } -> return tier * 3
            texts.any { it.startsWith(q) } -> return tier * 3 + 1
            texts.any { it.contains(q) } -> return tier * 3 + 2
        }
    }
    return null
}

/** SQLite's LIKE ignores case for ASCII letters only. */
private fun foldAscii(s: String): String =
    buildString(s.length) { s.forEach { append(if (it in 'A'..'Z') it + ('a' - 'A') else it) } }
