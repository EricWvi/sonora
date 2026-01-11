import { useState, useMemo } from "react";
import {
  useAlbums,
  useCreateAlbum,
  useUpdateAlbum,
  useDeleteAlbum,
  type Album,
} from "@/hooks/admin/use-albums";
import {
  useTracks,
  useCreateTrack,
  useUpdateTrack,
  useDeleteTrack,
  getLyric,
  createLyric,
  updateLyric,
  type Track,
} from "@/hooks/admin/use-tracks";
import { fileUpload } from "@/lib/fileUpload";
import { formatMediaUrl } from "@/lib/utils";
import { usePlayer } from "@/hooks/admin/use-miniplayer";

function TrackItemWithPlay({
  track,
  onEditTrack,
  onDeleteTrack,
  albumId,
  formatDuration,
}: {
  track: Track;
  onEditTrack: (track: Track) => void;
  onDeleteTrack: (trackId: number, trackName: string, albumId: number) => void;
  albumId: number;
  formatDuration: (seconds: number) => string;
}) {
  const { play } = usePlayer();

  const handlePlay = () => {
    play(track);
  };

  return (
    <div className="flex items-center justify-between py-2 px-2 border border-gray-200 dark:border-gray-600 rounded bg-white dark:bg-gray-800">
      <div className="flex items-center space-x-2">
        <span className="text-xs text-gray-500 dark:text-gray-400 w-6">
          #{track.trackNumber}
        </span>
        <div>
          <div className="font-medium text-sm">{track.name}</div>
          <div className="text-xs text-gray-500 dark:text-gray-400">
            {track.singer} • {formatDuration(track.duration)}
          </div>
        </div>
      </div>
      <div className="flex space-x-2 text-xs">
        <button
          onClick={handlePlay}
          className="text-green-600 dark:text-green-400 hover:underline"
        >
          Play
        </button>
        <button
          onClick={() => onEditTrack(track)}
          className="text-blue-600 dark:text-blue-400 hover:underline"
        >
          Edit
        </button>
        <button
          onClick={() => onDeleteTrack(track.id, track.name, albumId)}
          className="text-red-600 dark:text-red-400 hover:underline"
        >
          Delete
        </button>
      </div>
    </div>
  );
}

