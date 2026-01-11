import Layout from "./components/player/Layout";
import { useAlbums } from "./hooks/player/use-albums";
import { useTracks } from "./hooks/player/use-tracks";
import { useSingers } from "./hooks/player/use-singers";
import { Music, Disc, Users } from "lucide-react";

const i18nText = {
  welcome: "Welcome to Sonora",
  yourLibrary: "Your Music Library",
  tracks: "Tracks",
  albums: "Albums",
  artists: "Artists",
  loading: "Loading...",
};

function App() {
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
    },
    {
      icon: Disc,
      label: i18nText.albums,
      count: albums.length,
      color: "bg-purple-500",
      loading: albumsLoading,
    },
    {
      icon: Users,
      label: i18nText.artists,
      count: singers.length,
      color: "bg-pink-500",
      loading: singersLoading,
    },
  ];

  return (
    <Layout>
      <div className="space-y-10">
        {/* Welcome Section */}
        <div className="animate-in fade-in slide-in-from-top-4 duration-700">
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
              <div
                key={stat.label}
                className="group animate-in fade-in slide-in-from-bottom-4 rounded-2xl border border-gray-200/50 bg-white/80 p-6 shadow-xl backdrop-blur-sm transition-all duration-300 hover:scale-105 hover:shadow-2xl dark:border-gray-800/50 dark:bg-[#2A2A2A]/80"
                style={{ animationDelay: `${index * 100}ms` }}
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
              </div>
            );
          })}
        </div>

        {/* Recent Content - Placeholder for future implementation */}
        <div
          className="animate-in fade-in slide-in-from-bottom-4 rounded-2xl border border-gray-200/50 bg-white/80 p-8 shadow-xl backdrop-blur-sm dark:border-gray-800/50 dark:bg-[#2A2A2A]/80"
          style={{ animationDelay: "300ms" }}
        >
          <h2 className="mb-6 text-2xl font-bold text-gray-900 dark:text-white">
            Recent Tracks
          </h2>
          <div className="flex items-center justify-center py-12">
            <p className="text-gray-400 dark:text-gray-500">
              Track list will be displayed here...
            </p>
          </div>
        </div>
      </div>
    </Layout>
  );
}

export default App;
