package top.onlyquant.sonora.data.auth

import androidx.compose.material3.ColorScheme
import androidx.compose.ui.graphics.compositeOver
import androidx.compose.ui.graphics.toArgb

/**
 * The colors parameter of the device login: the app's current colour roles in the order the
 * confirm page expects (themeRoles in backend/internal/auth/device.go), so the page matches the
 * wallpaper colours. Translucent roles are flattened onto the background.
 */
fun loginColors(scheme: ColorScheme): String = formatLoginColors(
    listOf(
        scheme.primary,
        scheme.onPrimary,
        scheme.background,
        scheme.onSurface,
        scheme.onSurfaceVariant,
        scheme.surfaceContainerLowest,
        scheme.surfaceContainerLow,
        scheme.surfaceContainerHigh,
        scheme.secondaryContainer,
        scheme.onSecondaryContainer,
    ).map { it.compositeOver(scheme.background).toArgb() }
)

internal fun formatLoginColors(argb: List<Int>): String =
    argb.joinToString(",") { "%06x".format(it and 0xFFFFFF) }
