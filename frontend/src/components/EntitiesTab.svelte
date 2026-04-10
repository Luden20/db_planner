<script lang="ts">
  import {onMount, tick} from "svelte";
  import type {utils} from "../../wailsjs/go/models";
  import ButtonIcon from "./ButtonIcon.svelte";
  import CreateEntity from "./forms/CreateEntity.svelte";
  import DeleteEntity from "./forms/DeleteEntity.svelte";
  import EditIntersectionEntity from "./forms/EditIntersectionEntity.svelte";
  import {MarkEntityStatus, MoveEntity, Save} from "../../wailsjs/go/main/App";
  import {showToast} from "../lib/toast";

  export let onSave: () => Promise<void> = async () => {};
  export let project: utils.DbProject;
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
  let activeScope: "strong" | "intersection" = "strong";
  let entities: utils.Entity[] = [];
  let intersectionEntities: utils.IntersectionEntity[] = [];
  let filteredIntersectionEntities: utils.IntersectionEntity[] = [];
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
  const approvedCount = () => entities.filter((entity) => isApproved(entity)).length;
  const entityTypeLabel = (entity: utils.Entity) => entity.TableType === "intersection" ? "Interseccion" : "Fuerte";
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

  const intersectionMatchesSearch = (item: utils.IntersectionEntity) => {
    if (!normalizedSearchQuery) {
      return true;
    }
    const haystack = normalizeSearchText(`${item.Entity.Name} ${item.Entity.Description || ""} ${intersectionSourceLabel(item)}`);
    return haystack.includes(normalizedSearchQuery);
  };

  const intersectionSourceLabel = (item: utils.IntersectionEntity) => {
    const relation = project?.Relations?.find((current) => current.Id === item.RelationID);
    if (!relation) {
      return "Sin relación asociada";
    }
    const left = entities.find((entity) => entity.Id === relation.IdEntity1)?.Name ?? `Tabla ${relation.IdEntity1}`;
    const right = entities.find((entity) => entity.Id === relation.IdEntity2)?.Name ?? `Tabla ${relation.IdEntity2}`;
    return `${left} <-> ${right}`;
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
    if (activeScope !== "strong" || event.key !== "Enter" || searchMatchIds.length === 0) {
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

  const switchScope = (scope: "strong" | "intersection") => {
    activeScope = scope;
    clearDrag();
    closeContextMenu();
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

  let stickySentinel: HTMLDivElement | null = null;
  let stickyStack: HTMLDivElement | null = null;
  let stickyStackHeight = 0;
  let stickyStackPinned = false;

  const syncStickyState = () => {
    if (!stickySentinel) {
      stickyStackPinned = false;
      return;
    }
    stickyStackPinned = stickySentinel.getBoundingClientRect().top <= 0;
  };

  const syncStickyStackHeight = () => {
    stickyStackHeight = stickyStack?.offsetHeight ?? 0;
  };

  onMount(() => {
    const handleWindowBlur = () => closeContextMenu();
    window.addEventListener("blur", handleWindowBlur);

    syncStickyStackHeight();
    syncStickyState();

    let observer: ResizeObserver | null = null;
    if (typeof ResizeObserver !== "undefined" && stickyStack) {
      observer = new ResizeObserver(() => {
        syncStickyStackHeight();
      });
      observer.observe(stickyStack);
    }

    return () => {
      window.removeEventListener("blur", handleWindowBlur);
      observer?.disconnect();
    };
  });

  $: entities = project?.Entities ?? [];
  $: intersectionEntities = project?.IntersectionEntities ?? [];
  $: normalizedSearchQuery = normalizeSearchText(searchQuery);
  $: filteredIntersectionEntities = normalizedSearchQuery
    ? intersectionEntities.filter(intersectionMatchesSearch)
    : intersectionEntities;

  $: {
    const nextMatchIds = activeScope === "strong" && normalizedSearchQuery
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
  $: if (stickySentinel) {
    syncStickyState();
  }
  $: if (stickyStack) {
    syncStickyStackHeight();
  }
</script>

<svelte:window on:click={closeContextMenu} on:keydown={handleWindowKeydown} on:scroll={() => { closeContextMenu(); syncStickyState(); }} on:resize={syncStickyState}/>

<section class="entities-studio" style={`--entities-sticky-total-height: ${stickyStackHeight}px;`}>
  <div class="entities-sticky-sentinel" bind:this={stickySentinel} aria-hidden="true"></div>
  <div
    class:entities-sticky-stack={true}
    class:entities-sticky-stack--pinned={stickyStackPinned}
    bind:this={stickyStack}
  >
    <div class="tab-toolbar tab-toolbar--studio">
      <div>
        <p class="label">Entidades</p>
        <p class="muted">Gestiona entidades fuertes e intersecciones N:N desde el mismo estudio.</p>
      </div>
      <div class="entities-toolbar__side">
        <div class="scope-switch" role="tablist" aria-label="Tipo de entidad">
          <button
            class={`scope-switch__item ${activeScope === 'strong' ? 'scope-switch__item--active' : ''}`}
            type="button"
            role="tab"
            aria-selected={activeScope === "strong"}
            on:click={() => switchScope("strong")}
          >
            Fuertes
          </button>
          <button
            class={`scope-switch__item ${activeScope === 'intersection' ? 'scope-switch__item--active' : ''}`}
            type="button"
            role="tab"
            aria-selected={activeScope === "intersection"}
            on:click={() => switchScope("intersection")}
          >
            Intersección
          </button>
        </div>
        <div class="entities-toolbar__meta">
          <span class="studio-chip">{activeScope === "strong" ? entities.length : intersectionEntities.length} {activeScope === "strong" ? "fuertes" : "intersecciones"}</span>
          {#if activeScope === "strong"}
            <span class="studio-chip studio-chip--quiet">{approvedCount()} aprobadas</span>
          {/if}
          {#if normalizedSearchQuery && activeScope === "strong"}
            <span class="studio-chip studio-chip--quiet">{searchMatchIds.length} coincidencias</span>
          {:else if normalizedSearchQuery && activeScope === "intersection"}
            <span class="studio-chip studio-chip--quiet">{filteredIntersectionEntities.length} coincidencias</span>
          {/if}
        </div>
        {#if activeScope === "strong"}
          <CreateEntity onSave={onSave}/>
        {/if}
      </div>
    </div>
  </div>

  <CreateEntity bind:this={createEntityModal} onSave={onSave} showTrigger={false}/>

  <div class="entities-layout">
    <aside class="entities-deck">
      <div class="search-toolbar search-toolbar--studio">
        <div class="search-toolbar__copy">
          <p class="label">Filtro activo</p>
          <p class="muted">Búsqueda rápida de entidades.</p>
        </div>
        <div class="search-toolbar__controls">
          <div class="search-field">
            <input
              class="search-input"
              type="search"
              bind:value={searchQuery}
              placeholder={activeScope === "strong" ? "Buscar entidad por nombre o descripción" : "Buscar intersección por nombre, descripción o relación"}
              aria-label={activeScope === "strong" ? "Buscar entidad" : "Buscar entidad de intersección"}
              on:keydown={handleSearchKeydown}
            />
            {#if searchQuery}
              <button class="control control--sm control--ghost" on:click={clearSearch} aria-label="Limpiar búsqueda">
                <ButtonIcon name="clear"/>
                <span>Limpiar</span>
              </button>
            {/if}
          </div>

          <div class="search-meta">
            {#if normalizedSearchQuery && activeScope === "strong"}
              {#if searchMatchIds.length}
                <span class="search-count">
                  {activeSearchMatchIndex + 1} de {searchMatchIds.length}
                </span>
                <button class="control control--sm control--icon control--soft" on:click={() => cycleSearchMatch(-1)} aria-label="Coincidencia anterior">
                  <ButtonIcon name="chevron-left"/>
                </button>
                <button class="control control--sm control--icon control--soft" on:click={() => cycleSearchMatch(1)} aria-label="Siguiente coincidencia">
                  <ButtonIcon name="chevron-right"/>
                </button>
              {:else}
                <span class="search-empty">Sin coincidencias</span>
              {/if}
            {:else if normalizedSearchQuery}
              <span class="search-count">{filteredIntersectionEntities.length} coincidencias</span>
            {/if}
          </div>
        </div>
      </div>

      <section class="entities-side-card">
        <p class="label">Atajos</p>
        <div class="entities-side-card__stack">
          <div class="entities-side-stat">
            <span class="entities-side-stat__value">{activeScope === "strong" ? entities.length : intersectionEntities.length}</span>
            <span class="entities-side-stat__label">{activeScope === "strong" ? "entidades fuertes en el inventario" : "entidades de intersección detectadas"}</span>
          </div>
          {#if activeScope === "strong"}
            <div class="entities-side-stat">
              <span class="entities-side-stat__value">{approvedCount()}</span>
              <span class="entities-side-stat__label">listas para seguir en el flujo</span>
            </div>
            <p class="muted entities-side-note">Click derecho en una fila para saltar a atributos, relaciones o insertar arriba y abajo.</p>
          {:else}
            <p class="muted entities-side-note">Las entidades de intersección se generan automáticamente cuando una relación N:N existe o aparece al leer el JSON.</p>
          {/if}
        </div>
      </section>
    </aside>

    <section class="entities-panel">
      <div class="entities-panel__head">
        <div>
          <p class="label">Inventario</p>
          <p class="muted">{activeScope === "strong" ? "Gestión de entidades fuertes." : "Gestión de intersecciones."}</p>
        </div>
      </div>

      <div
        class="table-wrapper"
        bind:this={tableWrapper}
        on:dragover={activeScope === "strong" ? handleTableDragOver : undefined}
        on:dragleave={activeScope === "strong" ? handleTableDragLeave : undefined}
        on:drop={activeScope === "strong" ? stopAutoScroll : undefined}
      >
        <table class="entities-table">
          <thead>
          <tr>
            <th>Nombre</th>
            <th>Descripción</th>
            {#if activeScope === "intersection"}
              <th>Relación origen</th>
              <th style="width: 160px;">Acciones</th>
            {:else}
              <th style="width: 240px;">Acciones</th>
            {/if}
          </tr>
          </thead>

          <tbody class="draggable-body">
          {#if activeScope === "strong"}
            {#each entities as entity, index (entity.Id)}
              <tr
                class:approved={isApproved(entity)}
                class:search-match={isSearchMatch(entity.Id)}
                class:search-match-active={activeSearchMatchId === entity.Id}
                class:dragging={draggingIndex === index}
                class:drag-hover={hoverIndex === index && draggingIndex !== null && draggingIndex !== index}
                data-entity-row={entity.Id}
                draggable="true"
                style={`view-transition-name: entity-row-${entity.Id};`}
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
                    <span class="status-pill status-pill--type">{entityTypeLabel(entity)}</span>
                    {#if isApproved(entity)}
                      <span class="status-pill status-pill--approved">&#10003;</span>
                    {/if}
                  </div>
                </td>
                <td>{entity.Description}</td>
                <td>
                  <div class="row-actions">
                    <button
                      class={`control control--sm control--success ${isApproved(entity) ? 'control--active' : ''}`}
                      on:click={() => toggleEntityApproval(entity)}
                    >
                      <ButtonIcon name={isApproved(entity) ? "check-off" : "check"}/>
                      <span>{isApproved(entity) ? "Quitar aprobación" : "Aprobar"}</span>
                    </button>
                    <CreateEntity onSave={onSave} id={entity.Id}/>
                    <DeleteEntity onSave={onSave} id={entity.Id}/>
                  </div>
                </td>
              </tr>
            {/each}
          {:else if filteredIntersectionEntities.length === 0}
            <tr class="empty-row" draggable="false">
              <td colspan="4">{normalizedSearchQuery ? "No hay intersecciones que coincidan." : "No hay entidades de intersección todavía."}</td>
            </tr>
          {:else}
            {#each filteredIntersectionEntities as item (item.RelationID)}
              <tr data-entity-row={`intersection-${item.RelationID}`}>
                <td>
                  <div class="entity-cell">
                    <span>{item.Entity.Name}</span>
                    <span class="status-pill status-pill--type">Intersección</span>
                  </div>
                </td>
                <td>{item.Entity.Description}</td>
                <td>{intersectionSourceLabel(item)}</td>
                <td>
                  <div class="row-actions">
                    <EditIntersectionEntity item={item} onSave={onSave}/>
                  </div>
                </td>
              </tr>
            {/each}
          {/if}
          </tbody>
        </table>
      </div>
    </section>
  </div>
</section>

{#if contextMenu.open}
  <div
    class="context-menu"
    style={`left: ${contextMenu.x}px; top: ${contextMenu.y}px;`}
    on:click|stopPropagation
    on:keydown|stopPropagation
  >
    <p class="context-title">{contextMenu.entityName}</p>
    <button class="menu-action control control--sm control--block control--accent" on:click={() => handleJumpFromContext("tertiary")}>
      <ButtonIcon name="attributes"/>
      <span>Ir a atributos</span>
    </button>
    <button class="menu-action control control--sm control--block control--accent" on:click={() => handleJumpFromContext("relations")}>
      <ButtonIcon name="relations"/>
      <span>Ir a relaciones</span>
    </button>
    <div class="context-divider"></div>
    <button class="menu-action control control--sm control--block control--ghost" on:click={() => handleInsertFromContext("above")}>
      <ButtonIcon name="arrow-up"/>
      <span>Insertar arriba</span>
    </button>
    <button class="menu-action control control--sm control--block control--ghost" on:click={() => handleInsertFromContext("below")}>
      <ButtonIcon name="arrow-down"/>
      <span>Insertar abajo</span>
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

  .status-pill--type {
    color: var(--accent-strong);
    background: color-mix(in srgb, var(--accent) 12%, var(--surface-strong));
    border: 1px solid color-mix(in srgb, var(--accent) 22%, var(--border));
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

  .context-divider {
    height: 1px;
    background: rgba(255, 255, 255, 0.08);
    margin: 2px 0;
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

  .tab-toolbar {
    margin-bottom: 1rem;
    padding: 1.1rem 1.15rem;
    border: 1px solid var(--border);
    border-radius: calc(var(--radius-md) - 4px);
    background: var(--panel-surface);
  }

  .label {
    color: var(--accent);
    font-size: 0.74rem;
    letter-spacing: 0.16em;
    font-weight: 800;
  }

  .muted {
    color: var(--ink-faint);
    opacity: 1;
  }

  .search-toolbar {
    padding: 0.95rem 1rem;
    border-radius: calc(var(--radius-md) - 6px);
    border: 1px solid var(--border);
    background: var(--panel-surface-soft);
  }

  .search-input {
    border-color: var(--border);
    background: var(--field-surface);
    color: var(--ink);
  }

  .search-input::placeholder {
    color: var(--ink-faint);
  }

  .search-input:focus {
    border-color: var(--focus-border);
    box-shadow: var(--focus-ring);
    background: var(--field-surface-focus);
  }

  .search-count,
  .search-empty {
    color: var(--ink-faint);
  }

  .table-wrapper {
    background: var(--panel-surface-strong);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    padding: 0;
    box-shadow: var(--surface-highlight);
    max-height: calc(100vh - var(--entities-sticky-total-height, 0px) - 10rem);
    overflow-y: auto;
    scrollbar-gutter: stable;
  }

  .entities-table {
    color: var(--ink);
    border-collapse: collapse;
    width: 100%;
  }

  .entities-table thead th {
    position: sticky;
    top: 0;
    z-index: 10;
    background: var(--surface-strong);
    color: var(--ink-faint);
    border-bottom: 2px solid var(--border);
    font-size: 0.72rem;
    font-weight: 600;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    text-align: left;
    padding: 0.75rem 1rem;
    box-shadow: inset 0 -1px 0 var(--line-soft);
  }

  .entities-table thead th:first-child {
    border-top-left-radius: var(--radius-sm);
  }

  .entities-table thead th:last-child {
    border-top-right-radius: var(--radius-sm);
  }

  .entities-table tbody tr:last-child td:first-child {
    border-bottom-left-radius: var(--radius-sm);
  }

  .entities-table tbody tr:last-child td:last-child {
    border-bottom-right-radius: var(--radius-sm);
  }

  .entities-table tbody tr:nth-child(odd),
  .entities-table tbody tr:nth-child(even) {
    background: transparent;
  }

  .entities-table tbody tr:hover {
    background: var(--hover-soft);
  }

  .entities-table tbody tr.approved:nth-child(odd),
  .entities-table tbody tr.approved:nth-child(even) {
    background: var(--success-soft);
  }

  .entities-table tbody tr.approved:hover {
    background: var(--hover-success);
  }

  .entities-table tbody tr.search-match:nth-child(odd),
  .entities-table tbody tr.search-match:nth-child(even) {
    background: color-mix(in srgb, var(--accent) 10%, var(--surface));
  }

  .entities-table tbody tr.search-match-active:nth-child(odd),
  .entities-table tbody tr.search-match-active:nth-child(even) {
    background: color-mix(in srgb, var(--accent) 14%, var(--surface));
    box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--accent) 22%, transparent);
  }

  .status-pill {
    background: var(--chip-surface);
    border-color: var(--line-soft);
    color: var(--ink-faint);
  }

  .status-pill--approved {
    background: var(--chip-success-surface);
    border-color: color-mix(in srgb, var(--success) 24%, var(--border));
    color: var(--success);
  }

  .context-menu {
    border-color: var(--border);
    background: var(--popover-surface);
    box-shadow: var(--shadow-sm);
  }

  .context-title {
    color: var(--accent);
    font-size: 0.74rem;
    letter-spacing: 0.14em;
  }

  .context-divider {
    background: var(--line-soft);
  }

  @media (max-width: 720px) {
    .tab-toolbar,
    .search-toolbar {
      padding: 0.9rem;
    }
  }

  .entities-studio {
    display: grid;
    gap: 1rem;
  }

  .entities-layout {
    display: grid;
    grid-template-columns: minmax(17rem, 21rem) minmax(0, 1fr);
    align-items: start;
    gap: 1rem;
  }

  .entities-sticky-stack {
    margin-bottom: 0;
  }

  .entities-sticky-stack--pinned {
    position: sticky;
    top: 0;
    z-index: calc(var(--layer-ribbon, 100) - 2);
    background: var(--surface-strong);
  }

  .entities-sticky-sentinel {
    height: 1px;
    margin-top: -1px;
  }

  .entities-deck {
    display: grid;
    gap: 1rem;
    position: sticky;
    top: calc(var(--entities-sticky-total-height) + 1rem);
  }

  .tab-toolbar--studio,
  .search-toolbar--studio,
  .entities-panel,
  .entities-side-card {
    position: relative;
    overflow: clip;
  }

  .tab-toolbar--studio::before,
  .search-toolbar--studio::before,
  .entities-panel::before,
  .entities-side-card::before {
    content: "";
    position: absolute;
    inset: 0 auto auto 0;
    width: min(220px, 42%);
    height: 1px;
    background: linear-gradient(90deg, color-mix(in srgb, var(--accent) 34%, transparent), transparent);
    pointer-events: none;
  }

  .entities-toolbar__side {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 0.9rem;
    flex-wrap: wrap;
  }

  .scope-switch {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.25rem;
    border-radius: 999px;
    border: 1px solid var(--line-soft);
    background: color-mix(in srgb, var(--surface-strong) 88%, transparent);
  }

  .scope-switch__item {
    min-height: 2.2rem;
    padding: 0.45rem 0.85rem;
    border-radius: 999px;
    border: none;
    background: transparent;
    color: var(--ink-soft);
    font-size: 0.82rem;
    font-weight: 700;
    letter-spacing: 0.04em;
    transition: background 140ms ease, color 140ms ease, transform 140ms ease;
  }

  .scope-switch__item:hover {
    transform: translateY(-1px);
  }

  .scope-switch__item--active {
    background: color-mix(in srgb, var(--accent) 14%, var(--surface));
    color: var(--accent-strong);
  }

  .entities-toolbar__meta,
  .search-toolbar__controls {
    display: flex;
    align-items: center;
    gap: 0.7rem;
    flex-wrap: wrap;
  }

  .search-toolbar__copy {
    max-width: 34rem;
  }

  .search-toolbar__controls {
    justify-content: flex-end;
    flex: 1 1 24rem;
  }

  .studio-chip {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-height: 2rem;
    padding: 0.42rem 0.78rem;
    border-radius: 999px;
    border: 1px solid color-mix(in srgb, var(--accent) 16%, var(--border));
    background: color-mix(in srgb, var(--accent) 10%, var(--surface-strong));
    color: var(--accent-strong);
    font-size: 0.76rem;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    white-space: nowrap;
  }

  .studio-chip--quiet {
    border-color: var(--line-soft);
    background: color-mix(in srgb, var(--surface) 82%, transparent);
    color: var(--ink-soft);
  }

  .entities-panel {
    padding: 1rem;
    border: 1px solid var(--border);
    border-radius: calc(var(--radius-lg) - 2px);
    background:
      radial-gradient(circle at top right, color-mix(in srgb, var(--accent) 8%, transparent), transparent 34%),
      var(--panel-surface);
    box-shadow: var(--shadow-sm);
  }

  .entities-side-card {
    padding: 1rem;
    border: 1px solid var(--border);
    border-radius: calc(var(--radius-lg) - 2px);
    background:
      linear-gradient(180deg, color-mix(in srgb, var(--accent) 4%, transparent), transparent 42%),
      var(--panel-surface);
    box-shadow: var(--shadow-sm);
  }

  .entities-side-card__stack {
    display: grid;
    gap: 0.85rem;
    margin-top: 0.8rem;
  }

  .entities-side-stat {
    display: grid;
    gap: 0.2rem;
    padding: 0.85rem 0.9rem;
    border-radius: calc(var(--radius-md) - 4px);
    background: color-mix(in srgb, var(--surface-strong) 78%, transparent);
    border: 1px solid var(--line-soft);
  }

  .entities-side-stat__value {
    color: var(--accent-strong);
    font-size: clamp(1.4rem, 2vw, 1.8rem);
    font-weight: 800;
    line-height: 1;
  }

  .entities-side-stat__label,
  .entities-side-note {
    color: var(--ink-soft);
  }

  .entities-side-stat__label {
    font-size: 0.84rem;
    line-height: 1.35;
  }

  .entities-side-note {
    margin: 0;
    font-size: 0.86rem;
    line-height: 1.45;
  }

  .entities-panel__head {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 1rem;
    margin-bottom: 0.9rem;
  }

  .entities-panel__hint {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0.5rem 0.8rem;
    border-radius: 999px;
    background: color-mix(in srgb, var(--surface-strong) 82%, transparent);
    border: 1px solid var(--line-soft);
    color: var(--ink-soft);
    font-size: 0.78rem;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    white-space: nowrap;
  }

  .empty-row td {
    color: var(--ink-faint);
    text-align: center;
    padding-block: 1.2rem;
  }

  @media (max-width: 720px) {
    .entities-layout {
      grid-template-columns: 1fr;
    }

    .entities-deck {
      position: static;
    }

    .entities-toolbar__side,
    .search-toolbar__controls,
    .entities-panel__head {
      align-items: stretch;
      justify-content: flex-start;
      flex-direction: column;
    }

    .entities-panel {
      padding: 0.9rem;
    }

    .studio-chip,
    .entities-panel__hint {
      white-space: normal;
    }
  }
</style>
