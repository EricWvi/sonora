import { useState } from "react";
import { Home, Music, Disc, Users, Menu, X } from "lucide-react";
import { cn } from "@/lib/utils";
import SearchBar from "./SearchBar";
import { AppleMusicIcon, SearchIcon } from "./icons";

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

interface SidebarProps {
  currentView?: string;
  onNavigate?: (path: string) => void;
  searchQuery?: string;
  onSearchChange?: (value: string) => void;
  onSearchClear?: () => void;
}

export default function Sidebar({
  currentView,
  onNavigate,
  searchQuery = "",
  onSearchChange,
  onSearchClear,
}: SidebarProps) {
  const [isOpen, setIsOpen] = useState(false);

  const toggleSidebar = () => setIsOpen(!isOpen);

  return (
    <>
      {/* Mobile Menu Button */}
      <button
        onClick={toggleSidebar}
        className="fixed top-4 left-4 z-50 rounded-full bg-white/90 p-3 shadow-lg backdrop-blur-md transition-all hover:bg-white hover:shadow-xl md:hidden dark:bg-black/40 dark:hover:bg-black/60"
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
          "fixed inset-y-0 left-0 z-40 w-65 transform border-r border-[rgba(0,0,0,.15)] bg-[#f9f9f9] backdrop-blur-xl transition-transform duration-300 ease-in-out md:relative md:translate-x-0 dark:border-[hsla(0,0%,100%,.1)] dark:bg-[#282727]/95",
          isOpen ? "translate-x-0" : "-translate-x-full",
        )}
      >
        <div className="flex h-full flex-col">
          {/* Logo/Brand */}
          <div className="h-[55px] px-[30px] pt-[17px]">
            <div className="h-5 w-fit text-[#1e1e1e] dark:text-[#eeeeee]">
              <AppleMusicIcon />
            </div>
          </div>

          {/* Search Bar */}
          <div className="mx-[25px] mb-4 flex h-14 flex-col justify-center">
            <div className="relative">
              <div className="text-sidebar-icon absolute inset-y-[10px] start-[10px] size-3">
                <SearchIcon />
              </div>
              <input className="focus-glow border-searchbar-border bg-background text-primary h-8 w-full rounded-sm border pe-[5px] pt-[6px] pb-[5px] pl-7 text-xs leading-[1.25] font-normal tracking-normal focus:shadow-[0_0_0_4px_rgb(var(--keyColor))] focus:outline-none" />
            </div>
          </div>

          {/* <div className="px-6 pb-4">
            <SearchBar
              value={searchQuery}
              onChange={(value) => onSearchChange?.(value)}
              onClear={onSearchClear}
              className="shadow-md"
            />
          </div> */}

          {/* Navigation */}
          <nav className="flex-1 space-y-2 p-6">
            {navItems.map((item) => {
              const Icon = item.icon;
              const viewMap: Record<string, string> = {
                "/": "home",
                "/tracks": "songs",
                "/albums": "albums",
                "/artists": "singers",
              };
              const isActive = currentView === viewMap[item.href];

              return (
                <a
                  key={item.href}
                  href={item.href}
                  onClick={(e) => {
                    e.preventDefault();
                    onNavigate?.(item.href);
                    setIsOpen(false);
                  }}
                  className={cn(
                    "group flex items-center gap-4 rounded-xl px-4 py-3.5 text-base font-semibold transition-all duration-200",
                    isActive
                      ? "bg-gradient-to-r from-purple-600 to-pink-600 text-white shadow-lg shadow-purple-500/50"
                      : "text-gray-700 hover:bg-gray-100/80 dark:text-gray-300 dark:hover:bg-white/5",
                  )}
                >
                  <Icon
                    className={cn(
                      "h-5 w-5 transition-transform duration-200",
                      isActive ? "" : "group-hover:scale-110",
                    )}
                  />
                  <span>{item.label}</span>
                </a>
              );
            })}
          </nav>
        </div>
      </aside>
    </>
  );
}
