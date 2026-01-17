import Dexie, { type EntityTable } from "dexie";
import { getRequest } from "./queryClient";

// Define types based on the backend models
export interface Album {
  id: number;
  name: string;
  cover: string;
  year: number;
}

export interface Singer {
  id: number;
  name: string;
  avatar: string;
}

export interface Track {
  id: number;
  name: string;
  singer: string;
  album: number;
  cover: string;
  url: string;
  lyric: number;
  duration: number;
  year: number;
  trackNumber: number;
  genre: string;
  albumText: string;
}

export interface Lyric {
  id: number;
  content: string;
}

export interface SyncMetadata {
  id: number;
  lastSyncTimestamp: number;
  createdAt: number;
}

// Define the database schema
class SonoraDB extends Dexie {
  albums!: EntityTable<Album, "id">;
  singers!: EntityTable<Singer, "id">;
  tracks!: EntityTable<Track, "id">;
  lyrics!: EntityTable<Lyric, "id">;
  syncMetadata!: EntityTable<SyncMetadata, "id">;

  constructor() {
    super("SonoraDB");
    this.version(1).stores({
      albums: "id, name, year",
      singers: "id, name",
      tracks: "id, name, album, singer, genre, year",
      lyrics: "id",
      syncMetadata: "id",
    });
  }
}

export const db = new SonoraDB();

const SYNC_INTERVAL_MS = 28 * 24 * 60 * 60 * 1000; // 28 days

interface FullSyncResponse {
  albums: Album[];
  singers: Singer[];
  tracks: Track[];
  lyrics: Lyric[];
  timestamp: number;
}

interface UpdatesResponse {
  entries: Array<{
    tableName: string;
    stale: number[];
    deleted: number[];
  }>;
  timestamp: number;
}

interface ChangeLogItem {
  tableName: string;
  stale: number[];
  deleted: number[];
}

// Sync manager class
class SyncManager {
  private isSyncing = false;
  private syncPromise: Promise<void> | null = null;

  async initialize(): Promise<void> {
    if (this.syncPromise) {
      return this.syncPromise;
    }

    this.syncPromise = this._performSync();
    return this.syncPromise;
  }

  private async _performSync(): Promise<void> {
    if (this.isSyncing) return;

    this.isSyncing = true;

    try {
      const metadata = await db.syncMetadata.get(1);
      const now = Date.now();
      const shouldFullSync =
        !metadata || now - metadata.lastSyncTimestamp > SYNC_INTERVAL_MS;

      if (shouldFullSync) {
        await this.performFullSync();
      } else {
        try {
          await this.performIncrementalSync(metadata.lastSyncTimestamp);
        } catch (error) {
          // Incremental sync failures should not block initialization
          console.warn("Incremental sync failed, will retry later:", error);
        }
      }
    } catch (error) {
      console.error("Sync failed:", error);
      throw error;
    } finally {
      this.isSyncing = false;
      this.syncPromise = null;
    }
  }

  private async performFullSync(): Promise<void> {
    console.log("Performing full sync...");

    const response = (await getRequest(
      "/api/sync?Action=GetFullSync",
    )) as FullSyncResponse;
    const { albums, singers, tracks, lyrics, timestamp } = response;

    await db.transaction(
      "rw",
      [db.albums, db.singers, db.tracks, db.lyrics, db.syncMetadata],
      async () => {
        // Clear existing data
        await db.albums.clear();
        await db.singers.clear();
        await db.tracks.clear();
        await db.lyrics.clear();

        // Insert new data
        await db.albums.bulkAdd(albums);
        await db.singers.bulkAdd(singers);
        await db.tracks.bulkAdd(tracks);
        await db.lyrics.bulkAdd(lyrics);

        // Update sync metadata
        await db.syncMetadata.put({
          id: 1,
          lastSyncTimestamp: timestamp,
          createdAt: Date.now(),
        });
      },
    );

    console.log("Full sync completed:", {
      albums: albums.length,
      singers: singers.length,
      tracks: tracks.length,
      lyrics: lyrics.length,
    });
  }

  private async performIncrementalSync(since: number): Promise<void> {
    console.log("Performing incremental sync since:", since);

    // do not retry on failure
    const response = (await getRequest(
      `/api/sync?Action=GetUpdates&since=${since}`,
      1,
    )) as UpdatesResponse;
    const { entries, timestamp } = response;

    if (entries.length === 0) {
      console.log("No updates to sync");
      await db.syncMetadata.put({
        id: 1,
        lastSyncTimestamp: timestamp,
        createdAt: Date.now(),
      });
      return;
    }

    // Fetch all updated records BEFORE starting the transaction
    const updatedRecords = await this.fetchUpdatedRecords(entries);

    // Now perform the transaction with all the data we need
    await db.transaction(
      "rw",
      [db.albums, db.singers, db.tracks, db.lyrics, db.syncMetadata],
      async () => {
        // Apply deletions and updates
        await this.applyChanges(updatedRecords);

        // Update sync metadata
        await db.syncMetadata.put({
          id: 1,
          lastSyncTimestamp: timestamp,
          createdAt: Date.now(),
        });
      },
    );

    console.log(
      "Incremental sync completed:",
      entries.length,
      "table(s) changed",
    );
  }

