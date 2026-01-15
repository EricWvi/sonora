import { useState } from "react";
import Layout from "./components/player/Layout";
import SongList from "./components/player/SongList";
import AlbumGrid from "./components/player/AlbumGrid";
import AlbumDetail from "./components/player/AlbumDetail";
import SingerList from "./components/player/SingerList";
import SingerDetail from "./components/player/SingerDetail";
import SearchBar from "./components/player/SearchBar";
import SearchResults from "./components/player/SearchResults";
import { useAlbums } from "./hooks/player/use-albums";
import { useTracks } from "./hooks/player/use-tracks";
import { useSingers } from "./hooks/player/use-singers";
import { useKeyboardShortcuts } from "./hooks/player/use-keyboard-shortcuts";
import { Music, Disc, Users } from "lucide-react";

const i18nText = {
  welcome: "Welcome to Sonora",
  yourLibrary: "Your Music Library",
  tracks: "Tracks",
  albums: "Albums",
  artists: "Artists",
  loading: "Loading...",
  viewAllTracks: "View All Tracks",
  viewAllAlbums: "View All Albums",
  viewAllArtists: "View All Artists",
  backToHome: "Back to Home",
  searchResults: "Search Results",
};

type View =
  | "home"
  | "songs"
  | "albums"
  | "album-detail"
  | "singers"
  | "singer-detail"
  | "search";

