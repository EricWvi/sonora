import { Slot } from "@radix-ui/react-slot";
import { motion } from "motion/react";
import { type ComponentPropsWithoutRef } from "react";
import { cn } from "./lib/cn.js";

type Variant = "filled" | "ghost" | "destructive";

interface ButtonProps extends ComponentPropsWithoutRef<"button"> {
  variant?: Variant;
  asChild?: boolean;
}

const variantClass: Record<Variant, string> = {
  filled:
    "bg-[--color-accent] text-white hover:bg-[--color-accent-hover] active:scale-[0.97]",
  ghost:
    "bg-transparent text-[--color-neutral-700] hover:bg-[--color-neutral-100] active:bg-[--color-neutral-200]",
  destructive:
    "bg-[--color-destructive] text-white hover:bg-[--color-destructive-hover] active:scale-[0.97]",
};

export function Button({
  variant = "filled",
  asChild = false,
  className,
  children,
  ...props
}: ButtonProps) {
  const Comp = asChild ? Slot : "button";

  return (
    <motion.div
      whileTap={{ scale: variant === "ghost" ? 0.98 : 0.97 }}
      transition={{ duration: 0.1 }}
      className="inline-flex"
    >
      <Comp
        className={cn(
          "inline-flex items-center justify-center gap-1.5 rounded-[--radius-md]",
          "h-8 px-3 text-[13px] font-medium leading-none",
          "transition-colors duration-100 outline-none",
          "focus-visible:ring-2 focus-visible:ring-[--color-accent]/50",
          "disabled:pointer-events-none disabled:opacity-40",
          "cursor-default select-none",
          variantClass[variant],
          className,
        )}
        {...props}
      >
        {children}
      </Comp>
    </motion.div>
  );
}
