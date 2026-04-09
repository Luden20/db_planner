<script lang="ts">
  import {onMount, tick} from "svelte";
  import type {utils} from "../../wailsjs/go/models";
  import {GetCombinatory} from "../../wailsjs/go/main/App";
  import {GetEntity, MarkEntityStatus, MoveAttribute, MoveIntersectionAttribute, Save} from "../../wailsjs/go/main/App";
  import ButtonIcon from "./ButtonIcon.svelte";
  import EntityFocusCard from "./EntityFocusCard.svelte";
  import CreateEntity from "./forms/CreateEntity.svelte";
  import AttributeForm from "./forms/AttributeForm.svelte";
  import DeleteAttribute from "./forms/DeleteAttribute.svelte";
  import {showToast} from "../lib/toast";

  type RelationGroup = {
    type: string;
    label: string;
    items: string[];
  };
  type ViewTransitionDocument = Document & {
    startViewTransition?: (update: () => void | Promise<void>) => {
      finished: Promise<void>;
    };
  };

  export let project: utils.DbProject;
  export let onRefresh: () => Promise<void> = async () => {};
  export let focusEntityId: number | null = null;
  export let onJumpTo: (tab: "entities" | "relations" | "tertiary", entityId?: number | null) => void = () => {};

  let stickySentinel: HTMLDivElement | null = null;
  let stickyStack: HTMLDivElement | null = null;
  let tableWrapper: HTMLDivElement | null = null;
  let stickyStackHeight = 0;
  let stickyStackPinned = false;
  let activeScope: "strong" | "intersection" = "strong";
  let entities: utils.Entity[] = [];
  let intersectionEntities: utils.IntersectionEntity[] = [];
  let selectedId: number | null = null;
  let selectedIntersectionRelationId: number | null = null;
  let current: utils.Entity | null = null;
  let currentIntersection: utils.IntersectionEntity | null = null;
  let loading = false;
  let draggingIndex: number | null = null;
  let hoverIndex: number | null = null;
  let lastLoadedId: number | null = null;
  let lastSyncedFocusId: number | null = null;
  let relationSummary: RelationGroup[] = [];
  let activeRelationType: string | null = null;
  let activeRelationGroup: RelationGroup | null = null;
  let autoScrollFrame: number | null = null;
  let autoScrollDirection: -1 | 0 | 1 = 0;
  let approvalUpdating = false;
  let relationSummaryCount = 0;

  type InheritedPK = {
    entityName: string;
    attributeName: string | null;
    description: string | null;
    type: string | null;
    isIntersection?: boolean;
  };

  const getInheritedPKs = (entity: utils.Entity, project: utils.DbProject): InheritedPK[] => {
    if (!entity || !project || !project.Relations) return [];
    
    const inherited: InheritedPK[] = [];
    
    // Buscar relaciones donde esta entidad sea el lado "N" de una relación 1:N
    project.Relations.forEach(rel => {
      let otherEntityId: number | null = null;
      
      // En rel.Relation, "1:N" significa IdEntity1 es 1 y IdEntity2 es N
      // "N:1" significa IdEntity1 es N y IdEntity2 es 1
      if (rel.IdEntity1 === entity.Id && rel.Relation === "N:1") {
        otherEntityId = rel.IdEntity2;
      } else if (rel.IdEntity2 === entity.Id && rel.Relation === "1:N") {
        otherEntityId = rel.IdEntity1;
      }

      if (otherEntityId !== null) {
        const otherEntity = project.Entities?.find(e => e.Id === otherEntityId);
        if (otherEntity) {
          const pkAttributes = otherEntity.Attributes?.filter(a => a.KeyType === "pk") || [];
          if (pkAttributes.length > 0) {
            pkAttributes.forEach(pk => {
              inherited.push({
                entityName: otherEntity.Name,
                attributeName: pk.Name,
                description: pk.Description,
                type: pk.Type
              });
            });
          } else {
            inherited.push({
              entityName: otherEntity.Name,
              attributeName: null,
              description: null,
              type: null
            });
          }
        }
      }
    });

    return inherited;
  };

  const getIntersectionInheritedPKs = (intersection: utils.IntersectionEntity, project: utils.DbProject): InheritedPK[] => {
    if (!intersection || !project || !project.Relations) return [];
    
    const rel = project.Relations.find(r => r.Id === intersection.RelationID);
    if (!rel) return [];

    const inherited: InheritedPK[] = [];
    [rel.IdEntity1, rel.IdEntity2].forEach(id => {
      const otherEntity = project.Entities?.find(e => e.Id === id);
      if (otherEntity) {
        const pkAttributes = otherEntity.Attributes?.filter(a => a.KeyType === "pk") || [];
        if (pkAttributes.length > 0) {
          pkAttributes.forEach(pk => {
            inherited.push({
              entityName: otherEntity.Name,
              attributeName: pk.Name,
              description: pk.Description,
              type: pk.Type,
              isIntersection: true
            });
          });
        } else {
          inherited.push({
            entityName: otherEntity.Name,
            attributeName: null,
            description: null,
            type: null,
            isIntersection: true
          });
        }
      }
    });

    return inherited;
  };

  const AUTO_SCROLL_EDGE_PX = 72;
  const AUTO_SCROLL_STEP = 14;
  const relationTypeLabels: Record<string, string> = {
    "1:1": "Uno a uno",
    "1:N": "Uno a muchos",
    "N:1": "Muchos a uno",
    "N:N": "Muchos a muchos"
  };
  const relationTypeOrder = ["1:1", "1:N", "N:1", "N:N"];

  const prefersReducedMotion = () =>
    typeof window !== "undefined"
    && typeof window.matchMedia === "function"
    && window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  const runAttributeTransition = async (update: () => void | Promise<void>) => {
    const doc = typeof document !== "undefined" ? (document as ViewTransitionDocument) : null;
    if (doc?.startViewTransition && !prefersReducedMotion()) {
      try {
        const transition = doc.startViewTransition(update);
        await transition.finished;
        return;
      } catch (err) {
        console.warn("No se pudo aplicar la transicion de atributos:", err);
      }
    }
    await update();
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
    if (entities.length && selectedId === null) {
      selectedId = entities[0].Id;
    }
    if (intersectionEntities.length && selectedIntersectionRelationId === null) {
      selectedIntersectionRelationId = intersectionEntities[0].RelationID;
    }

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

  $: entities = project?.Entities ?? [];
  $: intersectionEntities = project?.IntersectionEntities ?? [];
  $: if (entities.length && selectedId === null) {
    selectedId = entities[0].Id;
  }
  $: if (!entities.some((entity) => entity.Id === selectedId)) {
    selectedId = entities[0]?.Id ?? null;
  }
  $: if (intersectionEntities.length && selectedIntersectionRelationId === null) {
    selectedIntersectionRelationId = intersectionEntities[0].RelationID;
  }
  $: if (!intersectionEntities.some((item) => item.RelationID === selectedIntersectionRelationId)) {
    selectedIntersectionRelationId = intersectionEntities[0]?.RelationID ?? null;
  }
  $: currentIntersection = intersectionEntities.find((item) => item.RelationID === selectedIntersectionRelationId) ?? null;

  $: if (
    activeScope === "strong" &&
    focusEntityId !== null &&
    focusEntityId !== lastSyncedFocusId &&
    entities.some(entity => entity.Id === focusEntityId)
  ) {
    void selectEntity(focusEntityId);
    lastSyncedFocusId = focusEntityId;
  }

  $: if (selectedId !== null && selectedId !== lastLoadedId) {
    loadEntity(selectedId);
  }

  $: relationSummaryCount = relationSummary.reduce((total, group) => total + group.items.length, 0);
  $: if (!relationSummary.length) {
    activeRelationType = null;
  } else if (!activeRelationType || !relationSummary.some((group) => group.type === activeRelationType)) {
    activeRelationType = relationSummary[0].type;
  }
  $: activeRelationGroup = relationSummary.find((group) => group.type === activeRelationType) ?? relationSummary[0] ?? null;
  $: if (stickySentinel) {
    syncStickyState();
  }
  $: if (stickyStack) {
    syncStickyStackHeight();
  }

  const loadEntity = async (id: number) => {
    loading = true;
    try {
      current = await GetEntity(id);
      if (current && current.Attributes) {
        // Sort PK to top
        current.Attributes.sort((a, b) => {
          if (a.KeyType === "pk") return -1;
          if (b.KeyType === "pk") return 1;
          return 0;
        });
      }
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

  const selectEntity = async (entityId: number | null) => {
    if (entityId === null || !entities.some(entity => entity.Id === entityId)) {
      return;
    }
    await runAttributeTransition(async () => {
      selectedId = entityId;
      await tick();
    });
  };

  const selectIntersection = async (relationId: number | null) => {
    if (relationId === null || !intersectionEntities.some((item) => item.RelationID === relationId)) {
      return;
    }
    await runAttributeTransition(async () => {
      selectedIntersectionRelationId = relationId;
      await tick();
    });
  };

  const handleSelectChange = async (event: Event) => {
    const target = event.target as HTMLSelectElement;
    if (activeScope === "strong") {
      await selectEntity(Number(target?.value ?? 0));
      return;
    }
    await selectIntersection(Number(target?.value ?? 0));
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

  const nextEntity = async () => {
    if (activeScope === "strong") {
      if (!entities.length) return;
      const currentIndex = entities.findIndex((ent) => ent.Id === selectedId);
      const nextIndex = currentIndex === -1 || currentIndex === entities.length - 1 ? 0 : currentIndex + 1;
      await selectEntity(entities[nextIndex].Id);
      return;
    }
    if (!intersectionEntities.length) return;
    const currentIndex = intersectionEntities.findIndex((item) => item.RelationID === selectedIntersectionRelationId);
    const nextIndex = currentIndex === -1 || currentIndex === intersectionEntities.length - 1 ? 0 : currentIndex + 1;
    await selectIntersection(intersectionEntities[nextIndex].RelationID);
  };

  const prevEntity = async () => {
    if (activeScope === "strong") {
      if (!entities.length) return;
      const currentIndex = entities.findIndex((ent) => ent.Id === selectedId);
      const prevIndex = currentIndex <= 0 ? entities.length - 1 : currentIndex - 1;
      await selectEntity(entities[prevIndex].Id);
      return;
    }
    if (!intersectionEntities.length) return;
    const currentIndex = intersectionEntities.findIndex((item) => item.RelationID === selectedIntersectionRelationId);
    const prevIndex = currentIndex <= 0 ? intersectionEntities.length - 1 : currentIndex - 1;
    await selectIntersection(intersectionEntities[prevIndex].RelationID);
  };

  const applyReorder = async (from: number, to: number) => {
    const attributes = activeScope === "strong"
      ? (current?.Attributes ?? [])
      : (currentIntersection?.Entity.Attributes ?? []);
    if (from === to || from < 0 || to < 0 || from >= attributes.length || to >= attributes.length) {
      return;
    }
    const direction: "up" | "down" = to < from ? "up" : "down";
    const steps = Math.abs(to - from);
    const attributeId = attributes[from].Id;
    try {
      for (let i = 0; i < steps; i++) {
        if (activeScope === "strong" && current) {
          await MoveAttribute(current.Id, attributeId, direction);
        } else if (activeScope === "intersection" && currentIntersection) {
          await MoveIntersectionAttribute(currentIntersection.RelationID, attributeId, direction);
        }
      }
      await Save();
      await onRefresh();
      if (activeScope === "strong" && current) {
        await loadEntity(current.Id);
      }
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
  const attributeKeyLabel = (attribute: utils.Attribute) => attribute.KeyType === "pk" ? "PK" : "";
  const intersectionOriginLabel = (item: utils.IntersectionEntity | null) => {
    if (!item) {
      return "Sin relación origen";
    }
    const relation = project?.Relations?.find((currentRelation) => currentRelation.Id === item.RelationID);
    if (!relation) {
      return "Sin relación origen";
    }
    const left = entities.find((entity) => entity.Id === relation.IdEntity1)?.Name ?? `Tabla ${relation.IdEntity1}`;
    const right = entities.find((entity) => entity.Id === relation.IdEntity2)?.Name ?? `Tabla ${relation.IdEntity2}`;
    return `${left} <-> ${right}`;
  };
  const switchScope = (scope: "strong" | "intersection") => {
    activeScope = scope;
    clearDrag();
  };

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
    onJumpTo(tab, activeScope === "strong" ? selectedId : null);
  };

  const handleRelationTypeChange = (event: Event) => {
    const target = event.target as HTMLSelectElement;
    activeRelationType = target?.value || relationSummary[0]?.type || null;
  };

</script>

<svelte:window on:scroll={syncStickyState} on:resize={syncStickyState}/>

<section class="attributes-studio" style={`--attributes-sticky-total-height: ${stickyStackHeight}px;`}>
  <div class="attributes-sticky-sentinel" bind:this={stickySentinel} aria-hidden="true"></div>
  <div
    class:attributes-sticky-stack={true}
    class:attributes-sticky-stack--pinned={stickyStackPinned}
    bind:this={stickyStack}
  >
    <div class="tab-toolbar attributes-toolbar">
      <div class="attributes-toolbar__copy">
        <p class="label">Atributos</p>
        <p class="muted">{activeScope === "strong" ? "Recorre la entidad activa, reordena atributos y mantén visible su mapa relacional." : "Administra atributos de las entidades de intersección sin definir PK manualmente."}</p>
      </div>
      <div class="attributes-toolbar__meta">
        <div class="scope-switch" role="tablist" aria-label="Tipo de atributos">
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
        <span class="studio-chip">{activeScope === "strong" ? entities.length : intersectionEntities.length} entidades</span>
        <span class="studio-chip studio-chip--quiet">{activeScope === "strong" ? (current?.Attributes?.length ?? 0) : (currentIntersection?.Entity.Attributes?.length ?? 0)} atributos</span>
        {#if activeScope === "strong"}
          <span class="studio-chip studio-chip--quiet">{relationSummaryCount} cruces</span>
        {/if}
      </div>
      <div class="toolbar-actions attributes-toolbar__actions">
        <div class="view-jumps">
          <button class="control control--ghost" on:click={() => jumpToTab("entities")} disabled={activeScope !== "strong" || !selectedId}>
            <ButtonIcon name="database"/>
            <span>Ir a definicion</span>
          </button>
          <button class="control control--accent" on:click={() => jumpToTab("relations")} disabled={activeScope !== "strong" || !selectedId}>
            <ButtonIcon name="relations"/>
            <span>Ir a combinatorio</span>
          </button>
        </div>
        <AttributeForm
          entityId={activeScope === "strong" ? (selectedId ?? (entities[0]?.Id ?? null)) : null}
          relationId={activeScope === "intersection" ? (selectedIntersectionRelationId ?? (intersectionEntities[0]?.RelationID ?? null)) : null}
          entity={activeScope === "strong" ? current : currentIntersection?.Entity ?? null}
          allowPrimaryKey={activeScope === "strong"}
          triggerClass="attributes-toolbar-trigger"
          onSaved={async () => {
            await onRefresh();
            if (activeScope === "strong" && selectedId !== null) {
              await loadEntity(selectedId);
            }
          }}
        />
        <div class="entity-switcher">
          <div class="entity-picker">
            <select
              class="entity-select"
              on:change={handleSelectChange}
              disabled={activeScope === "strong" ? !entities.length : !intersectionEntities.length}
            >
              {#if activeScope === "strong"}
                {#each entities as entity}
                  <option value={entity.Id} selected={entity.Id === selectedId}>{entity.Name}</option>
                {/each}
              {:else}
                {#each intersectionEntities as item}
                  <option value={item.RelationID} selected={item.RelationID === selectedIntersectionRelationId}>{item.Entity.Name}</option>
                {/each}
              {/if}
            </select>
            <div class="entity-nav">
              <button class="control control--icon control--soft" on:click={prevEntity} aria-label="Entidad anterior" disabled={activeScope === "strong" ? entities.length <= 1 : intersectionEntities.length <= 1}>
                <ButtonIcon name="chevron-left"/>
              </button>
              <button class="control control--icon control--soft" on:click={nextEntity} aria-label="Entidad siguiente" disabled={activeScope === "strong" ? entities.length <= 1 : intersectionEntities.length <= 1}>
                <ButtonIcon name="chevron-right"/>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
    {#if activeScope === "strong" && relationSummary.length}
      <div class="attributes-toolbar__relations">
        <div class="relation-bar">
          <div class="relation-bar__picker">
            <span class="banner-title">Relaciones</span>
            <select
              id="attribute-relation-type"
              class="entity-select relation-type-select"
              value={activeRelationGroup?.type ?? ""}
              on:change={handleRelationTypeChange}
            >
              {#each relationSummary as group}
                <option value={group.type}>{group.type}</option>
              {/each}
            </select>
          </div>
          {#if activeRelationGroup}
            <div class="relation-bar__tags" aria-label="Relaciones de la entidad activa">
              {#each activeRelationGroup.items as item}
                <span class="pill relation-pill" title={`${activeRelationGroup.type} ${item}`}>{item}</span>
              {/each}
            </div>
          {:else}
            <div class="relation-bar__empty">Sin relaciones visibles.</div>
          {/if}
        </div>
        {#if relationSummary.length > 1}
          <div class="relation-type-summary">
            {#each relationSummary as group}
              <span class="relation-type-summary__item">{group.type}</span>
            {/each}
          </div>
        {/if}
      </div>
    {/if}
  </div>

  {#if activeScope === "strong" && !entities.length}
    <div class="empty-panel">Crea entidades para gestionar atributos.</div>
  {:else if activeScope === "intersection" && !intersectionEntities.length}
    <div class="empty-panel">No hay entidades de intersección todavía. Se crearán cuando exista una relación N:N.</div>
  {:else if activeScope === "strong" && loading}
    <div class="empty-panel">Cargando atributos...</div>
  {:else if activeScope === "strong" && current}
    <section class="attributes-stage">
      <div class="attributes-layout">
        <aside class="attributes-deck">
          <div class="attributes-focus-card">
            <EntityFocusCard
              kicker="Entidad seleccionada"
              name={current.Name}
              description={current.Description || "Sin definición."}
              approved={isApproved(current)}
              transitionName={`attribute-entity-${current.Id}`}
            >
              <div slot="actions" class="entity-focus-actions">
                <CreateEntity
                  id={current.Id}
                  triggerLabel="Editar"
                  onSave={async () => {
                    await onRefresh();
                    await loadEntity(current.Id);
                  }}
                />
                <button
                  class={`control control--success ${isApproved(current) ? 'control--active' : ''}`}
                  on:click={toggleCurrentApproval}
                  disabled={approvalUpdating}
                >
                  <ButtonIcon name={isApproved(current) ? "check-off" : "check"}/>
                  <span>{isApproved(current) ? "Quitar" : "Aprobar"}</span>
                </button>
              </div>
            </EntityFocusCard>
          </div>
        </aside>

        <section class="attributes-panel" style={`view-transition-name: attribute-table-${current.Id};`}>
          <div class="attributes-panel__head">
            <div>
              <p class="label">Inventario de atributos</p>
              <p class="muted">Arrastra filas para cambiar el orden natural de la definición.</p>
            </div>
          </div>
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
                <th style="width: 72px;">Clave</th>
                <th style="width: 180px;">Acciones</th>
              </tr>
              </thead>
              <tbody class="draggable-body">
              {#if !current.Attributes?.some(a => a.KeyType === "pk")}
                {#each getInheritedPKs(current, project) as inherited}
                  <tr class="inherited-pk-row" draggable="false">
                    <td class="inherited-name">
                      <span class="inherited-tag">FK heredada</span>
                      {inherited.attributeName || `PK de ${inherited.entityName} pendiente por definir`}
                    </td>
                    <td>
                      {#if inherited.attributeName}
                        <span class="muted italic">{inherited.description || "Sin descripción."}</span>
                      {:else}
                        <span class="muted italic">Atributo heredado por relación con {inherited.entityName}</span>
                      {/if}
                    </td>
                    <td class="muted">{inherited.type || "—"}</td>
                    <td><span class="badge badge--fk">FK (H)</span></td>
                    <td class="muted">Lectura</td>
                  </tr>
                {/each}
              {/if}

              {#if (!current.Attributes || current.Attributes.length === 0) && getInheritedPKs(current, project).length === 0}
                <tr class="empty-row" draggable="false">
                  <td colspan="5">No hay atributos definidos aún.</td>
                </tr>
              {:else}
                {#each (current.Attributes || []) as attribute, index (attribute.Id)}
                  <tr
                    class:dragging={draggingIndex === index}
                    class:drag-hover={hoverIndex === index && draggingIndex !== null && draggingIndex !== index}
                    class:pk-row={attribute.KeyType === "pk"}
                    draggable={attribute.KeyType !== "pk"}
                    style={`view-transition-name: attribute-row-${attribute.Id};`}
                    on:dragstart={(event) => attribute.KeyType !== "pk" && startDrag(index, event)}
                    on:dragover={(event) => attribute.KeyType !== "pk" && handleDragOver(index, event)}
                    on:dragenter={(event) => attribute.KeyType !== "pk" && handleDragOver(index, event)}
                    on:drop={(event) => attribute.KeyType !== "pk" && handleDrop(index, event)}
                    on:dragend={clearDrag}
                  >
                    <td class:inherited-name={attribute.KeyType === "pk"}>
                      {#if attribute.KeyType === "pk"}
                        <span class="pk-tag">PK</span>
                      {/if}
                      {attribute.Name}
                    </td>
                    <td>{attribute.Description}</td>
                    <td>{attribute.Type || "Por definir"}</td>
                    <td>
                      {#if attribute.KeyType === "pk"}
                        <span class="badge badge--pk">PK</span>
                      {:else}
                        {attributeKeyLabel(attribute)}
                      {/if}
                    </td>
                    <td>
                      <div class="row-actions">
                        <AttributeForm
                          entityId={current.Id}
                          entity={current}
                          attribute={attribute}
                          allowPrimaryKey={true}
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

                  {#if attribute.KeyType === "pk"}
                    {#each getInheritedPKs(current, project) as inherited}
                      <tr class="inherited-pk-row" draggable="false">
                        <td class="inherited-name">
                          <span class="inherited-tag">FK heredada</span>
                          {inherited.attributeName || `PK de ${inherited.entityName} pendiente por definir`}
                        </td>
                        <td>
                          {#if inherited.attributeName}
                            <span class="muted italic">{inherited.description || "Sin descripción."}</span>
                          {:else}
                            <span class="muted italic">Atributo heredado por relación con {inherited.entityName}</span>
                          {/if}
                        </td>
                        <td class="muted">{inherited.type || "—"}</td>
                        <td><span class="badge badge--fk">FK (H)</span></td>
                        <td class="muted">Lectura</td>
                      </tr>
                    {/each}
                  {/if}
                {/each}
              {/if}
              </tbody>
            </table>
          </div>
        </section>
      </div>
    </section>
  {:else if activeScope === "intersection" && currentIntersection}
    <section class="attributes-stage">
      <div class="attributes-layout">
        <aside class="attributes-deck">
          <div class="attributes-focus-card">
            <EntityFocusCard
              kicker="Entidad de intersección"
              name={currentIntersection.Entity.Name}
              description={currentIntersection.Entity.Description || "Sin definición."}
              approved={false}
              transitionName={`attribute-entity-intersection-${currentIntersection.RelationID}`}
            >
              <div slot="actions" class="entity-focus-actions">
                <span class="intersection-origin">{intersectionOriginLabel(currentIntersection)}</span>
              </div>
            </EntityFocusCard>
          </div>
        </aside>

        <section class="attributes-panel" style={`view-transition-name: attribute-table-intersection-${currentIntersection.RelationID};`}>
          <div class="attributes-panel__head">
            <div>
              <p class="label">Inventario de atributos</p>
              <p class="muted">Las intersecciones permiten agregar atributos, pero no marcar PK por ahora.</p>
            </div>
          </div>
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
                <th style="width: 72px;">Clave</th>
                <th style="width: 180px;">Acciones</th>
              </tr>
              </thead>
              <tbody class="draggable-body">
              {#each getIntersectionInheritedPKs(currentIntersection, project) as inherited}
                <tr class="inherited-pk-row intersection-pk" draggable="false">
                  <td class="inherited-name">
                    <span class="inherited-tag">PK heredada</span>
                    {inherited.attributeName || `PK de ${inherited.entityName} pendiente por definir`}
                  </td>
                  <td>
                    {#if inherited.attributeName}
                      <span class="muted italic">{inherited.description || "Sin descripción."}</span>
                    {:else}
                      <span class="muted italic">Atributo heredado por relación con {inherited.entityName}</span>
                    {/if}
                  </td>
                  <td class="muted">{inherited.type || "—"}</td>
                  <td><span class="badge badge--pk">PK (H)</span></td>
                  <td class="muted">Lectura</td>
                </tr>
              {/each}

              {#if !currentIntersection.Entity.Attributes || currentIntersection.Entity.Attributes.length === 0}
                {#if getIntersectionInheritedPKs(currentIntersection, project).length === 0}
                  <tr class="empty-row" draggable="false">
                    <td colspan="5">No hay atributos definidos aún.</td>
                  </tr>
                {/if}
              {:else}
                {#each currentIntersection.Entity.Attributes as attribute, index (attribute.Id)}
                  <tr
                    class:dragging={draggingIndex === index}
                    class:drag-hover={hoverIndex === index && draggingIndex !== null && draggingIndex !== index}
                    draggable="true"
                    style={`view-transition-name: intersection-attribute-row-${attribute.Id};`}
                    on:dragstart={(event) => startDrag(index, event)}
                    on:dragover={(event) => handleDragOver(index, event)}
                    on:dragenter={(event) => handleDragOver(index, event)}
                    on:drop={(event) => handleDrop(index, event)}
                    on:dragend={clearDrag}
                  >
                    <td>{attribute.Name}</td>
                    <td>{attribute.Description}</td>
                    <td>{attribute.Type || "Por definir"}</td>
                    <td></td>
                    <td>
                      <div class="row-actions">
                        <AttributeForm
                          relationId={currentIntersection.RelationID}
                          entity={currentIntersection.Entity}
                          attribute={attribute}
                          allowPrimaryKey={false}
                          onSaved={async () => {
                            await onRefresh();
                          }}
                        />
                        <DeleteAttribute
                          relationId={currentIntersection.RelationID}
                          attributeId={attribute.Id}
                          onSaved={async () => {
                            await onRefresh();
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
        </section>
      </div>
    </section>
  {/if}
</section>

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

  .banner-title {
    margin: 0;
    font-size: 0.74rem;
    letter-spacing: 0.16em;
    color: var(--accent);
    text-transform: uppercase;
    font-weight: 800;
  }

  .relation-count-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-height: 2rem;
    padding: 0.38rem 0.78rem;
    border-radius: 999px;
    border: 1px solid color-mix(in srgb, var(--accent) 16%, var(--border));
    background: color-mix(in srgb, var(--accent) 10%, var(--surface-strong));
    color: var(--accent-strong);
    font-size: 0.78rem;
    font-weight: 800;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .attributes-toolbar__relations {
    display: grid;
    gap: 0.65rem;
    padding: 0.95rem 1.1rem 1rem;
    border: 1px solid var(--border);
    border-radius: calc(var(--radius-md) - 4px);
    background:
      linear-gradient(180deg, color-mix(in srgb, var(--surface-strong) 96%, var(--surface)), color-mix(in srgb, var(--surface) 98%, var(--surface-strong))),
      linear-gradient(90deg, color-mix(in srgb, var(--accent) 6%, var(--surface-strong)), transparent 38%);
    box-shadow: var(--shadow-sm);
  }

  .relation-bar {
    display: flex;
    align-items: flex-start;
    gap: 0.85rem;
    padding: 0.72rem 0.82rem;
    border-radius: 1rem;
    border: 1px solid var(--line-soft);
    background: color-mix(in srgb, var(--surface-strong) 74%, transparent);
  }

  .relation-bar__picker {
    display: grid;
    gap: 0.38rem;
    min-width: 0;
    flex: 0 0 auto;
    align-self: flex-start;
  }

  .relation-bar__tags {
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem;
    min-width: 0;
    flex: 1 1 auto;
  }

  .relation-type-select {
    min-width: 5.5rem;
    width: 5.5rem;
    padding-right: 2.4rem;
  }

  .relation-type-select:focus {
    outline: none;
  }

  .relation-bar__empty {
    display: inline-flex;
    align-items: center;
    min-height: 2.5rem;
    color: var(--ink-faint);
    font-size: 0.82rem;
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

  .relation-pill {
    max-width: none;
    background: color-mix(in srgb, var(--surface-strong) 92%, transparent);
  }

  .relation-type-summary {
    display: flex;
    flex-wrap: wrap;
    gap: 0.55rem;
  }

  .relation-type-summary__item {
    color: var(--ink-faint);
    font-size: 0.75rem;
    font-weight: 700;
  }
  .entities-table {
    width: 100%;
    border-collapse: collapse;
    color: #e8edf7;
  }

  .inherited-pk-row {
    background-color: rgba(59, 130, 246, 0.05);
    border-left: 3px solid rgba(59, 130, 246, 0.5);
  }

  .pk-row {
    background-color: rgba(16, 185, 129, 0.05);
    border-left: 3px solid rgba(16, 185, 129, 0.5);
  }

  .inherited-pk-row.intersection-pk {
    background-color: rgba(139, 92, 246, 0.05);
    border-left-color: rgba(139, 92, 246, 0.5);
  }

  .inherited-name {
    font-weight: 500;
  }

  .inherited-tag {
    display: inline-block;
    font-size: 0.65rem;
    text-transform: uppercase;
    background: rgba(59, 130, 246, 0.15);
    color: #60a5fa;
    padding: 1px 5px;
    border-radius: 4px;
    font-weight: 700;
    margin-right: 8px;
    vertical-align: middle;
  }

  .intersection-pk .inherited-tag {
    background: rgba(139, 92, 246, 0.15);
    color: #a78bfa;
  }

  .pk-tag {
    display: inline-block;
    font-size: 0.65rem;
    text-transform: uppercase;
    background: rgba(16, 185, 129, 0.15);
    color: #34d399;
    padding: 1px 5px;
    border-radius: 4px;
    font-weight: 700;
    margin-right: 8px;
    vertical-align: middle;
  }

  :global(.badge--fk) {
    background: #eab308;
    color: #422006;
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 600;
  }

  :global(.badge--pk) {
    background: #10b981;
    color: #064e3b;
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 600;
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

  .entity-switcher {
    margin-left: auto;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 0.7rem;
    flex-wrap: wrap;
  }

  .entity-picker {
    display: inline-flex;
    align-items: center;
    gap: 0.55rem;
    flex-wrap: wrap;
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

  .intersection-origin {
    display: inline-flex;
    align-items: center;
    min-height: 2rem;
    padding: 0.38rem 0.72rem;
    border-radius: 999px;
    border: 1px solid var(--line-soft);
    background: color-mix(in srgb, var(--surface-strong) 82%, transparent);
    color: var(--ink-soft);
    font-size: 0.78rem;
    font-weight: 700;
  }

  @media (max-width: 720px) {
    .entity-switcher {
      margin-left: 0;
      justify-content: flex-start;
    }

    .entity-picker {
      width: 100%;
    }

    .relation-type-select {
      min-width: 0;
      width: 100%;
    }

    .relation-bar {
      flex-direction: column;
    }

    .relation-bar__picker {
      min-width: 0;
      flex-basis: auto;
    }
  }

  .tab-toolbar {
    margin-bottom: 1rem;
    padding: 1.05rem 1.1rem;
    border: 1px solid var(--border);
    border-radius: calc(var(--radius-md) - 4px);
    background:
      linear-gradient(180deg, color-mix(in srgb, var(--surface-strong) 98%, var(--surface)), color-mix(in srgb, var(--surface) 100%, var(--surface-strong))),
      linear-gradient(90deg, color-mix(in srgb, var(--accent) 8%, var(--surface-strong)), transparent 38%);
    box-shadow: var(--shadow-sm);
    backdrop-filter: blur(18px);
  }

  .label,
  .banner-title {
    color: var(--accent);
    font-size: 0.74rem;
    letter-spacing: 0.16em;
    font-weight: 800;
  }

  .muted,
  .entity-description,
  .helper,
  .empty-panel {
    color: var(--ink-faint);
    opacity: 1;
  }

  .entity-select {
    border-color: var(--border);
    background: var(--field-surface);
    color: var(--ink);
    box-shadow: none;
  }

  .entity-select:focus {
    border-color: var(--focus-border);
    box-shadow: var(--focus-ring);
  }

  .entities-table {
    color: var(--ink);
  }

  .pill {
    background: var(--chip-surface);
    border-color: var(--line-soft);
    color: var(--ink-soft);
  }

  .table-wrapper.frosted {
    background: var(--panel-surface-strong);
    border-color: var(--border);
    border-radius: calc(var(--radius-md) - 4px);
    box-shadow: var(--surface-highlight);
  }

  .entities-table thead th {
    color: var(--ink-faint);
    border-bottom-color: var(--line-soft);
    font-size: 0.76rem;
    letter-spacing: 0.14em;
    text-transform: uppercase;
  }

  .entities-table tbody tr:nth-child(odd):not(.empty-row),
  .entities-table tbody tr:nth-child(even):not(.empty-row) {
    background: transparent;
  }

  .entities-table tbody tr:hover:not(.empty-row) {
    background: var(--hover-soft);
  }

  .empty-panel {
    border-color: var(--line-soft);
    background: color-mix(in srgb, var(--surface) 78%, transparent);
  }

  .attributes-studio {
    --attributes-sticky-total-height: 0px;
    display: grid;
    gap: 1rem;
  }

  .attributes-sticky-stack {
    margin-bottom: 18px;
  }

  .attributes-sticky-stack--pinned {
    position: sticky;
    top: 0;
    z-index: calc(var(--layer-ribbon) - 2);
  }

  .attributes-sticky-sentinel {
    height: 1px;
    margin-top: -1px;
  }

  .attributes-layout {
    display: grid;
    grid-template-columns: minmax(18rem, 24rem) minmax(0, 1fr);
    align-items: start;
    gap: 1rem;
  }

  .attributes-deck {
    display: grid;
    gap: 1rem;
    position: sticky;
    top: calc(var(--attributes-sticky-total-height) + 1rem);
    align-self: start;
  }

  .attributes-toolbar,
  .attributes-stage,
  .attributes-panel {
    position: relative;
    overflow: clip;
  }

  .attributes-toolbar::before,
  .attributes-stage::before,
  .attributes-panel::before {
    content: "";
    position: absolute;
    inset: 0 auto auto 0;
    width: min(220px, 42%);
    height: 1px;
    background: linear-gradient(90deg, color-mix(in srgb, var(--accent) 34%, transparent), transparent);
    pointer-events: none;
  }

  .attributes-toolbar {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: start;
    gap: 0.9rem 1rem;
    margin-bottom: 0;
  }

  .attributes-toolbar__copy {
    max-width: 38rem;
  }

  .attributes-toolbar__meta {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 0.65rem;
    flex-wrap: wrap;
  }

  .attributes-toolbar__actions {
    grid-column: 1 / -1;
  }

  :global(.attributes-toolbar-trigger) {
    min-height: 2.85rem;
    padding: 0.72rem 1rem;
    border-radius: 1rem;
    font-size: 0.92rem;
    box-shadow: 0 12px 22px color-mix(in srgb, var(--ink) 10%, transparent), inset 0 1px 0 color-mix(in srgb, white 30%, transparent);
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

  .attributes-stage {
    display: grid;
    gap: 1rem;
  }

  .attributes-panel {
    padding: 1rem;
    border: 1px solid var(--border);
    border-radius: calc(var(--radius-lg) - 2px);
    background:
      radial-gradient(circle at top right, color-mix(in srgb, var(--accent) 8%, transparent), transparent 34%),
      var(--panel-surface);
    box-shadow: var(--shadow-sm);
  }

  .attributes-panel__head {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 1rem;
    margin-bottom: 0.9rem;
  }

  @media (max-width: 720px) {
    .attributes-layout {
      grid-template-columns: 1fr;
    }

    .attributes-focus-card {
      position: static;
    }

    .attributes-toolbar,
    .attributes-panel__head {
      grid-template-columns: 1fr;
      align-items: stretch;
    }

    .attributes-toolbar__meta {
      justify-content: flex-start;
    }

    .entity-select {
      min-width: 0;
      flex: 1 1 14rem;
    }

    .attributes-panel {
      padding: 0.9rem;
    }
    .studio-chip {
      white-space: normal;
    }
  }

</style>
