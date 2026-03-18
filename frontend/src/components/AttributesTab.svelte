<script lang="ts">
  import {onMount} from "svelte";
  import type {utils} from "../../wailsjs/go/models";
  import {GetCombinatory} from "../../wailsjs/go/main/App";
  import {GetEntity, MarkEntityStatus, MoveAttribute, Save} from "../../wailsjs/go/main/App";
  import CreateEntity from "./forms/CreateEntity.svelte";
  import AttributeForm from "./forms/AttributeForm.svelte";
  import DeleteAttribute from "./forms/DeleteAttribute.svelte";
  import {showToast} from "../lib/toast";

  type RelationGroup = {
    type: string;
    label: string;
    items: string[];
  };

  export let entities: utils.Entity[] = [];
  export let onRefresh: () => Promise<void> = async () => {};
  export let focusEntityId: number | null = null;
  export let onJumpTo: (tab: "entities" | "relations" | "tertiary", entityId?: number | null) => void = () => {};

  let tableWrapper: HTMLDivElement | null = null;
  let selectedId: number | null = null;
  let current: utils.Entity | null = null;
  let loading = false;
  let draggingIndex: number | null = null;
  let hoverIndex: number | null = null;
  let lastLoadedId: number | null = null;
  let lastSyncedFocusId: number | null = null;
  let relationSummary: RelationGroup[] = [];
  let autoScrollFrame: number | null = null;
  let autoScrollDirection: -1 | 0 | 1 = 0;
  let approvalUpdating = false;

  const AUTO_SCROLL_EDGE_PX = 72;
  const AUTO_SCROLL_STEP = 14;
  const relationTypeLabels: Record<string, string> = {
    "1:1": "Uno a uno",
    "1:N": "Uno a muchos",
    "N:1": "Muchos a uno",
    "N:N": "Muchos a muchos"
  };
  const relationTypeOrder = ["1:1", "1:N", "N:1", "N:N"];

  const runAutoScroll = () => {
    if (!tableWrapper || autoScrollDirection === 0) {
      autoScrollFrame = null;
      return;
    }

    tableWrapper.scrollTop += autoScrollDirection * AUTO_SCROLL_STEP;
    autoScrollFrame = window.requestAnimationFrame(runAutoScroll);
  };

  const startAutoScroll = (direction: -1 | 1) => {
    if (autoScrollDirection === direction && autoScrollFrame !== null) {
      return;
    }

    autoScrollDirection = direction;
    if (autoScrollFrame === null) {
      autoScrollFrame = window.requestAnimationFrame(runAutoScroll);
    }
  };

  const stopAutoScroll = () => {
    autoScrollDirection = 0;
    if (autoScrollFrame !== null) {
      window.cancelAnimationFrame(autoScrollFrame);
      autoScrollFrame = null;
    }
  };

  const updateAutoScroll = (event: DragEvent) => {
    if (!tableWrapper) {
      return;
    }

    const bounds = tableWrapper.getBoundingClientRect();
    if (event.clientY <= bounds.top + AUTO_SCROLL_EDGE_PX) {
      startAutoScroll(-1);
      return;
    }
    if (event.clientY >= bounds.bottom - AUTO_SCROLL_EDGE_PX) {
      startAutoScroll(1);
      return;
    }

    stopAutoScroll();
  };

  onMount(() => {
    if (entities.length && selectedId === null) {
      selectedId = entities[0].Id;
    }
  });

  $: if (entities.length && selectedId === null) {
    selectedId = entities[0].Id;
  }

  $: if (
    focusEntityId !== null &&
    focusEntityId !== lastSyncedFocusId &&
    entities.some(entity => entity.Id === focusEntityId)
  ) {
    selectedId = focusEntityId;
    lastSyncedFocusId = focusEntityId;
  }

  $: if (selectedId !== null && selectedId !== lastLoadedId) {
    loadEntity(selectedId);
  }

  const loadEntity = async (id: number) => {
    loading = true;
    try {
      current = await GetEntity(id);
      lastLoadedId = id;
      await loadRelationSummary(current);
    } catch (err) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      showToast(`No se pudo cargar la entidad: ${message}`, "error");
    } finally {
      loading = false;
    }
  };

  const clearDrag = () => {
    stopAutoScroll();
    draggingIndex = null;
    hoverIndex = null;
  };

  const startDrag = (index: number, event: DragEvent) => {
    draggingIndex = index;
    hoverIndex = index;
    event.dataTransfer?.setData("text/plain", `${index}`);
  };

  const handleSelectChange = (event: Event) => {
    const target = event.target as HTMLSelectElement;
    selectedId = Number(target?.value ?? 0);
  };

  const handleDragOver = (index: number, event: DragEvent) => {
    event.preventDefault();
    hoverIndex = index;
    updateAutoScroll(event);
  };

  const handleTableDragOver = (event: DragEvent) => {
    event.preventDefault();
    updateAutoScroll(event);
  };

  const handleTableDragLeave = (event: DragEvent) => {
    const nextTarget = event.relatedTarget as Node | null;
    if (tableWrapper && nextTarget && tableWrapper.contains(nextTarget)) {
      return;
    }
    stopAutoScroll();
  };

  const nextEntity = () => {
    if (!entities.length) return;
    const currentIndex = entities.findIndex((ent) => ent.Id === selectedId);
    const nextIndex = currentIndex === -1 || currentIndex === entities.length - 1 ? 0 : currentIndex + 1;
    selectedId = entities[nextIndex].Id;
  };

  const prevEntity = () => {
    if (!entities.length) return;
    const currentIndex = entities.findIndex((ent) => ent.Id === selectedId);
    const prevIndex = currentIndex <= 0 ? entities.length - 1 : currentIndex - 1;
    selectedId = entities[prevIndex].Id;
  };

  const applyReorder = async (from: number, to: number) => {
    if (!current || from === to || from < 0 || to < 0 || from >= current.Attributes.length || to >= current.Attributes.length) {
      return;
    }
    const direction: "up" | "down" = to < from ? "up" : "down";
    const steps = Math.abs(to - from);
    const attributeId = current.Attributes[from].Id;
    try {
      for (let i = 0; i < steps; i++) {
        await MoveAttribute(current.Id, attributeId, direction);
      }
      await Save();
      await onRefresh();
      await loadEntity(current.Id);
    } catch (err) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      showToast(`No se pudo reordenar el atributo: ${message}`, "error");
    }
  };

  const handleDrop = async (targetIndex: number, event: DragEvent) => {
    event.preventDefault();
    if (draggingIndex === null) {
      clearDrag();
      return;
    }
    await applyReorder(draggingIndex, targetIndex);
    clearDrag();
  };

  const invertRelationValue = (relation: string) => {
    switch (relation) {
      case "1:1":
        return "1:1";
      case "1:N":
        return "N:1";
      case "N:1":
        return "1:N";
      default:
        return relation;
    }
  };

  const loadRelationSummary = async (ent: utils.Entity | null) => {
    if (!ent) {
      relationSummary = [];
      return;
    }
    try {
      const comb = await GetCombinatory();
      const groupedItems = new Map<string, string[]>();
      (comb || []).forEach(view => {
        if (!view.Relations) return;
        view.Relations.forEach(rel => {
          const val = rel.Relation || "";
          if (!val) return;
          let summaryType = val;
          let summaryTarget = rel.Entity2;
          if (view.IdPrincipalEntity === ent.Id) {
            summaryType = val;
            summaryTarget = rel.Entity2;
          } else if (rel.IdEntity2 === ent.Id) {
            summaryType = invertRelationValue(val);
            summaryTarget = view.PrincipalEntity;
          } else {
            return;
          }

          const currentItems = groupedItems.get(summaryType) || [];
          currentItems.push(`con ${summaryTarget}`);
          groupedItems.set(summaryType, currentItems);
        });
      });
      relationSummary = relationTypeOrder
        .filter(type => groupedItems.has(type))
        .map(type => ({
          type,
          label: relationTypeLabels[type] || type,
          items: groupedItems.get(type) || []
        }));
    } catch (err) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      console.error("No se pudo cargar resumen de relaciones:", message);
      relationSummary = [];
    }
  };

  const isApproved = (entity: utils.Entity | null) => entity?.Status === true;

  const toggleCurrentApproval = async () => {
    if (!current) {
      return;
    }

    approvalUpdating = true;
    try {
      await MarkEntityStatus(current.Id, !isApproved(current));
      await onRefresh();
      await loadEntity(current.Id);
    } catch (err) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      showToast(`No se pudo actualizar la aprobación: ${message}`, "error");
    } finally {
      approvalUpdating = false;
    }
  };

  const jumpToTab = (tab: "entities" | "relations") => {
    onJumpTo(tab, selectedId);
  };
