<script lang="ts">
  import type { utils } from "../../wailsjs/go/models";
  import Button from "./ui/Button.svelte";
  import Table from "./ui/Table.svelte";

  type RelationRow = {
    Id?: number;
    Entity2: string;
    IdEntity2: number;
    Relation: string;
  };

  type RelationMenuState = {
    open: boolean;
    targetId: number | null;
  };

  let {
    relations = [],
    relationOptions = [],
    updating = false,
    relationMenu,
    onRelationChange,
    onContextMenu,
    isApprovedEntity,
    getEntityDefinition,
    relationIdentifiers
  }: {
    relations: RelationRow[];
    relationOptions: string[];
    updating: boolean;
    relationMenu: RelationMenuState;
    onRelationChange: (relation: RelationRow) => void;
    onContextMenu: (relation: RelationRow, event: MouseEvent) => void;
    isApprovedEntity: (entityId: number | null | undefined) => boolean;
    getEntityDefinition: (entityId: number) => string;
    relationIdentifiers: (relation: RelationRow) => { principalId: number | null; relationId: number | null; targetId: number | null };
  } = $props();
</script>

<Table class="frosted">
  {#snippet header()}
    <th>Entidad</th>
    <th style="width: 160px;">Relación</th>
  {/snippet}
  {#snippet body()}
    {#if relations && relations.length > 0}
      {#each relations as relation}
        <tr
          class:approved-row={isApprovedEntity(relation.IdEntity2)}
          class:relation-row-menu-open={relationMenu.open && relationMenu.targetId === relation.IdEntity2}
          oncontextmenu={(event) => { event.preventDefault(); event.stopPropagation(); onContextMenu(relation, event); }}
        >
          <td>
            <div class="relation-name-cell">
              <span>{relation.Entity2}</span>
              <span class="relation-info">
                <Button
                  variant="ghost"
                  size="icon"
                  class="info-trigger"
                  aria-label="Ayuda de la relación"
                  onclick={(e) => e.stopPropagation()}
                  icon="eye"
                />
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
              onclick={(e) => e.stopPropagation()}
              onchange={() => onRelationChange(relation)}
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
        <td colspan="2">Sin relaciones para esta entidad.</td>
      </tr>
    {/if}
  {/snippet}
</Table>
