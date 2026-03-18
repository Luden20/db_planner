<script lang="ts">
  import ModalLauncher from "../ModalLauncher.svelte";
  import {AddEntity, EditEntity, GetActualProject, GetEntity, MoveEntity} from "../../../wailsjs/go/main/App";
  import {showToast} from "../../lib/toast";
  type InsertPlacement = "above" | "below";
  export let id:number|null=null;
  export let onSave: () => Promise<void> = async () => {};
  export let showTrigger = true;
  let name = "";
  let description = "";
  let error = "";
  let modalRef: ModalLauncher | null = null;
  let insertionTarget: { referenceId: number; placement: InsertPlacement } | null = null;

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
    if (!insertionTarget) {
      return;
    }

    const project = await GetActualProject();
    const entities = project?.Entities ?? [];
    const currentIndex = entities.findIndex((entity) => entity.Id === newEntityId);
    const referenceIndex = entities.findIndex((entity) => entity.Id === insertionTarget?.referenceId);
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

  export const openForInsert = async (referenceId: number, placement: InsertPlacement) => {
    if (id !== null) {
      return;
    }

    insertionTarget = {referenceId, placement};
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
      if(id==null){
        await AddEntity(trimmedName, trimmedDescription);
        const project = await GetActualProject();
        const entities = project?.Entities ?? [];
        const newEntity = entities.reduce<{ Id: number } | null>((latest, entity) => {
          if (latest === null || entity.Id > latest.Id) {
            return entity;
          }
          return latest;
        }, null);
        if (newEntity) {
          await placeCreatedEntity(newEntity.Id);
        }
      }else if(id!=null){
        await EditEntity(id,trimmedName, trimmedDescription);
      }
      await onSave();
      if (id === null) {
        resetForm();
      }
    } catch (err) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      showToast(`No se pudo guardar la entidad: ${message}`, "error");
      throw err;
    }
  };
</script>

<div class="toolbar-actions">
  <ModalLauncher
          bind:this={modalRef}
          triggerLabel={id === null ? "Nueva entidad" : "Editar entidad"}
          title={id === null ? (insertionTarget?.placement === "above" ? "Insertar entidad arriba" : insertionTarget?.placement === "below" ? "Insertar entidad abajo" : "Crear entidad") : "Editar entidad"}
          confirmLabel="Guardar"
          triggerVariant="primary"
          confirmVariant="primary"
          size="form"
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
      />
    </div>

    {#if error}
      <p class="form-error">{error}</p>
    {/if}
  </ModalLauncher>
</div>

<style>
  .toolbar-actions {
    display: inline-flex;
    gap: 8px;
  }

  .field {
    display: grid;
    gap: 10px;
    color: #cfd9e9;
    font-size: 14px;
    padding: 4px 2px;
  }

  .field input,
  .field textarea {
    width: 100%;
    box-sizing: border-box;
    border-radius: 12px;
    border: 1px solid rgba(255, 255, 255, 0.14);
    background: rgba(255, 255, 255, 0.04);
    color: #e8edf7;
    padding: 14px 16px;
    font-size: 15px;
    outline: none;
    transition: border 140ms ease, box-shadow 140ms ease;
  }

  .field textarea {
    min-height: 144px;
    resize: vertical;
  }

  .field input:focus,
  .field textarea:focus {
    border-color: rgba(90, 209, 255, 0.8);
    box-shadow: 0 0 0 2px rgba(90, 209, 255, 0.22);
  }

  .form-error {
    margin: 2px 0 0;
    color: #ffb4a2;
    font-weight: 600;
  }
</style>
