import * as RadixAlertDialog from "@radix-ui/react-alert-dialog";
import { AnimatePresence, motion } from "motion/react";
import { type ReactNode } from "react";
import { Button } from "./button.js";

interface AlertDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: string;
  confirmLabel?: string;
  cancelLabel?: string;
  onConfirm: () => void;
  /** Extra content between description and footer. */
  children?: ReactNode;
}

export function AlertDialog({
  open,
  onOpenChange,
  title,
  description,
  confirmLabel = "Confirm",
  cancelLabel = "Cancel",
  onConfirm,
  children,
}: AlertDialogProps) {
  return (
    <RadixAlertDialog.Root open={open} onOpenChange={onOpenChange}>
      <AnimatePresence>
        {open && (
          <RadixAlertDialog.Portal forceMount>
            <RadixAlertDialog.Overlay asChild>
              <motion.div
                className="fixed inset-0 z-50 bg-black/20 backdrop-blur-[2px]"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 0.15 }}
              />
            </RadixAlertDialog.Overlay>
            <RadixAlertDialog.Content asChild>
              <motion.div
                className="fixed left-1/2 top-1/2 z-50 w-full max-w-[400px] rounded-[--radius-lg] bg-white shadow-[--shadow-popover] outline-none"
                initial={{ opacity: 0, scale: 0.96, x: "-50%", y: "-48%" }}
                animate={{ opacity: 1, scale: 1, x: "-50%", y: "-50%" }}
                exit={{ opacity: 0, scale: 0.96, x: "-50%", y: "-48%" }}
                transition={{ duration: 0.15, ease: [0.16, 1, 0.3, 1] }}
              >
                <div className="px-5 pt-5 pb-4">
                  <RadixAlertDialog.Title className="text-[15px] font-semibold text-[--color-neutral-900]">
                    {title}
                  </RadixAlertDialog.Title>
                  <RadixAlertDialog.Description className="mt-1 text-[13px] text-[--color-neutral-500]">
                    {description}
                  </RadixAlertDialog.Description>
                  {children && <div className="mt-3">{children}</div>}
                </div>
                <div className="flex items-center justify-end gap-2 border-t border-[--color-neutral-100] px-5 py-3">
                  <RadixAlertDialog.Cancel asChild>
                    <Button variant="ghost">{cancelLabel}</Button>
                  </RadixAlertDialog.Cancel>
                  <RadixAlertDialog.Action asChild>
                    <Button variant="destructive" onClick={onConfirm}>
                      {confirmLabel}
                    </Button>
                  </RadixAlertDialog.Action>
                </div>
              </motion.div>
            </RadixAlertDialog.Content>
          </RadixAlertDialog.Portal>
        )}
      </AnimatePresence>
    </RadixAlertDialog.Root>
  );
}
