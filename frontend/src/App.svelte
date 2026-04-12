<script lang="ts">
  import {onMount} from "svelte";
  import "@xyflow/svelte/dist/style.css";
  // Forcing Vite rebuild to resolve ERD sync issues
  import {
    CreateExcelPath,
    CreateExcelPathWithoutRelations,
    CreateNew,
    CreateProjectJSONPath,
    EjectProject,
    ExportCombinationsToExcel,
    ExportToExcel,
    GetActualProject,
    OpenPath,
    PickProjectJSON,
    Save
  } from '../wailsjs/go/main/App.js';
  import type {utils} from "../wailsjs/go/models";
  import Hero from './components/Hero.svelte';
  import WorkspaceRibbon from "./components/WorkspaceRibbon.svelte";
  import ButtonIcon from "./components/ButtonIcon.svelte";
  import EntitiesTab from './components/EntitiesTab.svelte';
  import RelationsTab from "./components/RelationsTab.svelte";
  import AttributesTab from "./components/AttributesTab.svelte";
  import FlowsTab from "./components/FlowsTab.svelte";
  import HomeTab from "./components/HomeTab.svelte";
  import RolesTab from "./components/RolesTab.svelte";
  import ExportAISQL from "./components/forms/ExportAISQL.svelte";
  import ExportPowerDesigner from "./components/forms/ExportPowerDesigner.svelte";
  import Modal from "./components/ui/Modal.svelte";
  import {showToast, toastState} from "./lib/toast";

  type TabKey = 'home' | 'entities' | 'relations' | 'roles' | 'flows' | 'tertiary';
  type ThemeMode = "light" | "dark";
  const THEME_STORAGE_KEY = "db-planner-theme";

  let data = $state<utils.DbProject | null>(null);
  let activeTab = $state<TabKey>('entities');
  let themeMode = $state<ThemeMode>("dark");
  let focusEntityId = $state<number | null>(null);
  let ribbonHeight = $state(0);
  let exportAISQLModal: ExportAISQL | null = $state(null);
  let exportPowerDesignerModal: ExportPowerDesigner | null = $state(null);
  let showExitDialog = $state(false);
  let exitInProgress = $state(false);
  let showCreateDialog = $state(false);
  let createName = $state("");
  let createPath = $state("");
  let createBusy = $state(false);
  let createError = $state("");

  async function openProject() {
    try {
      const path = await PickProjectJSON();
      if (!path) return;
      const res= await OpenPath(path);
      data = res;
      activeTab = 'entities';
    } catch (e: any) {
      const message = e?.error ?? e?.message ?? e ?? "Error desconocido";
      showToast(`Error del dialogo: ${message}`, "error");
    }
  }
  const openCreateDialog = () => {
    createName = "";
    createPath = "";
    createError = "";
    showCreateDialog = true;
  };

  const closeCreateDialog = () => {
    if (createBusy) return;
    showCreateDialog = false;
    createError = "";
  };

  const selectCreatePath = async () => {
    try {
      const path = await CreateProjectJSONPath();
      if (!path) return;
      createPath = path;
      if (!createName.trim()) {
        const filename = path.split(/[\\/]/).pop() || "";
        createName = filename.replace(/\.json$/i, "");
      }
    } catch (e: any) {
      const message = e?.error ?? e?.message ?? e;
      showToast(`Error del dialogo: ${message}`, "error");
    }
  };

  const submitCreateProject = async () => {
    if (createBusy) return;
    const trimmedName = createName.trim();
    if (!trimmedName) {
      createError = "Ingresa un nombre de proyecto.";
      return;
    }
    if (!createPath) {
      createError = "Selecciona un archivo JSON para el proyecto.";
      return;
    }
    createError = "";
    createBusy = true;
    try {
      const res= await CreateNew(createPath, trimmedName);
      data = res;
      activeTab = 'entities';
      showCreateDialog = false;
    } catch (e: any) {
      const message = e?.error ?? e?.message ?? e ?? "Error desconocido";
      createError = `${message}`;
    } finally {
      createBusy = false;
    }
  };
   const handleSave:  () => Promise<void> = async () => {
    try{
      showToast("Guardando...", "info", 0);
      await Save();
      showToast("Proyecto guardado.", "success");
    }catch(e){
      showToast("Error al guardar.", "error");
    }
  };
  const handleExport = async () => {
    try {
      const path = await CreateExcelPath();
      if (!path) {
        return;
      }
      await ExportToExcel(path);
      showToast("Exportado a Excel con relaciones.", "success");
    } catch (e: any) {
      const message = e?.error ?? e?.message ?? e;
      showToast(`Error al exportar: ${message}`, "error");
    }
  };
  const handleExportWithoutRelations = async () => {
    try {
      const path = await CreateExcelPathWithoutRelations();
      if (!path) {
        return;
      }
      await ExportCombinationsToExcel(path);
      showToast("Exportado a Excel sin relaciones.", "success");
    } catch (e: any) {
      const message = e?.error ?? e?.message ?? e;
      showToast(`Error al exportar: ${message}`, "error");
    }
  };
  const handleExportScripts = () => {
    exportAISQLModal?.openDialog();
  };
  const handleExportPowerDesigner = () => {
    exportPowerDesignerModal?.openDialog();
  };

  const handleRefresh = async () => {
    try {
      const res = await GetActualProject();
      data = res;
    } catch (e: any) {
      console.error("Error al refrescar:", e);
    }
  };

  const handleTabSelect = (tab: TabKey) => {
    activeTab = tab;
    focusEntityId = null;
  };

  const handleExitRequest = () => {
    showExitDialog = true;
  };

  const exitWithoutSave = async () => {
    try {
      exitInProgress = true;
      await EjectProject();
      data = null;
      showExitDialog = false;
    } catch (e: any) {
      showToast("Error al salir.", "error");
    } finally {
      exitInProgress = false;
    }
  };

  const saveAndExit = async () => {
    try {
      exitInProgress = true;
      await Save();
      await EjectProject();
      data = null;
      showExitDialog = false;
      showToast("Guardado y sesión cerrada.", "success");
    } catch (e: any) {
      showToast("Error al guardar y salir.", "error");
    } finally {
      exitInProgress = false;
    }
  };

  const cancelExit = () => {
    showExitDialog = false;
  };

  onMount(() => {
    applyTheme("dark");
  });

  const applyTheme = (nextTheme: ThemeMode) => {
    if (typeof document !== "undefined") {
      document.documentElement.dataset.theme = "dark";
      document.documentElement.classList.add("dark");
    }
  };

  const handleJumpToEntityTab = (tab: TabKey, entityId: number | null = null) => {
    activeTab = tab;
    focusEntityId = entityId;
  };
