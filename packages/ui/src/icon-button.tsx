import { motion } from "motion/react";
import { type ComponentPropsWithoutRef } from "react";
import { cn } from "./lib/cn.js";

interface IconButtonProps extends ComponentPropsWithoutRef<"button"> {
  label: string;
}

export function IconButton({ label, className, children, ...props }: IconButtonProps) {
  return (
    <motion.button
      whileTap={{ scale: 0.92 }}
      transition={{ duration: 0.1 }}
      aria-label={label}
      className={cn(
        "inline-flex h-7 w-7 items-center justify-center rounded-[--radius-md]",
        "text-[--color-neutral-500] hover:text-[--color-neutral-800]",
        "hover:bg-[--color-neutral-100] active:bg-[--color-neutral-200]",
        "transition-colors duration-100 outline-none",
        "focus-visible:ring-2 focus-visible:ring-[--color-accent]/50",
        "disabled:pointer-events-none disabled:opacity-40",
        "cursor-default",
        className,
      )}
      {...(props as ComponentPropsWithoutRef<typeof motion.button>)}
    >
      {children}
    </motion.button>
  );
}
