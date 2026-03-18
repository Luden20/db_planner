<script lang="ts">
  import {onMount} from "svelte";
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

  export let entities: utils.Entity[] = [];
  export let onRefresh: () => Promise<void> = async () => {};
  export let focusEntityId: number | null = null;
  export let onJumpTo: (tab: "entities" | "relations" | "tertiary", entityId?: number | null) => void = () => {};
  let comb: CombView[] = [];
  let activeIndex = 0;
  let lastSyncedFocusId: number | null = null;
  let currentPrincipalEntity: utils.Entity | null = null;
  let stickyStack: HTMLDivElement | null = null;
  let stickyStackHeight = 0;
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

  onMount(() => {
    load();
    syncStickyStackHeight();
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
    syncFocusedEntity();
  }

  $: currentPrincipalEntity = entities.find(entity => entity.Id === comb[activeIndex]?.IdPrincipalEntity) ?? null;
  $: if (relationMenu.open && relationMenu.principalId !== comb[activeIndex]?.IdPrincipalEntity) {
    closeRelationMenu();
  }
  $: if (stickyStack) {
    syncStickyStackHeight();
  }

  const nextSlide = () => {
    closeRelationMenu();
    if (!comb.length) return;
    activeIndex = (activeIndex + 1) % comb.length;
  };

  const prevSlide = () => {
    closeRelationMenu();
    if (!comb.length) return;
    activeIndex = activeIndex === 0 ? comb.length - 1 : activeIndex - 1;
  };

  const handlePrincipalSelectChange = (event: Event) => {
    closeRelationMenu();
    const target = event.target as HTMLSelectElement;
    const nextId = Number(target?.value ?? 0);
    const nextIndex = comb.findIndex(view => view.IdPrincipalEntity === nextId);
    if (nextIndex !== -1) {
      activeIndex = nextIndex;
    }
  };

  const syncFocusedEntity = () => {
    if (!comb.length || focusEntityId === null) {
      return;
    }

    const nextIndex = comb.findIndex(view => view.IdPrincipalEntity === focusEntityId);
    if (nextIndex !== -1) {
      activeIndex = nextIndex;
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

  const goToRelatedEntityRelations = () => {
    if (relationMenu.targetId === null) {
      return;
    }

    const nextIndex = comb.findIndex(view => view.IdPrincipalEntity === relationMenu.targetId);
    if (nextIndex === -1) {
      showToast("No se pudo abrir esa relación.", "error");
      return;
    }
    activeIndex = nextIndex;
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

<svelte:window on:click={closeRelationMenu} on:keydown={handleWindowKeydown} on:scroll={closeRelationMenu}/>

<section class="relations-tab" style={`--relations-sticky-total-height: ${stickyStackHeight}px;`}>
  <div class="sticky-stack" bind:this={stickyStack}>
    <div class="tab-toolbar">
      <div>
        <p class="label">Relaciones</p>
        <p class="muted">Explora relaciones por entidad principal. Cada slide se recorre con las flechas.</p>
      </div>
      <div class="toolbar-actions">
        <div class="view-jumps">
          <button class="jump-btn" on:click={() => jumpToTab("entities")} disabled={!comb.length}>
            Ir a definicion
          </button>
          <button class="jump-btn" on:click={() => jumpToTab("tertiary")} disabled={!comb.length}>
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
          <button class="nav-btn" on:click={prevSlide} aria-label="Entidad anterior" disabled={comb.length <= 1}>&lt;</button>
          <button class="nav-btn" on:click={nextSlide} aria-label="Entidad siguiente" disabled={comb.length <= 1}>&gt;</button>
        </div>
      </div>
    </div>

    {#if comb.length !== 0 && comb[activeIndex]}
      <header class:slide-head={true} class:slide-head--approved={isApprovedEntity(comb[activeIndex].IdPrincipalEntity)}>
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
            class:approve-btn={true}
            class:approve-btn--approved={isApprovedEntity(comb[activeIndex].IdPrincipalEntity)}
            on:click={togglePrincipalApproval}
            disabled={updating}
          >
            {isApprovedEntity(comb[activeIndex].IdPrincipalEntity) ? "Quitar aprobación" : "Aprobar entidad"}
          </button>
        </div>
      </header>
    {/if}
  </div>

{#if comb.length === 0}
  <div class="empty-panel">Sin datos de relaciones.</div>
{:else}
  <div class="slide-shell">
    <article class:slide={true} class:slide--approved={isApprovedEntity(comb[activeIndex]?.IdPrincipalEntity)}>
      {#if comb[activeIndex]}
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

  <div class="slide-meta">
    <span class="counter">{activeIndex + 1} / {comb.length}</span>
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
    <button class="context-item context-item--accent" on:click={goToRelatedEntityAttributes}>
      Ir a atributos
    </button>
    <button class="context-item" on:click={goToRelatedEntityRelations}>
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
    position: sticky;
    top: var(--relations-sticky-top);
    z-index: 8;
    display: grid;
    gap: var(--relations-sticky-gap);
    margin-bottom: 18px;
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

  .nav-btn {
    border: 1px solid rgba(255, 255, 255, 0.14);
    background: rgba(255, 255, 255, 0.08);
    color: #d9e4f5;
    border-radius: 12px;
    height: 48px;
    width: 52px;
    cursor: pointer;
    transition: background 150ms ease, transform 120ms ease;
  }

  .nav-btn:hover:enabled {
    background: rgba(255, 255, 255, 0.12);
    transform: translateY(-1px);
  }

  .nav-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
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

  .approve-btn {
    border: 1px solid rgba(121, 205, 126, 0.32);
    background: rgba(48, 83, 52, 0.32);
    color: #dff7df;
    border-radius: 10px;
    padding: 10px 12px;
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

  .context-item:active {
    transform: translateY(1px);
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

    .jump-btn {
      flex-basis: 100%;
    }

    .entity-select {
      min-width: 0;
      width: 100%;
    }

    .slide-head {
      flex-direction: column;
      align-items: stretch;
    }

    .nav-btn {
      width: 40px;
      height: 44px;
    }
  }
</style>
