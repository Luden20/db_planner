<script lang="ts">
  import type { utils } from "../../wailsjs/go/models";
  import ButtonIcon from "./ButtonIcon.svelte";
  import AttributeForm from "./forms/AttributeForm.svelte";
  import DeleteAttribute from "./forms/DeleteAttribute.svelte";
  import Table from "./ui/Table.svelte";

  type InheritedPK = {
    entityName: string;
    attributeName: string | null;
    description: string | null;
    type: string | null;
    isIntersection?: boolean;
    isOptional?: boolean;
  };

  let {
    // Mode
    isIntersection = false,
    // Data for strong entities
    entity = null,
    entityId = null,
    attributes = [],
    inheritedPKs = [],
    // Data for intersection
    currentIntersection = null,
    intersectionAttributes = [],
    intersectionInheritedPKs = [],
    // Drag state
    draggingIndex = null,
    hoverIndex = null,
    tableRef = $bindable(),
    // Drag callbacks
    onDragStart,
    onDragOver,
    onDrop,
    onDragEnd,
    onTableDragOver,
    onTableDragLeave,
    onTableDrop,
    // Navigation
    onJumpToEntity,
    // Refresh callbacks
    onRefresh,
    onEntityReload
  }: {
    isIntersection?: boolean;
    entity?: utils.Entity | null;
    entityId?: number | null;
    attributes?: utils.Attribute[];
    inheritedPKs?: InheritedPK[];
    currentIntersection?: utils.IntersectionEntity | null;
    intersectionAttributes?: utils.Attribute[];
    intersectionInheritedPKs?: InheritedPK[];
    draggingIndex?: number | null;
    hoverIndex?: number | null;
    tableRef?: HTMLDivElement | null;
    onDragStart: (index: number, event: DragEvent) => void;
    onDragOver: (index: number, event: DragEvent) => void;
    onDrop: (index: number, event: DragEvent) => Promise<void>;
    onDragEnd: () => void;
    onTableDragOver: (event: DragEvent) => void;
    onTableDragLeave: (event: DragEvent) => void;
    onTableDrop: () => void;
    onJumpToEntity?: (entityName: string) => void;
    onRefresh: () => Promise<void>;
    onEntityReload?: () => Promise<void>;
  } = $props();

  // Active list of attributes based on mode
  const activeAttrs = $derived(isIntersection ? intersectionAttributes : attributes);
  const activeInherited = $derived(isIntersection ? intersectionInheritedPKs : inheritedPKs);

  // For the strong-entity table: show inherited PKs only if the entity has no own PK
  const hasPK = $derived(!isIntersection && attributes.some(a => a.KeyType === "pk"));
  // For intersection: always show inherited PKs
  const showInheritedPKs = $derived(isIntersection || !hasPK);

  const isEmpty = $derived(
    (!activeAttrs || activeAttrs.length === 0) &&
    (!showInheritedPKs || activeInherited.length === 0)
  );

  const canDrag = (attribute: utils.Attribute) =>
    isIntersection || attribute.KeyType !== "pk";
</script>

<!--
  This component renders the full attributes table for both "strong" and "intersection" scopes.
  It wraps the shared <Table> component and wires all drag-and-drop events.
-->
<Table
  class="frosted"
  tableClass={`${!isIntersection ? 'draggable' : ''}`}
  ref={tableRef}
  ondragover={onTableDragOver}
  ondragleave={onTableDragLeave}
  ondrop={onTableDrop}
