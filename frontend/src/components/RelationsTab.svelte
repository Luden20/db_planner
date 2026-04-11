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
  import Table from "./ui/Table.svelte";
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

  export let entities: utils.Entity[] = [];
  export let onRefresh: () => Promise<void> = async () => {};
  export let focusEntityId: number | null = null;
  export let onJumpTo: (tab: "entities" | "relations" | "tertiary", entityId: number | null) => void = () => {};
  let comb: CombView[] = [];
  let activeIndex = 0;
  let lastSyncedFocusId: number | null = null;
  let currentPrincipalEntity: utils.Entity | null = null;
  let stickyStackHeight = 0;
  let currentRelationCount = 0;
  let approvedRelationCount = 0;
  let relationOptions: string[] = [];
  let loadingOptions = true;
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
    relationOverrides[key] = selection; // Guardamos incluso vacío para respetar la selección
    applyOverrides(); // Optimista: refleja en UI mientras persiste

    // Advertencia si se intenta cambiar una relación N:N que ya tiene atributos reales
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
              // Revertir cambio en UI
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
      const message = getErrorMessage(err);
      showToast(`No se pudo actualizar la relación: ${message}`, "error");
    } finally {
      updating = false;
    }
  };
</script>

<svelte:window
  on:click={closeRelationMenu}
  on:keydown={handleWindowKeydown}
  on:scroll={closeRelationMenu}
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
          on:change={handlePrincipalSelectChange}
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
          <div slot="actions" class="entity-focus-actions">
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
  <div
    class="context-menu"
    style={`left: ${relationMenu.x}px; top: ${relationMenu.y}px;`}
    on:click|stopPropagation
    on:keydown|stopPropagation
  >
    <p class="context-title">{relationMenu.principalName} -> {relationMenu.targetName}</p>
    <Button class="menu-action" variant="accent" size="sm" icon="attributes" onclick={goToRelatedEntityAttributes}>
      Ir a atributos
    </Button>
    <Button class="menu-action" variant="ghost" size="sm" icon="relations" onclick={goToRelatedEntityRelations}>
      Ir a esta relacion
    </Button>
  </div>
{/if}
</section>
