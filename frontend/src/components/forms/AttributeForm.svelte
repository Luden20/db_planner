<script lang="ts">
  import { onDestroy, tick } from "svelte";
  import { scale } from "svelte/transition";
  import ButtonIcon from "../ButtonIcon.svelte";
  import ModalLauncher from "../ModalLauncher.svelte";
  import { AddAttribute, AddIntersectionAttribute, EditAttribute, EditIntersectionAttribute, Save } from "../../../wailsjs/go/main/App";
  import type { utils } from "../../../wailsjs/go/models";
  import { showToast } from "../../lib/toast";

  let { 
    entityId = null, 
    relationId = null, 
    entity = null, 
    attribute = null, 
    onSaved = async () => {}, 
    triggerClass = "", 
    allowPrimaryKey = true 
  } = $props<{
    entityId?: number | null;
    relationId?: number | null;
    entity?: utils.Entity | null;
    attribute?: utils.Attribute | null;
    onSaved?: () => Promise<void>;
    triggerClass?: string;
    allowPrimaryKey?: boolean;
  }>();

  const typeOptions = [
    "Por definir",
    "Numérico", "Cadena", "Fecha", "Tiempo", "Booleano", "N/A",
    "--- PowerDesigner Text ---",
    "Characters", "Variable Characters", "Long Varchar", "Text", "Multi-line Variable Characters",
    "--- PowerDesigner Numeric ---",
    "Integer", "Long Integer", "Short Integer", "Byte",
    "Decimal", "Float", "Double", "Number", "Money",
    "--- PowerDesigner Other ---",
    "Date & Time", "Timestamp", "Binary", "Variable Binary", "Long Binary", "Image", "UniqueID"
  ];

  let name = $state("");
  let description = $state("");
  let error = $state("");
  let typeSelection = $state("Por definir");
  let isPrimaryKey = $state(false);
  let isOptional = $state(false);
  let lengthInput = $state("");
  let domainValues = $state<string[]>([]);
  let domainDraft = $state("");
  let showTypeMenu = $state(false);
  let typeMenuEl = $state<HTMLDivElement | null>(null);
  let typeTriggerEl = $state<HTMLButtonElement | null>(null);
  let modalBodyEl = $state<HTMLElement | null>(null);
  let typeMenuStyle = $state("");

  const portal = (node: HTMLElement) => {
    if (typeof document === "undefined") return {};
    document.body.appendChild(node);
    return {
      destroy() {
        if (node.parentNode === document.body) document.body.removeChild(node);
      }
    };
  };

  const parseType = (value: string) => {
    const match = (value || "").match(/^\s*([^\(]+)\s*\(\s*([^\)]+)\s*\)\s*$/i);
    if (match) {
      return { type: match[1].trim(), length: match[2].trim() };
    }
    return { type: value || "Por definir", length: "" };
  };

  const prefill = () => {
    if (attribute) {
      name = attribute.Name;
      description = attribute.Description;
      const parsed = parseType(attribute.Type || "Por definir");
      typeSelection = parsed.type;
      isPrimaryKey = allowPrimaryKey && attribute.KeyType === "pk";
      isOptional = !!attribute.Optional;
      lengthInput = parsed.length;
      domainValues = Array.isArray(attribute.Domain) ? [...attribute.Domain] : [];
    } else {
      name = "";
      description = "";
      typeSelection = "Por definir";
      isPrimaryKey = false;
      isOptional = false;
      lengthInput = "";
      domainValues = [];
    }
    domainDraft = "";
    error = "";
    showTypeMenu = false;
  };

  const normalizeDomainValues = (values: string[]) => {
    const seen = new Set<string>();
    return values.map((v) => v.trim()).filter((v) => {
      if (!v || seen.has(v)) return false;
      seen.add(v);
      return true;
    });
  };

  const addDomainValue = () => {
    const nextValue = domainDraft.trim();
    if (!nextValue) return;
    domainValues = normalizeDomainValues([...domainValues, nextValue]);
    domainDraft = "";
  };

  const removeDomainValue = (value: string) => {
    domainValues = domainValues.filter((v) => v !== value);
  };

  const handleDomainKeydown = (event: KeyboardEvent) => {
    if (event.key === "Enter") {
      event.preventDefault();
      addDomainValue();
    }
  };

  const hasAnotherPrimaryKey = () => allowPrimaryKey && (entity?.Attributes || []).some((item) => item.KeyType === "pk" && item.Id !== attribute?.Id);

  const togglePrimaryKey = () => {
    if (!allowPrimaryKey) return;
    if (!isPrimaryKey && hasAnotherPrimaryKey()) {
      error = "La entidad fuerte ya tiene una PK.";
      showToast(error, "error");
      return;
    }
    error = "";
    isPrimaryKey = !isPrimaryKey;
    if (isPrimaryKey) isOptional = false;
  };

  const toggleOptional = () => {
    if (isPrimaryKey) return;
    isOptional = !isOptional;
  };

  const openTypeMenu = async () => {
    showTypeMenu = true;
    await tick();
    syncTypeMenuPosition();
  };

  const closeTypeMenu = () => { showTypeMenu = false; };
  const toggleTypeMenu = async () => {
    if (showTypeMenu) { closeTypeMenu(); return; }
    await openTypeMenu();
  };

  const isParameterized = (type: string) => {
    const t = type.toLowerCase();
    return t.includes("char") || t.includes("cadena") || t.includes("varchar") || t.includes("string") || t.includes("text") || t.includes("decimal") || t.includes("dec") || t.includes("numeric") || t.includes("number") || t.includes("float") || t.includes("double") || t.includes("binary");
  };

  const selectType = (option: string) => {
    if (option.startsWith("---")) return;
    typeSelection = option;
    if (!isParameterized(option)) lengthInput = "";
    closeTypeMenu();
  };

  const handleWindowPointerDown = (event: MouseEvent) => {
    if (!showTypeMenu) return;
    const target = event.target as Node;
    if (typeMenuEl?.contains(target) || typeTriggerEl?.contains(target)) return;
    closeTypeMenu();
  };

  const handleTypeTriggerKeydown = async (event: KeyboardEvent) => {
    if (event.key === "ArrowDown" || event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      await openTypeMenu();
    }
  };

  const handleTypeMenuKeydown = (event: KeyboardEvent) => { if (event.key === "Escape") { event.preventDefault(); closeTypeMenu(); } };

  const syncTypeMenuPosition = () => {
    if (!typeTriggerEl || typeof window === "undefined") return;
    const rect = typeTriggerEl.getBoundingClientRect();
    const viewportPadding = 16;
    const menuWidth = Math.max(rect.width * 0.62, 180);
    const estimatedHeight = 220;
    const spaceBelow = window.innerHeight - rect.bottom - viewportPadding;
    const spaceAbove = rect.top - viewportPadding;
    const openUpwards = spaceBelow < 220 && spaceAbove > spaceBelow;
    const top = openUpwards ? Math.max(viewportPadding, rect.top - Math.min(spaceAbove, estimatedHeight) - 10) : Math.max(viewportPadding, rect.bottom + 10);
    const left = Math.min(Math.max(viewportPadding, rect.left), Math.max(viewportPadding, window.innerWidth - menuWidth - viewportPadding));
    const maxHeight = Math.max(180, openUpwards ? spaceAbove - 14 : spaceBelow - 14);
    typeMenuStyle = `top:${top}px;left:${left}px;width:${menuWidth}px;max-height:${maxHeight}px;`;
  };

  const handleViewportChange = () => { if (showTypeMenu) syncTypeMenuPosition(); };

  const syncModalBodyRef = async () => {
    await tick();
    const nextModalBody = typeTriggerEl?.closest(".modal-body") as HTMLElement | null;
    if (modalBodyEl === nextModalBody) return;
    modalBodyEl?.removeEventListener("scroll", handleViewportChange);
    modalBodyEl = nextModalBody;
    modalBodyEl?.addEventListener("scroll", handleViewportChange, { passive: true });
  };

  $effect(() => { if (showTypeMenu) void syncModalBodyRef(); });

  onDestroy(() => { modalBodyEl?.removeEventListener("scroll", handleViewportChange); });

  const handleSave = async () => {
    const trimmedName = name.trim();
    if (!trimmedName) { error = "Ingresa un nombre para el atributo."; throw new Error(error); }
    const selType = typeSelection || "Por definir";
    let finalType = selType;
    if (isParameterized(selType) && lengthInput.trim()) finalType = `${selType}(${lengthInput.trim()})`;

    try {
      error = "";
      if (allowPrimaryKey && isPrimaryKey && hasAnotherPrimaryKey()) { error = "La entidad fuerte ya tiene una PK."; throw new Error(error); }
      const keyTypeSelection = isPrimaryKey ? "pk" : "nil";
      if (relationId !== null) {
        if (attribute) await EditIntersectionAttribute(relationId, attribute.Id, trimmedName, description.trim(), finalType, isOptional, normalizeDomainValues(domainValues));
        else await AddIntersectionAttribute(relationId, trimmedName, description.trim(), finalType, isOptional, normalizeDomainValues(domainValues));
      } else if (entityId !== null && attribute) {
        await EditAttribute(entityId, attribute.Id, trimmedName, description.trim(), finalType, keyTypeSelection, isOptional, normalizeDomainValues(domainValues));
      } else if (entityId !== null) {
        await AddAttribute(entityId, trimmedName, description.trim(), finalType, keyTypeSelection, isOptional, normalizeDomainValues(domainValues));
      } else throw new Error("No se encontro la entidad destino.");
      await Save();
      await onSaved();
      if (!attribute) { name = ""; description = ""; typeSelection = "Por definir"; isPrimaryKey = false; isOptional = false; lengthInput = ""; domainValues = []; domainDraft = ""; }
    } catch (err: any) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      showToast(`No se pudo guardar el atributo: ${message}`, "error");
      throw err;
    }
  };
