package top.onlyquant.sonora.data.auth

import android.net.Uri
import android.os.Build
import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.withContext
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import timber.log.Timber

/** A login failure with a message for the user. */
class AuthException(message: String) : Exception(message)

/** The user chose 取消 on the confirm page. */
class LoginCancelledException : Exception("登录已取消")

/**
 * The device login of backend/internal/auth/device.go: the system browser goes through the
 * server's OIDC login and confirm page, then comes back to sonora://auth/callback with a code the
 * app exchanges, with its PKCE verifier, for a device token.
 */
@Singleton
class AuthRepository @Inject constructor(
    private val store: SessionStore,
    private val http: OkHttpClient,
    private val json: Json,
) {
    val session: Flow<Session?> = store.session

    suspend fun lastServer(): String? = store.lastServer()

    /** Checks the server and returns the address to open in the browser. */
    suspend fun begin(serverInput: String, colors: String): Uri {
        val server = ServerUrl.normalize(serverInput)
            ?: throw AuthException("服务器地址无效，请填写 http:// 或 https:// 开头的地址")
        checkServer(server)
        val login = PendingLogin(server, state = Pkce.newSecret(), verifier = Pkce.newSecret())
        store.savePending(login)
        return Uri.parse("$server/auth/device/login").buildUpon()
            .appendQueryParameter("platform", "android")
            .appendQueryParameter("name", deviceName())
            .appendQueryParameter("redirect_uri", REDIRECT_URI)
            .appendQueryParameter("state", login.state)
            .appendQueryParameter("code_challenge", Pkce.challenge(login.verifier))
            .appendQueryParameter("code_challenge_method", "S256")
            .appendQueryParameter("colors", colors)
            .build()
    }

    /** Finishes a login from the browser's callback. */
    suspend fun complete(callback: Uri): Session {
        val login = store.pending() ?: throw AuthException("登录已过期，请重新登录")
        if (callback.getQueryParameter("state") != login.state) {
            throw AuthException("登录请求不匹配，请重新登录")
        }
        if (callback.getQueryParameter("error") != null) {
            store.clearPending()
            throw LoginCancelledException()
        }
        val code = callback.getQueryParameter("code") ?: throw AuthException("登录没有完成，请重新登录")
        val body = json.encodeToString(TokenRequest.serializer(), TokenRequest(code, login.verifier))
        val request = Request.Builder()
            .url("${login.server}/auth/device/token")
            .post(body.toRequestBody(JSON))
            .build()
        val response = call(request)
        if (response.code != 200) {
            store.clearPending()
            throw AuthException(errorMessage(response.body) ?: "登录失败（${response.code}），请重新登录")
        }
        val token = json.decodeFromString(TokenResponse.serializer(), response.body)
        val session = Session(login.server, token.token, token.device.id)
        store.saveSession(session)
        return session
    }

    /** Revokes this device on the server, then forgets the session even if that failed. */
    suspend fun signOut() {
        val session = store.current() ?: return
        try {
            call(
                Request.Builder()
                    .url("${session.server}/api/v1/devices/${session.deviceId}")
                    .header("Authorization", "Bearer ${session.token}")
                    .delete()
                    .build()
            )
        } catch (e: Exception) {
            Timber.w(e, "Revoking the device failed")
        }
        store.clearSession()
    }

    /** The server no longer accepts the token. */
    suspend fun onUnauthorized() = store.clearSession()

    private suspend fun checkServer(server: String) {
        val response = try {
            call(Request.Builder().url("$server/api/version").build())
        } catch (e: Exception) {
            Timber.w(e, "Server check failed")
            throw AuthException("无法连接服务器，请检查地址和网络")
        }
        val parsed = runCatching { json.decodeFromString(VersionResponse.serializer(), response.body) }
        val ok = response.code == 200 && parsed.isSuccess
        if (!ok) Timber.w(parsed.exceptionOrNull(), "Server check: %d %s", response.code, response.body.take(200))
        if (!ok) throw AuthException("这个地址不是 Sonora 服务器")
    }

    private class Response(val code: Int, val body: String)

    private suspend fun call(request: Request): Response = withContext(Dispatchers.IO) {
        http.newCall(request).execute().use { Response(it.code, it.body.string()) }
    }

    private fun errorMessage(body: String): String? =
        runCatching { json.decodeFromString(ErrorResponse.serializer(), body).error }.getOrNull()

    private fun deviceName(): String =
        (Build.MODEL?.takeIf { it.isNotBlank() } ?: "安卓设备").take(64)

    @Serializable
    private data class TokenRequest(val code: String, @kotlinx.serialization.SerialName("code_verifier") val verifier: String)

    @Serializable
    private data class TokenResponse(val token: String, val device: Device)

    @Serializable
    private data class Device(val id: Long)

    @Serializable
    private data class VersionResponse(val version: String)

    @Serializable
    private data class ErrorResponse(val error: String)

    companion object {
        const val REDIRECT_URI = "sonora://auth/callback"
        private val JSON = "application/json; charset=utf-8".toMediaType()

        fun isCallback(uri: Uri?): Boolean =
            uri != null && uri.scheme == "sonora" && uri.host == "auth" && uri.path == "/callback"
    }
}