  private async fetchUpdatedRecords(entries: ChangeLogItem[]): Promise<{
    albums: { toDelete: number[]; toUpdate: Album[] };
    singers: { toDelete: number[]; toUpdate: Singer[] };
    tracks: { toDelete: number[]; toUpdate: Track[] };
    lyrics: { toDelete: number[]; toUpdate: Lyric[] };
  }> {
    const result = {
      albums: { toDelete: [] as number[], toUpdate: [] as Album[] },
      singers: { toDelete: [] as number[], toUpdate: [] as Singer[] },
      tracks: { toDelete: [] as number[], toUpdate: [] as Track[] },
      lyrics: { toDelete: [] as number[], toUpdate: [] as Lyric[] },
    };

    const apiEndpointMap: Record<string, string> = {
      d_album: "album",
      d_singer: "singer",
      d_track: "track",
      d_lyric: "track",
    };

    for (const entry of entries) {
      const { tableName, stale, deleted } = entry;
      const apiEndpoint = apiEndpointMap[tableName];

      if (!apiEndpoint) {
        console.warn(`Unknown table: ${tableName}`);
        continue;
      }

      // Collect deletions
      if (tableName === "d_album") {
        result.albums.toDelete.push(...deleted);
      } else if (tableName === "d_singer") {
        result.singers.toDelete.push(...deleted);
      } else if (tableName === "d_track") {
        result.tracks.toDelete.push(...deleted);
      } else if (tableName === "d_lyric") {
        result.lyrics.toDelete.push(...deleted);
      }

      // Fetch updated records
      for (const recordId of stale) {
        try {
          if (tableName === "d_lyric") {
            const response = await getRequest(
              `/api/${apiEndpoint}?Action=GetLyric&id=${recordId}`,
            );
            result.lyrics.toUpdate.push({
              id: recordId,
              content: response.lyric,
            });
          } else if (tableName === "d_album") {
            const response = await getRequest(
              `/api/${apiEndpoint}?Action=GetAlbum&id=${recordId}`,
            );
            if (response.album) {
              result.albums.toUpdate.push(response.album);
            }
          } else if (tableName === "d_singer") {
            const response = await getRequest(
              `/api/${apiEndpoint}?Action=GetSinger&id=${recordId}`,
            );
            if (response.singer) {
              result.singers.toUpdate.push(response.singer);
            }
          } else if (tableName === "d_track") {
            const response = await getRequest(
              `/api/${apiEndpoint}?Action=GetTrack&id=${recordId}`,
            );
            if (response.track) {
              result.tracks.toUpdate.push(response.track);
            }
          }
        } catch (error) {
          console.error(
            `Failed to fetch ${tableName} record ${recordId}:`,
            error,
          );
        }
      }
    }

    return result;
  }

  private async applyChanges(records: {
    albums: { toDelete: number[]; toUpdate: Album[] };
    singers: { toDelete: number[]; toUpdate: Singer[] };
    tracks: { toDelete: number[]; toUpdate: Track[] };
    lyrics: { toDelete: number[]; toUpdate: Lyric[] };
  }): Promise<void> {
    // Apply deletions
    for (const id of records.albums.toDelete) {
      await db.albums.delete(id);
    }
    for (const id of records.singers.toDelete) {
      await db.singers.delete(id);
    }
    for (const id of records.tracks.toDelete) {
      await db.tracks.delete(id);
    }
    for (const id of records.lyrics.toDelete) {
      await db.lyrics.delete(id);
    }

    // Apply updates
    if (records.albums.toUpdate.length > 0) {
      await db.albums.bulkPut(records.albums.toUpdate);
    }
    if (records.singers.toUpdate.length > 0) {
      await db.singers.bulkPut(records.singers.toUpdate);
    }
    if (records.tracks.toUpdate.length > 0) {
      await db.tracks.bulkPut(records.tracks.toUpdate);
    }
    if (records.lyrics.toUpdate.length > 0) {
      await db.lyrics.bulkPut(records.lyrics.toUpdate);
    }
  }

  async forceSync(): Promise<void> {
    await this._performSync();
  }
}

export const syncManager = new SyncManager();

// Query functions for use in hooks
export const dbClient = {
  // Albums
  async getAllAlbums(): Promise<Album[]> {
    return await db.albums.toArray();
  },

  async getAlbum(id: number): Promise<Album | undefined> {
    return await db.albums.get(id);
  },

  async searchAlbums(query: string): Promise<Album[]> {
    const lowerQuery = query.toLowerCase();
    return await db.albums
      .filter((album) => album.name.toLowerCase().includes(lowerQuery))
      .toArray();
  },

  // Singers
  async getAllSingers(): Promise<Singer[]> {
    return await db.singers.toArray();
  },

  async getSinger(id: number): Promise<Singer | undefined> {
    return await db.singers.get(id);
  },

  async searchSingers(query: string): Promise<Singer[]> {
    const lowerQuery = query.toLowerCase();
    return await db.singers
      .filter((singer) => singer.name.toLowerCase().includes(lowerQuery))
      .toArray();
  },

  // Tracks
  async getAllTracks(): Promise<Track[]> {
    return await db.tracks.toArray();
  },

  async getTrack(id: number): Promise<Track | undefined> {
    return await db.tracks.get(id);
  },

  async getTracksByAlbum(albumId: number): Promise<Track[]> {
    return await db.tracks.where("album").equals(albumId).toArray();
  },

  async searchTracks(query: string): Promise<Track[]> {
    const lowerQuery = query.toLowerCase();
    return await db.tracks
      .filter(
        (track) =>
          track.name.toLowerCase().includes(lowerQuery) ||
          track.singer.toLowerCase().includes(lowerQuery),
      )
      .toArray();
  },

  async getSingles(): Promise<Track[]> {
    return await db.tracks.where("album").equals(0).toArray();
  },

  // Lyrics
  async getLyric(id: number): Promise<Lyric | undefined> {
    return await db.lyrics.get(id);
  },
};
