import { type ComponentPropsWithoutRef } from "react";
import { cn } from "./lib/cn.js";

export function Label({ className, ...props }: ComponentPropsWithoutRef<"label">) {
  return (
    <label
      className={cn(
        "block text-[11px] font-medium uppercase tracking-wide text-[--color-neutral-500]",
        className,
      )}
      {...props}
    />
  );
}
