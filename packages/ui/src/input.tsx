import { type ComponentPropsWithoutRef } from "react";
import { cn } from "./lib/cn.js";

interface InputProps extends ComponentPropsWithoutRef<"input"> {
  error?: string;
}

export function Input({ error, className, ...props }: InputProps) {
  return (
    <div className="flex flex-col gap-1">
      <input
        className={cn(
          "h-8 w-full rounded-[--radius-md] bg-[--color-neutral-100] px-3",
          "text-[13px] text-[--color-neutral-900] placeholder:text-[--color-neutral-400]",
          "border border-transparent outline-none transition-all duration-100",
          "hover:bg-[--color-neutral-200]",
          "focus:bg-white focus:border-[--color-accent]/50 focus:ring-2 focus:ring-[--color-accent]/20",
          error && "border-[--color-destructive]/50 focus:ring-[--color-destructive]/20",
          className,
        )}
        {...props}
      />
      {error && (
        <span className="text-[11px] text-[--color-destructive]">{error}</span>
      )}
    </div>
  );
}
