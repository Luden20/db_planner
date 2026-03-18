import {writable} from "svelte/store";

export type ToastTone = "info" | "success" | "error";

export type ToastPayload = {
  message: string;
  tone: ToastTone;
};

export const toastState = writable<ToastPayload | null>(null);

let toastTimer: ReturnType<typeof setTimeout> | null = null;

export const clearToast = () => {
  if (toastTimer !== null) {
    clearTimeout(toastTimer);
    toastTimer = null;
  }
  toastState.set(null);
};

export const showToast = (message: string, tone: ToastTone = "info", duration = 2000) => {
  if (toastTimer !== null) {
    clearTimeout(toastTimer);
    toastTimer = null;
  }

  toastState.set({message, tone});

  if (duration > 0) {
    toastTimer = setTimeout(() => {
      toastState.set(null);
      toastTimer = null;
    }, duration);
  }
};
