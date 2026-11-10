package top.onlyquant.sonora.data.api

import kotlinx.coroutines.runBlocking
import okhttp3.HttpUrl.Companion.toHttpUrlOrNull
import okhttp3.Interceptor
import okhttp3.Response
import top.onlyquant.sonora.data.auth.SessionStore

/**
 * Adds the device token to requests for the signed-in server, so that audio (Media3) and covers
 * (Coil) load with the same OkHttpClient as the API. Requests to other hosts, and ones that already
 * carry an Authorization header, go out unchanged.
 */
class AuthHeaderInterceptor(private val sessions: SessionStore) : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val request = chain.request()
        if (request.header("Authorization") != null) return chain.proceed(request)
        // DataStore keeps the last value in memory, so this does not touch the disk per request.
        val session = runBlocking { sessions.current() } ?: return chain.proceed(request)
        val server = session.server.toHttpUrlOrNull() ?: return chain.proceed(request)
        val url = request.url
        if (url.scheme != server.scheme || url.host != server.host || url.port != server.port) {
            return chain.proceed(request)
        }
        return chain.proceed(
            request.newBuilder().header("Authorization", "Bearer ${session.token}").build(),
        )
    }
}
