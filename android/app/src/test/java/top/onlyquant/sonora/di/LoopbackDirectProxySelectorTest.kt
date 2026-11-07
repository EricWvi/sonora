package top.onlyquant.sonora.di

import com.google.common.truth.Truth.assertThat
import java.net.InetSocketAddress
import java.net.Proxy
import java.net.ProxySelector
import java.net.URI
import org.junit.jupiter.api.Test

class LoopbackDirectProxySelectorTest {
    private val proxy = Proxy(Proxy.Type.HTTP, InetSocketAddress.createUnresolved("10.0.2.2", 7890))
    private val system = object : ProxySelector() {
        override fun select(uri: URI?) = listOf(proxy)
        override fun connectFailed(uri: URI?, sa: java.net.SocketAddress?, ioe: java.io.IOException?) {}
    }
    private val selector = LoopbackDirectProxySelector(system)

    @Test
    fun `loopback goes direct and the rest uses the system proxy`() {
        for (url in listOf("http://127.0.0.1:14560/api/version", "http://localhost:14560", "http://[::1]:14560")) {
            assertThat(selector.select(URI(url))).containsExactly(Proxy.NO_PROXY)
        }
        for (url in listOf("https://music.example.com", "http://192.168.1.2:14560", "http://deadbeef.example")) {
            assertThat(selector.select(URI(url))).containsExactly(proxy)
        }
    }
}
