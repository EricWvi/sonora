package top.onlyquant.sonora.presentation.main

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import top.onlyquant.sonora.data.api.ApiException
import top.onlyquant.sonora.data.api.SonoraApi
import top.onlyquant.sonora.data.model.Song
import top.onlyquant.sonora.presentation.components.MiniPlayerBottomSpacer
import top.onlyquant.sonora.presentation.components.MiniPlayerHeight
import top.onlyquant.sonora.presentation.viewmodel.PlayerViewModel

/**
 * Temporary, not from PixelPlayerOSS: the newest 500 tracks from `GET tracks`, so the player sheet
 * and the queue can be tried before the library screens exist. Replaced by the library step.
 */
@HiltViewModel
class TrackListPlaceholderViewModel @Inject constructor(private val api: SonoraApi) : ViewModel() {
    private val _songs = MutableStateFlow<List<Song>>(emptyList())
    val songs: StateFlow<List<Song>> = _songs.asStateFlow()
    private val _error = MutableStateFlow<String?>(null)
    val error: StateFlow<String?> = _error.asStateFlow()

    fun load() {
        viewModelScope.launch {
            try {
                _songs.value = api.songs()
                _error.value = null
            } catch (e: ApiException) {
                _error.value = e.message
            }
        }
    }
}

@Composable
fun TrackListPlaceholder(
    viewModel: TrackListPlaceholderViewModel,
    playerViewModel: PlayerViewModel,
    userName: String?,
    contentPadding: PaddingValues,
    onSignOut: () -> Unit,
) {
    val songs by viewModel.songs.collectAsStateWithLifecycle()
    val error by viewModel.error.collectAsStateWithLifecycle()
    LaunchedEffect(Unit) { viewModel.load() }

    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(
            top = contentPadding.calculateTopPadding() + 48.dp,
            bottom = contentPadding.calculateBottomPadding() + MiniPlayerHeight + MiniPlayerBottomSpacer * 2,
        ),
    ) {
        item {
            Row(
                modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 8.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween,
            ) {
                Text(userName ?: "…", style = MaterialTheme.typography.titleMedium)
                OutlinedButton(onClick = onSignOut) { Text("退出登录") }
            }
            error?.let { Text(it, modifier = Modifier.padding(16.dp), color = MaterialTheme.colorScheme.error) }
        }
        items(songs, key = { it.id }) { song ->
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .clickable { playerViewModel.showAndPlaySong(song, songs, "全部歌曲") }
                    .padding(horizontal = 16.dp, vertical = 10.dp),
            ) {
                Text(song.title, maxLines = 1, overflow = TextOverflow.Ellipsis)
                Text(
                    song.displayArtist,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
            }
        }
    }
}

/** Temporary blank page for the Search and Library tabs, until those screens are copied. */
@Composable
fun BlankTab() {
    Box(Modifier.fillMaxSize())
}
