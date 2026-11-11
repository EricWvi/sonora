package top.onlyquant.sonora.data.api

import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.KSerializer
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import okhttp3.OkHttpClient
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import top.onlyquant.sonora.data.auth.AuthRepository
import top.onlyquant.sonora.data.auth.SessionStore
import top.onlyquant.sonora.data.model.Song

/** A failed player API call; [status] is the HTTP status, 0 when the server was not reached. */
class ApiException(val status: Int, message: String) : Exception(message)

/** The radio (docs/api.md#漫游-fm): albums from the current one, the song to resume at, the groups. */
@Serializable
data class FmState(val albums: List<FmAlbum>, val trackId: Long? = null, val groupIds: List<Long> = emptyList())

@Serializable
data class FmAlbum(val id: Long, val tracks: List<Track>)

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

    /** One page of `GET library/changes` after [since] (docs/api.md#曲库变更). */
    suspend fun libraryChanges(since: Long, limit: Int = 500): LibraryChanges =
        get("library/changes?since=$since&limit=$limit", LibraryChanges.serializer())

    /** `GET fm?count=` (docs/api.md#漫游-fm): the radio's next [count] albums, the current one first. */
    suspend fun fm(count: Int = 2): FmState = get("fm?count=$count", FmState.serializer())

    /** `POST fm/albums/{id}/finish`: the album was heard to its end. */
    suspend fun fmFinish(albumId: Long): FmState =
        send("POST", "fm/albums/$albumId/finish", null, FmState.serializer())!!

    /** `POST fm/albums/{id}/skip`: skip the album without counting it. */
    suspend fun fmSkip(albumId: Long): FmState =
        send("POST", "fm/albums/$albumId/skip", null, FmState.serializer())!!

    /** `PUT fm/position`: the radio's song, to resume at. */
    suspend fun fmPosition(trackId: Long) {
        send<Unit>("PUT", "fm/position", "{\"trackId\":$trackId}", null)
    }

    /** `GET history`: recently played songs, most recent first (docs/api.md). */
    suspend fun history(): List<Song> {
        val server = sessions.current()?.server ?: throw ApiException(401, "未登录")
        return get("history", History.serializer()).items.map { it.track.toSong(server) }
    }

    private suspend fun <T> get(path: String, serializer: KSerializer<T>): T =
        send("GET", path, null, serializer)!!

    private suspend fun <T> send(method: String, path: String, jsonBody: String?, serializer: KSerializer<T>?): T? {
        val session = sessions.current() ?: throw ApiException(401, "未登录")
        val requestBody = when {
            jsonBody != null -> jsonBody.toRequestBody("application/json".toMediaType())
            method == "GET" -> null
            else -> ByteArray(0).toRequestBody(null)
        }
        val request = Request.Builder()
            .url("${session.server}/api/v1/$path")
            .method(method, requestBody)
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
        if (code !in 200..299) {
            val message = runCatching { json.decodeFromString(Error.serializer(), body).error }.getOrNull()
            throw ApiException(code, message ?: "请求失败（$code）")
        }
        return serializer?.let { json.decodeFromString(it, body) }
    }

    @Serializable
    private data class Error(val error: String)

    @Serializable
    private data class HistoryItem(val track: Track)

    @Serializable
    private data class History(val items: List<HistoryItem>)
}
