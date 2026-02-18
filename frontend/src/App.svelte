<script lang="ts">
  import {CreateProjectJSONPath, OpenPath, Save} from '../wailsjs/go/main/App.js';
  import {CreateNew, GetActualProject, PickProjectJSON} from "../wailsjs/go/main/App";
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
  async function createProject() {
    try {
      const path = await CreateProjectJSONPath();
      if (!path) return;
      const res= await CreateNew(path);
      data = res;
      activeTab = 'entities';
    } catch (e) {
      alert("Dialog error:"+e.error);
    }
  }
   const handleSave:  () => Promise<void> = async () => {
    try{
      alert("Guardando");
      await Save();
      alert("Guardado");
    }catch(e){
      alert("Error en guardado");
    }
  };
  const handleExport = () => {};
  const handleRefresh = async () => {
    try {
      data = await GetActualProject();
    } catch (err) {
      const message = err?.error ?? err?.message ?? err;
      alert(`No se pudo recargar el proyecto: ${message}`);
    }
  };
  const handleTabSelect = (tab:TabKey) => {
    activeTab = tab;
  };
</script>

<main class="app-shell">
  {#if data === null}
    <Hero onOpen={openProject} onCreate={createProject}/>
  {:else}
    <ActionsBar onSave={handleSave} onExport={handleExport}/>
  {/if}

  {#if data != null}
    <ProjectHeader name={data.Name} entityCount={data.Entities.length}/>

    <TabBar activeTab={activeTab} onSelect={handleTabSelect}/>

    <section class="tab-panel">
      {#if activeTab === 'entities'}
        <EntitiesTab entities={data.Entities} onSave={handleRefresh}/>
      {:else if activeTab === 'relations'}
        <RelationsTab onRefresh={handleRefresh}/>
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
    background: #111a2b;
    min-height: 240px;
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

  :global(.btn.danger) {
    background: linear-gradient(120deg, #ff6b6b, #ff416c);
    color: #0b0f1a;
    box-shadow: 0 12px 30px rgba(255, 99, 123, 0.35);
  }

  :global(.btn.danger:hover) {
    transform: translateY(-1px);
    box-shadow: 0 16px 36px rgba(255, 99, 123, 0.45);
  }

  :global(.btn.danger:active) {
    transform: translateY(0);
    box-shadow: 0 8px 24px rgba(255, 99, 123, 0.35);
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
