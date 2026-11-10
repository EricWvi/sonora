package top.onlyquant.sonora

import android.app.Application
import dagger.hilt.android.HiltAndroidApp
import coil.ImageLoader
import coil.ImageLoaderFactory
import javax.inject.Inject
import timber.log.Timber

@HiltAndroidApp
class SonoraApplication : Application(), ImageLoaderFactory {
    @Inject lateinit var imageLoader: dagger.Lazy<ImageLoader>

    /** As in PixelPlayerOSS: Coil's singleton is the injected loader (di/AppModule.kt). */
    override fun newImageLoader(): ImageLoader = imageLoader.get()

    override fun onCreate() {
        super.onCreate()
        Timber.plant(Timber.DebugTree())
    }
}
