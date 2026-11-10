package top.onlyquant.sonora

import android.Manifest
import android.content.Intent
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.result.contract.ActivityResultContracts
import androidx.activity.viewModels
import androidx.browser.customtabs.CustomTabsIntent
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.Surface
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import dagger.hilt.android.AndroidEntryPoint
import top.onlyquant.sonora.data.auth.AuthRepository
import top.onlyquant.sonora.presentation.auth.AuthViewModel
import top.onlyquant.sonora.presentation.auth.LoginScreen
import top.onlyquant.sonora.presentation.auth.SessionState
import top.onlyquant.sonora.presentation.main.BlankTab
import top.onlyquant.sonora.presentation.main.MainUI
import top.onlyquant.sonora.presentation.main.TrackListPlaceholder
import top.onlyquant.sonora.presentation.navigation.Screen
import top.onlyquant.sonora.presentation.viewmodel.PlayerViewModel
import top.onlyquant.sonora.ui.theme.PixelPlayerTheme

/**
 * Shows the login until a device token is stored, then the main shell (MainUI) with the player
 * sheet. The Home tab is a temporary track list and the other tabs are blank until the library
 * screens arrive.
 */
@AndroidEntryPoint
class MainActivity : ComponentActivity() {
    private val auth: AuthViewModel by viewModels()
    // Activity-scoped, as sharedPlayerViewModel() resolves it for the copied components.
    private val playerViewModel: PlayerViewModel by viewModels()

    private val notificationPermission =
        registerForActivityResult(ActivityResultContracts.RequestPermission()) {}

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
                        val navController = rememberNavController()
                        MainUI(playerViewModel, navController) { padding ->
                            NavHost(navController, startDestination = Screen.Home.route) {
                                composable(Screen.Home.route) {
                                    TrackListPlaceholder(
                                        viewModel = hiltViewModel(),
                                        playerViewModel = playerViewModel,
                                        userName = ui.userName,
                                        contentPadding = padding,
                                        onSignOut = auth::signOut,
                                    )
                                }
                                composable(Screen.Search.route) { BlankTab() }
                                composable(Screen.Library.route) { BlankTab() }
                            }
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
}
