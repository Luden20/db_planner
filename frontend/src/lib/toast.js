import { writable } from "svelte/store";
export const toastState = writable(null);
let toastTimer = null;
export const clearToast = () => {
    if (toastTimer !== null) {
        clearTimeout(toastTimer);
        toastTimer = null;
    }
    toastState.set(null);
};
export const showToast = (message, tone = "info", duration = 2000) => {
    if (toastTimer !== null) {
        clearTimeout(toastTimer);
        toastTimer = null;
    }
    toastState.set({ message, tone });
    if (duration > 0) {
        toastTimer = setTimeout(() => {
            toastState.set(null);
            toastTimer = null;
        }, duration);
    }
};
//# sourceMappingURL=toast.js.map