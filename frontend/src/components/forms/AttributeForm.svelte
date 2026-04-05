<script lang="ts">
  import {tick} from "svelte";
  import {scale} from "svelte/transition";
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
  let showTypeMenu = false;
  let typeSearch = "";
  let filteredTypeOptions = [...typeOptions];
  let typeMenuEl: HTMLDivElement | null = null;
  let typeSearchInput: HTMLInputElement | null = null;

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
    showTypeMenu = false;
    typeSearch = "";
  };

  const openTypeMenu = async () => {
    showTypeMenu = true;
    typeSearch = "";
    await tick();
    typeSearchInput?.focus();
  };

  const closeTypeMenu = () => {
    showTypeMenu = false;
    typeSearch = "";
  };

  const toggleTypeMenu = async () => {
    if (showTypeMenu) {
      closeTypeMenu();
      return;
    }
    await openTypeMenu();
  };

  const selectType = (option: string) => {
    typeSelection = option;
    if (option !== "VARCHAR2") {
      lengthInput = "";
    }
    closeTypeMenu();
  };

  const handleWindowPointerDown = (event: MouseEvent) => {
    if (!showTypeMenu || !typeMenuEl) {
      return;
    }

    if (!typeMenuEl.contains(event.target as Node)) {
      closeTypeMenu();
    }
  };

  const handleTypeTriggerKeydown = async (event: KeyboardEvent) => {
    if (event.key === "ArrowDown" || event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      await openTypeMenu();
    }
  };

  const handleTypeMenuKeydown = (event: KeyboardEvent) => {
    if (event.key === "Escape") {
      event.preventDefault();
      closeTypeMenu();
    }
  };

  $: filteredTypeOptions = typeOptions.filter((option) =>
    option.toLowerCase().includes(typeSearch.trim().toLowerCase())
  );

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

