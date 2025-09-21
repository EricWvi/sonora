import { useState } from "react";
import { SingerAdmin } from "./components/admin/singer";
import { AlbumAdmin } from "./components/admin/album";
import { TrackAdmin } from "./components/admin/track";

function AdminBoard() {
  const [activeSection, setActiveSection] = useState("singers");

  return (
    <div className="min-h-screen bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100">
      <div className="max-w-4xl mx-auto px-6 py-8">
        <header className="mb-8 pb-6 border-b border-dotted border-gray-300 dark:border-gray-600">
          <h1 className="text-2xl font-medium mb-4">Sonora Admin</h1>
          <nav>
            <ul className="flex space-x-6 text-sm">
              <li>
                <button
                  onClick={() => setActiveSection("singers")}
                  className={`hover:text-blue-600 dark:hover:text-blue-400 ${
                    activeSection === "singers" ? "text-blue-600 dark:text-blue-400" : "text-gray-600 dark:text-gray-400"
                  }`}
                >
                  Singers
                </button>
              </li>
              <li>
                <button
                  onClick={() => setActiveSection("albums")}
                  className={`hover:text-blue-600 dark:hover:text-blue-400 ${
                    activeSection === "albums" ? "text-blue-600 dark:text-blue-400" : "text-gray-600 dark:text-gray-400"
                  }`}
                >
                  Albums
                </button>
              </li>
              <li>
                <button
                  onClick={() => setActiveSection("tracks")}
                  className={`hover:text-blue-600 dark:hover:text-blue-400 ${
                    activeSection === "tracks" ? "text-blue-600 dark:text-blue-400" : "text-gray-600 dark:text-gray-400"
                  }`}
                >
                  Tracks
                </button>
              </li>
            </ul>
          </nav>
        </header>

        <main>
          {activeSection === "singers" && <SingerAdmin />}
          {activeSection === "albums" && <AlbumAdmin />}
          {activeSection === "tracks" && <TrackAdmin />}
        </main>
      </div>
    </div>
  );
}


export default AdminBoard;
