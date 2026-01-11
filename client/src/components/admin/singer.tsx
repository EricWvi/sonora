import { useState, useMemo } from "react";
import {
  useSingers,
  useCreateSinger,
  useUpdateSinger,
  useDeleteSinger,
  type Singer,
} from "@/hooks/admin/use-singers";
import { fileUpload } from "@/lib/fileUpload";
import { formatMediaUrl } from "@/lib/utils";

export function SingerAdmin() {
  const [showForm, setShowForm] = useState(false);
  const [editingSinger, setEditingSinger] = useState<Singer | null>(null);
  const [newSinger, setNewSinger] = useState({ name: "", avatar: "" });
  const [searchInput, setSearchInput] = useState("");
  const [searchTerm, setSearchTerm] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [isUploading, setIsUploading] = useState(false);
  const itemsPerPage = 10;

  const { data: singers, isLoading, error } = useSingers();
  const createSinger = useCreateSinger();
  const updateSinger = useUpdateSinger();
  const deleteSinger = useDeleteSinger();

  // Filter and paginate singers
  const filteredSingers = useMemo(() => {
    if (!singers) return [];
    return singers.filter((singer) =>
      singer.name.toLowerCase().includes(searchTerm.toLowerCase())
    );
  }, [singers, searchTerm]);

  const totalPages = Math.ceil(filteredSingers.length / itemsPerPage);
  const paginatedSingers = useMemo(() => {
    const startIndex = (currentPage - 1) * itemsPerPage;
    return filteredSingers.slice(startIndex, startIndex + itemsPerPage);
  }, [filteredSingers, currentPage, itemsPerPage]);

  // Apply search when Enter is pressed
  const handleSearchSubmit = () => {
    setSearchTerm(searchInput);
    setCurrentPage(1);
  };

  const handleAvatarUpload = async (
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
          setNewSinger({ ...newSinger, avatar: imageUrl });
        } catch {
          setNewSinger({ ...newSinger, avatar: response });
        }
        setIsUploading(false);
        setUploadProgress(0);
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
    if (!newSinger.name.trim()) return;

    try {
      if (editingSinger) {
        await updateSinger.mutateAsync({ id: editingSinger.id, ...newSinger });
        setEditingSinger(null);
      } else {
        await createSinger.mutateAsync(newSinger);
      }
      setNewSinger({ name: "", avatar: "" });
      setUploadProgress(0);
      setIsUploading(false);
      setShowForm(false);
      // Clear file input
      const fileInput = document.querySelector(
        'input[type="file"]'
      ) as HTMLInputElement;
      if (fileInput) fileInput.value = "";
    } catch (error) {
      console.error("Failed to save singer:", error);
    }
  };

  const handleEdit = (singer: Singer) => {
    setEditingSinger(singer);
    setNewSinger({ name: singer.name, avatar: singer.avatar });
    setShowForm(true);
  };

  const handleCancelEdit = () => {
    setEditingSinger(null);
    setNewSinger({ name: "", avatar: "" });
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
    if (!confirm(`Are you sure you want to delete singer "${name}"?`)) return;

    try {
      await deleteSinger.mutateAsync({ id });
    } catch (error) {
      console.error("Failed to delete singer:", error);
    }
  };

  if (isLoading) {
    return (
      <div className="text-center py-8 text-gray-500 dark:text-gray-400">
        Loading singers...
      </div>
    );
  }

  if (error) {
    return (
      <div className="text-center py-8 text-red-600 dark:text-red-400">
        Error loading singers: {error.message}
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-3">
        <h2 className="text-base font-medium">Singers</h2>
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
            className="grid grid-cols-1 md:grid-cols-3 gap-3"
          >
            <div>
              <label className="block text-xs font-medium mb-1">Name</label>
              <input
                type="text"
                value={newSinger.name}
                onChange={(e) =>
                  setNewSinger({ ...newSinger, name: e.target.value })
                }
                className="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
                placeholder="Singer name"
                required
              />
            </div>
            <div>
              <label className="block text-xs font-medium mb-1">
                Avatar Image
              </label>
              <div className="space-y-2">
                <input
                  type="file"
                  accept="image/*"
                  onChange={handleAvatarUpload}
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
                {newSinger.avatar && !isUploading && (
                  <div className="flex items-center space-x-2">
                    <img
                      src={formatMediaUrl(newSinger.avatar)}
                      alt="Avatar preview"
                      className="w-8 h-8 rounded object-cover"
                    />
                    <span className="text-xs text-green-600 dark:text-green-400">
                      Image uploaded
                    </span>
                  </div>
                )}
              </div>
            </div>
            <div className="flex items-end">
              <button
                type="submit"
                disabled={createSinger.isPending || updateSinger.isPending}
                className="px-3 py-1 text-xs bg-blue-600 hover:bg-blue-700 text-white rounded disabled:opacity-50 h-7"
              >
                {editingSinger
                  ? updateSinger.isPending
                    ? "Updating..."
                    : "Update"
                  : createSinger.isPending
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
          placeholder="Search singers... (Press Enter)"
          className="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
        />
      </div>

      <div className="space-y-1">
        {filteredSingers.length === 0 ? (
          <div className="text-center py-4 text-gray-500 dark:text-gray-400 text-sm">
            {searchTerm
              ? "No singers found."
              : "No singers found. Add one above."}
          </div>
        ) : (
          paginatedSingers.map((singer: Singer) => (
            <div
              key={singer.id}
              className="flex items-center justify-between py-2 border-b border-dotted border-gray-300 dark:border-gray-600"
            >
              <div className="flex items-center space-x-2">
                {singer.avatar && (
                  <img
                    src={formatMediaUrl(singer.avatar)}
                    alt={singer.name}
                    className="w-8 h-8 rounded object-cover"
                  />
                )}
                <div>
                  <div className="font-medium text-sm">{singer.name}</div>
                  <div className="text-xs text-gray-500 dark:text-gray-400">
                    ID: {singer.id}
                  </div>
                </div>
              </div>
              <div className="flex space-x-2 text-xs">
                <button
                  onClick={() => handleEdit(singer)}
                  className="text-blue-600 dark:text-blue-400 hover:underline"
                >
                  Edit
                </button>
                <button
                  onClick={() => handleDelete(singer.id, singer.name)}
                  disabled={deleteSinger.isPending}
                  className="text-red-600 dark:text-red-400 hover:underline disabled:opacity-50"
                >
                  {deleteSinger.isPending ? "Deleting..." : "Delete"}
                </button>
              </div>
            </div>
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
              {currentPage}/{totalPages} ({filteredSingers.length} singers)
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
