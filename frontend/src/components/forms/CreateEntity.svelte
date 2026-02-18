<script lang="ts">
  import ModalLauncher from "../ModalLauncher.svelte";
  import {AddEntity, EditEntity, GetCombinatory, GetEntity} from "../../../wailsjs/go/main/App";
  import {onMount} from "svelte";
  export let id:number|null=null;
  export let onSave: () => Promise<void> = async () => {};
  let name = "";
  let description = "";
  let error = "";

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
      }else if(id!=null){
        await EditEntity(id,trimmedName, trimmedDescription);
      }
      await onSave();
      name = "";
      description = "";
    } catch (err) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      alert(`No se pudo crear la entidad: ${message}`);
      throw err;
    }
  };
  async function load() {
    if(id!=null){
      const data=await GetEntity(id);
      name=data.Name;
      description=data.Description;
    }
  }

  onMount(load);
</script>

<div class="toolbar-actions">
  <ModalLauncher
          triggerLabel={id === null ? "Nueva entidad" : "Editar entidad"}
          title={id === null ? "Crear entidad" : "Editar entidad"}
          confirmLabel="Guardar"
          triggerVariant="primary"
          confirmVariant="primary"
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
    gap: 8px;
    color: #cfd9e9;
    font-size: 14px;
  }

  .field input,
  .field textarea {
    width: 100%;
    border-radius: 12px;
    border: 1px solid rgba(255, 255, 255, 0.14);
    background: rgba(255, 255, 255, 0.04);
    color: #e8edf7;
    padding: 12px 12px;
    font-size: 14px;
    outline: none;
    transition: border 140ms ease, box-shadow 140ms ease;
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
