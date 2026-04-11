<script lang="ts">
  import { SvelteFlow, Controls, Background, MiniMap } from '@xyflow/svelte';
  import '@xyflow/svelte/dist/style.css';
  import type { utils } from "../../wailsjs/go/models";
  import ButtonIcon from "./ButtonIcon.svelte";
  import DependencyAnalysisModal from "./forms/DependencyAnalysisModal.svelte";
  import { UpdateAllCoordinates, UpdateEntityCoordinates } from "../../wailsjs/go/main/App";

  let { project, onRefresh } = $props<{
    project: utils.DbProject;
    onRefresh: () => Promise<void>;
  }>();

  let analysisModal: DependencyAnalysisModal;

  let nodes = $state.raw<any[]>([]);
  let edges = $state.raw<any[]>([]);

  $effect(() => {
    if (!project) return;
    
    const newNodes: any[] = [];
    const newEdges: any[] = [];
    
    const GRID_GAP = 280;
    const COLS = 5;
    let nextGridIndex = 0;

    const entities = Array.isArray(project.Entities) ? project.Entities : [];
    const intersections = Array.isArray(project.IntersectionEntities) ? project.IntersectionEntities : [];
    const relations = Array.isArray(project.Relations) ? project.Relations : [];

    entities.forEach((ent: any) => {
      if (ent?.Id == null) return;
      const x = ent.coords?.x ?? (nextGridIndex % COLS) * GRID_GAP;
      const y = ent.coords?.y ?? Math.floor(nextGridIndex / COLS) * GRID_GAP;
      if (!ent.coords) nextGridIndex++;

      newNodes.push({
        id: `strong-${ent.Id}`,
        position: { x, y },
        data: { label: ent.Name, originalId: ent.Id, isIntersection: false }
      });
    });

    intersections.forEach((ie: any) => {
      const ent = ie?.Entity;
      if (ent?.Id == null) return;
      const x = ent.coords?.x ?? (nextGridIndex % COLS) * GRID_GAP;
      const y = ent.coords?.y ?? Math.floor(nextGridIndex / COLS) * GRID_GAP;
      if (!ent.coords) nextGridIndex++;

      newNodes.push({
        id: `intersection-${ent.Id}`,
        position: { x, y },
        data: { label: ent.Name, originalId: ent.Id, isIntersection: true }
      });
    });

    relations.forEach((rel: any) => {
      if (rel.Relation === "N:N") {
        const ie = intersections.find((i: any) => i?.RelationID === rel?.Id);
        if (ie?.Entity?.Id != null) {
          newEdges.push({
            id: `edge-${rel.Id}-1`,
            source: `strong-${rel.IdEntity1}`,
            target: `intersection-${ie.Entity.Id}`,
            label: "N-N",
            type: "smoothstep",
            style: "stroke: var(--ink-soft); stroke-width: 2;",
          });
          newEdges.push({
            id: `edge-${rel.Id}-2`,
            source: `strong-${rel.IdEntity2}`,
            target: `intersection-${ie.Entity.Id}`,
            label: "N-N",
            type: "smoothstep",
            style: "stroke: var(--ink-soft); stroke-width: 2;",
          });
        }
      } else if (rel.IdEntity1 != null && rel.IdEntity2 != null) {
        newEdges.push({
          id: `edge-${rel.Id}`,
          source: `strong-${rel.IdEntity1}`,
          target: `strong-${rel.IdEntity2}`,
          label: rel.Relation || "",
          type: "smoothstep",
          style: "stroke: var(--ink-soft); stroke-width: 2;",
        });
      }
    });

    nodes = newNodes;
    edges = newEdges;
  });

  async function handleNodeDragStop(event: any) {
    // Extraemos el node del evento
    const node = event.node;
    if (!node || node.data?.originalId == null) return;
    try {
      await UpdateEntityCoordinates(
        node.data.originalId,
        node.position.x,
        node.position.y,
        node.data.isIntersection
      );
    } catch (e) {
      console.error(e);
    }
  }

  async function handleSave() {
    try {
      const coordsToUpdate = nodes
        .filter((n: any) => n.data?.originalId != null)
        .map((n: any) => ({
          id: n.data.originalId,
          x: n.position.x,
          y: n.position.y,
          is_intersection: !!n.data.isIntersection,
        }));
      await UpdateAllCoordinates(coordsToUpdate);
      if (onRefresh) await onRefresh();
    } catch (e) {
      console.error(e);
    }
  }
</script>

<div class="erd-canvas-container" style="height: 100vh; display: flex; flex-direction: column;">
  <div style="display: flex; justify-content: flex-end; padding: 10px; gap: 10px; background: var(--surface);">
    <button type="button" class="btn accent" onclick={() => analysisModal?.openDialog()}>
      <ButtonIcon name="relations" />
      <span>Analizar dependencias</span>
    </button>
    <button type="button" class="btn secondary" onclick={handleSave}>
      <ButtonIcon name="save" />
      <span>Guardar cambios</span>
    </button>
  </div>

  <div style="flex: 1; width: 100%;">
    <SvelteFlow 
      bind:nodes 
      bind:edges 
      fitView 
      colorMode="dark"
      onnodedragstop={handleNodeDragStop}>
      <Controls />
      <Background />
      <MiniMap />
    </SvelteFlow>
  </div>
</div>

<DependencyAnalysisModal bind:this={analysisModal} />

<style>
  :global(.svelte-flow__node) {
    background: #fff;
    border: 1px solid #222;
    padding: 10px;
    border-radius: 5px;
    color: #222;
    font-weight: 600;
    min-width: 150px;
    text-align: center;
  }
</style>
