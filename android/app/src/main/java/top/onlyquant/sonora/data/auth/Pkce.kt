package top.onlyquant.sonora.data.auth

import java.security.MessageDigest
import java.security.SecureRandom

/** PKCE (RFC 7636, S256) for the device login, see backend/internal/auth/device.go. */
object Pkce {
    private val random = SecureRandom()

    /** A fresh random value for the code verifier or the state. */
    fun newSecret(): String {
        val bytes = ByteArray(32)
        random.nextBytes(bytes)
        return base64Url(bytes)
    }

    fun challenge(verifier: String): String =
        base64Url(MessageDigest.getInstance("SHA-256").digest(verifier.toByteArray(Charsets.US_ASCII)))

    private fun base64Url(bytes: ByteArray): String =
        java.util.Base64.getUrlEncoder().withoutPadding().encodeToString(bytes)
}
