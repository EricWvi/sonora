import { useState, useMemo } from "react";
import { useTracks, useCreateTrack, useUpdateTrack, useDeleteTrack, getLyric, createLyric, updateLyric, type Track } from "@/hooks/use-tracks";
import { fileUpload } from "@/lib/fileUpload";
import { formatMediaUrl } from "@/lib/utils";

export function TrackAdmin() {
  const [showForm, setShowForm] = useState(false);
  const [editingTrack, setEditingTrack] = useState<Track | null>(null);
  const [searchInput, setSearchInput] = useState("");
  const [searchTerm, setSearchTerm] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 10;
  const [newTrack, setNewTrack] = useState({
    name: "",
    singer: "",
    cover: "",
    url: "",
    lyric: 0,
    year: new Date().getFullYear(),
  });
  const [lyricContent, setLyricContent] = useState("");
  const [originalLyricContent, setOriginalLyricContent] = useState("");
  const [coverUploadProgress, setCoverUploadProgress] = useState(0);
  const [isCoverUploading, setIsCoverUploading] = useState(false);
  const [audioUploadProgress, setAudioUploadProgress] = useState(0);
  const [isAudioUploading, setIsAudioUploading] = useState(false);

  const { data: tracks, isLoading, error } = useTracks(0); // Only single tracks
  const createTrack = useCreateTrack();
  const updateTrack = useUpdateTrack();
  const deleteTrack = useDeleteTrack();

  // Filter and paginate tracks
  const filteredTracks = useMemo(() => {
    if (!tracks) return [];
    return tracks.filter(track =>
      track.name.toLowerCase().includes(searchTerm.toLowerCase())
    );
  }, [tracks, searchTerm]);

  const totalPages = Math.ceil(filteredTracks.length / itemsPerPage);
  const paginatedTracks = useMemo(() => {
    const startIndex = (currentPage - 1) * itemsPerPage;
    return filteredTracks.slice(startIndex, startIndex + itemsPerPage);
  }, [filteredTracks, currentPage, itemsPerPage]);

  // Apply search when Enter is pressed
  const handleSearchSubmit = () => {
    setSearchTerm(searchInput);
    setCurrentPage(1);
  };

  const handleSearchKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSearchSubmit();
    }
  };

  const handleCoverUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
    setIsCoverUploading(true);
    setCoverUploadProgress(0);

    await fileUpload({
      event,
      onProgress: (progress) => setCoverUploadProgress(progress),
      onSuccess: (response) => {
        try {
          const parsed = JSON.parse(response);
          const imageUrl = parsed.photos?.[0] || '';
          setNewTrack({ ...newTrack, cover: imageUrl });
        } catch {
          setNewTrack({ ...newTrack, cover: response });
        }
        setIsCoverUploading(false);
        setCoverUploadProgress(0);
      }
    });
  };

  const handleAudioUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
    setIsAudioUploading(true);
    setAudioUploadProgress(0);

    await fileUpload({
      event,
      onProgress: (progress) => setAudioUploadProgress(progress),
      onSuccess: (response) => {
        try {
          const parsed = JSON.parse(response);
          const audioUrl = parsed.photos?.[0] || '';
          setNewTrack({ ...newTrack, url: audioUrl });
        } catch {
          setNewTrack({ ...newTrack, url: response });
        }
        setIsAudioUploading(false);
        setAudioUploadProgress(0);
      }
    });
  };

  const handleSubmit = async (e: React.FormEvent) => {
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
        album: 0, // Always 0 for single tracks
        trackNumber: 0, // Always 0 for single tracks
        duration: editingTrack?.duration || 0, // Keep existing duration or 0 for new tracks
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
        cover: "",
        url: "",
        lyric: 0,
        year: new Date().getFullYear(),
      });
      setLyricContent("");
      setOriginalLyricContent("");
      setCoverUploadProgress(0);
      setIsCoverUploading(false);
      setAudioUploadProgress(0);
      setIsAudioUploading(false);
      setShowForm(false);
      // Clear file inputs
      const coverInput = document.querySelector('input[type="file"][accept="image/*"]') as HTMLInputElement;
      const audioInput = document.querySelector('input[type="file"][accept="audio/*"]') as HTMLInputElement;
      if (coverInput) coverInput.value = '';
      if (audioInput) audioInput.value = '';
    } catch (error) {
      console.error("Failed to save track:", error);
    }
  };

  const handleEdit = async (track: Track) => {
    setEditingTrack(track);
    setNewTrack({
      name: track.name,
      singer: track.singer,
      cover: track.cover,
      url: track.url,
      lyric: track.lyric,
      year: track.year,
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

    setShowForm(true);
  };

  const handleCancelEdit = () => {
    setEditingTrack(null);
    setNewTrack({
      name: "",
      singer: "",
      cover: "",
      url: "",
      lyric: 0,
      year: new Date().getFullYear(),
    });
    setLyricContent("");
    setOriginalLyricContent("");
    setCoverUploadProgress(0);
    setIsCoverUploading(false);
    setAudioUploadProgress(0);
    setIsAudioUploading(false);
    setShowForm(false);
    // Clear file inputs
    const coverInput = document.querySelector('input[type="file"][accept="image/*"]') as HTMLInputElement;
    const audioInput = document.querySelector('input[type="file"][accept="audio/*"]') as HTMLInputElement;
    if (coverInput) coverInput.value = '';
    if (audioInput) audioInput.value = '';
  };

  const handleDelete = async (id: number, name: string) => {
    if (!confirm(`Are you sure you want to delete track "${name}"?`)) return;

    try {
      await deleteTrack.mutateAsync({ id, album: 0 });
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
        Loading tracks...
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-8 text-red-600 dark:text-red-400">
        Error loading tracks: {error.message}
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-3">
        <h2 className="text-base font-medium">Single Tracks</h2>
        <button
          onClick={() => editingTrack ? handleCancelEdit() : setShowForm(!showForm)}
          className="px-2 py-1 text-xs bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 border border-gray-300 dark:border-gray-600 rounded"
        >
          {showForm ? "Cancel" : "+ Add"}
        </button>
      </div>

      {showForm && (
        <div className="mb-3 p-3 bg-gray-50 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded">
          <form onSubmit={handleSubmit} className="grid grid-cols-1 md:grid-cols-4 gap-3">
            <div>
              <label className="block text-xs font-medium mb-1">Track Name</label>
              <input
                type="text"
                value={newTrack.name}
                onChange={(e) => setNewTrack({ ...newTrack, name: e.target.value })}
                className="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
                placeholder="Track name"
                required
              />
            </div>
            <div>
              <label className="block text-xs font-medium mb-1">Singer</label>
              <input
                type="text"
                value={newTrack.singer}
                onChange={(e) => setNewTrack({ ...newTrack, singer: e.target.value })}
                className="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
                placeholder="Singer"
                required
              />
            </div>
            <div>
              <label className="block text-xs font-medium mb-1">Year</label>
              <input
                type="number"
                value={newTrack.year}
                onChange={(e) => setNewTrack({ ...newTrack, year: parseInt(e.target.value) || new Date().getFullYear() })}
                className="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
                placeholder={new Date().getFullYear().toString()}
                min="1900"
                max="2099"
              />
            </div>
            <div>
              <label className="block text-xs font-medium mb-1">Cover Image</label>
              <div className="space-y-2">
                <input
                  type="file"
                  accept="image/*"
                  onChange={handleCoverUpload}
                  disabled={isCoverUploading}
                  className="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 disabled:opacity-50"
                />
                {isCoverUploading && (
                  <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                    <div
                      className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                      style={{ width: `${coverUploadProgress}%` }}
                    ></div>
                  </div>
                )}
                {newTrack.cover && !isCoverUploading && (
                  <div className="flex items-center space-x-2">
                    <img
                      src={formatMediaUrl(newTrack.cover)}
                      alt="Cover preview"
                      className="w-8 h-8 rounded object-cover"
                    />
                    <span className="text-xs text-green-600 dark:text-green-400">Image uploaded</span>
                  </div>
                )}
              </div>
            </div>
            <div className="md:col-span-2">
              <label className="block text-xs font-medium mb-1">Audio File</label>
              <div className="space-y-2">
                <input
                  type="file"
                  accept="audio/*"
                  onChange={handleAudioUpload}
                  disabled={isAudioUploading}
                  className="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 disabled:opacity-50"
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
                    <span className="text-xs text-green-600 dark:text-green-400">Audio uploaded</span>
                  </div>
                )}
              </div>
            </div>
            <div className="md:col-span-2">
              <label className="block text-xs font-medium mb-1">Lyrics</label>
              <textarea
                value={lyricContent}
                onChange={(e) => setLyricContent(e.target.value)}
                className="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
                placeholder="Lyrics"
                rows={3}
              />
            </div>
            <div className="md:col-span-4 flex justify-start">
              <button
                type="submit"
                disabled={createTrack.isPending || updateTrack.isPending}
                className="px-3 py-1 text-xs bg-blue-600 hover:bg-blue-700 text-white rounded disabled:opacity-50"
              >
                {editingTrack
                  ? (updateTrack.isPending ? "Updating..." : "Update")
                  : (createTrack.isPending ? "Creating..." : "Create")
                }
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
          placeholder="Search single tracks... (Press Enter)"
          className="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
        />
      </div>

      <div className="space-y-1">
        {filteredTracks.length === 0 ? (
          <div className="text-center py-4 text-gray-500 dark:text-gray-400 text-sm">
            {searchTerm ? "No single tracks found." : "No single tracks found. Add one above."}
          </div>
        ) : (
          paginatedTracks.map((track: Track) => (
            <div
              key={track.id}
              className="flex items-center justify-between py-2 border-b border-dotted border-gray-300 dark:border-gray-600"
            >
              <div className="flex items-center space-x-2">
                {track.cover && (
                  <img
                    src={formatMediaUrl(track.cover)}
                    alt={track.name}
                    className="w-8 h-8 rounded object-cover"
                  />
                )}
                <div>
                  <div className="font-medium text-sm">{track.name}</div>
                  <div className="text-xs text-gray-500 dark:text-gray-400">
                    {track.singer} • {track.year} • {formatDuration(track.duration)}
                  </div>
                </div>
              </div>
              <div className="flex space-x-2 text-xs">
                <button
                  onClick={() => handleEdit(track)}
                  className="text-blue-600 dark:text-blue-400 hover:underline"
                >
                  Edit
                </button>
                <button
                  onClick={() => handleDelete(track.id, track.name)}
                  disabled={deleteTrack.isPending}
                  className="text-red-600 dark:text-red-400 hover:underline disabled:opacity-50"
                >
                  {deleteTrack.isPending ? "Deleting..." : "Delete"}
                </button>
              </div>
            </div>
          ))
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex justify-center items-center space-x-2 mt-3">
            <button
              onClick={() => setCurrentPage(prev => Math.max(prev - 1, 1))}
              disabled={currentPage === 1}
              className="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded disabled:opacity-50 hover:bg-gray-50 dark:hover:bg-gray-700"
            >
              Prev
            </button>

            <span className="text-xs text-gray-600 dark:text-gray-400">
              {currentPage}/{totalPages} ({filteredTracks.length} tracks)
            </span>

            <button
              onClick={() => setCurrentPage(prev => Math.min(prev + 1, totalPages))}
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