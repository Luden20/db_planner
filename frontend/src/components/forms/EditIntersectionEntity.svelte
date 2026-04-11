<script lang="ts">
  import ModalLauncher from "../ModalLauncher.svelte";
  import {EditIntersectionEntityDescription, Save} from "../../../wailsjs/go/main/App";
  import type {utils} from "../../../wailsjs/go/models";
  import {showToast} from "../../lib/toast";

  export let item: utils.IntersectionEntity;
  export let onSave: () => Promise<void> = async () => {};
  export let triggerLabel = "Editar";

  let description = "";
  let error = "";

  const loadValues = () => {
    description = item?.Entity?.Description ?? "";
    error = "";
  };

  const handleSave = async () => {
    try {
      error = "";
      await EditIntersectionEntityDescription(item.RelationID, description.trim());
      await Save();
      await onSave();
    } catch (err) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      error = `${message}`;
      showToast(`No se pudo guardar la descripción: ${message}`, "error");
      throw err;
    }
  };
</script>

<div class="toolbar-actions">
  <ModalLauncher
    triggerLabel={triggerLabel}
    title="Editar entidad de intersección"
    confirmLabel="Guardar"
    triggerVariant="edit"
    confirmVariant="primary"
    size="form"
    onOpen={loadValues}
    onSuccess={handleSave}
  >
    <div class="field">
      <label for="intersection-name">Nombre</label>
      <input
        id="intersection-name"
        type="text"
        value={item.Entity.Name}
        disabled
      />
      <p class="helper">El nombre se genera automáticamente a partir de las dos entidades relacionadas.</p>
    </div>

    <div class="field">
      <label for="intersection-description">Descripción</label>
      <textarea
        id="intersection-description"
        rows="3"
        placeholder="Describe el propósito de esta entidad de intersección"
        bind:value={description}
      ></textarea>
    </div>

    {#if error}
      <p class="form-error">{error}</p>
    {/if}
  </ModalLauncher>
</div>

<style>
  .toolbar-actions {
    display: inline-flex;
    gap: 0.5rem;
  }

  .field {
    display: grid;
    gap: 0.65rem;
    color: var(--ink-soft);
    font-size: 0.92rem;
    padding: 0.25rem 0.1rem;
  }

  .field input,
  .field textarea {
    width: 100%;
    box-sizing: border-box;
    border-radius: 1rem;
    border: 1px solid var(--border);
    background: var(--field-surface);
    color: var(--ink);
    padding: 0.9rem 1rem;
    font-size: 0.96rem;
    outline: none;
    transition: border 140ms ease, box-shadow 140ms ease, background 140ms ease;
  }

  .field input:disabled {
    color: var(--ink-faint);
    cursor: not-allowed;
    opacity: 0.86;
  }

  .field textarea {
    min-height: 144px;
    resize: vertical;
  }

  .field textarea:focus {
    border-color: var(--focus-border);
    box-shadow: var(--focus-ring);
    background: var(--field-surface-focus);
  }

  .helper {
    margin: 0;
    color: var(--ink-faint);
    font-size: 0.8rem;
  }

  .form-error {
    margin: 0.2rem 0 0;
    color: var(--danger);
    font-weight: 600;
  }
</style>