function App() {
  const [currentView, setCurrentView] = useState<View>("home");
  const [selectedAlbumId, setSelectedAlbumId] = useState<number | null>(null);
  const [selectedSingerId, setSelectedSingerId] = useState<number | null>(null);
  const [searchQuery, setSearchQuery] = useState("");

  // Enable keyboard shortcuts
  useKeyboardShortcuts();

  const { data: albums = [], isLoading: albumsLoading } = useAlbums();
  const { data: tracks = [], isLoading: tracksLoading } = useTracks();
  const { data: singers = [], isLoading: singersLoading } = useSingers();

  const stats = [
    {
      icon: Music,
      label: i18nText.tracks,
      count: tracks.length,
      color: "bg-blue-500",
      loading: tracksLoading,
      onClick: () => setCurrentView("songs"),
    },
    {
      icon: Disc,
      label: i18nText.albums,
      count: albums.length,
      color: "bg-purple-500",
      loading: albumsLoading,
      onClick: () => setCurrentView("albums"),
    },
    {
      icon: Users,
      label: i18nText.artists,
      count: singers.length,
      color: "bg-pink-500",
      loading: singersLoading,
      onClick: () => setCurrentView("singers"),
    },
  ];

  const handleSelectAlbum = (albumId: number) => {
    setSelectedAlbumId(albumId);
    setCurrentView("album-detail");
  };

  const handleSelectSinger = (singerId: number) => {
    setSelectedSingerId(singerId);
    setCurrentView("singer-detail");
  };

  const handleBackToAlbums = () => {
    setCurrentView("albums");
    setSelectedAlbumId(null);
  };

  const handleBackToSingers = () => {
    setCurrentView("singers");
    setSelectedSingerId(null);
  };

  const handleBackToHome = () => {
    setCurrentView("home");
    setSelectedAlbumId(null);
    setSelectedSingerId(null);
    setSearchQuery("");
  };

  const handleNavigate = (path: string) => {
    // Clear search when navigating
    setSearchQuery("");

    switch (path) {
      case "/":
        handleBackToHome();
        break;
      case "/tracks":
        setCurrentView("songs");
        break;
      case "/albums":
        setCurrentView("albums");
        break;
      case "/artists":
        setCurrentView("singers");
        break;
      default:
        break;
    }
  };

  const handleSearchChange = (value: string) => {
    setSearchQuery(value);
    if (value.trim()) {
      setCurrentView("search");
    }
  };

  const handleSearchClear = () => {
    setSearchQuery("");
    setCurrentView("home");
  };

  return (
    <Layout currentView={currentView} onNavigate={handleNavigate}>
      <div className="space-y-10">
        {/* Search Bar - Always visible except on home */}
        {currentView !== "home" && (
          <SearchBar
            value={searchQuery}
            onChange={handleSearchChange}
            onClear={handleSearchClear}
          />
        )}

        {/* Home View */}
        {currentView === "home" && (
          <>
            {/* Search Bar on Home */}
            <div className="animate-in fade-in slide-in-from-top-4 duration-500">
              <SearchBar
                value={searchQuery}
                onChange={handleSearchChange}
                onClear={handleSearchClear}
              />
            </div>

            {/* Welcome Section */}
            <div
              className="animate-in fade-in slide-in-from-top-4 duration-700"
              style={{ animationDelay: "100ms" }}
            >
              <h1 className="bg-gradient-to-r from-purple-600 via-pink-600 to-blue-600 bg-clip-text text-5xl font-black tracking-tight text-transparent">
                {i18nText.welcome}
              </h1>
              <p className="mt-3 text-lg text-gray-600 dark:text-gray-400">
                {i18nText.yourLibrary}
              </p>
            </div>

            {/* Stats Grid */}
            <div className="grid gap-6 md:grid-cols-3">
              {stats.map((stat, index) => {
                const Icon = stat.icon;
                const gradients = [
                  "from-blue-500 to-cyan-500",
                  "from-purple-500 to-pink-500",
                  "from-pink-500 to-rose-500",
                ];
                return (
                  <button
                    key={stat.label}
                    onClick={stat.onClick}
                    className="group animate-in fade-in slide-in-from-bottom-4 rounded-2xl border border-gray-200/50 bg-white/80 p-6 text-left shadow-xl backdrop-blur-sm transition-all duration-300 hover:scale-105 hover:shadow-2xl dark:border-gray-800/50 dark:bg-[#2A2A2A]/80"
                    style={{ animationDelay: `${(index + 2) * 100}ms` }}
                  >
                    <div className="flex items-center gap-5">
                      <div
                        className={`rounded-2xl bg-gradient-to-br ${gradients[index]} p-4 shadow-lg transition-transform duration-300 group-hover:scale-110`}
                      >
                        <Icon className="h-7 w-7 text-white" />
                      </div>
                      <div>
                        <p className="text-sm font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">
                          {stat.label}
                        </p>
                        <p className="mt-1 text-3xl font-black text-gray-900 dark:text-white">
                          {stat.loading ? "..." : stat.count}
                        </p>
                      </div>
                    </div>
                  </button>
                );
              })}
            </div>

            {/* Quick Actions */}
            <div className="grid gap-4 sm:grid-cols-3">
              <button
                onClick={() => setCurrentView("songs")}
                className="group rounded-2xl border border-gray-200/50 bg-gradient-to-r from-blue-50 to-cyan-50 p-6 text-left shadow-xl backdrop-blur-sm transition-all duration-300 hover:scale-105 hover:shadow-2xl dark:border-gray-800/50 dark:from-blue-900/20 dark:to-cyan-900/20"
              >
                <Music className="mb-3 h-8 w-8 text-blue-600 dark:text-blue-400" />
                <h3 className="text-xl font-bold text-gray-900 dark:text-white">
                  {i18nText.viewAllTracks}
                </h3>
              </button>
              <button
                onClick={() => setCurrentView("albums")}
                className="group rounded-2xl border border-gray-200/50 bg-gradient-to-r from-purple-50 to-pink-50 p-6 text-left shadow-xl backdrop-blur-sm transition-all duration-300 hover:scale-105 hover:shadow-2xl dark:border-gray-800/50 dark:from-purple-900/20 dark:to-pink-900/20"
              >
                <Disc className="mb-3 h-8 w-8 text-purple-600 dark:text-purple-400" />
                <h3 className="text-xl font-bold text-gray-900 dark:text-white">
                  {i18nText.viewAllAlbums}
                </h3>
              </button>
              <button
                onClick={() => setCurrentView("singers")}
                className="group rounded-2xl border border-gray-200/50 bg-gradient-to-r from-pink-50 to-rose-50 p-6 text-left shadow-xl backdrop-blur-sm transition-all duration-300 hover:scale-105 hover:shadow-2xl dark:border-gray-800/50 dark:from-pink-900/20 dark:to-rose-900/20"
              >
                <Users className="mb-3 h-8 w-8 text-pink-600 dark:text-pink-400" />
                <h3 className="text-xl font-bold text-gray-900 dark:text-white">
                  {i18nText.viewAllArtists}
                </h3>
              </button>
            </div>
          </>
        )}

        {/* Song List View */}
        {currentView === "songs" && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <h1 className="bg-gradient-to-r from-purple-600 via-pink-600 to-blue-600 bg-clip-text text-4xl font-black tracking-tight text-transparent">
                {i18nText.tracks}
              </h1>
              <button
                onClick={handleBackToHome}
                className="rounded-full px-6 py-2 text-sm font-semibold text-gray-600 transition-colors hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
              >
                {i18nText.backToHome}
              </button>
            </div>
            <SongList />
          </div>
        )}

        {/* Album Grid View */}
        {currentView === "albums" && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <h1 className="bg-gradient-to-r from-purple-600 via-pink-600 to-blue-600 bg-clip-text text-4xl font-black tracking-tight text-transparent">
                {i18nText.albums}
              </h1>
              <button
                onClick={handleBackToHome}
                className="rounded-full px-6 py-2 text-sm font-semibold text-gray-600 transition-colors hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
              >
                {i18nText.backToHome}
              </button>
            </div>
            <AlbumGrid onSelectAlbum={handleSelectAlbum} />
          </div>
        )}

        {/* Album Detail View */}
        {currentView === "album-detail" && selectedAlbumId !== null && (
          <AlbumDetail albumId={selectedAlbumId} onBack={handleBackToAlbums} />
        )}

        {/* Singer List View */}
        {currentView === "singers" && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <h1 className="bg-gradient-to-r from-purple-600 via-pink-600 to-blue-600 bg-clip-text text-4xl font-black tracking-tight text-transparent">
                {i18nText.artists}
              </h1>
              <button
                onClick={handleBackToHome}
                className="rounded-full px-6 py-2 text-sm font-semibold text-gray-600 transition-colors hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
              >
                {i18nText.backToHome}
              </button>
            </div>
            <SingerList onSelectSinger={handleSelectSinger} />
          </div>
        )}

        {/* Singer Detail View */}
        {currentView === "singer-detail" && selectedSingerId !== null && (
          <SingerDetail
            singerId={selectedSingerId}
            onBack={handleBackToSingers}
            onSelectAlbum={handleSelectAlbum}
          />
        )}

        {/* Search Results View */}
        {currentView === "search" && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <h1 className="bg-gradient-to-r from-purple-600 via-pink-600 to-blue-600 bg-clip-text text-4xl font-black tracking-tight text-transparent">
                {i18nText.searchResults}
              </h1>
              <button
                onClick={handleBackToHome}
                className="rounded-full px-6 py-2 text-sm font-semibold text-gray-600 transition-colors hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
              >
                {i18nText.backToHome}
              </button>
            </div>
            <SearchResults
              query={searchQuery}
              onSelectSinger={handleSelectSinger}
              onSelectAlbum={handleSelectAlbum}
            />
          </div>
        )}
      </div>
    </Layout>
  );
}

export default App;
