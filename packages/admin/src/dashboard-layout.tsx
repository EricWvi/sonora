import type { ReactNode } from "react";
import { Button } from "@heroui/react";
import { FolderOpen, Layers, ScrollText, Settings } from "lucide-react";

interface DashboardLayoutProps {
  title: string;
  actions?: ReactNode;
  children: ReactNode;
}

// App-level chrome: a persistent sidebar (brand mark + primary navigation) and
// a content column whose top bar hosts the active view's title and actions.
// Navigation is data-driven so wiring a future route is a one-line entry; until
// routes exist, non-Files items render as disabled placeholders so the sidebar
// shape stays stable as more views land.
export function DashboardLayout({ title, actions, children }: DashboardLayoutProps) {
  return (
    <div className="flex min-h-screen bg-background text-foreground">
      <Sidebar />
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex h-16 shrink-0 items-center justify-between gap-4 border-b border-default bg-surface px-6">
          <h1 className="text-lg font-semibold text-foreground">{title}</h1>
          {actions ? (
            <div className="flex items-center gap-2">{actions}</div>
          ) : null}
        </header>
        <main className="flex-1 overflow-auto">{children}</main>
      </div>
    </div>
  );
}

interface NavItem {
  id: string;
  label: string;
  icon: ReactNode;
  isActive?: boolean;
  isDisabled?: boolean;
}

// Files is the only live view today; Logs and Settings are listed (disabled) so
// the information architecture is visible before the routes ship.
const NAV_ITEMS: NavItem[] = [
  { id: "files", label: "Files", icon: <FolderOpen className="size-4" />, isActive: true },
  { id: "logs", label: "Logs", icon: <ScrollText className="size-4" />, isDisabled: true },
  { id: "settings", label: "Settings", icon: <Settings className="size-4" />, isDisabled: true },
];

function Sidebar() {
  return (
    <aside className="hidden w-60 shrink-0 flex-col gap-6 border-r border-default bg-surface-secondary p-4 md:flex">
      <div className="flex items-center gap-2 px-2 pt-2">
        <div className="flex size-7 items-center justify-center rounded-lg bg-accent text-accent-foreground">
          <Layers className="size-4" />
        </div>
        <span className="text-sm font-semibold tracking-wider text-foreground">
          SONORA
        </span>
      </div>
      <nav aria-label="Primary" className="flex flex-col gap-1">
        {NAV_ITEMS.map((item) => (
          <Button
            key={item.id}
            variant={item.isActive ? "secondary" : "ghost"}
            fullWidth
            isDisabled={item.isDisabled}
            className="justify-start"
          >
            <span className="flex items-center gap-2">
              {item.icon}
              {item.label}
            </span>
          </Button>
        ))}
      </nav>
    </aside>
  );
}