<svelte:window on:mousedown={handleWindowPointerDown} />

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
      <div class="type-picker" bind:this={typeMenuEl}>
        <button
          id="attr-type"
          class={`type-trigger ${showTypeMenu ? 'type-trigger--open' : ''}`}
          type="button"
          aria-haspopup="listbox"
          aria-expanded={showTypeMenu}
          aria-controls="attr-type-listbox"
          on:click={toggleTypeMenu}
          on:keydown={handleTypeTriggerKeydown}
        >
          <span class="type-trigger-copy">
            <span class="type-trigger-label">Tipo elegido</span>
            <strong>{typeSelection}</strong>
          </span>
          <svg class:chevron-open={showTypeMenu} viewBox="0 0 24 24" aria-hidden="true">
            <path d="M6.72 8.97a.75.75 0 0 1 1.06 0L12 13.19l4.22-4.22a.75.75 0 1 1 1.06 1.06l-4.75 4.75a.75.75 0 0 1-1.06 0L6.72 10.03a.75.75 0 0 1 0-1.06Z"/>
          </svg>
        </button>

        {#if showTypeMenu}
          <div
            id="attr-type-listbox"
            class="type-menu"
            role="listbox"
            tabindex="-1"
            transition:scale={{duration: 130, start: 0.97}}
            on:keydown={handleTypeMenuKeydown}
          >
            <div class="type-search-shell">
              <input
                bind:this={typeSearchInput}
                class="type-search"
                type="text"
                autocomplete="off"
                placeholder="Buscar tipo..."
                bind:value={typeSearch}
                aria-label="Buscar tipo de atributo"
              />
            </div>
            <div class="type-options">
              {#if filteredTypeOptions.length}
                {#each filteredTypeOptions as option}
                  <button
                    class={`type-option ${typeSelection === option ? 'type-option--active' : ''}`}
                    type="button"
                    role="option"
                    aria-selected={typeSelection === option}
                    on:click={() => selectType(option)}
                  >
                    <span>{option}</span>
                    {#if typeSelection === option}
                      <span class="type-option-check">Actual</span>
                    {/if}
                  </button>
                {/each}
              {:else}
                <div class="type-empty">No hay coincidencias para "{typeSearch}".</div>
              {/if}
            </div>
          </div>
        {/if}
      </div>
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
    appearance: none;
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

  .type-row {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: start;
    gap: 10px;
  }

  .type-picker {
    position: relative;
  }

  .type-trigger {
    width: 100%;
    min-height: 58px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 14px;
    text-align: left;
    border-radius: 1rem;
    border: 1px solid var(--border);
    background: linear-gradient(180deg, color-mix(in srgb, var(--field-surface) 94%, var(--surface)), var(--field-surface));
    color: var(--ink);
    padding: 0.85rem 1rem;
    outline: none;
    transition: border 140ms ease, box-shadow 140ms ease, transform 140ms ease, background 140ms ease;
  }

  .type-trigger:hover {
    transform: translateY(-1px);
    border-color: color-mix(in srgb, var(--accent) 28%, var(--border));
  }

  .type-trigger:focus-visible,
  .type-trigger--open {
    border-color: var(--focus-border);
    box-shadow: var(--focus-ring);
    background: var(--field-surface-focus);
  }

  .type-trigger-copy {
    display: grid;
    gap: 0.18rem;
  }

  .type-trigger-label {
    color: var(--ink-faint);
    font-size: 0.72rem;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    font-weight: 700;
  }

  .type-trigger strong {
    font-size: 0.98rem;
    font-weight: 700;
    color: var(--ink);
  }

  .type-trigger svg {
    width: 20px;
    height: 20px;
    flex: 0 0 auto;
    color: var(--ink-soft);
    transition: transform 140ms ease, color 140ms ease;
  }

  .type-trigger svg.chevron-open {
    transform: rotate(180deg);
    color: var(--accent);
  }

  .type-menu {
    position: absolute;
    top: calc(100% + 0.55rem);
    left: 0;
    right: 0;
    z-index: 12;
    display: grid;
    gap: 0.55rem;
    padding: 0.7rem;
    border-radius: 1.15rem;
    border: 1px solid color-mix(in srgb, var(--accent) 16%, var(--border));
    background: color-mix(in srgb, var(--popover-surface) 94%, var(--surface));
    box-shadow: var(--shadow-lg);
    transform-origin: top center;
  }

  .type-search-shell {
    padding-bottom: 0.05rem;
    border-bottom: 1px solid var(--line-soft);
  }

  .type-search {
    padding: 0.78rem 0.9rem;
    border-radius: 0.9rem;
    font-size: 0.92rem;
  }

  .type-options {
    display: grid;
    gap: 0.36rem;
    max-height: 240px;
    overflow-y: auto;
    padding-right: 0.08rem;
  }

  .type-option {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    text-align: left;
    border: 1px solid transparent;
    border-radius: 0.95rem;
    background: transparent;
    color: var(--ink);
    padding: 0.8rem 0.9rem;
    transition: background 130ms ease, border-color 130ms ease, transform 130ms ease;
  }

  .type-option:hover {
    background: var(--hover-soft);
    border-color: var(--line-soft);
    transform: translateX(2px);
  }

  .type-option--active {
    background: color-mix(in srgb, var(--accent) 12%, var(--surface));
    border-color: color-mix(in srgb, var(--accent) 26%, var(--border));
  }

  .type-option-check {
    display: inline-flex;
    align-items: center;
    padding: 0.25rem 0.55rem;
    border-radius: 999px;
    background: color-mix(in srgb, var(--accent) 14%, transparent);
    color: var(--accent);
    font-size: 0.72rem;
    font-weight: 700;
    letter-spacing: 0.04em;
  }

  .type-empty {
    padding: 0.8rem 0.9rem;
    border-radius: 0.95rem;
    background: color-mix(in srgb, var(--surface) 82%, transparent);
    color: var(--ink-faint);
    font-size: 0.88rem;
  }

  .length-input {
    width: 120px;
    min-width: 120px;
  }

  .helper {
    margin: 0.25rem 0 0;
    color: var(--ink-faint);
    font-size: 0.8rem;
  }

  .form-error {
    margin: 0.2rem 0 0;
    color: var(--danger);
    font-weight: 600;
  }

  @media (max-width: 640px) {
    .type-row {
      grid-template-columns: 1fr;
    }

    .length-input {
      width: 100%;
      min-width: 0;
    }
  }
</style>