</script>

<div class="tab-toolbar">
  <div>
    <p class="label">Atributos</p>
    <p class="muted">Selecciona una entidad para ver y reordenar sus atributos.</p>
  </div>
  <div class="toolbar-actions">
    <div class="view-jumps">
      <button class="jump-btn" on:click={() => jumpToTab("entities")} disabled={!selectedId}>
        Ir a definicion
      </button>
      <button class="jump-btn" on:click={() => jumpToTab("relations")} disabled={!selectedId}>
        Ir a combinatorio
      </button>
    </div>
    <AttributeForm
      entityId={selectedId ?? (entities[0]?.Id ?? 0)}
      onSaved={async () => {
        await onRefresh();
        if (selectedId !== null) {
          await loadEntity(selectedId);
        }
      }}
    />
    <select
      class="entity-select"
      bind:value={selectedId}
      on:change={handleSelectChange}
      disabled={!entities.length}
    >
      {#each entities as entity}
        <option value={entity.Id}>{entity.Name}</option>
      {/each}
    </select>
    <div class="entity-nav">
      <button class="nav-btn" on:click={prevEntity} aria-label="Entidad anterior" disabled={entities.length <= 1}>&lt;</button>
      <button class="nav-btn" on:click={nextEntity} aria-label="Entidad siguiente" disabled={entities.length <= 1}>&gt;</button>
    </div>
  </div>
</div>

{#if !entities.length}
  <div class="empty-panel">Crea entidades para gestionar atributos.</div>
{:else if loading}
  <div class="empty-panel">Cargando atributos...</div>
{:else if current}
  <div class:entity-status-card={true} class:entity-status-card--approved={isApproved(current)}>
    <div class="entity-primary-content">
      <p class="banner-title">Entidad seleccionada</p>
      <div class="entity-status-row">
        <h3>{current.Name}</h3>
        {#if isApproved(current)}
          <span class="status-pill status-pill--approved">&#10003;</span>
        {:else}
          <span class="status-pill">Pendiente</span>
        {/if}
      </div>
      <p class="entity-description">{current.Description || "Sin definición."}</p>
    </div>
    <div class="entity-card-actions">
      <CreateEntity
        id={current.Id}
        onSave={async () => {
          await onRefresh();
          await loadEntity(current.Id);
        }}
      />
    </div>
    <div class="entity-card-actions">
      <button
        class:approve-btn={true}
        class:approve-btn--approved={isApproved(current)}
        on:click={toggleCurrentApproval}
        disabled={approvalUpdating}
      >
        {isApproved(current) ? "Quitar aprobación" : "Aprobar entidad"}
      </button>
    </div>
  </div>
  {#if relationSummary.length}
    <div class:info-banner={true} class:info-banner--approved={isApproved(current)}>
      <p class="banner-title">Relaciones de esta entidad</p>
      <div class="relation-groups">
        {#each relationSummary as group}
          <section class="relation-group">
            <div class="relation-group-head">
              <span class="relation-group-type">{group.type}</span>
              <span class="relation-group-label">{group.label}</span>
            </div>
            <div class="pill-row">
              {#each group.items as item}
                <span class="pill">{item}</span>
              {/each}
            </div>
          </section>
        {/each}
      </div>
    </div>
  {/if}
  <div
    class="table-wrapper frosted"
    bind:this={tableWrapper}
    on:dragover={handleTableDragOver}
    on:dragleave={handleTableDragLeave}
    on:drop={stopAutoScroll}
  >
    <table class="entities-table">
      <thead>
      <tr>
        <th>Nombre</th>
        <th>Descripción</th>
        <th style="width: 120px;">Tipo</th>
        <th style="width: 180px;">Acciones</th>
      </tr>
      </thead>
      <tbody class="draggable-body">
      {#if !current.Attributes || current.Attributes.length === 0}
        <tr class="empty-row" draggable="false">
          <td colspan="4">No hay atributos definidos aún.</td>
        </tr>
      {:else}
        {#each current.Attributes as attribute, index (attribute.Id)}
          <tr
            class:dragging={draggingIndex === index}
            class:drag-hover={hoverIndex === index && draggingIndex !== null && draggingIndex !== index}
            draggable="true"
            on:dragstart={(event) => startDrag(index, event)}
            on:dragover={(event) => handleDragOver(index, event)}
            on:dragenter={(event) => handleDragOver(index, event)}
            on:drop={(event) => handleDrop(index, event)}
            on:dragend={clearDrag}
          >
            <td>{attribute.Name}</td>
            <td>{attribute.Description}</td>
            <td>{attribute.Type || "Por definir"}</td>
            <td>
              <div class="row-actions">
                <AttributeForm
                  entityId={current.Id}
                  attribute={attribute}
                  onSaved={async () => {
                    await onRefresh();
                    await loadEntity(current.Id);
                  }}
                />
                <DeleteAttribute
                  entityId={current.Id}
                  attributeId={attribute.Id}
                  onSaved={async () => {
                    await onRefresh();
                    await loadEntity(current.Id);
                  }}
                />
              </div>
            </td>
          </tr>
        {/each}
      {/if}
      </tbody>
    </table>
  </div>
{/if}

<style>
  .tab-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
    margin-bottom: 14px;
  }

  .label {
    margin: 0;
    color: #9ab5e4;
    font-size: 12px;
    letter-spacing: 0.6px;
    text-transform: uppercase;
  }

  .muted {
    margin: 6px 0 0;
    color: #cfd9e9;
    opacity: 0.75;
  }

  .table-wrapper {
    overflow: auto;
  }

  .toolbar-actions {
    display: flex;
    gap: 12px;
    align-items: center;
    flex-wrap: wrap;
    width: 100%;
  }

  .view-jumps {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
    width: 100%;
  }

  .jump-btn {
    border: 1px solid rgba(90, 209, 255, 0.22);
    background: rgba(90, 209, 255, 0.1);
    color: #dff5ff;
    border-radius: 10px;
    padding: 14px 16px;
    min-height: 56px;
    flex: 1 1 0;
    font-size: 18px;
    font-weight: 700;
    text-align: center;
    cursor: pointer;
    transition: background 140ms ease, transform 120ms ease, opacity 120ms ease;
  }

  .jump-btn:hover:enabled {
    background: rgba(90, 209, 255, 0.18);
  }

  .jump-btn:active:enabled {
    transform: translateY(1px);
  }

  .jump-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .info-banner {
    margin-bottom: 12px;
    padding: 12px;
    border-radius: 12px;
    background: linear-gradient(135deg, rgba(20, 32, 46, 0.55), rgba(27, 44, 63, 0.85));
    border: 1px solid rgba(255, 255, 255, 0.08);
  }

  .info-banner--approved {
    background: linear-gradient(135deg, rgba(18, 53, 27, 0.7), rgba(23, 67, 36, 0.88));
    border-color: rgba(113, 201, 118, 0.3);
  }

  .entity-status-card {
    margin-bottom: 12px;
    padding: 14px 16px;
    border-radius: 12px;
    background: linear-gradient(135deg, rgba(20, 32, 46, 0.7), rgba(24, 38, 54, 0.95));
    border: 1px solid rgba(255, 255, 255, 0.08);
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
  }

  .entity-primary-content {
    flex: 1;
    min-width: 0;
  }

  .entity-status-card--approved {
    background: linear-gradient(135deg, rgba(18, 53, 27, 0.78), rgba(23, 67, 36, 0.92));
    border-color: rgba(113, 201, 118, 0.3);
  }

  .banner-title {
    margin: 0 0 8px;
    font-size: 13px;
    letter-spacing: 0.4px;
    color: #9ab5e4;
    text-transform: uppercase;
  }

  .entity-status-row {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
  }

  .entity-status-row h3 {
    margin: 0;
    font-size: 20px;
  }

  .entity-description {
    margin: 10px 0 0;
    color: #d9e4f5;
    opacity: 0.82;
    line-height: 1.5;
  }

  .entity-card-actions {
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }

  .pill-row {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .relation-groups {
    display: grid;
    gap: 10px;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  }

  .relation-group {
    padding: 10px;
    border-radius: 12px;
    background: rgba(10, 18, 30, 0.28);
    border: 1px solid rgba(255, 255, 255, 0.08);
    display: grid;
    gap: 10px;
  }

  .relation-group-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    flex-wrap: wrap;
  }

  .relation-group-type {
    display: inline-flex;
    align-items: center;
    padding: 4px 8px;
    border-radius: 999px;
    background: rgba(90, 209, 255, 0.14);
    color: #dff5ff;
    border: 1px solid rgba(90, 209, 255, 0.2);
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0.3px;
  }

  .relation-group-label {
    color: #b9cbe6;
    font-size: 12px;
  }

  .pill {
    display: inline-flex;
    align-items: center;
    max-width: 220px;
    padding: 6px 10px;
    border-radius: 999px;
    background: rgba(90, 209, 255, 0.1);
    color: #d9e4f5;
    border: 1px solid rgba(90, 209, 255, 0.16);
    font-size: 13px;
    line-height: 1.35;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .status-pill {
    display: inline-flex;
    align-items: center;
    padding: 4px 10px;
    border-radius: 999px;
    background: rgba(255, 255, 255, 0.06);
    color: #d9e4f5;
    border: 1px solid rgba(255, 255, 255, 0.12);
    font-size: 12px;
    font-weight: 700;
  }

  .status-pill--approved {
    color: #dff7df;
    background: rgba(76, 175, 80, 0.2);
    border-color: rgba(113, 201, 118, 0.35);
    min-width: 30px;
    justify-content: center;
  }

  .entities-table {
    width: 100%;
    border-collapse: collapse;
    color: #e8edf7;
  }

  .entities-table th,
  .entities-table td {
    text-align: left;
    padding: 12px 10px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.07);
    font-size: 14px;
  }

  .entities-table thead th {
    font-size: 13px;
    color: #9ab5e4;
    letter-spacing: 0.3px;
    text-transform: uppercase;
  }

  .table-wrapper.frosted {
    background: linear-gradient(135deg, rgba(20, 32, 46, 0.7), rgba(20, 32, 46, 0.9));
    border-radius: 12px;
    border: 1px solid rgba(255, 255, 255, 0.08);
    padding: 8px;
  }

  .entities-table tbody tr:nth-child(odd):not(.empty-row) {
    background: rgba(255, 255, 255, 0.025);
  }

  .entities-table tbody tr:nth-child(even):not(.empty-row) {
    background: rgba(109, 216, 255, 0.045);
  }

  .draggable-body tr {
    cursor: grab;
    transition: background 120ms ease, transform 120ms ease, box-shadow 120ms ease;
  }

  .draggable-body tr:hover:not(.empty-row) {
    background: rgba(135, 202, 255, 0.1);
  }

  .empty-row {
    cursor: default;
    text-align: center;
    color: #cfd9e9;
  }

  .draggable-body tr.dragging {
    opacity: 0.75;
    background: rgba(255, 255, 255, 0.16);
  }

  .draggable-body tr.drag-hover {
    background: rgba(90, 209, 255, 0.12);
    box-shadow: inset 0 0 0 1px rgba(90, 209, 255, 0.4);
    transform: translateY(-1px);
  }

  .entity-nav {
    display: inline-flex;
    gap: 6px;
    align-items: center;
  }

  .nav-btn {
    border: 1px solid rgba(255, 255, 255, 0.14);
    background: rgba(255, 255, 255, 0.08);
    color: #d9e4f5;
    border-radius: 10px;
    height: 40px;
    width: 42px;
    cursor: pointer;
    transition: background 150ms ease, transform 120ms ease;
  }

  .nav-btn:hover:enabled {
    background: rgba(90, 209, 255, 0.16);
  }

  .nav-btn:active:enabled {
    transform: translateY(1px);
  }

  .nav-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .entity-select {
    border-radius: 10px;
    background: rgba(21, 32, 46, 0.82);
    border: 1px solid rgba(255, 255, 255, 0.14);
    color: #e8edf7;
    padding: 10px 12px;
    min-width: 220px;
    box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.03);
    appearance: none;
    transition: border 140ms ease, box-shadow 140ms ease;
  }

  .entity-select:focus {
    border-color: rgba(90, 209, 255, 0.8);
    box-shadow: 0 0 0 2px rgba(90, 209, 255, 0.22);
  }

  .empty-panel {
    padding: 24px 14px;
    border-radius: 12px;
    border: 1px dashed rgba(255, 255, 255, 0.12);
    color: #cfd9e9;
    text-align: center;
  }

  .row-actions {
    display: inline-flex;
    gap: 8px;
    align-items: center;
  }

  .approve-btn {
    border: 1px solid rgba(121, 205, 126, 0.32);
    background: rgba(48, 83, 52, 0.32);
    color: #dff7df;
    border-radius: 10px;
    padding: 10px 14px;
    cursor: pointer;
    transition: background 140ms ease, transform 120ms ease;
  }

  .approve-btn:hover:enabled {
    background: rgba(76, 175, 80, 0.28);
  }

  .approve-btn--approved {
    background: rgba(76, 175, 80, 0.18);
  }

  .approve-btn:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }

  @media (max-width: 720px) {
    .entity-status-card {
      flex-direction: column;
      align-items: stretch;
    }
  }
</style>
