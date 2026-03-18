<script lang="ts">
  import ModalLauncher from "../ModalLauncher.svelte";
  import {AddAttribute, EditAttribute, Save} from "../../../wailsjs/go/main/App";
  import type {utils} from "../../../wailsjs/go/models";
  import {showToast} from "../../lib/toast";

  export let entityId: number;
  export let attribute: utils.Attribute | null = null;
  export let onSaved: () => Promise<void> = async () => {};

  const typeOptions = ["Por definir", "VARCHAR2", "N/A", "Numérico", "Cadena", "Carácter", "Tiempo", "Fecha", "Booleano"];

  let name = "";
  let description = "";
  let error = "";
  let typeSelection = "Por definir";
  let lengthInput = "";

  const parseType = (value: string) => {
    const match = (value || "").match(/^\\s*varchar2\\s*\\(\\s*(\\d+)\\s*\\)\\s*$/i);
    if (match) {
      return {type: "VARCHAR2", length: match[1]};
    }
    return {type: value || "Por definir", length: ""};
  };

  const prefill = () => {
    if (attribute) {
      name = attribute.Name;
      description = attribute.Description;
      const parsed = parseType(attribute.Type || "Por definir");
      typeSelection = parsed.type;
      lengthInput = parsed.length;
    } else {
      name = "";
      description = "";
      typeSelection = "Por definir";
      lengthInput = "";
    }
    error = "";
  };

  const handleSave = async () => {
    const trimmedName = name.trim();
    if (!trimmedName) {
      error = "Ingresa un nombre para el atributo.";
      throw new Error(error);
    }
    const selectedType = typeSelection || "Por definir";
    let finalType = selectedType;
    if (selectedType === "VARCHAR2") {
      const len = parseInt(lengthInput || "0", 10);
      if (!len || len <= 0) {
        error = "Indica la longitud para VARCHAR2.";
        throw new Error(error);
      }
      finalType = `VARCHAR2(${len})`;
    }

    try {
      error = "";
      if (attribute) {
        await EditAttribute(entityId, attribute.Id, trimmedName, description.trim(), finalType);
      } else {
        await AddAttribute(entityId, trimmedName, description.trim(), finalType);
      }
      await Save();
      await onSaved();
      if (!attribute) {
        name = "";
        description = "";
        typeSelection = "Por definir";
        lengthInput = "";
      }
    } catch (err) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      showToast(`No se pudo guardar el atributo: ${message}`, "error");
      throw err;
    }
  };
</script>

<ModalLauncher
  triggerLabel={attribute ? "Editar" : "Nuevo atributo"}
  title={attribute ? "Editar atributo" : "Crear atributo"}
  confirmLabel="Guardar"
  triggerVariant={attribute ? "secondary" : "primary"}
  confirmVariant="primary"
  size="form"
  onOpen={prefill}
  onSuccess={handleSave}
>
  <div class="field">
    <label for="attr-name">Nombre</label>
    <input
      id="attr-name"
      type="text"
      autocomplete="off"
      placeholder="Código, Nombre, Estado..."
      bind:value={name}
    />
  </div>

  <div class="field">
    <label for="attr-description">Descripción</label>
    <textarea
      id="attr-description"
      rows="3"
      placeholder="Breve descripción"
      bind:value={description}
    />
  </div>

  <div class="field">
    <label for="attr-type">Tipo</label>
    <div class="type-row">
      <select id="attr-type" bind:value={typeSelection}>
        {#each typeOptions as option}
          <option value={option}>{option}</option>
        {/each}
      </select>
      {#if typeSelection === "VARCHAR2"}
        <input
          class="length-input"
          type="number"
          min="1"
          inputmode="numeric"
          placeholder="n"
          bind:value={lengthInput}
          aria-label="Longitud de VARCHAR2"
        />
      {/if}
    </div>
    {#if typeSelection === "VARCHAR2"}
      <p class="helper">Indica la longitud (n) para VARCHAR2.</p>
    {/if}
  </div>

  {#if error}
    <p class="form-error">{error}</p>
  {/if}
</ModalLauncher>

<style>
  .field {
    display: grid;
    gap: 10px;
    color: #cfd9e9;
    font-size: 14px;
    padding: 4px 2px;
  }

  .field input,
  .field textarea,
  .field select {
    width: 100%;
    box-sizing: border-box;
    border-radius: 12px;
    border: 1px solid rgba(255, 255, 255, 0.14);
    background: #0f1726;
    color: #e8edf7;
    padding: 14px 16px;
    font-size: 15px;
    outline: none;
    transition: border 140ms ease, box-shadow 140ms ease;
    appearance: none;
  }

  .field textarea {
    min-height: 144px;
    resize: vertical;
  }

  .field input:focus,
  .field textarea:focus,
  .field select:focus {
    border-color: rgba(90, 209, 255, 0.8);
    box-shadow: 0 0 0 2px rgba(90, 209, 255, 0.22);
  }

  .type-row {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .length-input {
    width: 120px;
    min-width: 120px;
  }

  .helper {
    margin: 4px 0 0;
    color: #9ab5e4;
    font-size: 12px;
  }

  .form-error {
    margin: 2px 0 0;
    color: #ffb4a2;
    font-weight: 600;
  }
</style>