function AlbumWithTracks({
  album,
  isExpanded,
  onToggleExpand,
  showTrackForm,
  newTrack,
  lyricContent,
  setLyricContent,
  editingTrack,
  onEdit,
  onDelete,
  onShowTrackForm,
  onTrackSubmit,
  onEditTrack,
  onCancelTrackEdit,
  onDeleteTrack,
  setNewTrack,
  formatDuration,
  createTrack,
  updateTrack,
  deleteAlbum,
  handleAudioUpload,
  audioUploadProgress,
  isAudioUploading,
}: {
  album: Album;
  isExpanded: boolean;
  onToggleExpand: () => void;
  showTrackForm: boolean;
  newTrack: any;
  lyricContent: string;
  setLyricContent: (content: string) => void;
  editingTrack: Track | null;
  onEdit: (album: Album) => void;
  onDelete: (id: number, name: string) => void;
  onShowTrackForm: (albumId: number, tracks: Track[] | undefined) => void;
  onTrackSubmit: (
    e: React.FormEvent,
    albumId: number,
    albumData: Album
  ) => void;
  onEditTrack: (track: Track) => void;
  onCancelTrackEdit: () => void;
  onDeleteTrack: (trackId: number, trackName: string, albumId: number) => void;
  setNewTrack: (track: any) => void;
  formatDuration: (seconds: number) => string;
  createTrack: any;
  updateTrack: any;
  deleteAlbum: any;
  handleAudioUpload: (event: React.ChangeEvent<HTMLInputElement>) => void;
  audioUploadProgress: number;
  isAudioUploading: boolean;
}) {
  // Only fetch tracks when album is expanded
  const { data: tracks } = useTracks(album.id, isExpanded);

  return (
    <div className="border border-gray-300 dark:border-gray-600 rounded mb-2">
      {/* Album header */}
      <div className="flex items-center justify-between py-2 px-3 bg-gray-50 dark:bg-gray-700">
        <div className="flex items-center space-x-2">
          {album.cover && (
            <img
              src={formatMediaUrl(album.cover)}
              alt={album.name}
              className="w-8 h-8 rounded object-cover"
            />
          )}
          <div>
            <div className="font-medium text-sm">{album.name}</div>
            <div className="text-xs text-gray-500 dark:text-gray-400">
              {album.year} • ID: {album.id}
              {isExpanded && tracks ? ` • ${tracks.length} tracks` : ""}
            </div>
          </div>
        </div>
        <div className="flex space-x-2 text-xs">
          <button
            onClick={onToggleExpand}
            className="text-blue-600 dark:text-blue-400 hover:underline"
          >
            {isExpanded ? "Hide" : "Show"}
          </button>
          <button
            onClick={() => onEdit(album)}
            className="text-blue-600 dark:text-blue-400 hover:underline"
          >
            Edit
          </button>
          <button
            onClick={() => onDelete(album.id, album.name)}
            disabled={deleteAlbum.isPending}
            className="text-red-600 dark:text-red-400 hover:underline disabled:opacity-50"
          >
            {deleteAlbum.isPending ? "Deleting..." : "Delete"}
          </button>
        </div>
      </div>

      {/* Tracks section */}
      {isExpanded && (
        <div className="p-3 border-t border-gray-300 dark:border-gray-600">
          <div className="flex items-center justify-between mb-2">
            <h4 className="text-sm font-medium">Album Tracks</h4>
            <button
              onClick={() =>
                showTrackForm
                  ? onCancelTrackEdit()
                  : onShowTrackForm(album.id, tracks)
              }
              className="px-2 py-1 text-xs bg-green-600 hover:bg-green-700 text-white rounded"
            >
              {showTrackForm ? "Cancel" : "+ Add Track"}
            </button>
          </div>

          {/* Track form */}
          {showTrackForm && (
            <div className="mb-3 p-3 bg-gray-100 dark:bg-gray-600 rounded">
              <form
                onSubmit={(e) => onTrackSubmit(e, album.id, album)}
                className="grid grid-cols-1 md:grid-cols-3 gap-3"
              >
                <div>
                  <label className="block text-xs font-medium mb-1">
                    Track Name
                  </label>
                  <input
                    type="text"
                    value={newTrack.name}
                    onChange={(e) =>
                      setNewTrack({ ...newTrack, name: e.target.value })
                    }
                    className="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800"
                    placeholder="Track name"
                    required
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1">
                    Singer
                  </label>
                  <input
                    type="text"
                    value={newTrack.singer}
                    onChange={(e) =>
                      setNewTrack({ ...newTrack, singer: e.target.value })
                    }
                    className="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800"
                    placeholder="Singer"
                    required
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium mb-1">
                    Track Number
                  </label>
                  <input
                    type="number"
                    value={newTrack.trackNumber}
                    onChange={(e) =>
                      setNewTrack({
                        ...newTrack,
                        trackNumber: parseInt(e.target.value) || 1,
                      })
                    }
                    className="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800"
                    min="1"
                    required
                  />
                </div>
                <div className="md:col-span-2">
                  <label className="block text-xs font-medium mb-1">
                    Audio File
                  </label>
                  <div className="space-y-2">
                    <input
                      type="file"
                      accept="audio/*"
                      onChange={handleAudioUpload}
                      disabled={isAudioUploading}
                      className="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 disabled:opacity-50"
                    />
                    {isAudioUploading && (
                      <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                        <div
                          className="bg-green-600 h-2 rounded-full transition-all duration-300"
                          style={{ width: `${audioUploadProgress}%` }}
                        ></div>
                      </div>
                    )}
                    {newTrack.url && !isAudioUploading && (
                      <div className="flex items-center space-x-2">
                        <span className="text-xs text-green-600 dark:text-green-400">
                          Audio uploaded
                        </span>
                      </div>
                    )}
                  </div>
                </div>
                <div className="md:col-span-1">
                  <label className="block text-xs font-medium mb-1">
                    Lyrics
                  </label>
                  <textarea
                    value={lyricContent}
                    onChange={(e) => setLyricContent(e.target.value)}
                    className="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800"
                    placeholder="Lyrics"
                    rows={2}
                  />
                </div>
                <div className="md:col-span-3 flex justify-start">
                  <button
                    type="submit"
                    disabled={createTrack.isPending || updateTrack.isPending}
                    className="px-3 py-1 text-xs bg-blue-600 hover:bg-blue-700 text-white rounded disabled:opacity-50"
                  >
                    {editingTrack
                      ? updateTrack.isPending
                        ? "Updating..."
                        : "Update Track"
                      : createTrack.isPending
                        ? "Creating..."
                        : "Create Track"}
                  </button>
                </div>
              </form>
            </div>
          )}

          {/* Track list */}
          <div className="space-y-1">
            {!tracks || tracks.length === 0 ? (
              <div className="text-center py-4 text-gray-500 dark:text-gray-400 text-sm">
                No tracks in this album. Add one above.
              </div>
            ) : (
              tracks
                .sort((a, b) => a.trackNumber - b.trackNumber)
                .map((track) => (
                  <TrackItemWithPlay
                    key={track.id}
                    track={track}
                    onEditTrack={onEditTrack}
                    onDeleteTrack={onDeleteTrack}
                    albumId={album.id}
                    formatDuration={formatDuration}
                  />
                ))
            )}
          </div>
        </div>
      )}
    </div>
  );
}

