package top.onlyquant.sonora.data.auth

import java.net.URI

/** Sonora is served at the root of its origin, so a server address is a bare http(s) origin. */
object ServerUrl {
    /** Returns the address as scheme://host[:port], or null when it is not one. */
    fun normalize(input: String): String? {
        val trimmed = input.trim().trimEnd('/')
        val uri = runCatching { URI(trimmed) }.getOrNull() ?: return null
        val scheme = uri.scheme?.lowercase() ?: return null
        if (scheme != "http" && scheme != "https") return null
        if (uri.host.isNullOrEmpty() || uri.rawUserInfo != null) return null
        if (!uri.rawPath.isNullOrEmpty() || uri.rawQuery != null || uri.rawFragment != null) return null
        val port = if (uri.port == -1) "" else ":${uri.port}"
        return "$scheme://${uri.host.lowercase()}$port"
    }
}
