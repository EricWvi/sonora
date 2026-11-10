package top.onlyquant.sonora.presentation.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import timber.log.Timber
import top.onlyquant.sonora.data.repository.LibraryRepository

/**
 * As the original's LibraryViewModel. Sonora: it also starts the library sync, which the original
 * leaves to SyncManager and its WorkManager workers; the paging flows come with the library tabs.
 */
@HiltViewModel
class LibraryViewModel @Inject constructor(
    private val libraryRepository: LibraryRepository,
) : ViewModel() {

    val isSyncing: StateFlow<Boolean> = libraryRepository.isSyncing

    private val _syncErrors = MutableSharedFlow<String>(extraBufferCapacity = 1)
    /** A failed sync, to show; the local library stays as it was. */
    val syncErrors: SharedFlow<String> = _syncErrors

    /** On opening the app and on pull to refresh (Eric, 2026-11-08). */
    fun sync() {
        viewModelScope.launch {
            try {
                libraryRepository.sync()
            } catch (e: CancellationException) {
                throw e
            } catch (e: Exception) {
                Timber.tag("LibraryViewModel").w(e, "Library sync failed")
                _syncErrors.tryEmit(e.message ?: "同步失败")
            }
        }
    }
}
