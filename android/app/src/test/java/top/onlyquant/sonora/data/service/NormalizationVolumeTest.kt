package top.onlyquant.sonora.data.service

import com.google.common.truth.Truth.assertThat
import org.junit.jupiter.api.Test

class NormalizationVolumeTest {
    @Test
    fun `a cut lowers the volume by the gain`() {
        assertThat(normalizationVolume(enabled = true, gainDb = -3.04, online = false))
            .isWithin(1e-4f).of(0.7047f)
        assertThat(normalizationVolume(enabled = true, gainDb = -0.58, online = false))
            .isWithin(1e-4f).of(0.9354f)
    }

    @Test
    fun `a boost is capped at full volume`() {
        assertThat(normalizationVolume(enabled = true, gainDb = 4.5, online = false)).isEqualTo(1f)
        assertThat(normalizationVolume(enabled = true, gainDb = 0.0, online = false)).isEqualTo(1f)
    }

    @Test
    fun `no normalization when off, unmeasured or online`() {
        assertThat(normalizationVolume(enabled = false, gainDb = -3.04, online = false)).isNull()
        assertThat(normalizationVolume(enabled = true, gainDb = null, online = false)).isNull()
        assertThat(normalizationVolume(enabled = true, gainDb = -3.04, online = true)).isNull()
    }
}
