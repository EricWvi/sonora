package top.onlyquant.sonora.data.database

import androidx.room.Database
import androidx.room.RoomDatabase

/**
 * As the original's PixelPlayerDatabase, with only the library tables for now. The library is a
 * copy of the server's, so until the app is released a schema change may drop it and sync again
 * (fallbackToDestructiveMigration in AppModule); after the release, add migrations as the
 * original does.
 */
@Database(
    entities = [
        SongEntity::class,
        AlbumEntity::class,
        ArtistEntity::class,
        SongArtistCrossRef::class,
        LibraryStateEntity::class,
    ],
    version = 2,
    exportSchema = true
)
abstract class SonoraDatabase : RoomDatabase() {
    abstract fun musicDao(): MusicDao
}
