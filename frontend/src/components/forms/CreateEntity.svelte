<script lang="ts">
  import ModalLauncher from "../ModalLauncher.svelte";
  import { AddEntity, EditEntity, GetActualProject, GetEntity, MoveEntity } from "../../../wailsjs/go/main/App";
  import { showToast } from "../../lib/toast";

  type InsertPlacement = "above" | "below";

  let { 
    id = null, 
    onSave = async () => {}, 
    showTrigger = true, 
    triggerLabel = null, 
    triggerSize = "default",
    title = null 
  } = $props<{
    id?: number | null;
    onSave?: () => Promise<void>;
    showTrigger?: boolean;
    triggerLabel?: string | null;
    triggerSize?: any;
    title?: string | null;
  }>();

  let name = $state("");
  let description = $state("");
  let error = $state("");
  let modalRef = $state<ModalLauncher | null>(null);
  let insertionTarget = $state<{ referenceId: number; placement: InsertPlacement } | null>(null);

  const resolvedTriggerLabel = $derived(triggerLabel ?? (id === null ? "Nueva entidad" : "Editar entidad"));
  const resolvedTitle = $derived(title ?? (id === null
    ? (insertionTarget?.placement === "above"
      ? "Insertar entidad arriba"
      : insertionTarget?.placement === "below"
        ? "Insertar entidad abajo"
        : "Crear entidad")
    : "Editar entidad"));

  const resetForm = () => {
    name = "";
    description = "";
    error = "";
  };

  const loadEntity = async () => {
    error = "";
    if (id === null) {
      resetForm();
      return;
    }
    const data = await GetEntity(id);
    name = data.Name;
    description = data.Description;
  };

  const placeCreatedEntity = async (newEntityId: number) => {
    if (!insertionTarget) return;
    const project = await GetActualProject();
    const entities = project?.Entities ?? [];
    const currentIndex = entities.findIndex((e) => e.Id === newEntityId);
    const referenceIndex = entities.findIndex((e) => e.Id === insertionTarget?.referenceId);
    if (currentIndex === -1 || referenceIndex === -1) {
      insertionTarget = null;
      return;
    }
    const desiredIndex = insertionTarget.placement === "above"
      ? referenceIndex
      : Math.min(referenceIndex + 1, entities.length - 1);

    for (let index = currentIndex; index > desiredIndex; index -= 1) {
      await MoveEntity(newEntityId, "up");
    }
    insertionTarget = null;
  };

  export const openForInsert = async (refId: number, placement: InsertPlacement) => {
    if (id !== null) return;
    insertionTarget = { referenceId: refId, placement };
    await loadEntity();
    modalRef?.openDialog();
  };

  const handleSave = async () => {
    const trimmedName = name.trim();
    const trimmedDescription = description.trim();
    if (!trimmedName) {
      error = "Ingresa un nombre para la entidad.";
      throw new Error(error);
    }

    try {
      error = "";
      if (id == null) {
        await AddEntity(trimmedName, trimmedDescription);
        const project = await GetActualProject();
        const entities = project?.Entities ?? [];
        const newEntity = entities.reduce<{ Id: number } | null>((latest, entity) => {
          if (latest === null || entity.Id > latest.Id) return entity;
          return latest;
        }, null);
        if (newEntity) await placeCreatedEntity(newEntity.Id);
      } else {
        await EditEntity(id, trimmedName, trimmedDescription);
      }
      await onSave();
      if (id === null) resetForm();
    } catch (err: any) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      showToast(`No se pudo guardar la entidad: ${message}`, "error");
      throw err;
    }
  };
</script>

<ModalLauncher
  bind:this={modalRef}
  triggerLabel={resolvedTriggerLabel}
  title={resolvedTitle}
  confirmLabel="Guardar"
  triggerVariant={id === null ? "primary" : "edit"}
  confirmVariant="primary"
  size="form"
  {triggerSize}
  {showTrigger}
  onOpen={loadEntity}
  onSuccess={handleSave}
>
  <div class="field">
    <label for="name">Nombre</label>
    <input
      id="name"
      type="text"
      autocomplete="off"
      placeholder="Cliente, Producto..."
      bind:value={name}
    />
  </div>

  <div class="field">
    <label for="description">Descripción</label>
    <textarea
      id="description"
      rows="3"
      placeholder="Breve contexto o notas opcionales"
      bind:value={description}
    ></textarea>
  </div>

  {#if error}
    <p class="form-error">{error}</p>
  {/if}
</ModalLauncher>

<style>
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
    transition: all 140ms ease;
  }

  .field textarea {
    min-height: 144px;
    resize: vertical;
  }

  .field input:focus,
  .field textarea:focus {
    border-color: var(--focus-border);
    box-shadow: var(--focus-ring);
    background: var(--field-surface-focus);
  }

  .form-error {
    margin: 0.2rem 0 0;
    color: var(--danger);
    font-weight: 600;
  }
</style>