</script>

<main class="app-shell">
  {#if data === null}
    <Hero onOpen={openProject} onCreate={openCreateDialog}/>
  {:else}
    <div class="workspace-ribbon-slot" bind:clientHeight={ribbonHeight}>
      <WorkspaceRibbon
        name={data.Name}
        activeTab={activeTab}
        onSelect={handleTabSelect}
        onSave={handleSave}
        onExport={handleExport}
        onExportWithoutRelations={handleExportWithoutRelations}
        onExportScripts={handleExportScripts}
        onExportPowerDesigner={handleExportPowerDesigner}
        onExit={handleExitRequest}
      />
    </div>
  {/if}

  {#if data != null}
    <ExportAISQL bind:this={exportAISQLModal} entities={data.Entities} intersectionEntities={data.IntersectionEntities}/>
    <ExportPowerDesigner bind:this={exportPowerDesignerModal} entities={data.Entities} intersectionEntities={data.IntersectionEntities}/>
    <section
      class="tab-panel tab-panel--workspace"
      style={`--workspace-ribbon-offset: ${Math.max(ribbonHeight + 18, 18)}px;`}
    >
      {#if activeTab === 'home'}
        <HomeTab project={data} onRefresh={handleRefresh}/>
      {:else if activeTab === 'entities'}
        <EntitiesTab project={data} onSave={handleRefresh} onJumpTo={handleJumpToEntityTab}/>
      {:else if activeTab === 'relations'}
        <RelationsTab entities={data.Entities} onRefresh={handleRefresh} focusEntityId={focusEntityId} onJumpTo={handleJumpToEntityTab}/>
      {:else if activeTab === 'roles'}
        <RolesTab project={data} entities={data.Entities} onRefresh={handleRefresh}/>
      {:else if activeTab === 'flows'}
        <FlowsTab project={data} entities={data.Entities} onRefresh={handleRefresh}/>
      {:else}
        <AttributesTab project={data} onRefresh={handleRefresh} focusEntityId={focusEntityId} onJumpTo={handleJumpToEntityTab}/>
      {/if}
    </section>
  {:else}
    <section class="empty-panel ghost">
      <p>Sin proyecto cargado. Usa el botón para empezar.</p>
    </section>
  {/if}

  {#if $toastState}
    <aside class={`toast toast--${$toastState.tone}`}>
      <span>{$toastState.message}</span>
    </aside>
  {/if}
  
  <Modal
    bind:open={showExitDialog}
    title="Salir del proyecto"
    description="¿Deseas salir del proyecto actual? Los cambios no guardados se perderán si no eliges 'Guardar y salir'."
    confirmLabel="Guardar y salir"
    confirmVariant="primary"
    cancelLabel="Sólo salir"
    busy={exitInProgress}
    onConfirm={saveAndExit}
    onCancel={exitWithoutSave}
  />

  <Modal
    bind:open={showCreateDialog}
    title="Nuevo proyecto"
    description="Define el nombre y la ubicación de tu nuevo esquema de base de datos."
    confirmLabel="Crear Proyecto"
    busy={createBusy}
    errorMessage={createError}
    onConfirm={submitCreateProject}
    onCancel={closeCreateDialog}
  >
    <div class="flex flex-col gap-4 py-4">
      <div class="flex flex-col gap-2">
        <label for="project-name" class="text-sm font-semibold text-foreground-alt">Nombre del proyecto</label>
        <input
          type="text"
          id="project-name"
          class="h-11 rounded-card-sm border border-border-input bg-background px-4 text-sm focus:ring-2 focus:ring-accent focus:outline-none"
          placeholder="Mi nuevo proyecto"
          bind:value={createName}
          disabled={createBusy}
        />
      </div>

      <div class="flex flex-col gap-2">
        <label for="create-path-btn" class="text-sm font-semibold text-foreground-alt">Archivo de destino</label>
        <div class="flex gap-2">
          <button 
            id="create-path-btn"
            class="h-11 rounded-card-sm bg-muted px-4 text-xs font-bold hover:bg-muted/80 transition-colors flex items-center gap-2"
            onclick={selectCreatePath} 
            disabled={createBusy}
          >
            <ButtonIcon name="folder"/>
            <span>Seleccionar</span>
          </button>
          <div class="flex-1 h-11 rounded-card-sm border border-dashed border-border-input bg-muted/30 px-4 flex items-center text-xs text-muted-foreground overflow-hidden whitespace-nowrap">
            {createPath || "Ruta del archivo..."}
          </div>
        </div>
      </div>
    </div>
  </Modal>
</main>

<style>
  main.app-shell {
    position: relative;
    max-width: 1680px;
    margin: 0 auto;
    padding: clamp(0.95rem, 2.2vw, 2rem) clamp(0.9rem, 2.4vw, 2.4rem) clamp(1.2rem, 3vw, 2.2rem);
    color: var(--ink);
    text-align: left;
  }

  .workspace-ribbon-slot {
    margin-bottom: 0.85rem;
  }

  .tab-panel--workspace {
    margin-top: calc(var(--workspace-ribbon-offset) * -1);
    padding-top: var(--workspace-ribbon-offset);
  }

  .empty-panel {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 60vh;
    color: var(--ink-soft);
  }

  .toast {
    position: fixed;
    bottom: 2rem;
    right: 2rem;
    padding: 1rem 1.5rem;
    background: var(--dark);
    color: var(--background);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-lg);
    z-index: 1000;
    font-weight: 600;
    font-size: 0.9rem;
    animation: toast-in 0.3s cubic-bezier(0.19, 1, 0.22, 1);
  }

  .toast--error {
    background: var(--destructive);
  }

  @keyframes toast-in {
    from { transform: translateY(1rem); opacity: 0; }
    to { transform: translateY(0); opacity: 1; }
  }
</style>