</script>

<svelte:window onmousedown={handleWindowPointerDown} onresize={handleViewportChange} onscroll={handleViewportChange} />

<ModalLauncher
  triggerLabel={attribute ? "Editar" : "Nuevo atributo"}
  title={attribute ? "Editar atributo" : "Crear atributo"}
  confirmLabel="Guardar"
  triggerVariant={attribute ? "edit" : "primary"}
  confirmVariant="primary"
  size="form"
  {triggerClass}
  onOpen={prefill}
  onSuccess={handleSave}
>
  <div class="field">
    <label for="attr-name">Nombre</label>
    <input id="attr-name" type="text" autocomplete="off" placeholder="Código, Nombre, Estado..." bind:value={name} />
  </div>

  <div class="field">
    <label for="attr-description">Descripción</label>
    <textarea id="attr-description" rows="1" placeholder="Breve descripción" bind:value={description}></textarea>
  </div>

  <div class="field">
    <label for="attr-type">Tipo</label>
    <div class="type-row">
      <div class="type-picker">
        <button
          id="attr-type"
          bind:this={typeTriggerEl}
          class={`type-trigger ${showTypeMenu ? 'type-trigger--open' : ''}`}
          type="button"
          aria-haspopup="listbox"
          aria-expanded={showTypeMenu}
          aria-controls="attr-type-listbox"
          onclick={toggleTypeMenu}
          onkeydown={handleTypeTriggerKeydown}
        >
          <span class="type-trigger-copy">
            <span class="type-trigger-label">Tipo elegido</span>
            <strong>{typeSelection}</strong>
          </span>
          <ButtonIcon name="chevron-down" />
        </button>
      </div>
      {#if isParameterized(typeSelection)}
        <input class="length-input" type="text" placeholder="Ej: 255 o 10,2" bind:value={lengthInput} aria-label="Parámetros del tipo" />
      {/if}
    </div>
    {#if isParameterized(typeSelection)}
      <p class="helper">Indica la longitud o precisión/escala para {typeSelection}.</p>
    {/if}
  </div>

  {#if allowPrimaryKey}
    <div class="field">
      <label for="attr-key-type">Clave primaria</label>
      <button
        id="attr-key-type"
        type="button"
        class={`pk-toggle ${isPrimaryKey ? 'pk-toggle--active' : ''}`}
        role="switch"
        aria-checked={isPrimaryKey}
        onclick={togglePrimaryKey}
      >
        <span class="pk-toggle__rail"><span class="pk-toggle__thumb"></span></span>
        <span class="pk-toggle__copy">
          <strong>{isPrimaryKey ? "PK activa" : "Sin PK"}</strong>
          <span>Las entidades fuertes solo pueden tener una PK.</span>
        </span>
      </button>
    </div>
  {/if}

  <div class="field">
    <label for="attr-optional">Nulabilidad</label>
    <button
      id="attr-optional"
      type="button"
      class={`pk-toggle ${isOptional ? 'pk-toggle--active' : ''}`}
      role="switch"
      disabled={isPrimaryKey}
      aria-checked={isOptional}
      onclick={toggleOptional}
    >
      <span class="pk-toggle__rail"><span class="pk-toggle__thumb"></span></span>
      <span class="pk-toggle__copy">
        <strong>{isOptional ? "Opcional" : "Mandatorio"}</strong>
        <span>{isPrimaryKey ? "Las PK siempre son mandatorias." : (isOptional ? "Este atributo puede ser nulo." : "Este atributo es obligatorio.")}</span>
      </span>
    </button>
  </div>

  <div class="field">
    <label for="attr-domain">Dominio</label>
    <div class="domain-editor">
      <div class="domain-input-row">
        <input id="attr-domain" type="text" autocomplete="off" placeholder="Activo, Inactivo, Pendiente..." bind:value={domainDraft} onkeydown={handleDomainKeydown} />
        <button class="domain-add control control--soft" type="button" onclick={addDomainValue} disabled={!domainDraft.trim()}>
          <ButtonIcon name="plus"/><span>Agregar</span>
        </button>
      </div>
      {#if domainValues.length}
        <div class="domain-list">
          {#each domainValues as dVal}
            <button class="domain-chip" type="button" onclick={() => removeDomainValue(dVal)} aria-label={`Quitar ${dVal} del dominio`}>
              <span>{dVal}</span><ButtonIcon name="close"/>
            </button>
          {/each}
        </div>
      {:else}
        <p class="helper">Sin opciones definidas. Si no agregas ninguna, se guarda como `[]`.</p>
      {/if}
    </div>
  </div>

  {#if error}
    <p class="form-error">{error}</p>
  {/if}
</ModalLauncher>

{#if showTypeMenu}
  <div
    id="attr-type-listbox"
    bind:this={typeMenuEl}
    class="type-menu"
    style={typeMenuStyle}
    role="listbox"
    tabindex="-1"
    use:portal
    transition:scale={{ duration: 130, start: 0.97 }}
    onkeydown={handleTypeMenuKeydown}
  >
    <div class="type-options">
      {#each typeOptions as option}
        <button
          class={`type-option ${typeSelection === option ? 'type-option--active' : ''} ${option.startsWith('---') ? 'type-option--separator' : ''}`}
          type="button"
          role="option"
          aria-selected={typeSelection === option}
          onclick={() => selectType(option)}
          disabled={option.startsWith('---')}
        >
          <span>{option.startsWith('---') ? option.replace(/---/g, '').trim() : option}</span>
          {#if typeSelection === option}<span class="type-option-check">Actual</span>{/if}
        </button>
      {/each}
    </div>
  </div>
{/if}

<style>
  .field { display: grid; gap: 0.65rem; color: var(--ink-soft); font-size: 0.92rem; padding: 0.25rem 0.1rem; }
  .field input, .field textarea, .field select { width: 100%; box-sizing: border-box; border-radius: 1rem; border: 1px solid var(--border); background: var(--field-surface); color: var(--ink); padding: 0.9rem 1rem; font-size: 0.96rem; outline: none; transition: all 140ms ease; appearance: none; }
  .field textarea { min-height: 144px; resize: vertical; }
  .field input:focus, .field textarea:focus, .field select:focus { border-color: var(--focus-border); box-shadow: var(--focus-ring); background: var(--field-surface-focus); }
  .type-row { display: grid; grid-template-columns: minmax(0, 1fr) auto; align-items: start; gap: 10px; }
  .pk-toggle { display: inline-flex; align-items: center; gap: 0.9rem; width: 100%; min-height: 58px; border-radius: 1rem; border: 1px solid var(--border); background: var(--field-surface); color: var(--ink); padding: 0.85rem 1rem; text-align: left; transition: all 140ms ease; cursor: pointer; }
  .pk-toggle:hover:not(:disabled) { transform: translateY(-1px); border-color: color-mix(in srgb, var(--accent) 28%, var(--border)); }
  .pk-toggle:focus-visible { border-color: var(--focus-border); box-shadow: var(--focus-ring); background: var(--field-surface-focus); }
  .pk-toggle--active { border-color: color-mix(in srgb, var(--accent) 30%, var(--border)); background: color-mix(in srgb, var(--accent) 10%, var(--field-surface)); }
  .pk-toggle__rail { position: relative; flex: 0 0 auto; width: 46px; height: 28px; border-radius: 999px; background: color-mix(in srgb, var(--ink-faint) 35%, var(--surface)); transition: background 140ms ease; }
  .pk-toggle--active .pk-toggle__rail { background: color-mix(in srgb, var(--accent) 72%, white 12%); }
  .pk-toggle__thumb { position: absolute; top: 3px; left: 3px; width: 22px; height: 22px; border-radius: 50%; background: white; box-shadow: 0 2px 8px rgba(15, 23, 42, 0.18); transition: transform 140ms ease; }
  .pk-toggle--active .pk-toggle__thumb { transform: translateX(18px); }
  .pk-toggle__copy { display: grid; gap: 0.14rem; min-width: 0; }
  .pk-toggle__copy strong { font-size: 0.95rem; color: var(--ink); }
  .pk-toggle__copy span { color: var(--ink-soft); font-size: 0.84rem; line-height: 1.35; }
  .type-picker { min-width: 0; }
  .type-trigger { width: 100%; min-height: 58px; display: flex; align-items: center; justify-content: space-between; gap: 14px; text-align: left; border-radius: 1rem; border: 1px solid var(--border); background: linear-gradient(180deg, color-mix(in srgb, var(--field-surface) 94%, var(--surface)), var(--field-surface)); color: var(--ink); padding: 0.85rem 1rem; outline: none; transition: all 140ms ease; cursor: pointer; }
  .type-trigger:hover { transform: translateY(-1px); border-color: color-mix(in srgb, var(--accent) 28%, var(--border)); }
  .type-trigger:focus-visible, .type-trigger--open { border-color: var(--focus-border); box-shadow: var(--focus-ring); background: var(--field-surface-focus); }
  .type-trigger-copy { display: grid; gap: 0.18rem; }
  .type-trigger-label { color: var(--ink-faint); font-size: 0.72rem; letter-spacing: 0.12em; text-transform: uppercase; font-weight: 700; }
  .type-trigger strong { font-size: 0.98rem; font-weight: 700; color: var(--ink); }
  .type-menu { position: fixed; z-index: calc(var(--layer-toast) + 1); display: grid; gap: 0.3rem; padding: 0.42rem; border-radius: 0.82rem; border: 1px solid color-mix(in srgb, var(--accent) 16%, var(--border)); background: var(--surface-strong); box-shadow: var(--shadow-lg); transform-origin: top left; overflow: hidden; }
  .type-options { display: grid; gap: 0.36rem; max-height: inherit; overflow-y: auto; padding-right: 0.08rem; }
  .type-option { width: 100%; display: flex; align-items: center; justify-content: space-between; gap: 10px; text-align: left; border: 1px solid transparent; border-radius: 0.72rem; background: transparent; color: var(--ink); padding: 0.5rem 0.66rem; transition: all 130ms ease; cursor: pointer; }
  .type-option:hover { background: var(--hover-soft); border-color: var(--line-soft); transform: translateX(2px); }
  .type-option--active { background: color-mix(in srgb, var(--accent) 12%, var(--surface)); border-color: color-mix(in srgb, var(--accent) 26%, var(--border)); }
  .type-option--separator { background: color-mix(in srgb, var(--accent) 10%, var(--surface-medium)); color: var(--accent); font-size: 0.7rem; font-weight: 800; text-transform: uppercase; letter-spacing: 0.1em; padding: 0.6rem 0.66rem; margin-top: 0.5rem; pointer-events: none; border: none; }
  .type-option--separator:first-child { margin-top: 0; }
  .type-option-check { display: inline-flex; align-items: center; padding: 0.25rem 0.55rem; border-radius: 999px; background: color-mix(in srgb, var(--accent) 14%, transparent); color: var(--accent); font-size: 0.72rem; font-weight: 700; letter-spacing: 0.04em; }
  .length-input { width: 120px; min-width: 120px; }
  .helper { margin: 0.25rem 0 0; color: var(--ink-faint); font-size: 0.8rem; }
  .domain-editor { display: grid; gap: 0.8rem; }
  .domain-input-row { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 0.7rem; align-items: center; }
  .domain-add { min-height: 3rem; padding-inline: 0.95rem; white-space: nowrap; }
  .domain-list { display: flex; flex-wrap: wrap; gap: 0.45rem; padding-top: 0.25rem; }
  .domain-chip { display: flex; align-items: center; gap: 0.5rem; background: var(--surface-medium); border: 1px solid var(--border); border-radius: 999px; padding: 0.35rem 0.8rem; font-size: 0.8rem; font-weight: 600; color: var(--ink-soft); transition: all 140ms ease; cursor: pointer; }
  .domain-chip:hover { border-color: var(--danger); color: var(--danger); background: color-mix(in srgb, var(--danger) 8%, var(--background)); }
  .form-error { color: var(--danger); font-size: 0.85rem; font-weight: 700; background: color-mix(in srgb, var(--danger) 10%, transparent); border-left: 3px solid var(--danger); padding: 0.75rem 1rem; border-radius: 4px; margin: 0; }
</style>
