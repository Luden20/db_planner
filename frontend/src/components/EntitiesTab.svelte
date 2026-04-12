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

  let { 
    onSave = async () => {}, 
    project, 
    onJumpTo = () => {} 
  } = $props<{
    onSave?: () => Promise<void>;
    project: utils.DbProject;
    onJumpTo?: (tab: "entities" | "relations" | "tertiary", entityId: number) => void;
  }>();

  let createEntityModal = $state<CreateEntity | null>(null);
  let tableWrapper = $state<HTMLDivElement | null>(null);
  let draggingIndex = $state<number | null>(null);
  let hoverIndex = $state<number | null>(null);
  let stickyStackHeight = $state(0);
  let searchQuery = $state("");
  let lastScrolledMatchId = $state<number | null>(null);
  let activeScope = $state<"strong" | "intersection">("strong");
  let activeSearchMatchId = $state<number | null>(null);

  let contextMenu = $state({
    open: false,
    x: 0,
    y: 0,
    entityId: null as number | null,
    entityName: ""
  });

  const normalizedSearchQuery = $derived(normalizeSearchText(searchQuery));
  const entities = $derived(project?.Entities ?? []);
  const intersectionEntities = $derived(project?.IntersectionEntities ?? []);

  const filteredIntersectionEntities = $derived(normalizedSearchQuery
    ? intersectionEntities.filter(intersectionMatchesSearch)
    : intersectionEntities);

  const searchMatchIds = $derived((activeScope === "strong" && normalizedSearchQuery)
    ? entities.filter(entityMatchesSearch).map((entity) => entity.Id)
    : []);

  const activeSearchMatchIndex = $derived(activeSearchMatchId === null
    ? -1
    : searchMatchIds.indexOf(activeSearchMatchId));

  const isApproved = (entity: utils.Entity) => entity.Status === true;
  const approvedCount = () => entities.filter((entity) => isApproved(entity)).length;
  const entityTypeLabel = (entity: utils.Entity) => entity.TableType === "intersection" ? "Interseccion" : "Fuerte";
  
  const autoScroller = createVerticalAutoScroller({
    edgePx: 72,
    stepPx: 14,
    getContainer: () => tableWrapper
  });

  function normalizeSearchText(value: string) {
    return value
      .normalize("NFD")
      .replace(/[\u0300-\u036f]/g, "")
      .toLowerCase()
      .trim();
  }

  function entityMatchesSearch(entity: utils.Entity) {
    if (!normalizedSearchQuery) return false;
    const haystack = normalizeSearchText(`${entity.Name} ${entity.Description || ""}`);
    return haystack.includes(normalizedSearchQuery);
  }

  function intersectionMatchesSearch(item: utils.IntersectionEntity) {
    if (!normalizedSearchQuery) return true;
    const haystack = normalizeSearchText(`${item.Entity.Name} ${item.Entity.Description || ""} ${intersectionSourceLabel(item)}`);
    return haystack.includes(normalizedSearchQuery);
  }

  function intersectionSourceLabel(item: utils.IntersectionEntity) {
    const relation = project?.Relations?.find((current) => current.Id === item.RelationID);
    if (!relation) return "Sin relación asociada";
    const left = entities.find((entity) => entity.Id === relation.IdEntity1)?.Name ?? `Tabla ${relation.IdEntity1}`;
    const right = entities.find((entity) => entity.Id === relation.IdEntity2)?.Name ?? `Tabla ${relation.IdEntity2}`;
    return `${left} <-> ${right}`;
  }

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
    if (searchMatchIds.length === 0) return;
    if (searchMatchIds.length === 1 || activeSearchMatchId === null) {
      activeSearchMatchId = searchMatchIds[0];
      return;
    }
    const currentIndex = searchMatchIds.indexOf(activeSearchMatchId);
    const nextIndex = (currentIndex + direction + searchMatchIds.length) % searchMatchIds.length;
    activeSearchMatchId = searchMatchIds[nextIndex];
  };

  const handleSearchKeydown = (event: KeyboardEvent) => {
    if (activeScope !== "strong" || event.key !== "Enter" || searchMatchIds.length === 0) return;
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
    if (tableWrapper && nextTarget && tableWrapper.contains(nextTarget)) return;
    autoScroller.stop();
  };

  const clearDrag = () => {
    autoScroller.stop();
    draggingIndex = null;
    hoverIndex = null;
  };

  const applyReorder = async (from: number, to: number) => {
    if (from === to || from < 0 || to < 0 || from >= entities.length || to >= entities.length) return;
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
    contextMenu.open = false;
    contextMenu.entityId = null;
    contextMenu.entityName = "";
  };

  const openContextMenu = (entity: utils.Entity, event: MouseEvent) => {
    event.preventDefault();
    const menuWidth = 220;
    const menuHeight = 196;
    contextMenu.open = true;
    contextMenu.x = Math.max(12, Math.min(event.clientX, window.innerWidth - menuWidth - 12));
    contextMenu.y = Math.max(12, Math.min(event.clientY, window.innerHeight - menuHeight - 12));
    contextMenu.entityId = entity.Id;
    contextMenu.entityName = entity.Name;
  };

  const handleInsertFromContext = async (placement: "above" | "below") => {
    if (contextMenu.entityId === null) return;
    const referenceId = contextMenu.entityId;
    closeContextMenu();
    await createEntityModal?.openForInsert(referenceId, placement);
  };

  const handleJumpFromContext = (tab: "entities" | "relations" | "tertiary") => {
    if (contextMenu.entityId === null) return;
    const referenceId = contextMenu.entityId;
    closeContextMenu();
    onJumpTo(tab, referenceId);
  };

  const handleWindowKeydown = (event: KeyboardEvent) => {
    if (event.key === "Escape") closeContextMenu();
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

  $effect(() => {
    if (searchMatchIds.length === 0) {
      activeSearchMatchId = null;
    } else if (activeSearchMatchId === null || !searchMatchIds.includes(activeSearchMatchId)) {
      activeSearchMatchId = searchMatchIds[0];
    }
  });

  $effect(() => {
    if (activeSearchMatchId !== null && activeSearchMatchId !== lastScrolledMatchId) {
      focusSearchMatch(activeSearchMatchId);
    }
  });
</script>

<svelte:window onclick={closeContextMenu} onkeydown={handleWindowKeydown} onscroll={closeContextMenu}/>

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
      <section class="entities-side-card">
        <p class="label">Atajos</p>
        <div class="entities-side-card__stack">
          <div class="entities-side-stat">
            <span class="entities-side-stat__value">{activeScope === "strong" ? entities.length : intersectionEntities.length}</span>
            <span class="entities-side-stat__label">{activeScope === "strong" ? "entidades fuertes" : "intersecciones N:N"}</span>
          </div>
          {#if activeScope === "strong"}
            <div class="entities-side-stat">
              <span class="entities-side-stat__value">{approvedCount()}</span>
              <span class="entities-side-stat__label">aprobadas</span>
            </div>
            <p class="muted entities-side-note">Click derecho en una fila para saltar a atributos, relaciones o insertar arriba y abajo.</p>
          {:else}
            <p class="muted entities-side-note">Las entidades de intersección se generan automáticamente al detectar relaciones N:N.</p>
          {/if}
        </div>
      </section>
    </aside>

    <section class="entities-panel">
      <div class="entities-panel__head">
        <div class="flex flex-col gap-1">
          <p class="label">Inventario</p>
          <p class="muted">{activeScope === "strong" ? "Gestión de entidades fuertes." : "Gestión de intersecciones."}</p>
        </div>

        <div class="search-controls-inline">
          <SearchInput
            bind:value={searchQuery}
            placeholder={activeScope === "strong" ? "Buscar entidad..." : "Buscar intersección..."}
            onkeydown={handleSearchKeydown}
            class="max-w-[300px]"
          />

          {#if normalizedSearchQuery && activeScope === "strong" && searchMatchIds.length}
            <div class="search-meta-inline">
              <span class="search-count">
                {activeSearchMatchIndex + 1}/{searchMatchIds.length}
              </span>
              <Button variant="ghost" size="icon" icon="chevron-left" onclick={() => cycleSearchMatch(-1)} />
              <Button variant="ghost" size="icon" icon="chevron-right" onclick={() => cycleSearchMatch(1)} />
            </div>
          {/if}
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
                ondragstart={(event) => startDrag(index, event)}
                ondragover={(event) => handleDragOver(index, event)}
                ondragenter={(event) => handleDragOver(index, event)}
                ondrop={(event) => handleDrop(index, event)}
                ondragend={clearDrag}
                oncontextmenu={(event) => openContextMenu(entity, event)}
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
                    <div class="inline-flex gap-2">
                       <CreateEntity onSave={onSave} id={entity.Id} triggerSize="sm"/>
                       <DeleteEntity onSave={onSave} id={entity.Id}/>
                    </div>
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
    onclick={(e) => e.stopPropagation()}
    onkeydown={(e) => e.stopPropagation()}
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

<style>
  .entities-studio {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .entities-layout {
    display: grid;
    grid-template-columns: 320px 1fr;
    gap: 1.5rem;
    align-items: start;
  }

  .entities-deck {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    position: sticky;
    top: calc(var(--entities-sticky-total-height) + 1.5rem);
  }

  .entities-side-card {
    padding: 1.5rem;
    background: var(--background);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-sm);
  }

  .entities-side-card__stack {
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
    margin-top: 1rem;
  }

  .entities-side-stat {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .entities-side-stat__value {
    font-size: 1.75rem;
    font-weight: 900;
    color: var(--accent);
    line-height: 1;
  }

  .entities-side-stat__label {
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--ink-soft);
    text-transform: uppercase;
    letter-spacing: 0.025em;
  }

  .entities-side-note {
    font-size: 0.7rem;
    line-height: 1.4;
  }

  .entities-panel {
    background: var(--background);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-sm);
    overflow: hidden;
  }

  .entities-panel__head {
    padding: 1rem 1.5rem;
    border-bottom: 1px solid var(--border-card);
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 2rem;
  }

  .search-controls-inline {
    display: flex;
    align-items: center;
    gap: 1rem;
    flex: 1;
    justify-content: flex-end;
  }

  .search-meta-inline {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.25rem;
    background: var(--muted-soft);
    border-radius: var(--radius-md);
  }

  .entity-cell {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    font-weight: 700;
  }

  .status-pill {
    font-size: 0.65rem;
    font-weight: 800;
    padding: 0.15rem 0.5rem;
    border-radius: 99px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .status-pill--type {
    background: var(--muted);
    color: var(--ink-soft);
  }

  .status-pill--approved {
    background: var(--success-soft);
    color: var(--success);
  }

  .search-toolbar--studio {
    padding: 1.25rem;
    background: var(--muted-soft);
    border-radius: var(--radius-lg);
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .search-toolbar__copy {
    display: flex;
    flex-direction: column;
    gap: 0.1rem;
  }

  .search-toolbar__controls {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .search-meta {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    min-height: 32px;
  }

  .search-count {
    font-size: 0.7rem;
    font-weight: 800;
    color: var(--accent);
    margin-right: auto;
  }

  .search-empty {
    font-size: 0.7rem;
    font-weight: 700;
    color: var(--destructive);
  }

  .row-actions {
    display: flex;
    gap: 0.5rem;
    justify-content: flex-end;
  }

  tr.approved {
    background: var(--success-ghost);
  }

  tr.search-match {
    background: var(--accent-ghost);
  }

  tr.search-match-active {
    box-shadow: inset 4px 0 0 var(--accent);
    background: color-mix(in srgb, var(--accent) 8%, transparent);
  }

  tr.dragging {
    opacity: 0.4;
    cursor: grabbing;
  }

  tr.drag-hover {
    border-top: 2px solid var(--accent);
  }

  .context-menu {
    position: fixed;
    z-index: 1000;
    background: var(--background);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-lg);
    padding: 0.5rem;
    min-width: 200px;
    animation: menu-in 0.15s cubic-bezier(0.19, 1, 0.22, 1);
  }

  .context-title {
    font-size: 0.7rem;
    font-weight: 900;
    text-transform: uppercase;
    color: var(--ink-soft);
    padding: 0.4rem 0.6rem;
    border-bottom: 1px solid var(--border-card);
    margin-bottom: 0.4rem;
  }

  .context-divider {
    height: 1px;
    background: var(--border-card);
    margin: 0.4rem 0;
  }

  @keyframes menu-in {
    from { transform: scale(0.95); opacity: 0; }
    to { transform: scale(1); opacity: 1; }
  }

  @media (max-width: 1024px) {
    .entities-layout {
      grid-template-columns: 1fr;
    }
    .entities-deck {
      position: static;
    }
  }
</style>
