<script lang="ts">
  import {onMount, tick} from "svelte";
  import type {utils} from "../../wailsjs/go/models";
  import {AddRelation, GetCombinatory, MarkEntityStatus, RemoveRelation} from "../../wailsjs/go/main/App";
  import CreateEntity from "./forms/CreateEntity.svelte";
  import {showToast} from "../lib/toast";

  type RelationRow = {
    Id?: number;
    Entity2: string;
    IdEntity2: number;
    Relation: string;
  };

  type CombView = {
    PrincipalEntity: string;
    IdPrincipalEntity: number;
    Relations: RelationRow[];
  };
  type ViewTransitionDocument = Document & {
    startViewTransition?: (update: () => void | Promise<void>) => {
      finished: Promise<void>;
    };
  };

  export let entities: utils.Entity[] = [];
  export let onRefresh: () => Promise<void> = async () => {};
  export let focusEntityId: number | null = null;
  export let onJumpTo: (tab: "entities" | "relations" | "tertiary", entityId?: number | null) => void = () => {};
  let comb: CombView[] = [];
  let activeIndex = 0;
  let lastSyncedFocusId: number | null = null;
  let currentPrincipalEntity: utils.Entity | null = null;
  let stickySentinel: HTMLDivElement | null = null;
  let stickyStack: HTMLDivElement | null = null;
  let stickyStackHeight = 0;
  let stickyStackPinned = false;
  let currentRelationCount = 0;
  let approvedRelationCount = 0;
  const relationOptions = ["", "1:1", "N:1", "1:N", "N:N"];
  let updating = false;
  let relationOverrides: Record<string, string> = {};
  let relationMenu: {
    open: boolean;
    x: number;
    y: number;
    principalId: number | null;
    principalName: string;
    targetId: number | null;
    targetName: string;
  } = {
    open: false,
    x: 0,
    y: 0,
    principalId: null,
    principalName: "",
    targetId: null,
    targetName: ""
  };

  const applyOverrides = () => {
    comb = comb.map(view => ({
      ...view,
      Relations: view.Relations.map(rel => {
        const key = `${view.IdPrincipalEntity}-${rel.IdEntity2}`;
        const override = relationOverrides[key];
        return override !== undefined ? {...rel, Relation: override} : rel;
      })
    }));
  };

  async function load(keepIndex = false) {
    const prevPrincipalId = keepIndex
      ? (comb[activeIndex]?.IdPrincipalEntity ?? null)
      : null;
    const data = await GetCombinatory();
    comb = (data || []).map(view => ({
      ...view,
      Relations: (view.Relations || []).map(rel => ({...rel, Relation: rel.Relation ?? ""}))
    }));
    applyOverrides();
    if (keepIndex && prevPrincipalId !== null) {
      const preservedIndex = comb.findIndex(view => view.IdPrincipalEntity === prevPrincipalId);
      activeIndex = preservedIndex !== -1
        ? preservedIndex
        : Math.min(activeIndex, Math.max(comb.length - 1, 0));
      return;
    }

    activeIndex = 0;
  }

  const syncStickyStackHeight = () => {
    stickyStackHeight = stickyStack?.offsetHeight ?? 0;
  };

  const prefersReducedMotion = () =>
    typeof window !== "undefined"
    && typeof window.matchMedia === "function"
    && window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  const runRelationTransition = async (update: () => void | Promise<void>) => {
    const doc = typeof document !== "undefined" ? (document as ViewTransitionDocument) : null;
    if (doc?.startViewTransition && !prefersReducedMotion()) {
      try {
        const transition = doc.startViewTransition(update);
        await transition.finished;
        return;
      } catch (err) {
        console.warn("No se pudo aplicar la transicion de relaciones:", err);
      }
    }
    await update();
  };

  const syncStickyState = () => {
    if (!stickySentinel) {
      stickyStackPinned = false;
      return;
    }

    stickyStackPinned = stickySentinel.getBoundingClientRect().top <= 0;
  };

  onMount(() => {
    load();
    syncStickyStackHeight();
    syncStickyState();
    if (typeof ResizeObserver === "undefined" || !stickyStack) {
      return;
    }

    const observer = new ResizeObserver(() => {
      syncStickyStackHeight();
    });
    observer.observe(stickyStack);

    return () => {
      observer.disconnect();
    };
  });

  $: if (activeIndex >= comb.length) {
    activeIndex = comb.length ? comb.length - 1 : 0;
  }

  $: if (comb.length && focusEntityId !== null && focusEntityId !== lastSyncedFocusId) {
    void syncFocusedEntity();
  }

  $: currentPrincipalEntity = entities.find(entity => entity.Id === comb[activeIndex]?.IdPrincipalEntity) ?? null;
  $: currentRelationCount = comb[activeIndex]?.Relations?.filter(relation => Boolean(relation.Relation)).length ?? 0;
  $: approvedRelationCount = comb[activeIndex]?.Relations?.filter(relation => isApprovedEntity(relation.IdEntity2)).length ?? 0;
  $: if (relationMenu.open && relationMenu.principalId !== comb[activeIndex]?.IdPrincipalEntity) {
    closeRelationMenu();
  }
  $: if (stickyStack) {
    syncStickyStackHeight();
  }
  $: if (stickySentinel) {
    syncStickyState();
  }

  const selectPrincipalIndex = async (nextIndex: number) => {
    if (!comb.length || nextIndex < 0 || nextIndex >= comb.length) {
      return;
    }
    await runRelationTransition(async () => {
      activeIndex = nextIndex;
      await tick();
    });
  };

  const nextSlide = async () => {
    closeRelationMenu();
    if (!comb.length) return;
    await selectPrincipalIndex((activeIndex + 1) % comb.length);
  };

  const prevSlide = async () => {
    closeRelationMenu();
    if (!comb.length) return;
    await selectPrincipalIndex(activeIndex === 0 ? comb.length - 1 : activeIndex - 1);
  };

  const handlePrincipalSelectChange = async (event: Event) => {
    closeRelationMenu();
    const target = event.target as HTMLSelectElement;
    const nextId = Number(target?.value ?? 0);
    const nextIndex = comb.findIndex(view => view.IdPrincipalEntity === nextId);
    if (nextIndex !== -1) {
      await selectPrincipalIndex(nextIndex);
    }
  };

  const syncFocusedEntity = async () => {
    if (!comb.length || focusEntityId === null) {
      return;
    }

    const nextIndex = comb.findIndex(view => view.IdPrincipalEntity === focusEntityId);
    if (nextIndex !== -1) {
      await selectPrincipalIndex(nextIndex);
    }
    lastSyncedFocusId = focusEntityId;
  };

  const jumpToTab = (tab: "entities" | "tertiary") => {
    const entityId = comb[activeIndex]?.IdPrincipalEntity ?? null;
    onJumpTo(tab, entityId);
  };

  const closeRelationMenu = () => {
    relationMenu = {
      open: false,
      x: 0,
      y: 0,
      principalId: null,
      principalName: "",
      targetId: null,
      targetName: ""
    };
  };

  const openRelationMenu = (relation: RelationRow, event: MouseEvent) => {
    if (updating) {
      return;
    }
    const principalId = comb[activeIndex]?.IdPrincipalEntity ?? null;
    const principalName = comb[activeIndex]?.PrincipalEntity ?? "";
    if (principalId === null) {
      return;
    }

    const menuWidth = 240;
    const menuHeight = 164;
    relationMenu = {
      open: true,
      x: Math.max(12, Math.min(event.clientX, window.innerWidth - menuWidth - 12)),
      y: Math.max(12, Math.min(event.clientY, window.innerHeight - menuHeight - 12)),
      principalId,
      principalName,
      targetId: relation.IdEntity2,
      targetName: relation.Entity2
    };
  };

  const handleWindowKeydown = (event: KeyboardEvent) => {
    if (event.key === "Escape") {
      closeRelationMenu();
    }
  };

  const goToRelatedEntityRelations = async () => {
    if (relationMenu.targetId === null) {
      return;
    }

    const nextIndex = comb.findIndex(view => view.IdPrincipalEntity === relationMenu.targetId);
    if (nextIndex === -1) {
      showToast("No se pudo abrir esa relación.", "error");
      return;
    }
    await selectPrincipalIndex(nextIndex);
    closeRelationMenu();
  };

  const goToRelatedEntityAttributes = () => {
    if (relationMenu.targetId === null) {
      return;
    }
    const targetId = relationMenu.targetId;
    closeRelationMenu();
    onJumpTo("tertiary", targetId);
  };

  const isApprovedEntity = (entityId: number | null | undefined) =>
    entityId != null && entities.some(entity => entity.Id === entityId && entity.Status === true);

  const getEntityDefinition = (entityId: number) =>
    entities.find(entity => entity.Id === entityId)?.Description || "Sin definición.";

  const relationIdentifiers = (relation: RelationRow) => ({
    principalId: comb[activeIndex]?.IdPrincipalEntity ?? null,
    // Usamos siempre el Id si existe, incluso cuando Relation está vacío
    relationId: relation?.Id ?? null,
    targetId: relation?.IdEntity2 ?? null
  });

  const togglePrincipalApproval = async () => {
    const principalId = comb[activeIndex]?.IdPrincipalEntity;
    if (principalId == null) {
      return;
    }

    updating = true;
    try {
      await MarkEntityStatus(principalId, !isApprovedEntity(principalId));
      await onRefresh();
    } catch (err) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      showToast(`No se pudo actualizar la aprobación: ${message}`, "error");
    } finally {
      updating = false;
    }
  };

  const handleRelationChange = async (relation: RelationRow) => {
    const ids = relationIdentifiers(relation);
    const selection = relation.Relation || "";
    if (ids.principalId == null || ids.targetId == null) {
      showToast("Faltan IDs de entidad para registrar la relación.", "error");
      return;
    }

    const key = `${ids.principalId}-${ids.targetId}`;
    relationOverrides[key] = selection; // Guardamos incluso vacío para respetar la selección
    applyOverrides(); // Optimista: refleja en UI mientras persiste

    updating = true;
    try {
      if (!selection) {
        // Borrar relación existente solo si hay ID persistido
        if (ids.relationId != null) {
          showToast("Eliminando relación...", "info", 1200);
          await RemoveRelation(ids.relationId);
        }
      } else {
        await AddRelation(ids.principalId, ids.targetId, selection);
      }
      await load(true);
      await onRefresh();
      applyOverrides();
      showToast(selection ? "Relación actualizada." : "Relación eliminada.", "success");
    } catch (err) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      showToast(`No se pudo actualizar la relación: ${message}`, "error");
    } finally {
      updating = false;
    }
  };
