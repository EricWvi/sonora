package top.onlyquant.sonora.data.model

import androidx.annotation.StringRes
import androidx.compose.runtime.Immutable
import top.onlyquant.sonora.R

/**
 * From LibraryTabId in PixelPlayerOSS v0.4.0-alpha.53 (GPL-3.0). Sonora: songs, albums, artists
 * and playlists (Eric, 2026-11-11); the original also has folders, liked and genres.
 */
@Immutable
enum class LibraryTabId(
    val storageKey: String,
    @StringRes val titleRes: Int,
    val defaultSort: SortOption
) {
    SONGS("SONGS", R.string.library_tab_songs, SortOption.SongTitleAZ),
    ALBUMS("ALBUMS", R.string.library_tab_albums, SortOption.AlbumTitleAZ),
    ARTISTS("ARTIST", R.string.library_tab_artists, SortOption.ArtistNameAZ),
    PLAYLISTS("PLAYLISTS", R.string.library_tab_playlists, SortOption.PlaylistNameAZ);

    companion object {
        fun fromStorageKey(key: String): LibraryTabId =
            entries.firstOrNull { it.storageKey == key } ?: SONGS
    }
}

fun String.toLibraryTabIdOrNull(): LibraryTabId? =
    LibraryTabId.entries.firstOrNull { it.storageKey == this }