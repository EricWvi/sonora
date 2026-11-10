package top.onlyquant.sonora.data.api

import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.KSerializer
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import okhttp3.OkHttpClient
import okhttp3.Request
import top.onlyquant.sonora.data.auth.AuthRepository
import top.onlyquant.sonora.data.auth.SessionStore
import top.onlyquant.sonora.data.model.Song

/** A failed player API call; [status] is the HTTP status, 0 when the server was not reached. */
class ApiException(val status: Int, message: String) : Exception(message)

@Serializable
data class Me(val name: String, val email: String = "", val canImport: Boolean = false)

/**
 * The player API (docs/api.md, /api/v1/) with the device token. A 401 means the token was revoked,
 * so the session is dropped and the app returns to the login.
 */
@Singleton
class SonoraApi @Inject constructor(
    private val http: OkHttpClient,
    private val json: Json,
    private val sessions: SessionStore,
    private val auth: AuthRepository,
) {
    suspend fun me(): Me = get("me", Me.serializer())

    /** One page of `GET tracks` in [sort] order (docs/api.md), as the original's songs. */
    suspend fun songs(offset: Int = 0, limit: Int = 500, sort: String = "default"): List<Song> {
        val server = sessions.current()?.server ?: throw ApiException(401, "未登录")
        return get("tracks?offset=$offset&limit=$limit&sort=$sort", Page.serializer(Track.serializer()))
            .items.map { it.toSong(server) }
    }

    private suspend fun <T> get(path: String, serializer: KSerializer<T>): T {
        val session = sessions.current() ?: throw ApiException(401, "未登录")
        val request = Request.Builder()
            .url("${session.server}/api/v1/$path")
            .build()
        val (code, body) = try {
            withContext(Dispatchers.IO) {
                http.newCall(request).execute().use { it.code to it.body.string() }
            }
        } catch (e: java.io.IOException) {
            throw ApiException(0, "无法连接服务器")
        }
        if (code == 401) {
            auth.onUnauthorized()
            throw ApiException(401, "登录已失效，请重新登录")
        }
        if (code != 200) {
            val message = runCatching { json.decodeFromString(Error.serializer(), body).error }.getOrNull()
            throw ApiException(code, message ?: "请求失败（$code）")
        }
        return json.decodeFromString(serializer, body)
    }

    @Serializable
    private data class Error(val error: String)
}
