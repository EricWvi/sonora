import { Breadcrumbs } from "@heroui/react";

export interface Crumb {
  id: string;
  name: string;
}

interface NodeBreadcrumbProps {
  // Ordered directory ids from root's first child down to the current folder.
  crumbs: Crumb[];
  // Navigate to a segment by stack index; -1 jumps to root.
  onNavigate: (index: number) => void;
}

// Renders the directory stack as a breadcrumb trail. Navigation is driven by
// the in-memory stack, so no path-string-to-id resolution is needed.
export function NodeBreadcrumb({ crumbs, onNavigate }: NodeBreadcrumbProps) {
  return (
    <Breadcrumbs>
      <Breadcrumbs.Item key="root" onPress={() => onNavigate(-1)}>
        Root
      </Breadcrumbs.Item>
      {crumbs.map((crumb, index) => (
        <Breadcrumbs.Item key={crumb.id} onPress={() => onNavigate(index)}>
          {crumb.name}
        </Breadcrumbs.Item>
      ))}
    </Breadcrumbs>
  );
}