>
  {#snippet header()}
    <th>Nombre</th>
    <th>Descripción</th>
    <th style="width: 120px;">Tipo</th>
    <th style="width: 80px;">Mandatorio</th>
    <th style="width: 120px;">Acciones</th>
  {/snippet}

  {#snippet body()}
    <!-- Inherited FK rows (strong: only when no own PK; intersection: always) -->
    {#if showInheritedPKs}
      {#each activeInherited as inherited}
        <tr class={`inherited-pk-row${isIntersection ? " intersection-pk" : ""}`} draggable="false">
          <td class="inherited-name">
            <span class="inherited-tag">{isIntersection ? "PK heredada" : "FK heredada"}</span>
            {inherited.attributeName || `PK de ${inherited.entityName} pendiente por definir`}
            {#if !isIntersection && onJumpToEntity}
              <button
                class="control control--icon control--xs jump-btn"
                onclick={() => onJumpToEntity!(inherited.entityName)}
                title={`Ir a ${inherited.entityName}`}
              >
                <ButtonIcon name="jump" />
              </button>
            {/if}
          </td>
          <td>
            {#if inherited.attributeName}
              <span class="muted italic">{inherited.description || "Sin descripción."}</span>
            {:else}
              <span class="muted italic">Atributo heredado por relación con {inherited.entityName}</span>
            {/if}
          </td>
          <td class="muted">{inherited.type || "—"}</td>
          <td>
            {#if inherited.isOptional}
              <span class="badge badge--optional" title="Esta columna puede ser NULL">No</span>
            {:else}
              <span class="badge badge--mandatory" title="Esta columna no puede ser NULL">Sí</span>
            {/if}
          </td>
          <td class="muted">
            <div class="row-actions" style="justify-content: flex-start; gap: 8px;">
              Lectura
            </div>
          </td>
        </tr>
      {/each}
    {/if}

    <!-- Empty state -->
    {#if isEmpty}
      <tr class="empty-row" draggable="false">
        <td colspan="5">No hay atributos definidos aún.</td>
      </tr>
    {:else}
      <!-- Own attribute rows -->
      {#each activeAttrs as attribute, index (attribute.Id)}
        <tr
          class:dragging={draggingIndex === index}
          class:drag-hover={hoverIndex === index && draggingIndex !== null && draggingIndex !== index}
          class:pk-row={!isIntersection && attribute.KeyType === "pk"}
          draggable={canDrag(attribute)}
          style={`view-transition-name: ${isIntersection ? "intersection-" : ""}attribute-row-${attribute.Id};`}
          ondragstart={(event) => canDrag(attribute) && onDragStart(index, event)}
          ondragover={(event) => canDrag(attribute) && onDragOver(index, event)}
          ondragenter={(event) => canDrag(attribute) && onDragOver(index, event)}
          ondrop={(event) => canDrag(attribute) && onDrop(index, event)}
          ondragend={onDragEnd}
        >
          <td>
            {#if !isIntersection && attribute.KeyType === "pk"}
              <div class="pk-name-cell">
                <span class="pk-tag">PK</span>
                <strong>{attribute.Name}</strong>
              </div>
            {:else}
              {attribute.Name}
            {/if}
          </td>
          <td>{attribute.Description}</td>
          <td>{attribute.Type || "Por definir"}</td>
          <td>
            {#if attribute.Optional}
              <span class="badge badge--optional" title="Esta columna puede ser NULL">No</span>
            {:else}
              <span class="badge badge--mandatory" title="Esta columna no puede ser NULL">Sí</span>
            {/if}
          </td>
          <td>
            <div class="row-actions">
              {#if isIntersection && currentIntersection}
                <AttributeForm
                  relationId={currentIntersection.RelationID}
                  entity={currentIntersection.Entity}
                  {attribute}
                  allowPrimaryKey={false}
                  triggerClass="control--sm"
                  onSaved={onRefresh}
                />
                <DeleteAttribute
                  relationId={currentIntersection.RelationID}
                  attributeId={attribute.Id}
                  triggerClass="control--sm"
                  onSaved={onRefresh}
                />
              {:else if !isIntersection && entity}
                <AttributeForm
                  {entityId}
                  {entity}
                  {attribute}
                  allowPrimaryKey={true}
                  triggerClass="control--sm"
                  onSaved={async () => {
                    await onRefresh();
                    if (onEntityReload) await onEntityReload();
                  }}
                />
                <DeleteAttribute
                  {entityId}
                  attributeId={attribute.Id}
                  triggerClass="control--sm"
                  onSaved={async () => {
                    await onRefresh();
                    if (onEntityReload) await onEntityReload();
                  }}
                />
              {/if}
            </div>
          </td>
        </tr>
      {/each}
    {/if}
  {/snippet}
</Table>
