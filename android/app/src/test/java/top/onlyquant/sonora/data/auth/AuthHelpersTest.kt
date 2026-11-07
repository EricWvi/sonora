package top.onlyquant.sonora.data.auth

import com.google.common.truth.Truth.assertThat
import org.junit.jupiter.api.Test

class AuthHelpersTest {
    @Test
    fun `challenge is the base64url SHA-256 of the verifier`() {
        // RFC 7636, appendix B.
        assertThat(Pkce.challenge("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"))
            .isEqualTo("E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM")
        val secret = Pkce.newSecret()
        assertThat(secret).matches("[A-Za-z0-9_-]{43}")
        assertThat(Pkce.newSecret()).isNotEqualTo(secret)
    }

    @Test
    fun `server addresses are bare origins`() {
        assertThat(ServerUrl.normalize(" https://Music.Example.com/ ")).isEqualTo("https://music.example.com")
        assertThat(ServerUrl.normalize("http://127.0.0.1:14560")).isEqualTo("http://127.0.0.1:14560")
        assertThat(ServerUrl.normalize("music.example.com")).isNull()
        assertThat(ServerUrl.normalize("ftp://music.example.com")).isNull()
        assertThat(ServerUrl.normalize("https://music.example.com/sonora")).isNull()
        assertThat(ServerUrl.normalize("https://user@music.example.com")).isNull()
        assertThat(ServerUrl.normalize("https://music.example.com?x=1")).isNull()
        assertThat(ServerUrl.normalize("")).isNull()
    }

    @Test
    fun `login colors are six hex digits without alpha`() {
        assertThat(formatLoginColors(listOf(0xFF6C4FF5.toInt(), 0x80FFFFFF.toInt(), 0xFF000000.toInt())))
            .isEqualTo("6c4ff5,ffffff,000000")
    }
}
