import {
  Button,
  Dropdown,
  Spinner,
  Table,
} from "@heroui/react";
import { File, Folder, MoreHorizontal, RotateCw } from "lucide-react";
import type { NodeView } from "@sonora/contracts";
import { formatBytes, formatTimestamp } from "../../../lib/format";

interface NodeTableProps {
  nodes: NodeView[];
  isLoading: boolean;
  isError: boolean;
  onOpenDir: (node: NodeView) => void;
  onRename: (node: NodeView) => void;
  onDelete: (node: NodeView) => void;
  onRetry: () => void;
}

export function NodeTable({
  nodes,
  isLoading,
  isError,
  onOpenDir,
  onRename,
  onDelete,
  onRetry,
}: NodeTableProps) {
  if (isLoading) {
    return (
      <div className="flex justify-center rounded-xl border border-default bg-surface py-16">
        <Spinner size="lg" />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="flex flex-col items-center gap-3 rounded-xl border border-default bg-surface py-16">
        <p className="text-danger">Failed to load nodes.</p>
        <Button variant="outline" onPress={onRetry}>
          <RotateCw /> Retry
        </Button>
      </div>
    );
  }

  if (nodes.length === 0) {
    return (
      <div className="rounded-xl border border-default bg-surface py-16 text-center text-muted">
        This folder is empty.
      </div>
    );
  }

  return (
    <Table className="overflow-hidden rounded-xl border border-default">
      <Table.ScrollContainer>
        <Table.Content
          aria-label="VFS nodes"
          selectionMode="none"
          onRowAction={(key) => {
            const node = nodes.find((n) => n.id === key);
            if (node && node.kind === "directory") {
              onOpenDir(node);
            }
          }}
        >
          <Table.Header>
            <Table.Column id="name" isRowHeader>
              Name
            </Table.Column>
            <Table.Column id="size" className="text-right">
              Size
            </Table.Column>
            <Table.Column id="updated">Updated</Table.Column>
            <Table.Column id="actions">Actions</Table.Column>
          </Table.Header>
          <Table.Body>
            {nodes.map((node) => (
              <Table.Row key={node.id} id={node.id}>
                <Table.Cell>
                  <div className="flex items-center gap-3">
                    <span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-surface-secondary text-muted">
                      {node.kind === "directory" ? (
                        <Folder className="size-4" />
                      ) : (
                        <File className="size-4" />
                      )}
                    </span>
                    <span className="flex flex-col">
                      <span className="font-medium text-foreground">
                        {node.name}
                      </span>
                      {/* Secondary line folds the old Kind/MIME columns into the
                          name cell: a folder is just "Folder", a file surfaces
                          its MIME type (falling back to "File" when unknown). */}
                      <span className="text-xs text-muted">
                        {node.kind === "directory"
                          ? "Folder"
                          : (node.mimeType ?? "File")}
                      </span>
                    </span>
                  </div>
                </Table.Cell>
                <Table.Cell className="whitespace-nowrap text-right tabular-nums text-muted">
                  {node.kind === "file" ? formatBytes(node.size) : "—"}
                </Table.Cell>
                <Table.Cell className="whitespace-nowrap text-muted">
                  {formatTimestamp(node.updatedAt)}
                </Table.Cell>
                <Table.Cell>
                  <Dropdown>
                    <Dropdown.Trigger>
                      <Button
                        variant="ghost"
                        isIconOnly
                        size="sm"
                        aria-label="Node actions"
                      >
                        <MoreHorizontal className="size-4" />
                      </Button>
                    </Dropdown.Trigger>
                    <Dropdown.Popover>
                      <Dropdown.Menu
                        onAction={(key) => {
                          if (key === "rename") {
                            onRename(node);
                          } else if (key === "delete") {
                            onDelete(node);
                          }
                        }}
                      >
                        <Dropdown.Item id="rename">Rename</Dropdown.Item>
                        <Dropdown.Item id="delete">Delete</Dropdown.Item>
                      </Dropdown.Menu>
                    </Dropdown.Popover>
                  </Dropdown>
                </Table.Cell>
              </Table.Row>
            ))}
          </Table.Body>
        </Table.Content>
      </Table.ScrollContainer>
    </Table>
  );
}
