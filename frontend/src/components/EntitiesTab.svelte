<script lang="ts">
  import {onMount, tick} from "svelte";
  import type {utils} from "../../wailsjs/go/models";
  import CreateEntity from "./forms/CreateEntity.svelte";
  import DeleteEntity from "./forms/DeleteEntity.svelte";
  import {MarkEntityStatus, MoveEntity, Save} from "../../wailsjs/go/main/App";
  import {showToast} from "../lib/toast";

  export let onSave: () => Promise<void> = async () => {};
  export let entities: utils.Entity[] = [];
  export let onJumpTo: (tab: "relations" | "tertiary", entityId: number) => void = () => {};

  let createEntityModal: CreateEntity | null = null;
  let tableWrapper: HTMLDivElement | null = null;
  let draggingIndex: number | null = null;
  let hoverIndex: number | null = null;
  let autoScrollFrame: number | null = null;
  let autoScrollDirection: -1 | 0 | 1 = 0;
  let searchQuery = "";
  let normalizedSearchQuery = "";
  let searchMatchIds: number[] = [];
  let activeSearchMatchId: number | null = null;
  let activeSearchMatchIndex = -1;
  let lastScrolledMatchId: number | null = null;
  let contextMenu: {
    open: boolean;
    x: number;
    y: number;
    entityId: number | null;
    entityName: string;
  } = {
    open: false,
    x: 0,
    y: 0,
    entityId: null,
    entityName: ""
  };

  const AUTO_SCROLL_EDGE_PX = 72;
  const AUTO_SCROLL_STEP = 14;
  const isApproved = (entity: utils.Entity) => entity.Status === true;
  const normalizeSearchText = (value: string) =>
    value
      .normalize("NFD")
      .replace(/[\u0300-\u036f]/g, "")
      .toLowerCase()
      .trim();

  const entityMatchesSearch = (entity: utils.Entity) => {
    if (!normalizedSearchQuery) {
      return false;
    }
    const haystack = normalizeSearchText(`${entity.Name} ${entity.Description || ""}`);
    return haystack.includes(normalizedSearchQuery);
  };

  const isSearchMatch = (entityId: number) => searchMatchIds.includes(entityId);

  const focusSearchMatch = async (entityId: number) => {
    await tick();
    const row = tableWrapper?.querySelector<HTMLTableRowElement>(`tr[data-entity-row="${entityId}"]`);
    row?.scrollIntoView({
      behavior: "smooth",
      block: "center"
    });
    lastScrolledMatchId = entityId;
  };

  const cycleSearchMatch = (direction: -1 | 1) => {
    if (searchMatchIds.length === 0) {
      return;
    }
    if (searchMatchIds.length === 1 || activeSearchMatchId === null) {
      activeSearchMatchId = searchMatchIds[0];
      return;
    }

    const currentIndex = searchMatchIds.indexOf(activeSearchMatchId);
    const nextIndex = currentIndex === -1
      ? 0
      : (currentIndex + direction + searchMatchIds.length) % searchMatchIds.length;
    activeSearchMatchId = searchMatchIds[nextIndex];
  };

  const handleSearchKeydown = (event: KeyboardEvent) => {
    if (event.key !== "Enter" || searchMatchIds.length === 0) {
      return;
    }
    event.preventDefault();
    cycleSearchMatch(event.shiftKey ? -1 : 1);
  };

  const clearSearch = () => {
    searchQuery = "";
  };

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

  const startDrag = (index: number, event: DragEvent) => {
    closeContextMenu();
    draggingIndex = index;
    hoverIndex = index;
    event.dataTransfer?.setData("text/plain", `${index}`);
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

  const clearDrag = () => {
    stopAutoScroll();
    draggingIndex = null;
    hoverIndex = null;
  };

  const applyReorder = async (from: number, to: number) => {
    if (from === to || from < 0 || to < 0 || from >= entities.length || to >= entities.length) {
      return;
    }
    const direction: "up" | "down" = to < from ? "up" : "down";
    const steps = Math.abs(to - from);
    const id = entities[from].Id;
    try {
      for (let i = 0; i < steps; i++) {
        await MoveEntity(id, direction);
      }
      if (steps > 0) {
        await Save();
        await onSave();
      }
    } catch (err) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      showToast(`No se pudo reordenar la entidad: ${message}`, "error");
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

  const closeContextMenu = () => {
    contextMenu = {
      open: false,
      x: 0,
      y: 0,
      entityId: null,
      entityName: ""
    };
  };

  const openContextMenu = (entity: utils.Entity, event: MouseEvent) => {
    event.preventDefault();
    const menuWidth = 220;
    const menuHeight = 196;
    contextMenu = {
      open: true,
      x: Math.max(12, Math.min(event.clientX, window.innerWidth - menuWidth - 12)),
      y: Math.max(12, Math.min(event.clientY, window.innerHeight - menuHeight - 12)),
      entityId: entity.Id,
      entityName: entity.Name
    };
  };

  const handleInsertFromContext = async (placement: "above" | "below") => {
    if (contextMenu.entityId === null) {
      return;
    }

    const referenceId = contextMenu.entityId;
    closeContextMenu();
    await createEntityModal?.openForInsert(referenceId, placement);
  };

  const handleJumpFromContext = (tab: "relations" | "tertiary") => {
    if (contextMenu.entityId === null) {
      return;
    }

    const referenceId = contextMenu.entityId;
    closeContextMenu();
    onJumpTo(tab, referenceId);
  };

  const handleWindowKeydown = (event: KeyboardEvent) => {
    if (event.key === "Escape") {
      closeContextMenu();
    }
  };

  const toggleEntityApproval = async (entity: utils.Entity) => {
    try {
      await MarkEntityStatus(entity.Id, !isApproved(entity));
      await onSave();
    } catch (err) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      showToast(`No se pudo actualizar la aprobación: ${message}`, "error");
    }
  };

  onMount(() => {
    const handleWindowBlur = () => closeContextMenu();
    window.addEventListener("blur", handleWindowBlur);
    return () => {
      window.removeEventListener("blur", handleWindowBlur);
    };
  });

  $: normalizedSearchQuery = normalizeSearchText(searchQuery);

  $: {
    const nextMatchIds = normalizedSearchQuery
      ? entities.filter(entityMatchesSearch).map((entity) => entity.Id)
      : [];
    searchMatchIds = nextMatchIds;

    if (nextMatchIds.length === 0) {
      activeSearchMatchId = null;
    } else if (activeSearchMatchId === null || !nextMatchIds.includes(activeSearchMatchId)) {
      activeSearchMatchId = nextMatchIds[0];
    }
  }

  $: activeSearchMatchIndex = activeSearchMatchId === null
    ? -1
    : searchMatchIds.indexOf(activeSearchMatchId);

  $: if (activeSearchMatchId === null) {
    lastScrolledMatchId = null;
  }

  $: if (activeSearchMatchId !== null && activeSearchMatchId !== lastScrolledMatchId) {
    focusSearchMatch(activeSearchMatchId);
  }
</script>

<svelte:window on:click={closeContextMenu} on:keydown={handleWindowKeydown} on:scroll={closeContextMenu}/>

<div class="tab-toolbar">
  <div>
    <p class="label">Entidades</p>
    <p class="muted">Vista general del proyecto. Agrega nuevas entidades desde aquí.</p>
  </div>
  <CreateEntity onSave={onSave}/>
</div>

<div class="search-toolbar">
  <div class="search-field">
    <input
      class="search-input"
      type="search"
      bind:value={searchQuery}
      placeholder="Buscar entidad por nombre o descripción"
      aria-label="Buscar entidad"
      on:keydown={handleSearchKeydown}
    />
    {#if searchQuery}
      <button class="search-clear" on:click={clearSearch} aria-label="Limpiar búsqueda">
        Limpiar
      </button>
    {/if}
  </div>

  <div class="search-meta">
    {#if normalizedSearchQuery}
      {#if searchMatchIds.length}
        <span class="search-count">
          {activeSearchMatchIndex + 1} de {searchMatchIds.length}
        </span>
        <button class="search-nav" on:click={() => cycleSearchMatch(-1)} aria-label="Coincidencia anterior">
          &lt;
        </button>
        <button class="search-nav" on:click={() => cycleSearchMatch(1)} aria-label="Siguiente coincidencia">
          &gt;
        </button>
      {:else}
        <span class="search-empty">Sin coincidencias</span>
      {/if}
    {/if}
  </div>
</div>

<CreateEntity bind:this={createEntityModal} onSave={onSave} showTrigger={false}/>

<div
  class="table-wrapper"
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
      <th style="width: 240px;">Acciones</th>
    </tr>
    </thead>

    <tbody class="draggable-body">
    {#each entities as entity, index (entity.Id)}
      <tr
        class:approved={isApproved(entity)}
        class:search-match={isSearchMatch(entity.Id)}
        class:search-match-active={activeSearchMatchId === entity.Id}
        class:dragging={draggingIndex === index}
        class:drag-hover={hoverIndex === index && draggingIndex !== null && draggingIndex !== index}
        data-entity-row={entity.Id}
        draggable="true"
        on:dragstart={(event) => startDrag(index, event)}
        on:dragover={(event) => handleDragOver(index, event)}
        on:dragenter={(event) => handleDragOver(index, event)}
        on:drop={(event) => handleDrop(index, event)}
        on:dragend={clearDrag}
        on:contextmenu={(event) => openContextMenu(entity, event)}
      >
        <td>
          <div class="entity-cell">
            <span>{entity.Name}</span>
            {#if isApproved(entity)}
              <span class="status-pill status-pill--approved">&#10003;</span>
            {/if}
          </div>
        </td>
        <td>{entity.Description}</td>
        <td>
          <div class="row-actions">
            <button
              class:approve-btn={true}
              class:approve-btn--approved={isApproved(entity)}
              on:click={() => toggleEntityApproval(entity)}
            >
              {isApproved(entity) ? "Quitar aprobación" : "Aprobar"}
            </button>
            <CreateEntity onSave={onSave} id={entity.Id}/>
            <DeleteEntity onSave={onSave} id={entity.Id}/>
          </div>
        </td>
      </tr>
    {/each}
    </tbody>
  </table>
</div>

{#if contextMenu.open}
  <div
    class="context-menu"
    style={`left: ${contextMenu.x}px; top: ${contextMenu.y}px;`}
    on:click|stopPropagation
    on:keydown|stopPropagation
  >
    <p class="context-title">{contextMenu.entityName}</p>
    <button class="context-item context-item--accent" on:click={() => handleJumpFromContext("tertiary")}>
      Ir a atributos
    </button>
    <button class="context-item context-item--accent" on:click={() => handleJumpFromContext("relations")}>
      Ir a relaciones
    </button>
    <div class="context-divider"></div>
    <button class="context-item" on:click={() => handleInsertFromContext("above")}>
      Insertar arriba
    </button>
    <button class="context-item" on:click={() => handleInsertFromContext("below")}>
      Insertar abajo
    </button>
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
    background: linear-gradient(135deg, rgba(18, 29, 44, 0.72), rgba(15, 23, 38, 0.92));
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 14px;
    padding: 8px;
  }

  .search-toolbar {
    display: flex;
    gap: 12px;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 14px;
    flex-wrap: wrap;
  }

  .search-field {
    display: flex;
    gap: 8px;
    align-items: center;
    flex: 1 1 320px;
  }

  .search-input {
    width: 100%;
    min-height: 48px;
    border-radius: 12px;
    border: 1px solid rgba(90, 209, 255, 0.22);
    background: rgba(9, 15, 26, 0.7);
    color: #e8edf7;
    padding: 0 14px;
    outline: none;
    transition: border-color 120ms ease, box-shadow 120ms ease, background 120ms ease;
  }

  .search-input::placeholder {
    color: rgba(232, 237, 247, 0.5);
  }

  .search-input:focus {
    border-color: rgba(90, 209, 255, 0.45);
    box-shadow: 0 0 0 3px rgba(90, 209, 255, 0.16);
    background: rgba(12, 20, 34, 0.9);
  }

  .search-clear,
  .search-nav {
    border: 1px solid rgba(90, 209, 255, 0.22);
    background: rgba(90, 209, 255, 0.1);
    color: #dff5ff;
    border-radius: 10px;
    min-height: 42px;
    padding: 0 12px;
    cursor: pointer;
    transition: background 120ms ease, transform 120ms ease, opacity 120ms ease;
  }

  .search-clear:hover,
  .search-nav:hover {
    background: rgba(90, 209, 255, 0.18);
  }

  .search-clear:active,
  .search-nav:active {
    transform: translateY(1px);
  }

  .search-meta {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .search-count,
  .search-empty {
    color: #cfe2ff;
    font-size: 13px;
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

  .entities-table tbody tr:nth-child(odd) {
    background: rgba(255, 255, 255, 0.025);
  }

  .entities-table tbody tr:nth-child(even) {
    background: rgba(109, 216, 255, 0.045);
  }

  .entities-table tbody tr:hover {
    background: rgba(135, 202, 255, 0.1);
  }

  .entities-table tbody tr.approved:nth-child(odd),
  .entities-table tbody tr.approved:nth-child(even) {
    background: rgba(76, 175, 80, 0.14);
  }

  .entities-table tbody tr.approved:hover {
    background: rgba(98, 201, 110, 0.2);
  }

  .entities-table tbody tr.search-match:nth-child(odd),
  .entities-table tbody tr.search-match:nth-child(even) {
    background: rgba(255, 196, 77, 0.12);
    box-shadow: inset 0 0 0 1px rgba(255, 196, 77, 0.22);
  }

  .entities-table tbody tr.search-match:hover {
    background: rgba(255, 196, 77, 0.18);
  }

  .entities-table tbody tr.search-match-active:nth-child(odd),
  .entities-table tbody tr.search-match-active:nth-child(even) {
    background: rgba(255, 196, 77, 0.22);
    box-shadow: inset 0 0 0 2px rgba(255, 211, 102, 0.62);
  }

  .draggable-body tr {
    cursor: grab;
    transition: background 120ms ease, transform 120ms ease, box-shadow 120ms ease;
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

  .row-actions {
    display: inline-flex;
    gap: 8px;
    align-items: center;
    flex-wrap: wrap;
  }

  .entity-cell {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .status-pill {
    display: inline-flex;
    align-items: center;
    padding: 4px 10px;
    border-radius: 999px;
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0.2px;
  }

  .status-pill--approved {
    color: #dff7df;
    background: rgba(76, 175, 80, 0.18);
    border: 1px solid rgba(113, 201, 118, 0.35);
    min-width: 30px;
    justify-content: center;
  }

  .approve-btn {
    border: 1px solid rgba(121, 205, 126, 0.32);
    background: rgba(48, 83, 52, 0.32);
    color: #dff7df;
    border-radius: 10px;
    padding: 9px 12px;
    cursor: pointer;
    transition: background 140ms ease, transform 120ms ease;
  }

  .approve-btn:hover {
    background: rgba(76, 175, 80, 0.28);
  }

  .approve-btn--approved {
    background: rgba(76, 175, 80, 0.18);
  }

  .context-menu {
    position: fixed;
    z-index: 60;
    min-width: 220px;
    padding: 10px;
    border-radius: 14px;
    border: 1px solid rgba(255, 255, 255, 0.1);
    background: linear-gradient(135deg, rgba(15, 23, 38, 0.98), rgba(22, 34, 52, 0.98));
    box-shadow: 0 16px 40px rgba(0, 0, 0, 0.35);
    display: grid;
    gap: 8px;
  }

  .context-title {
    margin: 0;
    padding: 4px 6px 8px;
    color: #9ab5e4;
    font-size: 12px;
    letter-spacing: 0.4px;
    text-transform: uppercase;
  }

  .context-item {
    width: 100%;
    border: 0;
    border-radius: 10px;
    padding: 11px 12px;
    text-align: left;
    background: rgba(255, 255, 255, 0.05);
    color: #e8edf7;
    cursor: pointer;
    transition: background 140ms ease, transform 120ms ease;
  }

  .context-item:hover {
    background: rgba(90, 209, 255, 0.16);
  }

  .context-item--accent {
    background: rgba(90, 209, 255, 0.12);
    border: 1px solid rgba(90, 209, 255, 0.16);
  }

  .context-divider {
    height: 1px;
    background: rgba(255, 255, 255, 0.08);
    margin: 2px 0;
  }

  .context-item:active {
    transform: translateY(1px);
  }

  @media (max-width: 720px) {
    .tab-toolbar {
      flex-direction: column;
      align-items: stretch;
    }

    .search-toolbar {
      align-items: stretch;
    }

    .search-field {
      flex-wrap: wrap;
    }
  }
</style>
