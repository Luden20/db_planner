<script lang="ts">
  import {onMount, tick} from "svelte";
  import type {utils} from "../../wailsjs/go/models";
  import Button from "./ui/Button.svelte";
  import Badge from "./ui/Badge.svelte";
  import Toolbar from "./ui/Toolbar.svelte";
  import SearchInput from "./ui/SearchInput.svelte";
  import Table from "./ui/Table.svelte";
  import CreateEntity from "./forms/CreateEntity.svelte";
  import DeleteEntity from "./forms/DeleteEntity.svelte";
  import EditIntersectionEntity from "./forms/EditIntersectionEntity.svelte";
  import ScopeSwitch from "./studio/ScopeSwitch.svelte";
  import StickyStack from "./studio/StickyStack.svelte";
  import {MarkEntityStatus, MoveEntity, Save} from "../../wailsjs/go/main/App";
  import {showToast} from "../lib/toast";
  import {createVerticalAutoScroller, getErrorMessage} from "../lib/ui-helpers";

  export let onSave: () => Promise<void> = async () => {};
  export let project: utils.DbProject;
  export let onJumpTo: (tab: "entities" | "relations" | "tertiary", entityId: number) => void = () => {};

  let createEntityModal: CreateEntity | null = null;
  let tableWrapper: HTMLDivElement | null = null;
  let draggingIndex: number | null = null;
  let hoverIndex: number | null = null;
  let stickyStackHeight = 0;
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

  const isApproved = (entity: utils.Entity) => entity.Status === true;
  const approvedCount = () => entities.filter((entity) => isApproved(entity)).length;
  const entityTypeLabel = (entity: utils.Entity) => entity.TableType === "intersection" ? "Interseccion" : "Fuerte";
  const autoScroller = createVerticalAutoScroller({
    edgePx: 72,
    stepPx: 14,
    getContainer: () => tableWrapper
  });
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

  const updateAutoScroll = (event: DragEvent) => autoScroller.updateFromDragEvent(event);

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
    autoScroller.stop();
  };

  const clearDrag = () => {
    autoScroller.stop();
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
      const message = getErrorMessage(err);
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

  const handleJumpFromContext = (tab: "entities" | "relations" | "tertiary") => {
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
      const message = getErrorMessage(err);
      showToast(`No se pudo actualizar la aprobación: ${message}`, "error");
    }
  };

  onMount(() => {
    const handleWindowBlur = () => closeContextMenu();
    window.addEventListener("blur", handleWindowBlur);

    return () => {
      autoScroller.stop();
      window.removeEventListener("blur", handleWindowBlur);
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
</script>

<svelte:window on:click={closeContextMenu} on:keydown={handleWindowKeydown} on:scroll={closeContextMenu}/>

<section class="entities-studio" style={`--entities-sticky-total-height: ${stickyStackHeight}px;`}>
  <StickyStack bind:height={stickyStackHeight}>
    <Toolbar 
      title="Entidades" 
      description="Gestiona entidades fuertes e intersecciones N:N desde el mismo estudio."
      class="tab-toolbar--studio"
    >
      {#snippet meta()}
        <Badge>{activeScope === "strong" ? entities.length : intersectionEntities.length} {activeScope === "strong" ? "fuertes" : "intersecciones"}</Badge>
        {#if activeScope === "strong"}
          <Badge variant="quiet">{approvedCount()} aprobadas</Badge>
        {/if}
        {#if normalizedSearchQuery && activeScope === "strong"}
          <Badge variant="quiet">{searchMatchIds.length} coincidencias</Badge>
        {:else if normalizedSearchQuery && activeScope === "intersection"}
          <Badge variant="quiet">{filteredIntersectionEntities.length} coincidencias</Badge>
        {/if}
      {/snippet}

      {#snippet actions()}
        <ScopeSwitch {activeScope} onSwitch={switchScope}/>
        {#if activeScope === "strong"}
          <CreateEntity onSave={onSave}/>
        {/if}
      {/snippet}
    </Toolbar>
  </StickyStack>

  <CreateEntity bind:this={createEntityModal} onSave={onSave} showTrigger={false}/>

  <div class="entities-layout">
    <aside class="entities-deck">
      <div class="search-toolbar search-toolbar--studio">
        <div class="search-toolbar__copy">
          <p class="label">Filtro activo</p>
          <p class="muted">Búsqueda rápida de entidades.</p>
        </div>
        <div class="search-toolbar__controls">
          <SearchInput
            bind:value={searchQuery}
            placeholder={activeScope === "strong" ? "Buscar entidad por nombre o descripción" : "Buscar intersección por nombre, descripción o relación"}
            onkeydown={handleSearchKeydown}
          />

          <div class="search-meta">
            {#if normalizedSearchQuery && activeScope === "strong"}
              {#if searchMatchIds.length}
                <span class="search-count">
                  {activeSearchMatchIndex + 1} de {searchMatchIds.length}
                </span>
                <Button variant="soft" size="icon" icon="chevron-left" onclick={() => cycleSearchMatch(-1)} aria-label="Coincidencia anterior" />
                <Button variant="soft" size="icon" icon="chevron-right" onclick={() => cycleSearchMatch(1)} aria-label="Siguiente coincidencia" />
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

      <Table
        bind:ref={tableWrapper}
        ondragover={activeScope === "strong" ? handleTableDragOver : undefined}
        ondragleave={activeScope === "strong" ? handleTableDragLeave : undefined}
        ondrop={activeScope === "strong" ? autoScroller.stop : undefined}
        tableClass="draggable"
      >
        {#snippet header()}
          <th>Nombre</th>
          <th>Descripción</th>
          {#if activeScope === "intersection"}
            <th>Relación origen</th>
            <th style="width: 160px;">Acciones</th>
          {:else}
            <th style="width: 240px;">Acciones</th>
          {/if}
        {/snippet}

        {#snippet body()}
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
                    <Button
                      variant="success"
                      size="sm"
                      class={isApproved(entity) ? 'control--active' : ''}
                      icon={isApproved(entity) ? "check-off" : "check"}
                      onclick={() => toggleEntityApproval(entity)}
                    >
                      {isApproved(entity) ? "Quitar aprobación" : "Aprobar"}
                    </Button>
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
        {/snippet}
      </Table>
    </section>
  </div>
</section>

{#if contextMenu.open}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="context-menu"
    style={`left: ${contextMenu.x}px; top: ${contextMenu.y}px;`}
    role="menu"
    tabindex="-1"
    on:click|stopPropagation
    on:keydown|stopPropagation
  >
    <p class="context-title">{contextMenu.entityName}</p>
    <Button class="menu-action control--block" size="sm" variant="accent" icon="attributes" onclick={() => handleJumpFromContext("tertiary")}>
      Ir a atributos
    </Button>
    <Button class="menu-action control--block" size="sm" variant="accent" icon="relations" onclick={() => handleJumpFromContext("relations")}>
      Ir a relaciones
    </Button>
    <div class="context-divider"></div>
    <Button class="menu-action control--block" size="sm" variant="ghost" icon="arrow-up" onclick={() => handleInsertFromContext("above")}>
      Insertar arriba
    </Button>
    <Button class="menu-action control--block" size="sm" variant="ghost" icon="arrow-down" onclick={() => handleInsertFromContext("below")}>
      Insertar abajo
    </Button>
  </div>
{/if}
