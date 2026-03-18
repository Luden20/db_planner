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
        <button class="icon-btn" on:click={closeModal} aria-label="Cerrar modal">
          x
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
    background: rgba(6, 12, 24, 0.7);
    backdrop-filter: blur(4px);
    z-index: 40;
    padding: 18px;
  }

  .modal {
    width: min(560px, 100%);
    max-height: min(88vh, 920px);
    background: linear-gradient(135deg, rgba(17, 25, 40, 0.95), rgba(20, 32, 52, 0.95));
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 16px;
    box-shadow: 0 18px 48px rgba(0, 0, 0, 0.45);
    color: #e8edf7;
    padding: 18px 18px 16px;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    animation: rise 160ms ease-out;
  }

  .modal--form {
    width: min(760px, 100%);
    padding: 24px 24px 20px;
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
    font-size: 18px;
    font-weight: 700;
    letter-spacing: 0.3px;
    color: #f1f5ff;
  }

  .icon-btn {
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.14);
    border-radius: 10px;
    color: #dbe8ff;
    width: 34px;
    height: 34px;
    cursor: pointer;
    font-weight: 700;
    display: grid;
    place-items: center;
    transition: background 160ms ease, transform 140ms ease;
  }

  .icon-btn:hover {
    background: rgba(255, 255, 255, 0.14);
    transform: translateY(-1px);
  }

  .modal-body {
    padding: 8px 2px 14px;
    color: #dbe8ff;
    overflow-y: auto;
    min-height: 0;
  }

  .modal--form .modal-body {
    padding: 14px 4px 20px;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    margin-top: 4px;
  }

  .error {
    margin: 6px 0 0;
    color: #ffb4a2;
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
      padding: 16px 16px 14px;
    }

    .modal--form {
      width: 100%;
      padding: 18px 16px 16px;
    }

    .modal--form .modal-body {
      padding: 12px 2px 16px;
    }
  }
</style>
