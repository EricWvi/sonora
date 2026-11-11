package top.onlyquant.sonora.data.preferences

import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.floatPreferencesKey
import androidx.datastore.preferences.core.intPreferencesKey
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.media3.common.Player
import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.flow.map
import top.onlyquant.sonora.data.model.SortOption
import top.onlyquant.sonora.utils.PlaybackSpeedUtils

/**
 * Reduced from UserPreferencesRepository in PixelPlayerOSS v0.4.0-alpha.53 (GPL-3.0): only the
 * settings the playback core reads, with the original keys and defaults. The rest come with the
 * screens that use them.
 */
@Singleton
class UserPreferencesRepository @Inject constructor(
    @param:SettingsDataStore private val dataStore: DataStore<Preferences>,
) {
    private object PreferencesKeys {
        val REPEAT_MODE = intPreferencesKey("repeat_mode")
        val IS_SHUFFLE_ON = booleanPreferencesKey("is_shuffle_on")
        val PERSISTENT_SHUFFLE_ENABLED = booleanPreferencesKey("persistent_shuffle_enabled")
        val AUTOPLAY_ON_SKIP = booleanPreferencesKey("autoplay_on_skip")
        val PREVIOUS_BUTTON_BEHAVIOR = stringPreferencesKey("previous_button_behavior")
        val PLAYBACK_SPEED = floatPreferencesKey("playback_speed")
        val SHOW_QUEUE_HISTORY = booleanPreferencesKey("show_queue_history")
        val SONGS_SORT_OPTION = stringPreferencesKey("songs_sort_option")
        val ALBUMS_SORT_OPTION = stringPreferencesKey("albums_sort_option")
        val ARTISTS_SORT_OPTION = stringPreferencesKey("artists_sort_option")
        val LAST_LIBRARY_TAB_INDEX =
                intPreferencesKey("last_library_tab_index")
        val IS_ALBUMS_LIST_VIEW = booleanPreferencesKey("is_albums_list_view")
    }

    suspend fun setRepeatMode(@Player.RepeatMode mode: Int) {
        dataStore.edit { preferences -> preferences[PreferencesKeys.REPEAT_MODE] = mode }
    }

    suspend fun setShuffleOn(on: Boolean) {
        dataStore.edit { preferences -> preferences[PreferencesKeys.IS_SHUFFLE_ON] = on }
    }

    val persistentShuffleEnabledFlow: Flow<Boolean> =
            dataStore.data.map { preferences ->
                preferences[PreferencesKeys.PERSISTENT_SHUFFLE_ENABLED] ?: false
            }

    val autoplayOnSkipFlow: Flow<Boolean> =
        dataStore.data.map { preferences ->
            preferences[PreferencesKeys.AUTOPLAY_ON_SKIP] ?: true
        }

    val previousButtonBehaviorFlow: Flow<PreviousButtonBehavior> =
        dataStore.data.map { preferences ->
            PreviousButtonBehavior.fromStorageKey(preferences[PreferencesKeys.PREVIOUS_BUTTON_BEHAVIOR])
        }

    /**
     * Normalized to hundredths on read and write, so legacy noisy values such as 0.74999994f,
     * NaN or backup imports outside the range come back as a clean, supported speed.
     */
    val playbackSpeedFlow: Flow<Float> =
            dataStore.data.map { preferences ->
                PlaybackSpeedUtils.normalize(
                    preferences[PreferencesKeys.PLAYBACK_SPEED] ?: PlaybackSpeedUtils.DEFAULT_SPEED
                )
            }.distinctUntilChanged()

    suspend fun setPlaybackSpeed(speed: Float) {
        dataStore.edit { preferences ->
            preferences[PreferencesKeys.PLAYBACK_SPEED] = PlaybackSpeedUtils.normalize(speed)
        }
    }

    val showQueueHistoryFlow: Flow<Boolean> =
            dataStore.data.map { preferences ->
                preferences[PreferencesKeys.SHOW_QUEUE_HISTORY] ?: false
            }

    val isAlbumsListViewFlow: Flow<Boolean> =
        dataStore.data.map { preferences ->
            preferences[PreferencesKeys.IS_ALBUMS_LIST_VIEW] ?: false
        }

    suspend fun setAlbumsListView(isList: Boolean) {
        dataStore.edit { preferences ->
            preferences[PreferencesKeys.IS_ALBUMS_LIST_VIEW] = isList
        }
    }

    val songsSortOptionFlow: Flow<String> =
            dataStore.data.map { preferences ->
                SortOption.fromStorageKey(
                                preferences[PreferencesKeys.SONGS_SORT_OPTION],
                                SortOption.SONGS,
                                SortOption.SongTitleAZ
                        )
                        .storageKey
            }

    val albumsSortOptionFlow: Flow<String> =
            dataStore.data.map { preferences ->
                SortOption.fromStorageKey(
                                preferences[PreferencesKeys.ALBUMS_SORT_OPTION],
                                SortOption.ALBUMS,
                                SortOption.AlbumTitleAZ
                        )
                        .storageKey
            }

    val artistsSortOptionFlow: Flow<String> =
            dataStore.data.map { preferences ->
                SortOption.fromStorageKey(
                                preferences[PreferencesKeys.ARTISTS_SORT_OPTION],
                                SortOption.ARTISTS,
                                SortOption.ArtistNameAZ
                        )
                        .storageKey
            }

    // Sonora: the original also sets songs_sort_option_migrated_v2, a one-off migration of its
    // older sort keys; Sonora has no older keys.
    suspend fun setSongsSortOption(optionKey: String) {
        dataStore.edit { preferences ->
            preferences[PreferencesKeys.SONGS_SORT_OPTION] = optionKey
        }
    }

    suspend fun setAlbumsSortOption(optionKey: String) {
        dataStore.edit { preferences ->
            preferences[PreferencesKeys.ALBUMS_SORT_OPTION] = optionKey
        }
    }

    suspend fun setArtistsSortOption(optionKey: String) {
        dataStore.edit { preferences ->
            preferences[PreferencesKeys.ARTISTS_SORT_OPTION] = optionKey
        }
    }

    val lastLibraryTabIndexFlow: Flow<Int> =
            dataStore.data.map { preferences ->
                preferences[PreferencesKeys.LAST_LIBRARY_TAB_INDEX] ?: 0
            }

    suspend fun saveLastLibraryTabIndex(tabIndex: Int) {
        dataStore.edit { preferences ->
            preferences[PreferencesKeys.LAST_LIBRARY_TAB_INDEX] = tabIndex
        }
    }
}
