package top.onlyquant.sonora.di

import android.content.Context
import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.preferencesDataStore
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import java.io.IOException
import java.net.InetAddress
import java.net.Proxy
import java.net.ProxySelector
import java.net.SocketAddress
import java.net.URI
import java.util.concurrent.TimeUnit
import javax.inject.Singleton
import kotlinx.serialization.json.Json
import okhttp3.OkHttpClient

private val Context.sessionDataStore: DataStore<Preferences> by preferencesDataStore(name = "session")

@Module
@InstallIn(SingletonComponent::class)
object AppModule {
    @Provides
    @Singleton
    fun dataStore(@ApplicationContext context: Context): DataStore<Preferences> = context.sessionDataStore

    @Provides
    @Singleton
    fun json(): Json = Json { ignoreUnknownKeys = true }

    @Provides
    @Singleton
    fun http(): OkHttpClient = OkHttpClient.Builder()
        .proxySelector(LoopbackDirectProxySelector(ProxySelector.getDefault()))
        .connectTimeout(10, TimeUnit.SECONDS)
        .readTimeout(30, TimeUnit.SECONDS)
        .build()
}

/**
 * Reaches loopback addresses directly, as Chrome does. Through a system proxy (the emulator often
 * has the host's), 127.0.0.1 is resolved wherever the proxy decides, which during development
 * reached an unrelated server instead of the port forwarded by adb reverse.
 */
internal class LoopbackDirectProxySelector(private val fallback: ProxySelector?) : ProxySelector() {
    override fun select(uri: URI?): List<Proxy> {
        if (isLoopback(uri?.host)) return listOf(Proxy.NO_PROXY)
        return fallback?.select(uri)?.takeIf { it.isNotEmpty() } ?: listOf(Proxy.NO_PROXY)
    }

    override fun connectFailed(uri: URI?, sa: SocketAddress?, ioe: IOException?) {
        fallback?.connectFailed(uri, sa, ioe)
    }

    companion object {
        internal fun isLoopback(host: String?): Boolean {
            val h = host?.trim('[', ']')?.lowercase() ?: return false
            if (h == "localhost") return true
            // Only literal addresses: looking a name up here would block.
            if (!h.all { it.isDigit() || it == '.' || it == ':' || it in 'a'..'f' }) return false
            return runCatching { InetAddress.getByName(h).isLoopbackAddress }.getOrDefault(false)
        }
    }
}
