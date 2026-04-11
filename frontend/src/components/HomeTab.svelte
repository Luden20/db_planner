<script lang="ts">
  import type { utils } from "../../wailsjs/go/models";
  import "@xyflow/svelte/dist/style.css";
  import { SvelteFlow, Background, Controls } from "@xyflow/svelte";
  import DependencyAnalysisModal from "./forms/DependencyAnalysisModal.svelte";
  import ButtonIcon from "./ButtonIcon.svelte";

  export let project: utils.DbProject;
  // @ts-ignore
  export let onRefresh: () => Promise<void>;

  let analysisModal: DependencyAnalysisModal;

  let nodes = [
    {
      id: "1",
      type: "input",
      data: { label: "Hola" },
      position: { x: 100, y: 100 },
    },
    {
      id: "2",
      data: { label: "Mundo" },
      position: { x: 300, y: 100 },
    },
    {
      id: "3",
      data: { label: "Mundo" },
      position: { x: 500, y: 100 },
    },
  ];

  let edges = [
    {
      id: "e1-2",
      source: "1",
      target: "2",
    },
    {
      id: "e2-3",
      source: "2",
      target: "3",
    },
  ];

  function handleNodeDragStop() {
    console.log("Nodos actuales:", nodes);
  }

  function showCoordinates() {
    const text = nodes
      .map((n) => `Nodo ${n.id}: x=${n.position.x}, y=${n.position.y}`)
      .join("\n");

    alert(text || "No hay nodos");
  }
</script>

<section class="home-sandbox">
  <header class="home-header">
    <div class="home-copy">
      <h1>Home</h1>
      <p class="muted">Visualiza tu esquema y analiza la integridad de las dependencias.</p>
    </div>
    <div class="home-actions">
      <button type="button" class="btn accent" on:click={() => analysisModal.openDialog()}>
        <ButtonIcon name="relations" />
        <span>Analizar dependencias</span>
      </button>
      <button type="button" class="btn secondary" on:click={showCoordinates}>
        <ButtonIcon name="frame" />
        <span>Ver coordenadas</span>
      </button>
    </div>
  </header>

  <div class="flow-wrapper">
    <SvelteFlow
      bind:nodes
      bind:edges
      fitView
      onnodedragstop={handleNodeDragStop}
    >
      <Background />
      <Controls />
    </SvelteFlow>
  </div>
</section>

<DependencyAnalysisModal bind:this={analysisModal} />

<style>
  .home-sandbox {
    min-height: 18rem;
    padding: 0.5rem;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .home-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .home-copy h1 {
    margin: 0;
    font-size: 2.5rem;
    letter-spacing: -0.04em;
    line-height: 1;
  }

  .muted {
    margin: 0.5rem 0 0;
    color: var(--ink-soft);
    font-size: 1.1rem;
  }

  .home-actions {
    display: flex;
    gap: 0.85rem;
  }

  .flow-wrapper {
    width: 100%;
    height: 600px;
    background: var(--surface-strong);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    overflow: hidden;
    box-shadow: var(--shadow-sm);
  }
</style>
