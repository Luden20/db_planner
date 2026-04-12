<script lang="ts">
  import {onMount, tick} from "svelte";
  import type {utils} from "../../wailsjs/go/models";
  import {
    AddRelation,
    GetCombinatory,
    GetRelationTypes,
    IntersectionHasAttributes,
    MarkEntityStatus,
    RemoveRelation
  } from "../../wailsjs/go/main/App";
  import EntityFocusCard from "./EntityFocusCard.svelte";
  import CreateEntity from "./forms/CreateEntity.svelte";
  import StudioToolbar from "./studio/StudioToolbar.svelte";
  import StickyStack from "./studio/StickyStack.svelte";
  import EmptyPanel from "./studio/EmptyPanel.svelte";
  import Button from "./ui/Button.svelte";
  import RelationsTable from "./RelationsTable.svelte";
  import Badge from "./ui/Badge.svelte";
  import {showToast} from "../lib/toast";
  import {getErrorMessage, runViewTransition} from "../lib/ui-helpers";

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

  let { 
    entities = [], 
    onRefresh = async () => {}, 
    focusEntityId = null, 
    onJumpTo = () => {} 
  } = $props<{
    entities?: utils.Entity[];
    onRefresh?: () => Promise<void>;
    focusEntityId?: number | null;
    onJumpTo?: (tab: "entities" | "relations" | "tertiary", entityId: number | null) => void;
  }>();

  let comb = $state<CombView[]>([]);
  let activeIndex = $state(0);
  let lastSyncedFocusId = $state<number | null>(null);
  let stickyStackHeight = $state(0);
  let relationOptions = $state<string[]>([]);
  let loadingOptions = $state(true);
  let updating = $state(false);
  let relationOverrides = $state<Record<string, string>>({});
  
  let relationMenu = $state({
    open: false,
    x: 0,
    y: 0,
    principalId: null as number | null,
    principalName: "",
    targetId: null as number | null,
    targetName: ""
  });

  const currentPrincipalEntity = $derived(entities.find(entity => entity.Id === comb[activeIndex]?.IdPrincipalEntity) ?? null);
  const currentRelationCount = $derived(comb[activeIndex]?.Relations?.filter(relation => Boolean(relation.Relation)).length ?? 0);
  const approvedRelationCount = $derived(comb[activeIndex]?.Relations?.filter(relation => isApprovedEntity(relation.IdEntity2)).length ?? 0);

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
    try {
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
    } catch (err) {
      console.error("Error al cargar combinatorio:", err);
    }
  }

  const runRelationTransition = (update: () => void | Promise<void>) =>
    runViewTransition(update, "No se pudo aplicar la transicion de relaciones:");

  onMount(async () => {
    try {
      const types = await GetRelationTypes();
      relationOptions = types || ["", "1:1", "N:1", "1:N", "N:N"];
    } catch (err) {
      console.error("Error al cargar tipos de relación:", err);
      relationOptions = ["", "1:1", "N:1", "1:N", "N:N"];
    } finally {
      loadingOptions = false;
    }
    load();
  });

  $effect(() => {
    if (activeIndex >= comb.length) {
      activeIndex = comb.length ? comb.length - 1 : 0;
    }
  });

  $effect(() => {
    if (comb.length && focusEntityId !== null && focusEntityId !== lastSyncedFocusId) {
      void syncFocusedEntity();
    }
  });

  $effect(() => {
    if (relationMenu.open && relationMenu.principalId !== comb[activeIndex]?.IdPrincipalEntity) {
      closeRelationMenu();
    }
  });

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

  const jumpToTab = (tab: "entities" | "relations" | "tertiary") => {
    const entityId = comb[activeIndex]?.IdPrincipalEntity ?? null;
    onJumpTo(tab, entityId);
  };

  const closeRelationMenu = () => {
    relationMenu.open = false;
    relationMenu.principalId = null;
    relationMenu.principalName = "";
    relationMenu.targetId = null;
    relationMenu.targetName = "";
  };

  const openRelationMenu = (relation: RelationRow, event: MouseEvent) => {
    if (updating) return;
    const principalId = comb[activeIndex]?.IdPrincipalEntity ?? null;
    const principalName = comb[activeIndex]?.PrincipalEntity ?? "";
    if (principalId === null) return;

    const menuWidth = 240;
    const menuHeight = 164;
    relationMenu.open = true;
    relationMenu.x = Math.max(12, Math.min(event.clientX, window.innerWidth - menuWidth - 12));
    relationMenu.y = Math.max(12, Math.min(event.clientY, window.innerHeight - menuHeight - 12));
    relationMenu.principalId = principalId;
    relationMenu.principalName = principalName;
    relationMenu.targetId = relation.IdEntity2;
    relationMenu.targetName = relation.Entity2;
  };

  const handleWindowKeydown = (event: KeyboardEvent) => {
    if (event.key === "Escape") closeRelationMenu();
  };

  const goToRelatedEntityRelations = async () => {
    if (relationMenu.targetId === null) return;
    const nextIndex = comb.findIndex(view => view.IdPrincipalEntity === relationMenu.targetId);
    if (nextIndex === -1) {
      showToast("No se pudo abrir esa relación.", "error");
      return;
    }
    await selectPrincipalIndex(nextIndex);
    closeRelationMenu();
  };

  const goToRelatedEntityAttributes = () => {
    if (relationMenu.targetId === null) return;
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
    relationId: relation?.Id ?? null,
    targetId: relation?.IdEntity2 ?? null
  });

  const togglePrincipalApproval = async () => {
    const principalId = comb[activeIndex]?.IdPrincipalEntity;
    if (principalId == null) return;
    updating = true;
    try {
      await MarkEntityStatus(principalId, !isApprovedEntity(principalId));
      await onRefresh();
    } catch (err) {
      const message = getErrorMessage(err);
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
    relationOverrides[key] = selection;
    applyOverrides();

    if (ids.relationId != null) {
      const prevRelation = comb[activeIndex]?.Relations.find(r => r.Id === ids.relationId)?.Relation;
      if (prevRelation === "N:N" && selection !== "N:N") {
        try {
          const hasAttrs = await IntersectionHasAttributes(ids.relationId);
          if (hasAttrs) {
            const confirm = window.confirm(
              "Esta intersección ya tiene datos (atributos reales). ¿Estás seguro de que quieres cambiar o eliminar la relación? Los atributos de la intersección se perderán."
            );
            if (!confirm) {
              await load(true);
              return;
            }
          }
        } catch (e) {
          console.error("Error al verificar atributos de intersección:", e);
        }
      }
    }

    updating = true;
    try {
      if (!selection) {
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
      const message = getErrorMessage(err);
      showToast(`No se pudo actualizar la relación: ${message}`, "error");
    } finally {
      updating = false;
    }
  };
</script>

<svelte:window
  onclick={closeRelationMenu}
  onkeydown={handleWindowKeydown}
  onscroll={closeRelationMenu}
/>

<section class="relations-tab relations-studio" style={`--relations-sticky-total-height: ${stickyStackHeight}px;`}>
  <StickyStack bind:height={stickyStackHeight}>
    <StudioToolbar 
      title="Relaciones" 
      description="Recorre el combinatorio por entidad principal y ajusta el cruce sin perder contexto."
    >
      {#snippet meta()}
        <Badge>{comb.length} entidades</Badge>
        <Badge variant="quiet">{currentRelationCount} relaciones activas</Badge>
        <Badge variant="quiet">{approvedRelationCount} relacionadas aprobadas</Badge>
      {/snippet}
      {#snippet actions()}
        <div class="view-jumps">
          <Button variant="ghost" disabled={!comb.length} icon="database" onclick={() => jumpToTab("entities")}>
            Ir a definicion
          </Button>
          <Button variant="accent" disabled={!comb.length} icon="attributes" onclick={() => jumpToTab("tertiary")}>
            Ir a atributos
          </Button>
        </div>
        <select
          class="entity-select"
          value={comb[activeIndex]?.IdPrincipalEntity ?? ""}
          onchange={handlePrincipalSelectChange}
          disabled={!comb.length}
        >
          {#each comb as view}
            <option value={view.IdPrincipalEntity}>{view.PrincipalEntity}</option>
          {/each}
        </select>
        <div class="entity-nav">
          <Button variant="soft" size="icon" icon="chevron-left" disabled={comb.length <= 1} onclick={prevSlide} aria-label="Entidad anterior" />
          <Button variant="soft" size="icon" icon="chevron-right" disabled={comb.length <= 1} onclick={nextSlide} aria-label="Entidad siguiente" />
        </div>
      {/snippet}
    </StudioToolbar>
  </StickyStack>

  {#if comb.length === 0}
    <EmptyPanel message="Sin datos de relaciones." />
  {:else}
    <div class="relations-layout">
      <aside class="relations-deck">
        {#if comb[activeIndex]}
          <EntityFocusCard
            kicker="Entidad principal"
            name={comb[activeIndex].PrincipalEntity}
            description={currentPrincipalEntity?.Description || "Sin definición."}
            approved={isApprovedEntity(comb[activeIndex].IdPrincipalEntity)}
            transitionName={`relation-head-${comb[activeIndex].IdPrincipalEntity}`}
          >
            {#snippet actions()}
              <div class="entity-focus-actions">
                <CreateEntity
                  id={comb[activeIndex].IdPrincipalEntity}
                  triggerLabel="Editar"
                  onSave={async () => {
                    await load(true);
                    await onRefresh();
                  }}
                />
                <Button
                  variant={isApprovedEntity(comb[activeIndex].IdPrincipalEntity) ? 'active' : 'success'}
                  icon={isApprovedEntity(comb[activeIndex].IdPrincipalEntity) ? "check-off" : "check"}
                  disabled={updating}
                  onclick={togglePrincipalApproval}
                >
                  {isApprovedEntity(comb[activeIndex].IdPrincipalEntity) ? "Quitar" : "Aprobar"}
                </Button>
              </div>
            {/snippet}
          </EntityFocusCard>
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
                <p class="muted">Configura cardinalidades.</p>
              </div>
              <span class="slide-panel__hint">Auto-guardado</span>
            </div>
            <RelationsTable
              relations={comb[activeIndex].Relations}
              {relationOptions}
              {updating}
              {relationMenu}
              onRelationChange={handleRelationChange}
              onContextMenu={openRelationMenu}
              {isApprovedEntity}
              {getEntityDefinition}
              {relationIdentifiers}
            />
          {/if}
        </article>
      </div>
    </div>
  {/if}

  {#if relationMenu.open}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      class="context-menu"
      style={`left: ${relationMenu.x}px; top: ${relationMenu.y}px;`}
      onclick={(e) => e.stopPropagation()}
    >
      <p class="context-title">{relationMenu.principalName} -> {relationMenu.targetName}</p>
      <Button class="menu-action control--block" variant="accent" size="sm" icon="attributes" onclick={goToRelatedEntityAttributes}>
        Ir a atributos
      </Button>
      <Button class="menu-action control--block" variant="ghost" size="sm" icon="relations" onclick={goToRelatedEntityRelations}>
        Ir a esta relacion
      </Button>
    </div>
  {/if}
</section>

<style>
  .relations-studio {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .relations-layout {
    display: grid;
    grid-template-columns: 320px 1fr;
    gap: 1.5rem;
    align-items: start;
  }

  .relations-deck {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    position: sticky;
    top: calc(var(--relations-sticky-total-height) + 1.5rem);
  }

  .slide-panel__head {
    padding: 1.5rem;
    border-bottom: 1px solid var(--border-card);
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: var(--background);
  }

  .slide-panel__hint {
    font-size: 0.65rem;
    font-weight: 800;
    text-transform: uppercase;
    color: var(--ink-soft);
    letter-spacing: 0.05em;
    opacity: 0.6;
  }

  .view-jumps {
    display: flex;
    gap: 0.5rem;
  }

  .entity-nav {
    display: flex;
    gap: 0.5rem;
  }

  .slide-shell {
    border: 1px solid var(--border-card);
    border-radius: var(--radius-lg);
    background: var(--surface);
    overflow: hidden;
    box-shadow: var(--shadow-sm);
  }

  .slide--approved {
    border-color: var(--success-soft);
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

  @keyframes menu-in {
    from { transform: scale(0.95); opacity: 0; }
    to { transform: scale(1); opacity: 1; }
  }

  @media (max-width: 1024px) {
    .relations-layout {
      grid-template-columns: 1fr;
    }
    .relations-deck {
      position: static;
    }
  }
</style>
