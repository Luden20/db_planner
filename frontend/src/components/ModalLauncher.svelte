<script lang="ts">
  export let triggerLabel = "Abrir modal";
  export let title = "Acción requerida";
  export let confirmLabel = "Confirmar";
  export let triggerVariant: "primary" | "danger" | "secondary" = "primary";
  export let confirmVariant: "primary" | "danger" | "secondary" = "primary";
  export let size: "default" | "form" = "default";
  export let showTrigger = true;
  export let onSuccess: () => void | Promise<void> = () => {};
  export let onOpen: () => void | Promise<void> = () => {};

  let isOpen = false;
  let busy = false;
  let errorMessage = "";

  const openModal = () => {
    errorMessage = "";
    try {
      const result = onOpen();
      if (result instanceof Promise) {
        result.catch((err) => {
          const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
          errorMessage = `${message}`;
        });
      }
    } catch (err) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      errorMessage = `${message}`;
    } finally {
      isOpen = true;
    }
  };

  const closeModal = () => {
    if (busy) return;
    isOpen = false;
  };

  const handleBackdropKey = (event: KeyboardEvent) => {
    if (busy) return;
    const key = event.key;
    if (key === "Escape" || key === "Enter" || key === " ") {
      event.preventDefault();
      closeModal();
    }
  };

  const handleSuccess = async () => {
    if (busy) return;
    busy = true;
    errorMessage = "";

    try {
      await onSuccess();
      isOpen = false;
    } catch (err) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      errorMessage = `${message}`;
      console.error("Modal action failed:", err);
    } finally {
      busy = false;
    }
  };

  export const openDialog = () => {
    openModal();
  };

  export const closeDialog = () => {
    closeModal();
  };
</script>

{#if showTrigger}
  <button class={`btn ${triggerVariant}`} on:click={openModal}>
    {triggerLabel}
  </button>
{/if}

{#if isOpen}
  <div
    class="modal-backdrop"
    role="presentation"
    tabindex="-1"
    on:click={closeModal}
    on:keydown={handleBackdropKey}
  >
    <div class={`modal modal--${size}`} tabindex="-1" on:click|stopPropagation on:keydown|stopPropagation>
      <header class="modal-header">
        <p class="modal-title">{title}</p>
        <button class="icon-btn control control--icon control--soft" on:click={closeModal} aria-label="Cerrar modal">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d="M6.72 6.72a.75.75 0 0 1 1.06 0L12 10.94l4.22-4.22a.75.75 0 1 1 1.06 1.06L13.06 12l4.22 4.22a.75.75 0 0 1-1.06 1.06L12 13.06l-4.22 4.22a.75.75 0 1 1-1.06-1.06L10.94 12 6.72 7.78a.75.75 0 0 1 0-1.06Z"/>
          </svg>
        </button>
      </header>

      <section class="modal-body">
        <slot/>
        {#if errorMessage}
          <p class="error">{errorMessage}</p>
        {/if}
      </section>

      <footer class="modal-footer">
        <button class="btn secondary" on:click={closeModal} disabled={busy}>Cancelar</button>
        <button class={`btn ${confirmVariant}`} on:click={handleSuccess} disabled={busy}>
          {busy ? "Procesando..." : confirmLabel}
        </button>
      </footer>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    display: grid;
    place-items: center;
    background: var(--overlay-scrim);
    backdrop-filter: blur(10px);
    z-index: 40;
    padding: 1.2rem;
  }

  .modal {
    width: min(560px, 100%);
    max-height: min(88vh, 920px);
    background: var(--popover-surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-lg);
    color: var(--ink);
    padding: 1.15rem 1.15rem 1rem;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    animation: rise 180ms cubic-bezier(.19,1,.22,1);
  }

  .modal--form {
    width: min(760px, 100%);
    padding: 1.45rem 1.45rem 1.2rem;
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    margin-bottom: 12px;
  }

  .modal-title {
    margin: 0;
    font-size: 1.4rem;
    font-weight: 700;
    letter-spacing: -0.02em;
    color: var(--ink);
    font-family: var(--font-display);
  }

  .icon-btn {
    color: var(--ink-soft);
    display: grid;
    place-items: center;
  }

  .modal-body {
    padding: 0.4rem 0.15rem 0.9rem;
    color: var(--ink-soft);
    overflow-y: auto;
    min-height: 0;
  }

  .modal--form .modal-body {
    padding: 0.8rem 0.25rem 1.15rem;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    margin-top: 4px;
  }

  .error {
    margin: 0.4rem 0 0;
    color: var(--danger);
    font-weight: 600;
  }

  @keyframes rise {
    from {
      opacity: 0;
      transform: translateY(8px) scale(0.98);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }

  @media (max-width: 540px) {
    .modal {
      padding: 1rem 1rem 0.9rem;
    }

    .modal--form {
      width: 100%;
      padding: 1.1rem 1rem 1rem;
    }

    .modal--form .modal-body {
      padding: 0.75rem 0.1rem 0.9rem;
    }
  }
</style>
