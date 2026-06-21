import * as Radix from "@radix-ui/react-context-menu";
import { motion } from "motion/react";
import { type ReactNode } from "react";
import { cn } from "./lib/cn.js";

interface ContextMenuItem {
  label: string;
  icon?: ReactNode;
  onSelect: () => void;
  destructive?: boolean;
  disabled?: boolean;
}

interface ContextMenuProps {
  items: ContextMenuItem[];
  children: ReactNode;
}

export function ContextMenu({ items, children }: ContextMenuProps) {
  return (
    <Radix.Root>
      <Radix.Trigger asChild>{children}</Radix.Trigger>
      <Radix.Portal>
        <Radix.Content asChild>
          <motion.div
            className="z-50 min-w-[200px] overflow-hidden rounded-[--radius-md] bg-white py-1 shadow-[--shadow-popover]"
            initial={{ opacity: 0, scale: 0.96, y: -4 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            transition={{ duration: 0.12, ease: [0.16, 1, 0.3, 1] }}
          >
            {items.map((item) => (
              <Radix.Item
                key={item.label}
                disabled={item.disabled}
                onSelect={item.onSelect}
                className={cn(
                  "flex h-8 cursor-default select-none items-center gap-2 px-3 text-[13px] outline-none",
                  "transition-colors duration-75",
                  item.destructive
                    ? "text-[--color-destructive] data-[highlighted]:bg-red-50"
                    : "text-[--color-neutral-700] data-[highlighted]:bg-[--color-neutral-100]",
                  item.disabled && "opacity-40",
                )}
              >
                {item.icon && (
                  <span className="flex h-4 w-4 items-center justify-center text-current/60">
                    {item.icon}
                  </span>
                )}
                {item.label}
              </Radix.Item>
            ))}
          </motion.div>
        </Radix.Content>
      </Radix.Portal>
    </Radix.Root>
  );
}
