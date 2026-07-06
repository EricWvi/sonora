import { toast } from "@heroui/react";
import { ContractTransportError } from "@sonora/contracts";

// Surfaces an error from a mutation/query as a danger toast. Transport errors
// carry a stable `code` plus a human-readable `message` from the backend error
// envelope; everything else falls back to the generic error message.
export function notifyError(error: unknown): void {
  if (error instanceof ContractTransportError) {
    toast.danger(`${error.code}: ${error.message}`);
    return;
  }

  toast.danger(error instanceof Error ? error.message : "Unexpected error");
}