</script>

<svelte:window
  on:click={closeRelationMenu}
  on:keydown={handleWindowKeydown}
  on:scroll={() => {
    closeRelationMenu();
    syncStickyState();
  }}
  on:resize={syncStickyState}
/>

<section class="relations-tab relations-studio" style={`--relations-sticky-total-height: ${stickyStackHeight}px;`}>
  <div class="sticky-sentinel" bind:this={stickySentinel} aria-hidden="true"></div>
  <div class:sticky-stack={true} class:sticky-stack--pinned={stickyStackPinned} bind:this={stickyStack}>
    <div class="tab-toolbar relations-toolbar">
      <div class="relations-toolbar__copy">
        <p class="label">Relaciones</p>
        <p class="muted">Recorre el combinatorio por entidad principal y ajusta el cruce sin perder contexto.</p>
      </div>
      <div class="relations-toolbar__meta">
        <span class="studio-chip">{comb.length} entidades</span>
        <span class="studio-chip studio-chip--quiet">{currentRelationCount} relaciones activas</span>
        <span class="studio-chip studio-chip--quiet">{approvedRelationCount} relacionadas aprobadas</span>
      </div>
      <div class="toolbar-actions relations-toolbar__actions">
        <div class="view-jumps">
          <button class="control control--ghost" on:click={() => jumpToTab("entities")} disabled={!comb.length}>
            Ir a definicion
          </button>
          <button class="control control--accent" on:click={() => jumpToTab("tertiary")} disabled={!comb.length}>
            Ir a atributos
          </button>
        </div>
        <select
          class="entity-select"
          value={comb[activeIndex]?.IdPrincipalEntity ?? ""}
          on:change={handlePrincipalSelectChange}
          disabled={!comb.length}
        >
          {#each comb as view}
            <option value={view.IdPrincipalEntity}>{view.PrincipalEntity}</option>
          {/each}
        </select>
        <div class="entity-nav">
          <button class="control control--icon control--soft" on:click={prevSlide} aria-label="Entidad anterior" disabled={comb.length <= 1}>
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M14.78 5.47a.75.75 0 0 1 0 1.06L10.31 11l4.47 4.47a.75.75 0 0 1-1.06 1.06l-5-5a.75.75 0 0 1 0-1.06l5-5a.75.75 0 0 1 1.06 0Z"/>
            </svg>
          </button>
          <button class="control control--icon control--soft" on:click={nextSlide} aria-label="Entidad siguiente" disabled={comb.length <= 1}>
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M9.22 5.47a.75.75 0 0 1 1.06 0l5 5a.75.75 0 0 1 0 1.06l-5 5a.75.75 0 1 1-1.06-1.06L13.69 11 9.22 6.53a.75.75 0 0 1 0-1.06Z"/>
            </svg>
          </button>
        </div>
      </div>
    </div>

  </div>

{#if comb.length === 0}
  <div class="empty-panel">Sin datos de relaciones.</div>
{:else}
  <div class="relations-layout">
    <aside class="relations-deck">
      {#if comb[activeIndex]}
        <header
          class:slide-head={true}
          class:slide-head--approved={isApprovedEntity(comb[activeIndex].IdPrincipalEntity)}
          style={`view-transition-name: relation-head-${comb[activeIndex].IdPrincipalEntity};`}
        >
          <div class="slide-head-copy">
            <p class="label">Entidad principal</p>
            <div class="entity-title-row">
              <h3>{comb[activeIndex].PrincipalEntity}</h3>
              {#if isApprovedEntity(comb[activeIndex].IdPrincipalEntity)}
                <span class="status-pill status-pill--approved">&#10003;</span>
              {/if}
            </div>
            <p class="entity-description">{currentPrincipalEntity?.Description || "Sin definición."}</p>
          </div>
          <div class="head-meta head-meta--actions">
            <div>
              <p class="mini-label">ID</p>
              <p class="id-pill">{comb[activeIndex].IdPrincipalEntity}</p>
            </div>
            <CreateEntity
              id={comb[activeIndex].IdPrincipalEntity}
              onSave={async () => {
                await load(true);
                await onRefresh();
              }}
            />
            <button
              class={`control control--success ${isApprovedEntity(comb[activeIndex].IdPrincipalEntity) ? 'control--active' : ''}`}
              on:click={togglePrincipalApproval}
              disabled={updating}
            >
              {isApprovedEntity(comb[activeIndex].IdPrincipalEntity) ? "Quitar aprobación" : "Aprobar entidad"}
            </button>
          </div>
        </header>
      {/if}
    </aside>

    <div class="slide-shell">
      <article
        class:slide={true}
        class:slide--approved={isApprovedEntity(comb[activeIndex]?.IdPrincipalEntity)}
        style={`view-transition-name: relation-stage-${comb[activeIndex]?.IdPrincipalEntity ?? "empty"};`}
      >
        {#if comb[activeIndex]}
          <div class="slide-panel__head">
            <div>
              <p class="label">Matriz activa</p>
              <p class="muted">Cada fila representa una entidad relacionada. El selector aplica el tipo de cardinalidad.</p>
            </div>
            <span class="slide-panel__hint">Edicion directa</span>
          </div>
          <div class="table-wrapper">
            <table class="entities-table compact">
              <tbody>
              {#if comb[activeIndex].Relations && comb[activeIndex].Relations.length > 0}
                {#each comb[activeIndex].Relations as relation}
                  <tr
                    class:approved-row={isApprovedEntity(relation.IdEntity2)}
                    class:relation-row-menu-open={relationMenu.open && relationMenu.targetId === relation.IdEntity2}
                    on:contextmenu|preventDefault|stopPropagation={(event) => openRelationMenu(relation, event)}
                  >
                    <td>
                      <div class="relation-name-cell">
                        <span>{relation.Entity2}</span>
                        <span class="relation-info">
                          <button
                            type="button"
                            class="info-trigger"
                            aria-label="Ayuda de la relación"
                            on:click|stopPropagation
                          >
                            ...
                          </button>
                          <span class="relation-tooltip">
                            {getEntityDefinition(relation.IdEntity2)}
                          </span>
                        </span>
                      </div>
                    </td>
                    <td>
                      <select
                        class="relation-select"
                        bind:value={relation.Relation}
                        data-principal={relationIdentifiers(relation).principalId}
                        data-relation-id={relationIdentifiers(relation).relationId}
                        data-target={relationIdentifiers(relation).targetId}
                        on:click|stopPropagation
                        on:change={() => handleRelationChange(relation)}
                        disabled={updating}
                      >
                        {#each relationOptions as option}
                          <option value={option}>{option || "-"}</option>
                        {/each}
                      </select>
                    </td>

                  </tr>
                {/each}
              {:else}
                <tr class="muted-row">
                  <td colspan="5">Sin relaciones para esta entidad.</td>
                </tr>
              {/if}
              </tbody>
            </table>
          </div>
        {/if}
      </article>
    </div>
  </div>
{/if}

{#if relationMenu.open}
  <div
    class="context-menu"
    style={`left: ${relationMenu.x}px; top: ${relationMenu.y}px;`}
    on:click|stopPropagation
    on:keydown|stopPropagation
  >
    <p class="context-title">{relationMenu.principalName} -> {relationMenu.targetName}</p>
    <button class="menu-action control control--sm control--block control--accent" on:click={goToRelatedEntityAttributes}>
      Ir a atributos
    </button>
    <button class="menu-action control control--sm control--block control--ghost" on:click={goToRelatedEntityRelations}>
      Ir a esta relacion
    </button>
  </div>
{/if}
</section>

<style>
  .relations-tab {
    --relations-sticky-top: 0px;
    --relations-sticky-gap: 0px;
    --relations-sticky-total-height: 0px;
  }

  .sticky-stack {
    z-index: 8;
    display: grid;
    gap: var(--relations-sticky-gap);
    margin-bottom: 18px;
  }

  .sticky-stack--pinned {
    position: sticky;
    top: 0;
  }

  .sticky-sentinel {
    height: 1px;
    margin-top: -1px;
  }

  .tab-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
    padding: 12px 14px;
    border-radius: 14px;
    border: 1px solid rgba(255, 255, 255, 0.1);
    background: linear-gradient(135deg, rgba(11, 18, 31, 0.96), rgba(19, 29, 45, 0.96));
    box-shadow: 0 16px 34px rgba(0, 0, 0, 0.3);
    backdrop-filter: blur(8px);
    border-bottom-left-radius: 0;
    border-bottom-right-radius: 0;
  }

  .toolbar-actions {
    display: flex;
    gap: 10px;
    align-items: center;
    flex: 1;
    flex-wrap: wrap;
  }

  .view-jumps {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
    justify-content: stretch;
    flex: 1 1 320px;
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

  .entity-nav {
    display: inline-flex;
    gap: 8px;
    align-items: center;
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

  .slide-shell {
    display: block;
  }

  .slide {
    background: rgba(255, 255, 255, 0.02);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 14px;
    padding: 14px 14px 16px;
    box-shadow: none;
    display: flex;
    flex-direction: column;
    gap: 12px;
    min-height: 220px;
    position: relative;
  }

  .slide--approved {
    border-color: rgba(102, 194, 108, 0.34);
    box-shadow: inset 0 0 0 1px rgba(73, 150, 78, 0.18);
  }

  .slide-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
    padding: 12px 14px;
    background: #0f1726;
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 12px;
    box-shadow: 0 10px 24px rgba(0, 0, 0, 0.35);
    border-top: 0;
    border-top-left-radius: 0;
    border-top-right-radius: 0;
  }

  .slide-head--approved {
    background: linear-gradient(135deg, rgba(17, 45, 25, 0.96), rgba(21, 57, 31, 0.96));
    border-color: rgba(113, 201, 118, 0.35);
  }

  .slide-head h3 {
    margin: 4px 0 0;
    font-size: 18px;
  }

  .slide-head-copy {
    flex: 1;
    min-width: 0;
  }

  .entity-title-row {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
  }

  .entity-description {
    margin: 10px 0 0;
    color: #d9e4f5;
    opacity: 0.82;
    line-height: 1.45;
  }

  .head-meta {
    text-align: right;
  }

  .head-meta--actions {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
    justify-content: flex-end;
  }

  .mini-label {
    margin: 0;
    color: #9ab5e4;
    font-size: 11px;
    letter-spacing: 0.6px;
    text-transform: uppercase;
  }

  .id-pill {
    margin: 4px 0 0;
    padding: 6px 10px;
    background: rgba(109, 216, 255, 0.12);
    border: 1px solid rgba(109, 216, 255, 0.35);
    color: #cfeeff;
    border-radius: 10px;
    font-weight: 700;
    font-size: 13px;
  }

  .table-wrapper {
    overflow: visible;
  }

  .entities-table {
    width: 100%;
    border-collapse: collapse;
    color: #e8edf7;
  }

  .entities-table td {
    text-align: left;
    padding: 10px 10px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.07);
    font-size: 13px;
  }

  .entities-table tbody tr:nth-child(odd) {
    background: rgba(255, 255, 255, 0.025);
  }

  .entities-table tbody tr:nth-child(even) {
    background: rgba(109, 216, 255, 0.045);
  }

  .entities-table tbody tr.approved-row {
    background: rgba(76, 175, 80, 0.12);
  }

  .entities-table tbody tr:hover {
    background: rgba(135, 202, 255, 0.1);
  }

  .entities-table tbody tr.relation-row-menu-open {
    background: rgba(90, 209, 255, 0.16);
    box-shadow: inset 0 0 0 1px rgba(90, 209, 255, 0.34);
  }

  .status-pill {
    display: inline-flex;
    align-items: center;
    padding: 4px 10px;
    border-radius: 999px;
    font-size: 12px;
    font-weight: 700;
  }

  .status-pill--approved {
    color: #dff7df;
    background: rgba(76, 175, 80, 0.2);
    border: 1px solid rgba(113, 201, 118, 0.35);
    min-width: 30px;
    justify-content: center;
  }

  .entities-table.compact td {
    font-size: 13px;
  }

  .relation-name-cell {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
  }

  .relation-info {
    position: relative;
    display: inline-flex;
    align-items: center;
    flex-shrink: 0;
  }

  .info-trigger {
    border: 1px solid rgba(255, 255, 255, 0.14);
    background: rgba(255, 255, 255, 0.06);
    color: #cfe2ff;
    border-radius: 999px;
    min-width: 30px;
    height: 30px;
    padding: 0 8px;
    cursor: help;
    font-size: 12px;
    letter-spacing: 0.8px;
    transition: background 140ms ease, border-color 140ms ease, transform 120ms ease;
  }

  .info-trigger:hover,
  .info-trigger:focus-visible {
    background: rgba(90, 209, 255, 0.14);
    border-color: rgba(90, 209, 255, 0.28);
    outline: none;
  }

  .relation-tooltip {
    position: absolute;
    right: 0;
    top: calc(100% + 8px);
    width: min(260px, 70vw);
    padding: 9px 10px;
    border-radius: 10px;
    background: rgba(10, 16, 27, 0.96);
    border: 1px solid rgba(255, 255, 255, 0.1);
    color: #d9e4f5;
    font-size: 12px;
    line-height: 1.4;
    box-shadow: 0 14px 32px rgba(0, 0, 0, 0.28);
    opacity: 0;
    transform: translateY(-4px);
    pointer-events: none;
    transition: opacity 140ms ease, transform 140ms ease;
    z-index: 4;
  }

  .relation-info:hover .relation-tooltip,
  .relation-info:focus-within .relation-tooltip {
    opacity: 1;
    transform: translateY(0);
  }

  .muted-row td {
    background: rgba(255, 255, 255, 0.02);
    color: #cfd9e9;
    opacity: 0.8;
    text-align: center;
  }

  .slide-meta {
    margin-top: 8px;
    text-align: center;
    color: #cfd9e9;
    opacity: 0.8;
    font-size: 13px;
    letter-spacing: 0.6px;
  }

  .counter {
    display: inline-block;
    padding: 6px 10px;
    border: 1px solid rgba(255, 255, 255, 0.14);
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.04);
  }

  .relation-select {
    width: 100%;
    min-width: 120px;
    border-radius: 10px;
    border: 1px solid rgba(255, 255, 255, 0.14);
    background: #0f1726;
    color: #e8edf7;
    padding: 8px 10px;
    font-size: 13px;
    outline: none;
    transition: border 140ms ease, box-shadow 140ms ease;
    appearance: none;
  }

  .relation-select:focus {
    border-color: rgba(90, 209, 255, 0.8);
    box-shadow: 0 0 0 2px rgba(90, 209, 255, 0.22);
  }

  .context-menu {
    position: fixed;
    z-index: 60;
    min-width: 240px;
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

  @media (max-width: 720px) {
    .relations-tab {
      --relations-sticky-gap: 0px;
    }

    .tab-toolbar {
      align-items: stretch;
      flex-direction: column;
    }

    .toolbar-actions {
      width: 100%;
    }

    .view-jumps {
      width: 100%;
    }

    .entity-select {
      min-width: 0;
      width: 100%;
    }

    .slide-head {
      flex-direction: column;
      align-items: stretch;
    }

  }

  .tab-toolbar {
    padding: 1.05rem 1.1rem;
    border: 1px solid var(--border);
    background: var(--panel-surface);
    box-shadow: var(--shadow-sm);
  }

  .label,
  .mini-label,
  .context-title {
    color: var(--accent);
    letter-spacing: 0.14em;
  }

  .label,
  .mini-label {
    font-size: 0.74rem;
    font-weight: 800;
  }

  .muted,
  .entity-description,
  .counter {
    color: var(--ink-faint);
    opacity: 1;
  }

  .entity-select,
  .relation-select {
    border-color: var(--border);
    background: var(--field-surface);
    color: var(--ink);
    box-shadow: none;
  }

  .entity-select:focus,
  .relation-select:focus {
    border-color: var(--focus-border);
    box-shadow: var(--focus-ring);
  }

  .slide-head {
    background: var(--panel-surface-strong);
    border-color: var(--border);
    box-shadow: var(--shadow-sm);
    border-top: 1px solid var(--border);
    border-radius: calc(var(--radius-md) - 4px);
  }

  .slide-head--approved {
    border-color: color-mix(in srgb, var(--success) 24%, var(--border));
    background: var(--panel-surface-success);
  }

  .slide-head h3,
  .relation-name-cell,
  .entities-table {
    color: var(--ink);
  }

  .status-pill,
  .pill {
    background: var(--chip-surface);
    border-color: var(--line-soft);
    color: var(--ink-soft);
  }

  .id-pill {
    background: var(--chip-accent-surface);
    border-color: color-mix(in srgb, var(--accent) 24%, var(--border));
    color: var(--accent-strong);
  }

  .status-pill--approved {
    background: var(--chip-success-surface);
    border-color: color-mix(in srgb, var(--success) 24%, var(--border));
    color: var(--success);
  }

  .slide {
    background: var(--panel-surface-strong);
    border-color: var(--border);
    border-radius: calc(var(--radius-md) - 4px);
  }

  .slide--approved {
    border-color: rgba(61, 114, 81, 0.18);
    box-shadow: inset 0 0 0 1px rgba(61, 114, 81, 0.08);
  }

  .table-wrapper {
    background: color-mix(in srgb, var(--surface-strong) 68%, transparent);
    border-color: transparent;
  }

  .entities-table tbody tr:nth-child(odd),
  .entities-table tbody tr:nth-child(even) {
    background: transparent;
  }

  .entities-table tbody tr.approved-row {
    background: var(--success-soft);
  }

  .entities-table tbody tr:hover,
  .entities-table tbody tr.relation-row-menu-open {
    background: var(--hover-soft);
  }

  .muted-row td {
    background: color-mix(in srgb, var(--surface) 72%, transparent);
    color: var(--ink-faint);
  }

  .info-trigger {
    border-color: var(--line-soft);
    background: color-mix(in srgb, var(--surface-strong) 72%, transparent);
    color: var(--ink-faint);
  }

  .info-trigger:hover,
  .info-trigger:focus-visible {
    background: color-mix(in srgb, var(--accent) 12%, var(--surface-strong));
    border-color: color-mix(in srgb, var(--accent) 24%, var(--border));
  }

  .relation-tooltip {
    background: var(--popover-surface);
    border-color: var(--border);
    color: var(--ink-soft);
    box-shadow: var(--shadow-sm);
  }

  .counter {
    border-color: var(--line-soft);
    background: color-mix(in srgb, var(--surface-strong) 74%, transparent);
  }

  .context-menu {
    border-color: var(--border);
    background: var(--popover-surface);
    box-shadow: var(--shadow-sm);
  }

  .relations-studio {
    display: grid;
    gap: 1rem;
  }

  .relations-layout {
    display: grid;
    grid-template-columns: minmax(18rem, 24rem) minmax(0, 1fr);
    align-items: start;
    gap: 1rem;
  }

  .relations-deck {
    display: grid;
    gap: 1rem;
    position: sticky;
    top: calc(var(--relations-sticky-total-height) + 1rem);
    align-self: start;
  }

  .relations-toolbar,
  .slide-shell,
  .slide-head {
    position: relative;
    overflow: clip;
  }

  .relations-toolbar::before,
  .slide-shell::before,
  .slide-head::before {
    content: "";
    position: absolute;
    inset: 0 auto auto 0;
    width: min(220px, 42%);
    height: 1px;
    background: linear-gradient(90deg, color-mix(in srgb, var(--accent) 36%, transparent), transparent);
    pointer-events: none;
  }

  .relations-toolbar {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: start;
    gap: 0.9rem 1rem;
  }

  .relations-toolbar__copy {
    max-width: 38rem;
  }

  .relations-toolbar__meta {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 0.65rem;
    flex-wrap: wrap;
  }

  .relations-toolbar__actions {
    grid-column: 1 / -1;
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

  .slide-shell {
    padding: 1rem;
    border: 1px solid var(--border);
    border-radius: calc(var(--radius-lg) - 2px);
    background:
      radial-gradient(circle at top right, color-mix(in srgb, var(--accent) 8%, transparent), transparent 34%),
      var(--panel-surface);
    box-shadow: var(--shadow-sm);
  }

  .slide-panel__head {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 1rem;
    margin-bottom: 0.9rem;
  }

  .slide-panel__hint {
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

  .slide-meta {
    display: flex;
    justify-content: center;
    margin-top: 0;
  }

  @media (max-width: 720px) {
    .relations-layout {
      grid-template-columns: 1fr;
    }

    .relations-deck {
      position: static;
    }

    .relations-toolbar,
    .slide-panel__head {
      grid-template-columns: 1fr;
      align-items: stretch;
    }

    .relations-toolbar__meta {
      justify-content: flex-start;
    }

    .slide-shell {
      padding: 0.9rem;
    }

    .slide-panel__hint,
    .studio-chip {
      white-space: normal;
    }
  }
</style>
