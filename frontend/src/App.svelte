<script lang="ts">
  import {OpenPath} from '../wailsjs/go/main/App.js';
  import {PickProjectJSON} from "../wailsjs/go/main/App";
  import {utils} from "../wailsjs/go/models";
  import Hero from './components/Hero.svelte';
  import ActionsBar from './components/ActionsBar.svelte';
  import ProjectHeader from './components/ProjectHeader.svelte';
  import TabBar from './components/TabBar.svelte';
  import EntitiesTab from './components/EntitiesTab.svelte';
  import RelationsTab from "./components/RelationsTab.svelte";
  import PlaceholderTab from './components/PlaceholderTab.svelte';

  type TabKey = 'entities' | 'relations' | 'tertiary';

  let data:utils.DbProject | null = null;
  let activeTab:TabKey = 'entities';

  async function openProject() {
    try {
      const path = await PickProjectJSON();
      if (!path) return;
      const res= await OpenPath(path);
      data = res;
      activeTab = 'entities';
    } catch (e) {
      alert("Dialog error:"+e.error);
    }
  }

  const handleSave = () => {};
  const handleExport = () => {};
  const handleCreate = () => {};
  const handleTabSelect = (tab:TabKey) => {
    activeTab = tab;
  };
</script>

<main class="app-shell">
  {#if data === null}
    <Hero onOpen={openProject}/>
  {:else}
    <ActionsBar onSave={handleSave} onExport={handleExport}/>
  {/if}

  {#if data != null}
    <ProjectHeader name={data.Name} entityCount={data.Entities.length}/>

    <TabBar activeTab={activeTab} onSelect={handleTabSelect}/>

    <section class="tab-panel">
      {#if activeTab === 'entities'}
        <EntitiesTab entities={data.Entities} onCreate={handleCreate}/>
      {:else if activeTab === 'relations'}
        <RelationsTab />
      {:else}
        <PlaceholderTab message="Contenido pendiente para esta pestaña."/>
      {/if}
    </section>
  {:else}
    <section class="empty-panel ghost">
      <p>Sin proyecto cargado. Usa el botón para empezar.</p>
    </section>
  {/if}
</main>

<style>
  :global(body) {
    background: radial-gradient(120% 120% at 10% 20%, rgba(255,255,255,0.08), rgba(27,38,54,1));
  }

  main.app-shell {
    max-width: 1080px;
    margin: 0 auto;
    padding: 32px 24px 48px;
    color: #e8edf7;
    text-align: left;
  }

  .tab-panel {
    margin-top: 16px;
    padding: 18px;
    border-radius: 14px;
    border: 1px solid rgba(255, 255, 255, 0.08);
    background: rgba(255, 255, 255, 0.02);
    min-height: 240px;
    box-shadow: 0 10px 30px rgba(0,0,0,0.3);
    max-height: calc(100vh - 260px);
    overflow: auto;
  }

  :global(.btn) {
    border: none;
    border-radius: 12px;
    padding: 14px 18px;
    font-weight: 700;
    cursor: pointer;
    transition: transform 150ms ease, box-shadow 180ms ease, background 180ms ease;
  }

  :global(.btn.primary) {
    background: linear-gradient(120deg, #5ad1ff, #6287f6);
    color: #0b1a30;
    box-shadow: 0 12px 30px rgba(82, 158, 255, 0.35);
  }

  :global(.btn.primary:hover) {
    transform: translateY(-1px);
    box-shadow: 0 16px 36px rgba(82, 158, 255, 0.45);
  }

  :global(.btn.primary:active) {
    transform: translateY(0);
    box-shadow: 0 8px 24px rgba(82, 158, 255, 0.35);
  }

  :global(.btn.secondary) {
    background: rgba(255, 255, 255, 0.08);
    color: #d9e4f5;
    border: 1px solid rgba(255, 255, 255, 0.14);
  }

  :global(.btn.secondary:hover) {
    background: rgba(255, 255, 255, 0.12);
  }

  :global(.empty-panel) {
    display: flex;
    align-items: center;
    justify-content: center;
    color: #cfd9e9;
    opacity: 0.85;
    font-size: 15px;
    height: 180px;
    border: 1px dashed rgba(255, 255, 255, 0.12);
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.02);
  }

  :global(.empty-panel.ghost) {
    margin-top: 18px;
  }

  @media (max-width: 720px) {
    main.app-shell {
      padding: 22px 16px 36px;
    }
  }
</style>
