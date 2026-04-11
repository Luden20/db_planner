<script lang="ts">
  import { Dialog } from "bits-ui";
  import ButtonIcon from "../ButtonIcon.svelte";

  let { 
    title = "", 
    open = $bindable(false),
    triggerLabel = "Abrir",
    triggerIcon = null,
    triggerVariant = "primary",
    triggerSize = "default", // default, sm, icon
    triggerClass = "",
    showTrigger = true,
    size = "default", // default, form
    busy = false,
    errorMessage = "",
    confirmLabel = "Confirmar",
    confirmIcon = null,
    confirmVariant = "primary",
    cancelIcon = "close",
    children, // main body
    onConfirm = null,
    onCancel = null,
    triggerDisabled = false
  } = $props();

  const handleConfirm = async () => {
     if (onConfirm) {
         await onConfirm();
     }
  }

  const handleCancel = () => {
    if (onCancel) {
      onCancel();
    }
    open = false;
  }

  const getTriggerClass = () => {
    if (triggerSize === "sm") return `control control--sm control--${triggerVariant} ${triggerClass}`.trim();
    if (triggerSize === "icon") return `control control--sm control--icon control--${triggerVariant} ${triggerClass}`.trim();
    return `btn ${triggerVariant} ${triggerClass}`.trim();
  }
</script>

<Dialog.Root bind:open>
  {#if showTrigger}
  <Dialog.Trigger
    disabled={triggerDisabled}
    class={getTriggerClass()}
  >
    {#if triggerIcon}<ButtonIcon name={triggerIcon}/>{/if}
    {#if triggerSize !== 'icon'}<span>{triggerLabel}</span>{/if}
  </Dialog.Trigger>
  {/if}

  <Dialog.Portal>
    <Dialog.Overlay class="modal-backdrop" />
    <Dialog.Content class="modal modal--{size}">
      <header class="modal-header">
        <Dialog.Title class="modal-title">{title}</Dialog.Title>
        <Dialog.Close class="icon-btn control control--icon control--soft" aria-label="Cerrar modal">
          <ButtonIcon name="close"/>
        </Dialog.Close>
      </header>

      <section class="modal-body">
        {#if children}
          {@render children()}
        {/if}
        {#if errorMessage}
          <p class="error">{errorMessage}</p>
        {/if}
      </section>

      <footer class="modal-footer">
        <Dialog.Close class="btn secondary" disabled={busy} onclick={handleCancel}>
           <ButtonIcon name={cancelIcon}/>
           <span>Cancelar</span>
        </Dialog.Close>
        <button class="btn {confirmVariant}" onclick={handleConfirm} disabled={busy}>
           <ButtonIcon name={busy ? "spark" : confirmIcon}/>
           <span>{busy ? "Procesando..." : confirmLabel}</span>
        </button>
      </footer>
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    display: grid;
    place-items: center;
    background: var(--overlay-scrim);
    backdrop-filter: blur(10px);
    z-index: var(--layer-modal);
    padding: 1.2rem;
  }

  :global(.modal) {
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

  :global(.modal--form) {
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

  :global(.modal-title) {
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

  :global(.modal--form) .modal-body {
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
    :global(.modal) {
      padding: 1rem 1rem 0.9rem;
    }
    :global(.modal--form) {
      width: 100%;
      padding: 1.1rem 1rem 1rem;
    }
    :global(.modal--form) .modal-body {
      padding: 0.75rem 0.1rem 0.9rem;
    }
  }
</style>
