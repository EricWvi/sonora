import { useState } from "react";
import { Home, Music, Disc, Users, Menu, X } from "lucide-react";
import { cn } from "@/lib/utils";

const i18nText = {
  appName: "Sonora",
  nav: {
    home: "Home",
    tracks: "Tracks",
    albums: "Albums",
    singers: "Artists",
  },
};

const navItems = [
  { icon: Home, label: i18nText.nav.home, href: "/" },
  { icon: Music, label: i18nText.nav.tracks, href: "/tracks" },
  { icon: Disc, label: i18nText.nav.albums, href: "/albums" },
  { icon: Users, label: i18nText.nav.singers, href: "/artists" },
];

export default function Sidebar() {
  const [isOpen, setIsOpen] = useState(false);
  const [activeItem, setActiveItem] = useState("/");

  const toggleSidebar = () => setIsOpen(!isOpen);

  return (
    <>
      {/* Mobile Menu Button */}
      <button
        onClick={toggleSidebar}
        className="fixed left-4 top-4 z-50 rounded-full bg-white/90 p-3 shadow-lg backdrop-blur-md transition-all hover:bg-white hover:shadow-xl dark:bg-black/40 dark:hover:bg-black/60 md:hidden"
        aria-label="Toggle menu"
      >
        {isOpen ? (
          <X className="h-5 w-5 text-gray-900 dark:text-white" />
        ) : (
          <Menu className="h-5 w-5 text-gray-900 dark:text-white" />
        )}
      </button>

      {/* Overlay for mobile */}
      {isOpen && (
        <div
          className="fixed inset-0 z-30 bg-black/50 md:hidden"
          onClick={toggleSidebar}
        />
      )}

      {/* Sidebar */}
      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-40 w-72 transform border-r border-gray-200/50 bg-white/95 backdrop-blur-xl transition-transform duration-300 ease-in-out dark:border-gray-800/50 dark:bg-[#1E1E1E]/95 md:relative md:translate-x-0",
          isOpen ? "translate-x-0" : "-translate-x-full"
        )}
      >
        <div className="flex h-full flex-col">
          {/* Logo/Brand */}
          <div className="border-b border-gray-200/50 px-6 py-8 dark:border-gray-800/50">
            <h1 className="bg-gradient-to-r from-purple-600 via-pink-500 to-blue-600 bg-clip-text text-3xl font-black tracking-tight text-transparent">
              {i18nText.appName}
            </h1>
            <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
              Your music, everywhere
            </p>
          </div>

          {/* Navigation */}
          <nav className="flex-1 space-y-2 p-6">
            {navItems.map((item) => {
              const Icon = item.icon;
              const isActive = activeItem === item.href;

              return (
                <a
                  key={item.href}
                  href={item.href}
                  onClick={(e) => {
                    e.preventDefault();
                    setActiveItem(item.href);
                    setIsOpen(false);
                  }}
                  className={cn(
                    "group flex items-center gap-4 rounded-xl px-4 py-3.5 text-base font-semibold transition-all duration-200",
                    isActive
                      ? "bg-gradient-to-r from-purple-600 to-pink-600 text-white shadow-lg shadow-purple-500/50"
                      : "text-gray-700 hover:bg-gray-100/80 dark:text-gray-300 dark:hover:bg-white/5"
                  )}
                >
                  <Icon
                    className={cn(
                      "h-5 w-5 transition-transform duration-200",
                      isActive ? "" : "group-hover:scale-110"
                    )}
                  />
                  <span>{item.label}</span>
                </a>
              );
            })}
          </nav>

          {/* Footer */}
          <div className="border-t border-gray-200/50 p-6 dark:border-gray-800/50">
            <div className="text-xs text-gray-400 dark:text-gray-500">
              <p className="font-medium">Sonora Music</p>
              <p className="mt-0.5 text-gray-400/70 dark:text-gray-500/70">
                Version 0.1.0
              </p>
            </div>
          </div>
        </div>
      </aside>
    </>
  );
}
