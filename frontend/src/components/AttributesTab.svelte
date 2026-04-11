<script lang="ts">
  import {onMount, tick} from "svelte";
  import type {utils} from "../../wailsjs/go/models";
  import {GetCombinatory} from "../../wailsjs/go/main/App";
  import {GetEntity, MarkEntityStatus, MoveAttribute, MoveIntersectionAttribute, Save} from "../../wailsjs/go/main/App";
  import ButtonIcon from "./ButtonIcon.svelte";
  import EntityFocusCard from "./EntityFocusCard.svelte";
  import ScopeSwitch from "./studio/ScopeSwitch.svelte";
  import StickyStack from "./studio/StickyStack.svelte";
  import CreateEntity from "./forms/CreateEntity.svelte";
  import AttributeForm from "./forms/AttributeForm.svelte";
  import AttributesTable from "./AttributesTable.svelte";
  import {showToast} from "../lib/toast";
  import {createVerticalAutoScroller, getErrorMessage, runViewTransition} from "../lib/ui-helpers";

  type RelationGroup = {
    type: string;
    label: string;
    items: string[];
  };

  export let project: utils.DbProject;
  export let onRefresh: () => Promise<void> = async () => {};
  export let focusEntityId: number | null = null;
  export let onJumpTo: (tab: "entities" | "relations" | "tertiary", entityId?: number | null) => void = () => {};

  let tableWrapper: HTMLDivElement | null = null;
  let stickyStackHeight = 0;
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
  let approvalUpdating = false;
  let relationSummaryCount = 0;
  let relationRules: utils.RelationRule[] = [];

  type InheritedPK = {
    entityName: string;
    attributeName: string | null;
    description: string | null;
    type: string | null;
    isIntersection?: boolean;
    isOptional?: boolean;
  };

  const getInheritedPKs = (entity: utils.Entity, project: utils.DbProject): InheritedPK[] => {
    if (!entity || !project || !project.Relations) return [];
    
    const inherited: InheritedPK[] = [];
    
    project.Relations.forEach(rel => {
      let otherEntityId: number | null = null;
      let isOptional = false;

      // Determinamos si se hereda a la Entidad 1
      if (rel.IdEntity1 === entity.Id) {
        // Tipos que heredan a la E1
        if (rel.Relation === "N:1" || rel.Relation === "Np:1") {
          otherEntityId = rel.IdEntity2;
          isOptional = false; // El lado "1" siempre es mandatorio para el N
        }
      } 
      // Determinamos si se hereda a la Entidad 2
      else if (rel.IdEntity2 === entity.Id) {
        if (rel.Relation === "1:N") {
          otherEntityId = rel.IdEntity1;
          isOptional = false;
        } else if (rel.Relation === "1:Np") {
          otherEntityId = rel.IdEntity1;
          isOptional = true;
        }
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
                type: pk.Type,
                isOptional: isOptional
              });
            });
          } else {
            inherited.push({
              entityName: otherEntity.Name,
              attributeName: null,
              description: null,
              type: null,
              isOptional: isOptional
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
    
    // En intersección (N:N), las PKs vienen de ambas entidades.
    // Analizamos la opcionalidad basándonos en el tipo de relación.
    // N:N -> Ambas mandatorias en la tabla de intersección (PK compuesta)
    // N:Np -> E1 es mandatoria, E2 es opcional? 
    // Realmente en una tabla de intersección pura, ambas suelen ser parte de la PK, por lo que son mandatorias.
    // Pero si el usuario quiere marcar opcionalidad en el modelo lógico:
    
    const isE1Optional = false;
    const isE2Optional = false;

    [
      { id: rel.IdEntity1, optional: isE1Optional },
      { id: rel.IdEntity2, optional: isE2Optional }
    ].forEach(target => {
      const otherEntity = project.Entities?.find(e => e.Id === target.id);
      if (otherEntity) {
        const pkAttributes = otherEntity.Attributes?.filter(a => a.KeyType === "pk") || [];
        if (pkAttributes.length > 0) {
          pkAttributes.forEach(pk => {
            inherited.push({
              entityName: otherEntity.Name,
              attributeName: pk.Name,
              description: pk.Description,
              type: pk.Type,
              isIntersection: true,
              isOptional: target.optional
            });
          });
        } else {
          inherited.push({
            entityName: otherEntity.Name,
            attributeName: null,
            description: null,
            type: null,
            isIntersection: true,
            isOptional: target.optional
          });
        }
      }
    });

    return inherited;
  };

  const autoScroller = createVerticalAutoScroller({
    edgePx: 72,
    stepPx: 14,
    getContainer: () => tableWrapper
  });
  const relationTypeLabels: Record<string, string> = {
    "1:1": "Uno a uno",
    "1:N": "Uno a muchos",
    "N:1": "Muchos a uno",
    "N:N": "Muchos a muchos",
    "1:Np": "Uno a muchos (Opcional)",
    "Np:1": "Muchos (Opcional) a uno"
  };
  const relationTypeOrder = ["1:1", "1:N", "N:1", "N:N", "1:Np", "Np:1"];

  const runAttributeTransition = (update: () => void | Promise<void>) =>
    runViewTransition(update, "No se pudo aplicar la transicion de atributos:");

  const updateAutoScroll = (event: DragEvent) => autoScroller.updateFromDragEvent(event);

  onMount(() => {
    if (entities.length && selectedId === null) {
      selectedId = entities[0].Id;
    }
    if (intersectionEntities.length && selectedIntersectionRelationId === null) {
      selectedIntersectionRelationId = intersectionEntities[0].RelationID;
    }

    return () => {
      autoScroller.stop();
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

  $: if (activeScope === "intersection" && selectedIntersectionRelationId !== null) {
    loadIntersectionRelationSummary(selectedIntersectionRelationId);
  } else if (activeScope === "strong" && current) {
    void loadRelationSummary(current);
  }

  $: relationSummaryCount = relationSummary.reduce((total, group) => total + group.items.length, 0);
  $: if (!relationSummary.length) {
    activeRelationType = null;
  } else if (!activeRelationType || !relationSummary.some((group) => group.type === activeRelationType)) {
    activeRelationType = relationSummary[0].type;
  }
  $: activeRelationGroup = relationSummary.find((group) => group.type === activeRelationType) ?? relationSummary[0] ?? null;

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
    } catch (err) {
      const message = getErrorMessage(err);
      showToast(`No se pudo cargar la entidad: ${message}`, "error");
    } finally {
      loading = false;
    }
  };

  const clearDrag = () => {
    autoScroller.stop();
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
    autoScroller.stop();
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
      const message = getErrorMessage(err);
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
      const message = getErrorMessage(err);
      console.error("No se pudo cargar resumen de relaciones:", message);
      relationSummary = [];
    }
  };

  const loadIntersectionRelationSummary = (relationId: number) => {
    if (!project?.Relations) {
      relationSummary = [];
      return;
    }
    const rel = project.Relations.find(r => r.Id === relationId);
    if (!rel) {
      relationSummary = [];
      return;
    }

    const left = entities.find(e => e.Id === rel.IdEntity1);
    const right = entities.find(e => e.Id === rel.IdEntity2);

    const items: string[] = [];
    if (left) items.push(left.Name);
    if (right) items.push(right.Name);

    if (items.length === 0) {
      relationSummary = [];
      return;
    }

    relationSummary = [{
      type: "Padres",
      label: "Entidades relacionadas",
      items: items
    }];
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
    const typeLabel = relationTypeLabels[relation.Relation] || relation.Relation;
    return `${left} --[${typeLabel}]-- ${right}`;
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
      const message = getErrorMessage(err);
      showToast(`No se pudo actualizar la aprobación: ${message}`, "error");
    } finally {
      approvalUpdating = false;
    }
  };

  const jumpToTab = (tab: "entities" | "relations" | "tertiary") => {
    onJumpTo(tab, activeScope === "strong" ? selectedId : null);
  };

  const jumpToEntity = (entityName: string) => {
    const target = entities.find(e => e.Name === entityName);
    if (target) {
      onJumpTo("tertiary", target.Id);
    }
  };

  const handleRelationTypeChange = (event: Event) => {
    const target = event.target as HTMLSelectElement;
    activeRelationType = target?.value || relationSummary[0]?.type || null;
  };

</script>

<section class="attributes-studio" style={`--attributes-sticky-total-height: ${stickyStackHeight}px;`}>
  <StickyStack bind:height={stickyStackHeight}>
    <div class="tab-toolbar attributes-toolbar">
      <div class="attributes-toolbar__copy">
        <p class="label">Atributos</p>
        <p class="muted">{activeScope === "strong" ? "Gestión de atributos y relaciones." : "Administra atributos de intersección."}</p>
      </div>
      <div class="attributes-toolbar__meta">
        <ScopeSwitch {activeScope} onSwitch={switchScope}/>
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
    {#if relationSummary.length}
      <div class="attributes-toolbar__relations">
        <div class="relation-bar">
          <div class="relation-bar__picker">
            <span class="banner-title">{activeScope === "strong" ? "Relaciones" : "Origen"}</span>
            {#if activeScope === "strong"}
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
            {:else}
              <span class="banner-subtitle">Tablas relacionadas</span>
            {/if}
          </div>
          {#if activeRelationGroup}
            <div class="relation-bar__tags" aria-label="Relaciones de la entidad activa">
              {#each activeRelationGroup.items as item}
                <div class="pill relation-pill" title={`${activeRelationGroup.type} ${item}`}>
                  <span class="pill-text">{item}</span>

                </div>
              {/each}
            </div>
          {:else}
            <div class="relation-bar__empty">Sin relaciones visibles.</div>
          {/if}
        </div>
        {#if activeScope === "strong" && relationSummary.length > 1}
          <div class="relation-type-summary">
            {#each relationSummary as group}
              <span class="relation-type-summary__item">{group.type}</span>
            {/each}
          </div>
        {/if}
      </div>
    {/if}
  </StickyStack>

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
          <AttributesTable
            entity={current}
            entityId={current.Id}
            attributes={current.Attributes || []}
            inheritedPKs={getInheritedPKs(current, project)}
            draggingIndex={draggingIndex}
            hoverIndex={hoverIndex}
            bind:tableRef={tableWrapper}
            onDragStart={startDrag}
            onDragOver={handleDragOver}
            onDrop={handleDrop}
            onDragEnd={clearDrag}
            onTableDragOver={handleTableDragOver}
            onTableDragLeave={handleTableDragLeave}
            onTableDrop={autoScroller.stop}
            onJumpToEntity={jumpToEntity}
            onRefresh={onRefresh}
            onEntityReload={async () => { await loadEntity(current.Id); }}
          />
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
          <AttributesTable
            isIntersection={true}
            currentIntersection={currentIntersection}
            intersectionAttributes={currentIntersection.Entity.Attributes || []}
            intersectionInheritedPKs={getIntersectionInheritedPKs(currentIntersection, project)}
            draggingIndex={draggingIndex}
            hoverIndex={hoverIndex}
            bind:tableRef={tableWrapper}
            onDragStart={startDrag}
            onDragOver={handleDragOver}
            onDrop={handleDrop}
            onDragEnd={clearDrag}
            onTableDragOver={handleTableDragOver}
            onTableDragLeave={handleTableDragLeave}
            onTableDrop={autoScroller.stop}
            onRefresh={onRefresh}
          />
        </section>
      </div>
    </section>
  {/if}
</section>
