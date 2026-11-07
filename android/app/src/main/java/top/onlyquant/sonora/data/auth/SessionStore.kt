package top.onlyquant.sonora.data.auth

import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.longPreferencesKey
import androidx.datastore.preferences.core.stringPreferencesKey
import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map

/** A signed-in Sonora server. */
data class Session(val server: String, val token: String, val deviceId: Long)

/** A device login waiting for the browser to come back. */
data class PendingLogin(val server: String, val state: String, val verifier: String)

/**
 * Keeps the session in the app's private storage. The pending login is stored too, as the app may
 * be killed while the browser is open.
 */
@Singleton
class SessionStore @Inject constructor(private val store: DataStore<Preferences>) {
    private object Keys {
        val server = stringPreferencesKey("server")
        val token = stringPreferencesKey("token")
        val deviceId = longPreferencesKey("device_id")
        val pendingServer = stringPreferencesKey("pending_server")
        val pendingState = stringPreferencesKey("pending_state")
        val pendingVerifier = stringPreferencesKey("pending_verifier")
    }

    val session: Flow<Session?> = store.data.map { p ->
        val server = p[Keys.server]
        val token = p[Keys.token]
        val id = p[Keys.deviceId]
        if (server != null && token != null && id != null) Session(server, token, id) else null
    }

    suspend fun current(): Session? = session.first()

    suspend fun pending(): PendingLogin? = store.data.first().let { p ->
        val server = p[Keys.pendingServer] ?: return null
        val state = p[Keys.pendingState] ?: return null
        val verifier = p[Keys.pendingVerifier] ?: return null
        PendingLogin(server, state, verifier)
    }

    suspend fun savePending(login: PendingLogin) {
        store.edit {
            it[Keys.pendingServer] = login.server
            it[Keys.pendingState] = login.state
            it[Keys.pendingVerifier] = login.verifier
        }
    }

    suspend fun clearPending() {
        store.edit {
            it.remove(Keys.pendingServer)
            it.remove(Keys.pendingState)
            it.remove(Keys.pendingVerifier)
        }
    }

    suspend fun saveSession(session: Session) {
        store.edit {
            it[Keys.server] = session.server
            it[Keys.token] = session.token
            it[Keys.deviceId] = session.deviceId
            it.remove(Keys.pendingServer)
            it.remove(Keys.pendingState)
            it.remove(Keys.pendingVerifier)
        }
    }

    /** Forgets the session; the server address stays for the next login. */
    suspend fun clearSession() {
        store.edit {
            it.remove(Keys.token)
            it.remove(Keys.deviceId)
        }
    }

    suspend fun lastServer(): String? = store.data.first()[Keys.server]
}
