package top.onlyquant.sonora

import android.Manifest
import android.content.ComponentName
import android.content.Intent
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.viewModels
import androidx.browser.customtabs.CustomTabsIntent
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.Button
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
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
import dagger.hilt.android.AndroidEntryPoint
import top.onlyquant.sonora.data.auth.AuthRepository
import top.onlyquant.sonora.data.service.MusicService
import top.onlyquant.sonora.presentation.auth.AuthViewModel
import top.onlyquant.sonora.presentation.auth.LoginScreen
import top.onlyquant.sonora.presentation.auth.SessionState
import top.onlyquant.sonora.ui.theme.PixelPlayerTheme

/**
 * Shows the login until a device token is stored. The signed-in screen is a placeholder for
 * checking the API and the playback core (one fixed song bundled with debug builds); it goes away
 * when the library screens arrive.
 */
@AndroidEntryPoint
class MainActivity : ComponentActivity() {
    private val auth: AuthViewModel by viewModels()

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
        enableEdgeToEdge()
        notificationPermission.launch(Manifest.permission.POST_NOTIFICATIONS)
        handleCallback(intent)
        setContent {
            PixelPlayerTheme {
                val session by auth.session.collectAsState()
                val ui by auth.ui.collectAsState()
                LaunchedEffect(Unit) {
                    auth.openBrowser.collect { uri ->
                        CustomTabsIntent.Builder().build().launchUrl(this@MainActivity, uri)
                    }
                }
                when (session) {
                    SessionState.Loading -> Surface(Modifier.fillMaxSize()) {}
                    SessionState.SignedOut -> LoginScreen(
                        initialServer = ui.server,
                        isLoading = ui.isLoading,
                        message = ui.message,
                        onMessageShown = auth::messageShown,
                        onConnect = auth::connect,
                    )
                    is SessionState.SignedIn -> Surface(Modifier.fillMaxSize()) {
                        Column(
                            horizontalAlignment = Alignment.CenterHorizontally,
                            verticalArrangement = Arrangement.Center,
                        ) {
                            Text("Sonora", style = MaterialTheme.typography.headlineMedium)
                            Text(ui.userName ?: "…")
                            Button(enabled = ready, onClick = ::togglePlayback) {
                                Text(if (isPlaying) "暂停" else "播放")
                            }
                            OutlinedButton(onClick = auth::signOut) { Text("退出登录") }
                        }
                    }
                }
            }
        }
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        handleCallback(intent)
    }

    private fun handleCallback(intent: Intent?) {
        val uri = intent?.data
        if (AuthRepository.isCallback(uri)) {
            auth.handleCallback(uri!!)
            // A recreated activity must not finish the same login twice.
            intent.data = null
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
