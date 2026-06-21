import * as RadixDialog from "@radix-ui/react-dialog";
import { AnimatePresence, motion } from "motion/react";
import { type ReactNode } from "react";
import { cn } from "./lib/cn.js";

interface DialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description?: string;
  children: ReactNode;
  footer?: ReactNode;
  className?: string;
}

export function Dialog({
  open,
  onOpenChange,
  title,
  description,
  children,
  footer,
  className,
}: DialogProps) {
  return (
    <RadixDialog.Root open={open} onOpenChange={onOpenChange}>
      <AnimatePresence>
        {open && (
          <RadixDialog.Portal forceMount>
            <RadixDialog.Overlay asChild>
              <motion.div
                className="fixed inset-0 z-50 bg-black/20 backdrop-blur-[2px]"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 0.15 }}
              />
            </RadixDialog.Overlay>
            <RadixDialog.Content asChild>
              <motion.div
                className={cn(
                  "fixed left-1/2 top-1/2 z-50 w-full max-w-[480px]",
                  "rounded-[--radius-lg] bg-white shadow-[--shadow-popover]",
                  "outline-none",
                  className,
                )}
                initial={{ opacity: 0, scale: 0.96, x: "-50%", y: "-48%" }}
                animate={{ opacity: 1, scale: 1, x: "-50%", y: "-50%" }}
                exit={{ opacity: 0, scale: 0.96, x: "-50%", y: "-48%" }}
                transition={{ duration: 0.15, ease: [0.16, 1, 0.3, 1] }}
              >
                <div className="px-5 pt-5 pb-4">
                  <RadixDialog.Title className="text-[15px] font-semibold text-[--color-neutral-900]">
                    {title}
                  </RadixDialog.Title>
                  {description && (
                    <RadixDialog.Description className="mt-1 text-[13px] text-[--color-neutral-500]">
                      {description}
                    </RadixDialog.Description>
                  )}
                </div>
                <div className="px-5 pb-4">{children}</div>
                {footer && (
                  <div className="flex items-center justify-end gap-2 border-t border-[--color-neutral-100] px-5 py-3">
                    {footer}
                  </div>
                )}
              </motion.div>
            </RadixDialog.Content>
          </RadixDialog.Portal>
        )}
      </AnimatePresence>
    </RadixDialog.Root>
  );
}

export { RadixDialog as DialogPrimitive };
