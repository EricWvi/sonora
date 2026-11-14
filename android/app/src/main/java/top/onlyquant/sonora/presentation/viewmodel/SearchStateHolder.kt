package top.onlyquant.sonora.presentation.viewmodel

import top.onlyquant.sonora.data.model.SearchFilterType
import top.onlyquant.sonora.data.model.SearchResultItem
import top.onlyquant.sonora.data.repository.LibraryRepository
import kotlinx.collections.immutable.ImmutableList
import kotlinx.collections.immutable.persistentListOf
import kotlinx.collections.immutable.toImmutableList
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Job
import kotlinx.coroutines.channels.BufferOverflow
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.flow.debounce
import kotlinx.coroutines.launch
import timber.log.Timber
import java.util.concurrent.atomic.AtomicLong
import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.coroutines.FlowPreview

/**
 * Copied from SearchStateHolder in PixelPlayerOSS v0.4.0-alpha.53 (GPL-3.0).
 *
 * Manages search state and operations.
 * Extracted from PlayerViewModel to improve modularity.
 *
 * Responsibilities:
 * - Search query execution
 * - Search filter management
 *
 * Sonora: searches through [LibraryRepository] (the server, or the local copy when it cannot be
 * reached). The original's search history is left out (Eric, 2026-11-14): it saves every
 * submitted query, but the search screen never shows them.
 */
@Singleton
class SearchStateHolder @Inject constructor(
    private val libraryRepository: LibraryRepository,
) {
    private companion object {
        const val SEARCH_DEBOUNCE_MS = 300L
    }

    private data class SearchRequest(
        val query: String,
        val filter: SearchFilterType,
        val requestId: Long,
        // A filter change re-runs a query the user already finished typing, so it skips the
        // typing debounce.
        val immediate: Boolean = false,
    )

    private val _searchResults = MutableStateFlow<ImmutableList<SearchResultItem>>(persistentListOf())
    val searchResults = _searchResults.asStateFlow()

    private val _selectedSearchFilter = MutableStateFlow(SearchFilterType.ALL)
    val selectedSearchFilter = _selectedSearchFilter.asStateFlow()

    // Replays the latest request so a collector restarted for another PlayerViewModel resumes
    // the current search; the request id check below drops a replayed query that is stale.
    private val searchRequests = MutableSharedFlow<SearchRequest>(
        replay = 1,
        extraBufferCapacity = 1,
        onBufferOverflow = BufferOverflow.DROP_OLDEST
    )
    private val latestSearchRequestId = AtomicLong(0L)

    private var scope: CoroutineScope? = null
    private var searchJob: Job? = null
    private val owners = OwnerStack<CoroutineScope> { it }

    /**
     * Initialize with ViewModel scope.
     */
    fun initialize(scope: CoroutineScope) {
        owners.push(scope)
        this.scope = scope
        observeSearchRequests()
    }

    @OptIn(FlowPreview::class)
    private fun observeSearchRequests() {
        searchJob?.cancel()
        searchJob = scope?.launch {
            searchRequests
                .debounce { request -> if (request.immediate) 0L else SEARCH_DEBOUNCE_MS }
                .collectLatest { request ->
                    val normalizedQuery = request.query

                    if (normalizedQuery.isBlank()) {
                        if (_searchResults.value.isNotEmpty()) {
                            _searchResults.value = persistentListOf()
                        }
                        return@collectLatest
                    }

                    try {
                        libraryRepository.searchAll(normalizedQuery, request.filter).collect { resultsList ->
                            val sortedResults = resultsList.sortedWith(
                                compareBy { result ->
                                    when (result) {
                                        is SearchResultItem.SongItem -> 0
                                        is SearchResultItem.AlbumItem -> 1
                                        is SearchResultItem.ArtistItem -> 2
                                    }
                                }
                            )

                            if (request.requestId != latestSearchRequestId.get()) {
                                return@collect
                            }

                            val immutableResults = sortedResults.toImmutableList()
                            if (_searchResults.value != immutableResults) {
                                _searchResults.value = immutableResults
                            }
                        }
                    } catch (_: CancellationException) {
                    } catch (e: Exception) {
                        if (request.requestId == latestSearchRequestId.get()) {
                            Timber.e(e, "Error performing search for query: $normalizedQuery")
                            _searchResults.value = persistentListOf()
                        }
                    }
                }
        }
    }

    /**
     * Selects [filterType] and re-runs the current query with it straight away, so the results
     * switch together with the chip instead of after the screen re-issues the query and the
     * typing debounce runs out.
     */
    fun updateSearchFilter(filterType: SearchFilterType) {
        if (_selectedSearchFilter.value == filterType) return
        _selectedSearchFilter.value = filterType

        val currentQuery = searchRequests.replayCache.lastOrNull()?.query
        if (!currentQuery.isNullOrBlank()) {
            requestSearch(currentQuery, immediate = true)
        }
    }

    fun performSearch(query: String) {
        requestSearch(query.trim(), immediate = false)
    }

    private fun requestSearch(normalizedQuery: String, immediate: Boolean) {
        val requestId = latestSearchRequestId.incrementAndGet()

        if (normalizedQuery.isBlank()) {
            if (_searchResults.value.isNotEmpty()) {
                _searchResults.value = persistentListOf()
            }
        }

        // The request carries its filter, so a request replayed for another PlayerViewModel
        // searches with the filter it was made for.
        searchRequests.tryEmit(
            SearchRequest(
                query = normalizedQuery,
                filter = _selectedSearchFilter.value,
                requestId = requestId,
                immediate = immediate,
            )
        )
    }

    /**
     * A stale PlayerViewModel clearing leaves the newer binding alone; the current one clearing
     * moves the search collector to the previous live PlayerViewModel's scope, where it resumes
     * the current query, so search keeps answering after ExternalPlayerActivity finishes over
     * MainActivity.
     */
    fun onCleared(owner: CoroutineScope) {
        owners.remove(owner) { next ->
            scope = next
            if (next != null) observeSearchRequests() else searchJob?.cancel()
        }
    }
}