export function AlbumAdmin() {
  const [showForm, setShowForm] = useState(false);
  const [editingAlbum, setEditingAlbum] = useState<Album | null>(null);
  const [newAlbum, setNewAlbum] = useState({
    name: "",
    cover: "",
    year: new Date().getFullYear(),
  });
  const [searchInput, setSearchInput] = useState("");
  const [searchTerm, setSearchTerm] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [isUploading, setIsUploading] = useState(false);
  const [audioUploadProgress, setAudioUploadProgress] = useState(0);
  const [isAudioUploading, setIsAudioUploading] = useState(false);
  const [expandedAlbum, setExpandedAlbum] = useState<number | null>(null);
  const [showTrackForm, setShowTrackForm] = useState<number | null>(null);
  const [editingTrack, setEditingTrack] = useState<Track | null>(null);
  const [newTrack, setNewTrack] = useState({
    name: "",
    singer: "",
    url: "",
    lyric: 0,
    trackNumber: 1,
  });
  const [lyricContent, setLyricContent] = useState("");
  const [originalLyricContent, setOriginalLyricContent] = useState("");
  const itemsPerPage = 10;

  const { data: albums, isLoading, error } = useAlbums();
  const createAlbum = useCreateAlbum();
  const updateAlbum = useUpdateAlbum();
  const deleteAlbum = useDeleteAlbum();

  const createTrack = useCreateTrack();
  const updateTrack = useUpdateTrack();
  const deleteTrack = useDeleteTrack();

  // Filter and paginate albums
  const filteredAlbums = useMemo(() => {
    if (!albums) return [];
    return albums.filter((album) =>
      album.name.toLowerCase().includes(searchTerm.toLowerCase())
    );
  }, [albums, searchTerm]);

  const totalPages = Math.ceil(filteredAlbums.length / itemsPerPage);
  const paginatedAlbums = useMemo(() => {
    const startIndex = (currentPage - 1) * itemsPerPage;
    return filteredAlbums.slice(startIndex, startIndex + itemsPerPage);
  }, [filteredAlbums, currentPage, itemsPerPage]);

  // Apply search when Enter is pressed
  const handleSearchSubmit = () => {
    setSearchTerm(searchInput);
    setCurrentPage(1);
  };

  const handleCoverUpload = async (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    setIsUploading(true);
    setUploadProgress(0);

    await fileUpload({
      event,
      onProgress: (progress) => setUploadProgress(progress),
      onSuccess: (response) => {
        try {
          const parsed = JSON.parse(response);
          const imageUrl = parsed.photos?.[0] || "";
          setNewAlbum({ ...newAlbum, cover: imageUrl });
        } catch {
          setNewAlbum({ ...newAlbum, cover: response });
        }
        setIsUploading(false);
        setUploadProgress(0);
      },
    });
  };

  const handleAudioUpload = async (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    setIsAudioUploading(true);
    setAudioUploadProgress(0);

    await fileUpload({
      event,
      onProgress: (progress) => setAudioUploadProgress(progress),
      onSuccess: (response) => {
        try {
          const parsed = JSON.parse(response);
          const audioUrl = parsed.photos?.[0] || "";
          setNewTrack({ ...newTrack, url: audioUrl });
        } catch {
          setNewTrack({ ...newTrack, url: response });
        }
        setIsAudioUploading(false);
        setAudioUploadProgress(0);
      },
    });
  };

  const handleSearchKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      handleSearchSubmit();
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newAlbum.name.trim()) return;

    try {
      if (editingAlbum) {
        await updateAlbum.mutateAsync({ id: editingAlbum.id, ...newAlbum });
        setEditingAlbum(null);
      } else {
        await createAlbum.mutateAsync(newAlbum);
      }
      setNewAlbum({ name: "", cover: "", year: new Date().getFullYear() });
      setUploadProgress(0);
      setIsUploading(false);
      setShowForm(false);
      // Clear file input
      const fileInput = document.querySelector(
        'input[type="file"]'
      ) as HTMLInputElement;
      if (fileInput) fileInput.value = "";
    } catch (error) {
      console.error("Failed to save album:", error);
    }
  };

  const handleEdit = (album: Album) => {
    setEditingAlbum(album);
    setNewAlbum({ name: album.name, cover: album.cover, year: album.year });
    setShowForm(true);
  };

  const handleCancelEdit = () => {
    setEditingAlbum(null);
    setNewAlbum({ name: "", cover: "", year: new Date().getFullYear() });
    setUploadProgress(0);
    setIsUploading(false);
    setShowForm(false);
    // Clear file input
    const fileInput = document.querySelector(
      'input[type="file"]'
    ) as HTMLInputElement;
    if (fileInput) fileInput.value = "";
  };

  const handleDelete = async (id: number, name: string) => {
    if (!confirm(`Are you sure you want to delete album "${name}"?`)) return;

    try {
      await deleteAlbum.mutateAsync({ id });
    } catch (error) {
      console.error("Failed to delete album:", error);
    }
  };

  // Track management functions
  const getNextTrackNumber = (albumId: number, tracks: Track[] | undefined) => {
    if (!tracks) return 1;
    const albumTracks = tracks.filter((t) => t.album === albumId);
    return albumTracks.length > 0
      ? Math.max(...albumTracks.map((t) => t.trackNumber)) + 1
      : 1;
  };

  const handleShowTrackForm = (
    albumId: number,
    tracks: Track[] | undefined
  ) => {
    setShowTrackForm(albumId);
    setNewTrack({
      name: "",
      singer: "",
      url: "",
      lyric: 0,
      trackNumber: getNextTrackNumber(albumId, tracks),
    });
    setLyricContent("");
    setOriginalLyricContent("");
  };

  const handleTrackSubmit = async (
    e: React.FormEvent,
    albumId: number,
    albumData: Album
  ) => {
    e.preventDefault();
    if (!newTrack.name.trim() || !newTrack.singer.trim()) return;

    try {
      let lyricId = newTrack.lyric;

      // Handle lyric creation/update if content exists and changed
      if (lyricContent.trim()) {
        if (editingTrack && editingTrack.lyric > 0) {
          // Only update if content changed
          if (lyricContent !== originalLyricContent) {
            await updateLyric(editingTrack.lyric, lyricContent);
          }
          lyricId = editingTrack.lyric;
        } else {
          // Create new lyric
          const response = await createLyric(lyricContent);
          lyricId = response.id;
        }
      }

      const trackData = {
        ...newTrack,
        lyric: lyricId,
        album: albumId,
        cover: albumData.cover,
        year: albumData.year,
        duration: editingTrack?.duration || 0, // Keep existing duration or 0 for new tracks
        genre: editingTrack?.genre || "",
        albumText: albumData.name,
      };

      if (editingTrack) {
        await updateTrack.mutateAsync({ id: editingTrack.id, ...trackData });
        setEditingTrack(null);
      } else {
        await createTrack.mutateAsync(trackData);
      }

      setNewTrack({
        name: "",
        singer: "",
        url: "",
        lyric: 0,
        trackNumber: 1,
      });
      setLyricContent("");
      setOriginalLyricContent("");
      setAudioUploadProgress(0);
      setIsAudioUploading(false);
      setShowTrackForm(null);
      // Clear audio file input
      const audioInput = document.querySelector(
        'input[type="file"][accept="audio/*"]'
      ) as HTMLInputElement;
      if (audioInput) audioInput.value = "";
    } catch (error) {
      console.error("Failed to save track:", error);
    }
  };

  const handleEditTrack = async (track: Track) => {
    setEditingTrack(track);
    setNewTrack({
      name: track.name,
      singer: track.singer,
      url: track.url,
      lyric: track.lyric,
      trackNumber: track.trackNumber,
    });

    // Fetch lyric content if track has a lyric ID
    if (track.lyric > 0) {
      try {
        const content = await getLyric(track.lyric);
        setLyricContent(content);
        setOriginalLyricContent(content);
      } catch (error) {
        console.error("Failed to fetch lyric:", error);
        setLyricContent("");
        setOriginalLyricContent("");
      }
    } else {
      setLyricContent("");
      setOriginalLyricContent("");
    }

    setShowTrackForm(track.album);
  };

  const handleCancelTrackEdit = () => {
    setEditingTrack(null);
    setNewTrack({
      name: "",
      singer: "",
      url: "",
      lyric: 0,
      trackNumber: 1,
    });
    setLyricContent("");
    setOriginalLyricContent("");
    setAudioUploadProgress(0);
    setIsAudioUploading(false);
    setShowTrackForm(null);
    // Clear audio file input
    const audioInput = document.querySelector(
      'input[type="file"][accept="audio/*"]'
    ) as HTMLInputElement;
    if (audioInput) audioInput.value = "";
  };

  const handleDeleteTrack = async (
    trackId: number,
    trackName: string,
    albumId: number
  ) => {
    if (!confirm(`Are you sure you want to delete track "${trackName}"?`))
      return;

    try {
      await deleteTrack.mutateAsync({ id: trackId, album: albumId });
    } catch (error) {
      console.error("Failed to delete track:", error);
    }
  };

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  };

  if (isLoading) {
    return (
      <div className="text-center py-8 text-gray-500 dark:text-gray-400">
        Loading albums...
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-8 text-red-600 dark:text-red-400">
        Error loading albums: {error.message}
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-3">
        <h2 className="text-base font-medium">Albums</h2>
        <button
          onClick={() => {
            if (showForm) {
              handleCancelEdit();
            } else {
              setShowForm(true);
            }
          }}
          className="px-2 py-1 text-xs bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 border border-gray-300 dark:border-gray-600 rounded"
        >
          {showForm ? "Cancel" : "+ Add"}
        </button>
      </div>

      {showForm && (
        <div className="mb-3 p-3 bg-gray-50 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded">
          <form
            onSubmit={handleSubmit}
            className="grid grid-cols-1 md:grid-cols-4 gap-3"
          >
            <div>
              <label className="block text-xs font-medium mb-1">Name</label>
              <input
                type="text"
                value={newAlbum.name}
                onChange={(e) =>
                  setNewAlbum({ ...newAlbum, name: e.target.value })
                }
                className="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
                placeholder="Album name"
                required
              />
            </div>
            <div>
              <label className="block text-xs font-medium mb-1">
                Cover Image
              </label>
              <div className="space-y-2">
                <input
                  type="file"
                  accept="image/*"
                  onChange={handleCoverUpload}
                  disabled={isUploading}
                  className="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 disabled:opacity-50"
                />
                {isUploading && (
                  <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                    <div
                      className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                      style={{ width: `${uploadProgress}%` }}
                    ></div>
                  </div>
                )}
                {newAlbum.cover && !isUploading && (
                  <div className="flex items-center space-x-2">
                    <img
                      src={formatMediaUrl(newAlbum.cover)}
                      alt="Cover preview"
                      className="w-8 h-8 rounded object-cover"
                    />
                    <span className="text-xs text-green-600 dark:text-green-400">
                      Image uploaded
                    </span>
                  </div>
                )}
              </div>
            </div>
            <div>
              <label className="block text-xs font-medium mb-1">Year</label>
              <input
                type="number"
                value={newAlbum.year}
                onChange={(e) =>
                  setNewAlbum({
                    ...newAlbum,
                    year: parseInt(e.target.value) || new Date().getFullYear(),
                  })
                }
                className="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
                placeholder={new Date().getFullYear().toString()}
                min="1900"
                max="2099"
              />
            </div>
            <div className="flex items-end">
              <button
                type="submit"
                disabled={createAlbum.isPending || updateAlbum.isPending}
                className="px-3 py-1 text-xs bg-blue-600 hover:bg-blue-700 text-white rounded disabled:opacity-50 h-7"
              >
                {editingAlbum
                  ? updateAlbum.isPending
                    ? "Updating..."
                    : "Update"
                  : createAlbum.isPending
                    ? "Creating..."
                    : "Create"}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Search */}
      <div className="mb-3">
        <input
          type="text"
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          onKeyDown={handleSearchKeyDown}
          placeholder="Search albums... (Press Enter)"
          className="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
        />
      </div>

      <div className="space-y-1">
        {filteredAlbums.length === 0 ? (
          <div className="text-center py-4 text-gray-500 dark:text-gray-400 text-sm">
            {searchTerm
              ? "No albums found."
              : "No albums found. Add one above."}
          </div>
        ) : (
          paginatedAlbums.map((album: Album) => (
            <AlbumWithTracks
              key={album.id}
              album={album}
              isExpanded={expandedAlbum === album.id}
              onToggleExpand={() =>
                setExpandedAlbum(expandedAlbum === album.id ? null : album.id)
              }
              showTrackForm={showTrackForm === album.id}
              newTrack={newTrack}
              lyricContent={lyricContent}
              setLyricContent={setLyricContent}
              editingTrack={editingTrack}
              onEdit={handleEdit}
              onDelete={handleDelete}
              onShowTrackForm={handleShowTrackForm}
              onTrackSubmit={handleTrackSubmit}
              onEditTrack={handleEditTrack}
              onCancelTrackEdit={handleCancelTrackEdit}
              onDeleteTrack={handleDeleteTrack}
              setNewTrack={setNewTrack}
              formatDuration={formatDuration}
              createTrack={createTrack}
              updateTrack={updateTrack}
              deleteAlbum={deleteAlbum}
              handleAudioUpload={handleAudioUpload}
              audioUploadProgress={audioUploadProgress}
              isAudioUploading={isAudioUploading}
            />
          ))
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex justify-center items-center space-x-2 mt-3">
            <button
              onClick={() => setCurrentPage((prev) => Math.max(prev - 1, 1))}
              disabled={currentPage === 1}
              className="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded disabled:opacity-50 hover:bg-gray-50 dark:hover:bg-gray-700"
            >
              Prev
            </button>

            <span className="text-xs text-gray-600 dark:text-gray-400">
              {currentPage}/{totalPages} ({filteredAlbums.length} albums)
            </span>

            <button
              onClick={() =>
                setCurrentPage((prev) => Math.min(prev + 1, totalPages))
              }
              disabled={currentPage === totalPages}
              className="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded disabled:opacity-50 hover:bg-gray-50 dark:hover:bg-gray-700"
            >
              Next
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
