type ViewTransitionDocument = Document & {
  startViewTransition?: (update: () => void | Promise<void>) => {
    finished: Promise<void>;
  };
};

type VerticalAutoScrollOptions = {
  edgePx: number;
  stepPx: number;
  getContainer?: () => HTMLElement | null;
  useWindow?: boolean;
};

export const getErrorMessage = (error: unknown, fallback = "Error desconocido") => {
  if (typeof error === "string") {
    return error;
  }

  if (error && typeof error === "object") {
    const record = error as Record<string, unknown>;
    return String(record.error ?? record.message ?? fallback);
  }

  return fallback;
};

const prefersReducedMotion = () =>
  typeof window !== "undefined"
  && typeof window.matchMedia === "function"
  && window.matchMedia("(prefers-reduced-motion: reduce)").matches;

export const runViewTransition = async (
  update: () => void | Promise<void>,
  warningMessage = "No se pudo aplicar la transicion de vista:"
) => {
  const doc = typeof document !== "undefined" ? (document as ViewTransitionDocument) : null;

  if (doc?.startViewTransition && !prefersReducedMotion()) {
    try {
      const transition = doc.startViewTransition(update);
      await transition.finished;
      return;
    } catch (error) {
      console.warn(warningMessage, error);
    }
  }

  await update();
};

export const createVerticalAutoScroller = ({
  edgePx,
  stepPx,
  getContainer,
  useWindow = false
}: VerticalAutoScrollOptions) => {
  let frame: number | null = null;
  let direction: -1 | 0 | 1 = 0;

  const stop = () => {
    direction = 0;
    if (frame !== null) {
      window.cancelAnimationFrame(frame);
      frame = null;
    }
  };

  const run = () => {
    if (direction === 0) {
      frame = null;
      return;
    }

    if (useWindow) {
      window.scrollBy({top: direction * stepPx, behavior: "auto"});
    } else {
      const container = getContainer?.();
      if (!container) {
        stop();
        return;
      }
      container.scrollTop += direction * stepPx;
    }

    frame = window.requestAnimationFrame(run);
  };

  const start = (nextDirection: -1 | 1) => {
    if (direction === nextDirection && frame !== null) {
      return;
    }

    direction = nextDirection;
    if (frame === null) {
      frame = window.requestAnimationFrame(run);
    }
  };

  const update = (clientY: number) => {
    if (useWindow) {
      if (typeof window === "undefined") {
        return;
      }

      if (clientY <= edgePx) {
        start(-1);
        return;
      }

      if (clientY >= window.innerHeight - edgePx) {
        start(1);
        return;
      }

      stop();
      return;
    }

    const container = getContainer?.();
    if (!container) {
      return;
    }

    const bounds = container.getBoundingClientRect();
    if (clientY <= bounds.top + edgePx) {
      start(-1);
      return;
    }

    if (clientY >= bounds.bottom - edgePx) {
      start(1);
      return;
    }

    stop();
  };

  return {
    stop,
    updateFromDragEvent(event: DragEvent) {
      update(event.clientY);
    }
  };
};
