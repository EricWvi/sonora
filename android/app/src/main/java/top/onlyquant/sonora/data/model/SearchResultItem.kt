package top.onlyquant.sonora.data.model

import androidx.compose.runtime.Immutable

/**
 * From SearchResultItem in PixelPlayerOSS v0.4.0-alpha.53 (GPL-3.0). Sonora: no PlaylistItem yet;
 * the 歌单 filter is shown and finds playlists once they come with step 5 (Eric, 2026-11-14).
 */
@Immutable
sealed interface SearchResultItem {
    data class SongItem(val song: Song) : SearchResultItem
    data class AlbumItem(val album: Album) : SearchResultItem
    data class ArtistItem(val artist: Artist) : SearchResultItem
}
