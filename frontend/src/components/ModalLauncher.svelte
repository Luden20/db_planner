<script lang="ts">
  export let triggerLabel = "Abrir modal";
  export let title = "Acción requerida";
  export let confirmLabel = "Confirmar";
  export let triggerVariant: "primary" | "danger" | "secondary" = "primary";
  export let confirmVariant: "primary" | "danger" | "secondary" = "primary";
  export let onSuccess: () => void | Promise<void> = () => {};

  let open = false;
  let busy = false;

  const openModal = () => {
    open = true;
  };

  const closeModal = () => {
    if (busy) return;
    open = false;
  };

  const handleSuccess = async () => {
    busy = true;
    try {
      await onSuccess();
      open = false;
    } catch (err) {
      console.error("Modal action failed:", err);
    } finally {
      busy = false;
    }
  };
</script>

<button class={`btn ${triggerVariant}`} on:click={openModal}>
  {triggerLabel}
</button>

{#if open}
  <div class="modal-backdrop" on:click={closeModal}>
    <div class="modal" on:click|stopPropagation>
      <header class="modal-header">
        <p class="modal-title">{title}</p>
        <button class="icon-btn" on:click={closeModal} aria-label="Cerrar modal">
          x
        </button>
      </header>

      <section class="modal-body">
        <slot/>
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
    background: linear-gradient(135deg, rgba(17, 25, 40, 0.95), rgba(20, 32, 52, 0.95));
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 16px;
    box-shadow: 0 18px 48px rgba(0, 0, 0, 0.45);
    color: #e8edf7;
    padding: 18px 18px 16px;
    animation: rise 160ms ease-out;
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
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    margin-top: 4px;
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
  }
</style>
