// Formatting helpers for node metadata shown in the table.

const BYTE_UNITS = ["B", "KB", "MB", "GB", "TB"] as const;

// Renders a byte count as a human-readable string; `null` (unknown size for a
// not-yet-uploaded file) renders as an em dash.
export function formatBytes(bytes: number | null): string {
  if (bytes === null) {
    return "—";
  }
  if (bytes < 1) {
    return "0 B";
  }
  const exponent = Math.min(
    Math.floor(Math.log(bytes) / Math.log(1024)),
    BYTE_UNITS.length - 1,
  );
  const value = bytes / Math.pow(1024, exponent);
  return `${value.toFixed(value >= 10 || exponent === 0 ? 0 : 1)} ${BYTE_UNITS[exponent]}`;
}

// Renders a millisecond epoch timestamp (the wire form of NodeView.updatedAt)
// using the user's locale.
export function formatTimestamp(ms: number): string {
  return new Date(ms).toLocaleString();
}
