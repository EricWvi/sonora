package top.onlyquant.sonora.presentation.auth

import android.net.Uri
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.receiveAsFlow
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import top.onlyquant.sonora.data.api.ApiException
import top.onlyquant.sonora.data.api.SonoraApi
import top.onlyquant.sonora.data.auth.AuthException
import top.onlyquant.sonora.data.auth.AuthRepository
import top.onlyquant.sonora.data.auth.LoginCancelledException
import top.onlyquant.sonora.data.auth.Session

sealed interface SessionState {
    data object Loading : SessionState
    data object SignedOut : SessionState
    data class SignedIn(val session: Session) : SessionState
}

data class LoginUiState(
    val server: String = "",
    val isLoading: Boolean = false,
    val message: String? = null,
    /** The signed-in user's name from GET me, once loaded. */
    val userName: String? = null,
)

@HiltViewModel
class AuthViewModel @Inject constructor(
    private val auth: AuthRepository,
    private val api: SonoraApi,
) : ViewModel() {
    val session: StateFlow<SessionState> = auth.session
        .map { if (it == null) SessionState.SignedOut else SessionState.SignedIn(it) }
        .stateIn(viewModelScope, SharingStarted.Eagerly, SessionState.Loading)

    private val _ui = MutableStateFlow(LoginUiState())
    val ui: StateFlow<LoginUiState> = _ui.asStateFlow()

    private val _openBrowser = Channel<Uri>(Channel.BUFFERED)
    /** Addresses to open in a Custom Tab. */
    val openBrowser = _openBrowser.receiveAsFlow()

    init {
        viewModelScope.launch { auth.lastServer()?.let { s -> _ui.update { it.copy(server = s) } } }
        viewModelScope.launch {
            session.collect { if (it is SessionState.SignedIn) loadMe() else _ui.update { u -> u.copy(userName = null) } }
        }
    }

    fun connect(server: String, colors: String) {
        if (_ui.value.isLoading) return
        _ui.update { it.copy(isLoading = true, server = server) }
        viewModelScope.launch {
            try {
                _openBrowser.send(auth.begin(server, colors))
            } catch (e: AuthException) {
                _ui.update { it.copy(message = e.message) }
            } finally {
                _ui.update { it.copy(isLoading = false) }
            }
        }
    }

    fun handleCallback(uri: Uri) {
        _ui.update { it.copy(isLoading = true) }
        viewModelScope.launch {
            try {
                auth.complete(uri)
            } catch (e: LoginCancelledException) {
                _ui.update { it.copy(message = "已取消登录") }
            } catch (e: AuthException) {
                _ui.update { it.copy(message = e.message) }
            } catch (e: Exception) {
                _ui.update { it.copy(message = "登录失败，请重新登录") }
            } finally {
                _ui.update { it.copy(isLoading = false) }
            }
        }
    }

    fun signOut() {
        viewModelScope.launch { auth.signOut() }
    }

    fun messageShown() = _ui.update { it.copy(message = null) }

    private suspend fun loadMe() {
        try {
            val me = api.me()
            _ui.update { it.copy(userName = me.name) }
        } catch (e: ApiException) {
            _ui.update { it.copy(message = e.message) }
        }
    }
}
