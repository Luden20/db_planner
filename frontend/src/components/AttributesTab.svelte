<script lang="ts">
  import {onMount, tick} from "svelte";
  import type {utils} from "../../wailsjs/go/models";
  import {GetCombinatory} from "../../wailsjs/go/main/App";
  import {GetEntity, MarkEntityStatus, MoveAttribute, MoveIntersectionAttribute, Save} from "../../wailsjs/go/main/App";
  import Button from "./ui/Button.svelte";
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

  let { 
    project, 
    onRefresh = async () => {}, 
    focusEntityId = null, 
    onJumpTo = () => {} 
  } = $props<{
    project: utils.DbProject;
    onRefresh?: () => Promise<void>;
    focusEntityId?: number | null;
    onJumpTo?: (tab: "entities" | "relations" | "tertiary", entityId?: number | null) => void;
  }>();

  let tableWrapper = $state<HTMLDivElement | null>(null);
  let stickyStackHeight = $state(0);
  let activeScope = $state<"strong" | "intersection">("strong");
  let selectedId = $state<number | null>(null);
  let selectedIntersectionRelationId = $state<number | null>(null);
  let current = $state<utils.Entity | null>(null);
  let currentIntersection = $state<utils.IntersectionEntity | null>(null);
  let loading = $state(false);
  let draggingIndex = $state<number | null>(null);
  let hoverIndex = $state<number | null>(null);
  let lastLoadedId = $state<number | null>(null);
  let lastSyncedFocusId = $state<number | null>(null);
  let relationSummary = $state<RelationGroup[]>([]);
  let activeRelationType = $state<string | null>(null);
  let approvalUpdating = $state(false);

  const entities = $derived(project?.Entities ?? []);
  const intersectionEntities = $derived(project?.IntersectionEntities ?? []);
  const relationSummaryCount = $derived(relationSummary.reduce((total, group) => total + group.items.length, 0));
  const activeRelationGroup = $derived(relationSummary.find((group) => group.type === activeRelationType) ?? relationSummary[0] ?? null);

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
      if (rel.IdEntity1 === entity.Id) {
        if (rel.Relation === "N:1" || rel.Relation === "Np:1") {
          otherEntityId = rel.IdEntity2;
          isOptional = false;
        }
      } else if (rel.IdEntity2 === entity.Id) {
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
            inherited.push({ entityName: otherEntity.Name, attributeName: null, description: null, type: null, isOptional: isOptional });
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
    [ { id: rel.IdEntity1, optional: false }, { id: rel.IdEntity2, optional: false } ].forEach(target => {
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
          inherited.push({ entityName: otherEntity.Name, attributeName: null, description: null, type: null, isIntersection: true, isOptional: target.optional });
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
    "1:1": "Uno a uno", "1:N": "Uno a muchos", "N:1": "Muchos a uno", "N:N": "Muchos a muchos", "1:Np": "Uno a muchos (Opcional)", "Np:1": "Muchos (Opcional) a uno"
  };
  const relationTypeOrder = ["1:1", "1:N", "N:1", "N:N", "1:Np", "Np:1"];

  const runAttributeTransition = (update: () => void | Promise<void>) =>
    runViewTransition(update, "No se pudo aplicar la transicion de atributos:");

  const updateAutoScroll = (event: DragEvent) => autoScroller.updateFromDragEvent(event);

  onMount(() => {
    if (entities.length && selectedId === null) selectedId = entities[0].Id;
    if (intersectionEntities.length && selectedIntersectionRelationId === null) selectedIntersectionRelationId = intersectionEntities[0].RelationID;
    return () => autoScroller.stop();
  });

  $effect(() => {
    if (entities.length && selectedId === null) selectedId = entities[0].Id;
    if (!entities.some((entity) => entity.Id === selectedId)) selectedId = entities[0]?.Id ?? null;
    if (intersectionEntities.length && selectedIntersectionRelationId === null) selectedIntersectionRelationId = intersectionEntities[0].RelationID;
    if (!intersectionEntities.some((item) => item.RelationID === selectedIntersectionRelationId)) selectedIntersectionRelationId = intersectionEntities[0]?.RelationID ?? null;
  });

  $effect(() => {
    currentIntersection = intersectionEntities.find((item) => item.RelationID === selectedIntersectionRelationId) ?? null;
  });

  $effect(() => {
    if (activeScope === "strong" && focusEntityId !== null && focusEntityId !== lastSyncedFocusId && entities.some(entity => entity.Id === focusEntityId)) {
      void selectEntity(focusEntityId);
      lastSyncedFocusId = focusEntityId;
    }
  });

  $effect(() => {
    if (selectedId !== null && selectedId !== lastLoadedId) loadEntity(selectedId);
  });

  $effect(() => {
    if (activeScope === "intersection" && selectedIntersectionRelationId !== null) {
      loadIntersectionRelationSummary(selectedIntersectionRelationId);
    } else if (activeScope === "strong" && current) {
      void loadRelationSummary(current);
    }
  });

  $effect(() => {
    if (!relationSummary.length) {
      activeRelationType = null;
    } else if (!activeRelationType || !relationSummary.some((group) => group.type === activeRelationType)) {
      activeRelationType = relationSummary[0].type;
    }
  });

  const loadEntity = async (id: number) => {
    loading = true;
    try {
      current = await GetEntity(id);
      if (current && current.Attributes) {
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
    if (entityId === null || !entities.some(entity => entity.Id === entityId)) return;
    await runAttributeTransition(async () => {
      selectedId = entityId;
      await tick();
    });
  };

  const selectIntersection = async (relationId: number | null) => {
    if (relationId === null || !intersectionEntities.some((item) => item.RelationID === relationId)) return;
    await runAttributeTransition(async () => {
      selectedIntersectionRelationId = relationId;
      await tick();
    });
  };

  const handleSelectChange = async (event: Event) => {
    const target = event.target as HTMLSelectElement;
    if (activeScope === "strong") {
      await selectEntity(Number(target?.value ?? 0));
    } else {
      await selectIntersection(Number(target?.value ?? 0));
    }
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

  const nextEntity = async () => {
    if (activeScope === "strong") {
      if (!entities.length) return;
      const currentIndex = entities.findIndex((ent) => ent.Id === selectedId);
      const nextIndex = currentIndex === -1 || currentIndex === entities.length - 1 ? 0 : currentIndex + 1;
      await selectEntity(entities[nextIndex].Id);
    } else {
      if (!intersectionEntities.length) return;
      const currentIndex = intersectionEntities.findIndex((item) => item.RelationID === selectedIntersectionRelationId);
      const nextIndex = currentIndex === -1 || currentIndex === intersectionEntities.length - 1 ? 0 : currentIndex + 1;
      await selectIntersection(intersectionEntities[nextIndex].RelationID);
    }
  };

  const prevEntity = async () => {
    if (activeScope === "strong") {
      if (!entities.length) return;
      const currentIndex = entities.findIndex((ent) => ent.Id === selectedId);
      const prevIndex = currentIndex <= 0 ? entities.length - 1 : currentIndex - 1;
      await selectEntity(entities[prevIndex].Id);
    } else {
      if (!intersectionEntities.length) return;
      const currentIndex = intersectionEntities.findIndex((item) => item.RelationID === selectedIntersectionRelationId);
      const prevIndex = currentIndex <= 0 ? intersectionEntities.length - 1 : currentIndex - 1;
      await selectIntersection(intersectionEntities[prevIndex].RelationID);
    }
  };

  const applyReorder = async (from: number, to: number) => {
    const attributes = activeScope === "strong" ? (current?.Attributes ?? []) : (currentIntersection?.Entity.Attributes ?? []);
    if (from === to || from < 0 || to < 0 || from >= attributes.length || to >= attributes.length) return;
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
      if (activeScope === "strong" && current) await loadEntity(current.Id);
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
      case "1:1": return "1:1";
      case "1:N": return "N:1";
      case "N:1": return "1:N";
      default: return relation;
    }
  };

  const loadRelationSummary = async (ent: utils.Entity | null) => {
    if (!ent) {
      relationSummary = [];
      return;
    }
    try {
      const combData = await GetCombinatory();
      const groupedItems = new Map<string, string[]>();
      (combData || []).forEach(view => {
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
      console.error("No se pudo cargar resumen de relaciones:", getErrorMessage(err));
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
    relationSummary = items.length === 0 ? [] : [{ type: "Padres", label: "Entidades relacionadas", items: items }];
  };

  const isApproved = (entity: utils.Entity | null) => entity?.Status === true;

  const intersectionOriginLabel = (item: utils.IntersectionEntity | null) => {
    if (!item) return "Sin relación origen";
    const relation = project?.Relations?.find((currentRelation) => currentRelation.Id === item.RelationID);
    if (!relation) return "Sin relación origen";
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
    if (!current) return;
    approvalUpdating = true;
    try {
      await MarkEntityStatus(current.Id, !isApproved(current));
      await onRefresh();
      await loadEntity(current.Id);
    } catch (err) {
      showToast(`No se pudo actualizar la aprobación: ${getErrorMessage(err)}`, "error");
    } finally {
      approvalUpdating = false;
    }
  };

  const jumpToTab = (tab: "entities" | "relations" | "tertiary") => {
    onJumpTo(tab, activeScope === "strong" ? selectedId : null);
  };

  const jumpToEntity = (entityName: string) => {
    const target = entities.find(e => e.Name === entityName);
    if (target) onJumpTo("tertiary", target.Id);
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
          <Button variant="ghost" icon="database" onclick={() => jumpToTab("entities")} disabled={activeScope !== "strong" || !selectedId}>
            Ir a definicion
          </Button>
          <Button variant="accent" icon="relations" onclick={() => jumpToTab("relations")} disabled={activeScope !== "strong" || !selectedId}>
            Ir a combinatorio
          </Button>
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
              onchange={handleSelectChange}
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
              <Button variant="soft" size="icon" icon="chevron-left" onclick={prevEntity} aria-label="Entidad anterior" disabled={activeScope === "strong" ? entities.length <= 1 : intersectionEntities.length <= 1} />
              <Button variant="soft" size="icon" icon="chevron-right" onclick={nextEntity} aria-label="Entidad siguiente" disabled={activeScope === "strong" ? entities.length <= 1 : intersectionEntities.length <= 1} />
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
                onchange={handleRelationTypeChange}
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
              {#snippet actions()}
                <div class="entity-focus-actions">
                  <CreateEntity
                    id={current.Id}
                    triggerLabel="Editar"
                    onSave={async () => {
                      await onRefresh();
                      if (current) await loadEntity(current.Id);
                    }}
                  />
                  <Button
                    variant={isApproved(current) ? 'active' : 'success'}
                    icon={isApproved(current) ? "check-off" : "check"}
                    onclick={toggleCurrentApproval}
                    disabled={approvalUpdating}
                  >
                    {isApproved(current) ? "Quitar" : "Aprobar"}
                  </Button>
                </div>
              {/snippet}
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
            onEntityReload={async () => { if (current) await loadEntity(current.Id); }}
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
              {#snippet actions()}
                <div class="entity-focus-actions">
                  <span class="intersection-origin">{intersectionOriginLabel(currentIntersection)}</span>
                </div>
              {/snippet}
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

<style>
  .attributes-studio {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .attributes-layout {
    display: grid;
    grid-template-columns: 320px 1fr;
    gap: 1.5rem;
    align-items: start;
  }

  .attributes-deck {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    position: sticky;
    top: calc(var(--attributes-sticky-total-height) + 1.5rem);
  }

  .attributes-focus-card {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .entity-focus-actions {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .attributes-panel {
    background: var(--background);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-sm);
    overflow: hidden;
  }

  .attributes-panel__head {
    padding: 1.5rem;
    border-bottom: 1px solid var(--border-card);
  }

  .intersection-origin {
    font-size: 0.7rem;
    font-weight: 700;
    color: var(--ink-soft);
    font-style: italic;
  }

  .attributes-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.25rem 1.5rem;
    background: var(--surface);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-mini);
    gap: 2rem;
  }

  .attributes-toolbar__meta {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .attributes-toolbar__actions {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .view-jumps {
    display: flex;
    gap: 0.5rem;
  }

  .entity-switcher {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .entity-picker {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .entity-nav {
    display: flex;
    gap: 0.25rem;
  }

  .attributes-toolbar__relations {
    margin-top: 1rem;
    background: var(--panel-surface-soft);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-lg);
    padding: 1rem;
  }

  .relation-bar {
    display: flex;
    align-items: center;
    gap: 1.5rem;
  }

  .relation-bar__picker {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    min-width: 140px;
  }

  .relation-bar__tags {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    flex: 1;
  }

  .relation-pill {
    background: var(--accent-soft);
    border: 1px solid var(--accent);
    color: var(--accent);
    padding: 0.25rem 0.75rem;
    border-radius: 99px;
    font-size: 0.75rem;
    font-weight: 700;
  }

  .relation-type-summary {
    display: flex;
    gap: 1rem;
    margin-top: 0.75rem;
    padding-top: 0.75rem;
    border-top: 1px solid var(--border-card);
  }

  .relation-type-summary__item {
    font-size: 0.65rem;
    font-weight: 800;
    color: var(--ink-faint);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  @media (max-width: 1200px) {
    .attributes-toolbar {
      flex-direction: column;
      align-items: stretch;
      gap: 1rem;
    }
  }

  @media (max-width: 1024px) {
    .attributes-layout {
      grid-template-columns: 1fr;
    }
    .attributes-deck {
      position: static;
    }
  }
</style>
