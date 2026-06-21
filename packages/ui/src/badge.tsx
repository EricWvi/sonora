import { cn } from "./lib/cn.js";

type BadgeVariant = "default" | "pending" | "available" | "muted";

interface BadgeProps {
  children: string;
  variant?: BadgeVariant;
  className?: string;
}

const variantClass: Record<BadgeVariant, string> = {
  default: "bg-[--color-neutral-100] text-[--color-neutral-600]",
  pending: "bg-[--color-amber-50] text-[--color-amber-700]",
  available: "bg-[--color-green-50] text-[--color-green-700]",
  muted: "bg-transparent text-[--color-neutral-400]",
};

export function Badge({ children, variant = "default", className }: BadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-[--radius-full] px-2 py-0.5",
        "text-[11px] font-medium leading-none",
        variantClass[variant],
        className,
      )}
    >
      {children}
    </span>
  );
}
