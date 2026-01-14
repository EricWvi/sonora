import { useState, useEffect, useRef } from "react";
import { Search, X } from "lucide-react";
import { cn } from "@/lib/utils";

const i18nText = {
  placeholder: "Search for songs, albums, or artists...",
  clear: "Clear search",
};

interface SearchBarProps {
  value: string;
  onChange: (value: string) => void;
  onFocus?: () => void;
  onClear?: () => void;
  className?: string;
}

export default function SearchBar({
  value,
  onChange,
  onFocus,
  onClear,
  className,
}: SearchBarProps) {
  const [localValue, setLocalValue] = useState(value);
  const inputRef = useRef<HTMLInputElement>(null);

  // Debounce the onChange callback
  useEffect(() => {
    const timer = setTimeout(() => {
      onChange(localValue);
    }, 300);

    return () => clearTimeout(timer);
  }, [localValue, onChange]);

  // Sync with external value changes
  useEffect(() => {
    setLocalValue(value);
  }, [value]);

  // Keyboard shortcut: / to focus
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "/" && !e.ctrlKey && !e.metaKey && !e.altKey) {
        const target = e.target as HTMLElement;
        // Only trigger if not already in an input
        if (target.tagName !== "INPUT" && target.tagName !== "TEXTAREA") {
          e.preventDefault();
          inputRef.current?.focus();
        }
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

  const handleClear = () => {
    setLocalValue("");
    onChange("");
    onClear?.();
    inputRef.current?.focus();
  };

  return (
    <div
      className={cn(
        "relative flex items-center gap-3 rounded-2xl border border-gray-200/50 bg-white/80 px-4 py-3 shadow-lg backdrop-blur-sm transition-all duration-200 focus-within:border-purple-500/50 focus-within:shadow-xl dark:border-gray-800/50 dark:bg-[#2A2A2A]/80 dark:focus-within:border-purple-500/50",
        className
      )}
    >
      {/* Search Icon */}
      <Search className="h-5 w-5 flex-shrink-0 text-gray-400 dark:text-gray-500" />

      {/* Input */}
      <input
        ref={inputRef}
        type="text"
        value={localValue}
        onChange={(e) => setLocalValue(e.target.value)}
        onFocus={onFocus}
        placeholder={i18nText.placeholder}
        className="flex-1 bg-transparent text-gray-900 placeholder-gray-400 outline-none dark:text-white dark:placeholder-gray-500"
      />

      {/* Clear Button */}
      {localValue && (
        <button
          onClick={handleClear}
          className="flex-shrink-0 rounded-full p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-gray-700 dark:hover:text-gray-300"
          aria-label={i18nText.clear}
        >
          <X className="h-4 w-4" />
        </button>
      )}

      {/* Keyboard Shortcut Hint */}
      {!localValue && (
        <kbd className="hidden flex-shrink-0 rounded border border-gray-300 bg-gray-100 px-2 py-1 text-xs font-semibold text-gray-600 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-400 md:block">
          /
        </kbd>
      )}
    </div>
  );
}
