package top.onlyquant.sonora.presentation.model

import top.onlyquant.sonora.data.model.Song
import android.os.Parcelable
import kotlinx.parcelize.Parcelize

// Sonora: only the model. The original maps its on-device playback statistics to it
// (mapRecentlyPlayedSongs and the helpers after it); Sonora's history comes from the server's
// `GET history`, newest first with one entry per song (docs/api.md).
@Parcelize
data class RecentlyPlayedSongUiModel(
    val song: Song,
    val lastPlayedTimestamp: Long
) : Parcelable
