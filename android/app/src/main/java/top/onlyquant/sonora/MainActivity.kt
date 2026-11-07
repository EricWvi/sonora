package top.onlyquant.sonora

import android.Manifest
import android.content.ComponentName
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.core.content.ContextCompat
import androidx.media3.common.MediaItem
import androidx.media3.common.MediaMetadata
import androidx.media3.common.Player
import androidx.media3.session.MediaController
import androidx.media3.session.SessionToken
import com.google.common.util.concurrent.ListenableFuture
import top.onlyquant.sonora.data.service.MusicService

/**
 * Skeleton screen for checking the playback core: plays one fixed song bundled with debug builds.
 * It goes away when the library screens arrive.
 */
class MainActivity : ComponentActivity() {
    private var controllerFuture: ListenableFuture<MediaController>? = null
    private var controller: MediaController? = null
    private var isPlaying by mutableStateOf(false)
    private var ready by mutableStateOf(false)

    private val notificationPermission =
        registerForActivityResult(ActivityResultContracts.RequestPermission()) {}

    private val playerListener = object : Player.Listener {
        override fun onIsPlayingChanged(playing: Boolean) {
            isPlaying = playing
        }
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        notificationPermission.launch(Manifest.permission.POST_NOTIFICATIONS)
        setContent {
            MaterialTheme {
                Surface(Modifier.fillMaxSize()) {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.Center,
                    ) {
                        Text("Sonora", style = MaterialTheme.typography.headlineMedium)
                        Button(enabled = ready, onClick = ::togglePlayback) {
                            Text(if (isPlaying) "暂停" else "播放")
                        }
                    }
                }
            }
        }
    }

    override fun onStart() {
        super.onStart()
        val token = SessionToken(this, ComponentName(this, MusicService::class.java))
        val future = MediaController.Builder(this, token).buildAsync()
        controllerFuture = future
        future.addListener(
            {
                val connected = future.get()
                connected.addListener(playerListener)
                controller = connected
                isPlaying = connected.isPlaying
                ready = true
            },
            ContextCompat.getMainExecutor(this),
        )
    }

    override fun onStop() {
        ready = false
        controller = null
        controllerFuture?.let(MediaController::releaseFuture)
        controllerFuture = null
        super.onStop()
    }

    private fun togglePlayback() {
        val player = controller ?: return
        if (player.isPlaying) {
            player.pause()
            return
        }
        if (player.mediaItemCount == 0) {
            player.setMediaItem(
                MediaItem.Builder()
                    .setMediaId("fixture")
                    .setUri("asset:///fixture.mp3")
                    .setMediaMetadata(MediaMetadata.Builder().setTitle("Fixture").build())
                    .build(),
            )
            player.prepare()
        }
        player.play()
    }
}
