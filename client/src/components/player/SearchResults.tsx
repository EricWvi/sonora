import { useSearchSingers, useSinger } from "@/hooks/player/use-singers";
import { useSearchAlbums, useAlbum } from "@/hooks/player/use-albums";
import { useSearchTracks, useTrack } from "@/hooks/player/use-tracks";
import { formatMediaUrl } from "@/lib/utils";
import { Users, Disc, Music, Clock } from "lucide-react";

const i18nText = {
  singers: "Artists",
  albums: "Albums",
  tracks: "Songs",
  noResults: "No results found",
  searchPrompt: "Start typing to search...",
  viewAll: "View all",
};

function formatDuration(seconds: number): string {
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins}:${secs.toString().padStart(2, "0")}`;
}

interface SingerResultProps {
  singerId: number;
  onClick: (singerId: number) => void;
}

function SingerResult({ singerId, onClick }: SingerResultProps) {
  const { data: singer } = useSinger(singerId);
  if (!singer) return null;

  return (
    <div
      onClick={() => onClick(singerId)}
      className="group flex cursor-pointer items-center gap-4 rounded-lg border border-transparent px-4 py-3 transition-all duration-200 hover:border-gray-200/50 hover:bg-gradient-to-r hover:from-pink-50/50 hover:to-purple-50/50 hover:shadow-md dark:hover:border-gray-700/50 dark:hover:from-pink-900/10 dark:hover:to-purple-900/10"
    >
      <div className="h-12 w-12 flex-shrink-0 overflow-hidden rounded-full bg-gradient-to-br from-pink-500 to-purple-500">
        {singer.avatar ? (
          <img
            src={formatMediaUrl(singer.avatar)}
            alt={singer.name}
            className="h-full w-full object-cover"
          />
        ) : (
          <div className="flex h-full w-full items-center justify-center">
            <Users className="h-6 w-6 text-white opacity-70" />
          </div>
        )}
      </div>
      <div className="min-w-0 flex-1">
        <p className="truncate font-semibold text-gray-900 dark:text-white">
          {singer.name}
        </p>
        <p className="text-sm text-gray-600 dark:text-gray-400">Artist</p>
      </div>
    </div>
  );
}

interface AlbumResultProps {
  albumId: number;
  onClick: (albumId: number) => void;
}

function AlbumResult({ albumId, onClick }: AlbumResultProps) {
  const { data: album } = useAlbum(albumId);
  if (!album) return null;

  return (
    <div
      onClick={() => onClick(albumId)}
      className="group flex cursor-pointer items-center gap-4 rounded-lg border border-transparent px-4 py-3 transition-all duration-200 hover:border-gray-200/50 hover:bg-gradient-to-r hover:from-purple-50/50 hover:to-pink-50/50 hover:shadow-md dark:hover:border-gray-700/50 dark:hover:from-purple-900/10 dark:hover:to-pink-900/10"
    >
      <div className="h-12 w-12 flex-shrink-0 overflow-hidden rounded-lg bg-gradient-to-br from-purple-500 to-pink-500">
        {album.cover ? (
          <img
            src={formatMediaUrl(album.cover)}
            alt={album.name}
            className="h-full w-full object-cover"
          />
        ) : (
          <div className="flex h-full w-full items-center justify-center">
            <Disc className="h-6 w-6 text-white opacity-70" />
          </div>
        )}
      </div>
      <div className="min-w-0 flex-1">
        <p className="truncate font-semibold text-gray-900 dark:text-white">
          {album.name}
        </p>
        <p className="text-sm text-gray-600 dark:text-gray-400">{album.year}</p>
      </div>
    </div>
  );
}

interface TrackResultProps {
  trackId: number;
  onClick: (trackId: number) => void;
}

function TrackResult({ trackId, onClick }: TrackResultProps) {
  const { data: track } = useTrack(trackId);
  if (!track) return null;

  return (
    <div
      onClick={() => onClick(trackId)}
      className="group flex cursor-pointer items-center gap-4 rounded-lg border border-transparent px-4 py-3 transition-all duration-200 hover:border-gray-200/50 hover:bg-gradient-to-r hover:from-blue-50/50 hover:to-cyan-50/50 hover:shadow-md dark:hover:border-gray-700/50 dark:hover:from-blue-900/10 dark:hover:to-cyan-900/10"
    >
      <div className="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-cyan-500">
        <Music className="h-6 w-6 text-white" />
      </div>
      <div className="min-w-0 flex-1">
        <p className="truncate font-semibold text-gray-900 dark:text-white">
          {track.name}
        </p>
        <p className="truncate text-sm text-gray-600 dark:text-gray-400">
          {track.singer}
        </p>
      </div>
      <div className="flex items-center gap-1 text-sm text-gray-500 dark:text-gray-400">
        <Clock className="h-3 w-3" />
        {formatDuration(track.duration)}
      </div>
    </div>
  );
}

interface SearchResultsProps {
  query: string;
  onSelectSinger?: (singerId: number) => void;
  onSelectAlbum?: (albumId: number) => void;
  onSelectTrack?: (trackId: number) => void;
}

export default function SearchResults({
  query,
  onSelectSinger,
  onSelectAlbum,
  onSelectTrack,
}: SearchResultsProps) {
  const { data: singerIds = [] } = useSearchSingers(query);
  const { data: albumIds = [] } = useSearchAlbums(query);
  const { data: trackIds = [] } = useSearchTracks(query);

  const hasResults =
    singerIds.length > 0 || albumIds.length > 0 || trackIds.length > 0;

  if (!query.trim()) {
    return (
      <div className="flex items-center justify-center py-24">
        <div className="text-center">
          <Music className="mx-auto h-16 w-16 text-gray-300 dark:text-gray-600" />
          <p className="mt-4 text-lg text-gray-500 dark:text-gray-400">
            {i18nText.searchPrompt}
          </p>
        </div>
      </div>
    );
  }

  if (!hasResults) {
    return (
      <div className="flex items-center justify-center py-24">
        <div className="text-center">
          <Music className="mx-auto h-16 w-16 text-gray-300 dark:text-gray-600" />
          <p className="mt-4 text-lg font-semibold text-gray-700 dark:text-gray-300">
            {i18nText.noResults}
          </p>
          <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
            Try searching with different keywords
          </p>
        </div>
      </div>
    );
  }

  const maxResults = 5;

  return (
    <div className="animate-in fade-in slide-in-from-bottom-4 space-y-8">
      {/* Artists Section */}
      {singerIds.length > 0 && (
        <div>
          <h2 className="mb-4 text-xl font-bold text-gray-900 dark:text-white">
            {i18nText.singers}
          </h2>
          <div className="space-y-2">
            {singerIds.slice(0, maxResults).map((id) => (
              <SingerResult
                key={id}
                singerId={id}
                onClick={(id) => onSelectSinger?.(id)}
              />
            ))}
          </div>
          {singerIds.length > maxResults && (
            <p className="mt-3 text-sm text-gray-500 dark:text-gray-400">
              +{singerIds.length - maxResults} more artists
            </p>
          )}
        </div>
      )}

      {/* Albums Section */}
      {albumIds.length > 0 && (
        <div>
          <h2 className="mb-4 text-xl font-bold text-gray-900 dark:text-white">
            {i18nText.albums}
          </h2>
          <div className="space-y-2">
            {albumIds.slice(0, maxResults).map((id) => (
              <AlbumResult
                key={id}
                albumId={id}
                onClick={(id) => onSelectAlbum?.(id)}
              />
            ))}
          </div>
          {albumIds.length > maxResults && (
            <p className="mt-3 text-sm text-gray-500 dark:text-gray-400">
              +{albumIds.length - maxResults} more albums
            </p>
          )}
        </div>
      )}

      {/* Tracks Section */}
      {trackIds.length > 0 && (
        <div>
          <h2 className="mb-4 text-xl font-bold text-gray-900 dark:text-white">
            {i18nText.tracks}
          </h2>
          <div className="space-y-2">
            {trackIds.slice(0, maxResults).map((id) => (
              <TrackResult
                key={id}
                trackId={id}
                onClick={(id) => onSelectTrack?.(id)}
              />
            ))}
          </div>
          {trackIds.length > maxResults && (
            <p className="mt-3 text-sm text-gray-500 dark:text-gray-400">
              +{trackIds.length - maxResults} more tracks
            </p>
          )}
        </div>
      )}
    </div>
  );
}
