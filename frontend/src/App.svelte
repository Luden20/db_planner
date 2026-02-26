<script lang="ts">
  import {
    CreateExcelPath,
    CreateNew,
    CreateProjectJSONPath,
    EjectProject,
    ExportToExcel,
    GetActualProject,
    OpenPath,
    PickProjectJSON,
    Save
  } from '../wailsjs/go/main/App.js';
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
  let showExitDialog = false;
  let exitInProgress = false;
  let showCreateDialog = false;
  let createName = "";
  let createPath = "";
  let createBusy = false;
  let createError = "";

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
    } catch (e) {
      const message = e?.error ?? e?.message ?? e;
      alert("Dialog error:" + message);
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
    } catch (e) {
      const message = e?.error ?? e?.message ?? e ?? "Error desconocido";
      createError = `${message}`;
    } finally {
      createBusy = false;
    }
  };
   const handleSave:  () => Promise<void> = async () => {
    try{
      alert("Guardando");
      await Save();
      alert("Guardado");
    }catch(e){
      alert("Error en guardado");
    }
  };
  const handleExport = async () => {
    try {
      const path = await CreateExcelPath();
      if (!path) {
        return;
      }
      await ExportToExcel(path);
      alert("Exportado a Excel.");
    } catch (e) {
      const message = e?.error ?? e?.message ?? e;
      alert(`Error al exportar: ${message}`);
    }
  };
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

  const handleExitRequest = () => {
    showExitDialog = true;
  };

  const resetProjectState = () => {
    data = null;
    activeTab = 'entities';
  };

  const exitWithoutSave = async () => {
    exitInProgress = true;
    try {
      await EjectProject();
      resetProjectState();
    } catch (err) {
      const message = err?.error ?? err?.message ?? err;
      alert(`No se pudo salir del proyecto: ${message}`);
    } finally {
      exitInProgress = false;
      showExitDialog = false;
    }
  };

  const saveAndExit = async () => {
    exitInProgress = true;
    try {
      await Save();
      await EjectProject();
      resetProjectState();
    } catch (err) {
      const message = err?.error ?? err?.message ?? err;
      alert(`No se pudo guardar y salir: ${message}`);
    } finally {
      exitInProgress = false;
      showExitDialog = false;
    }
  };

  const cancelExit = () => {
    showExitDialog = false;
  };
</script>

<main class="app-shell">
  {#if data === null}
    <Hero onOpen={openProject} onCreate={openCreateDialog}/>
  {:else}
    <ActionsBar onSave={handleSave} onExport={handleExport} onExit={handleExitRequest}/>
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

  {#if showExitDialog}
    <div class="modal-backdrop">
      <div class="modal">
        <p class="modal-kicker">Salir del proyecto</p>
        <h2>¿Deseas salir sin guardar?</h2>
        <p class="modal-hint">Puedes salir sin guardar, guardar y salir o cancelar para continuar editando.</p>
        <div class="modal-actions">
          <button class="btn danger" on:click={exitWithoutSave} disabled={exitInProgress}>
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M5.25 3A2.25 2.25 0 0 0 3 5.25v13.5A2.25 2.25 0 0 0 5.25 21h8.5A2.25 2.25 0 0 0 16 18.75V15.5a.75.75 0 0 0-1.5 0v3.25c0 .414-.336.75-.75.75h-8a.75.75 0 0 1-.75-.75V5.25c0-.414.336-.75.75-.75h8c.414 0 .75.336.75.75V9.5a.75.75 0 0 0 1.5 0V5.25A2.25 2.25 0 0 0 13.75 3h-8.5Zm11.53 6.22a.75.75 0 0 0-1.06 1.06l1.97 1.97H11.5a.75.75 0 0 0 0 1.5h6.19l-1.97 1.97a.75.75 0 1 0 1.06 1.06l3.25-3.25a.75.75 0 0 0 0-1.06l-3.25-3.25Z"/>
            </svg>
            <span>Salir sin guardar</span>
          </button>
          <button class="btn primary" on:click={saveAndExit} disabled={exitInProgress}>
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M5 4a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V8.414a2 2 0 0 0-.586-1.414l-3.414-3.414A2 2 0 0 0 15.586 3H5Zm10 2.5V8H5V6a1 1 0 0 1 1-1h8a1 1 0 0 1 1 1Zm-10 6h14V18a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1v-5.5Z"/>
            </svg>
            <span>Guardar y salir</span>
          </button>
          <button class="btn secondary" on:click={cancelExit} disabled={exitInProgress}>
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M12 4a8 8 0 1 0 0 16 8 8 0 0 0 0-16Zm2.53 4.47a.75.75 0 0 1 0 1.06L13.06 11l1.47 1.47a.75.75 0 0 1-1.06 1.06L12 12.06l-1.47 1.47a.75.75 0 1 1-1.06-1.06L10.94 11 9.47 9.53a.75.75 0 0 1 1.06-1.06L12 9.94l1.47-1.47a.75.75 0 0 1 1.06 0Z"/>
            </svg>
            <span>Cancelar</span>
          </button>
        </div>
      </div>
    </div>
  {/if}

  {#if showCreateDialog}
    <div class="modal-backdrop">
      <div class="modal">
        <p class="modal-kicker">Nuevo proyecto</p>
        <h2>Define el proyecto y el archivo</h2>
        <div class="field">
          <label for="project-name">Nombre del proyecto</label>
          <input
            type="text"
            id="project-name"
            placeholder="Mi proyecto"
            bind:value={createName}
            disabled={createBusy}
          />
        </div>

        <div class="field">
          <p class="field-label">Archivo JSON</p>
          <div class="file-row">
            <button class="btn secondary" on:click={selectCreatePath} disabled={createBusy}>
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <path d="M7.25 4A2.25 2.25 0 0 0 5 6.25v11.5A2.25 2.25 0 0 0 7.25 20h9.5A2.25 2.25 0 0 0 19 17.75V9.06a1.75 1.75 0 0 0-.51-1.24l-3.31-3.31A1.75 1.75 0 0 0 13.94 4h-6.7Zm0 1.5h6.44c.1 0 .2.04.28.12l3.66 3.66c.07.08.12.18.12.28v8.19c0 .41-.34.75-.75.75h-9.5a.75.75 0 0 1-.75-.75V5.75c0-.41.34-.75.75-.75Zm1.5 3.5a.75.75 0 0 0 0 1.5h6.5a.75.75 0 0 0 0-1.5h-6.5Zm0 3a.75.75 0 0 0 0 1.5h6.5a.75.75 0 0 0 0-1.5h-6.5Zm0 3a.75.75 0 0 0 0 1.5h4.5a.75.75 0 0 0 0-1.5h-4.5Z"/>
              </svg>
              <span>Seleccionar archivo</span>
            </button>
            <span class="path-label">
              {createPath || "Ningún archivo seleccionado"}
            </span>
          </div>
        </div>

        {#if createError}
          <p class="form-error">{createError}</p>
        {/if}

        <div class="modal-actions">
          <button class="btn secondary" on:click={closeCreateDialog} disabled={createBusy}>Cancelar</button>
          <button class="btn primary" on:click={submitCreateProject} disabled={createBusy}>
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M12 4a.75.75 0 0 0-.75.75V11H5.75a.75.75 0 0 0 0 1.5h5.5v6.25a.75.75 0 0 0 1.5 0V12.5h5.5a.75.75 0 0 0 0-1.5h-5.5V4.75A.75.75 0 0 0 12 4Z"/>
            </svg>
            <span>{createBusy ? "Creando..." : "Crear proyecto"}</span>
          </button>
        </div>
      </div>
    </div>
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
    padding: 12px 16px;
    font-weight: 700;
    cursor: pointer;
    transition: transform 150ms ease, box-shadow 180ms ease, background 180ms ease;
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  :global(.btn svg) {
    width: 14px;
    height: 14px;
    fill: currentColor;
    flex-shrink: 0;
    opacity: 0.92;
  }

  :global(.btn.primary) {
    background: linear-gradient(120deg, #5ad1ff, #6287f6);
    color: #0b1a30;
    box-shadow: 0 12px 30px rgba(82, 158, 255, 0.35);
  }

  :global(.btn:disabled) {
    opacity: 0.6;
    cursor: not-allowed;
    transform: none;
    box-shadow: none;
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

  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(7, 13, 26, 0.7);
    display: grid;
    place-items: center;
    padding: 18px;
    z-index: 10;
    backdrop-filter: blur(2px);
  }

  .modal {
    width: min(520px, 100%);
    background: #0d1627;
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 14px;
    padding: 20px 22px;
    box-shadow: 0 22px 50px rgba(0, 0, 0, 0.38);
  }

  .modal-kicker {
    margin: 0;
    text-transform: uppercase;
    letter-spacing: 1px;
    color: #9ab5e4;
    font-size: 12px;
  }

  .modal h2 {
    margin: 8px 0 6px;
    font-size: 22px;
  }

  .modal-hint {
    margin: 0 0 18px;
    color: #cfd9e9;
    opacity: 0.85;
  }

  .modal-actions {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }

  .modal-actions .btn {
    min-width: 140px;
  }

  .field {
    display: grid;
    gap: 8px;
    color: #cfd9e9;
    font-size: 14px;
    margin: 12px 0 0;
  }

  .field input {
    width: 100%;
    border-radius: 12px;
    border: 1px solid rgba(255, 255, 255, 0.14);
    background: rgba(255, 255, 255, 0.04);
    color: #e8edf7;
    padding: 12px 12px;
    font-size: 14px;
    outline: none;
    transition: border 140ms ease, box-shadow 140ms ease;
  }

  .field input:focus {
    border-color: rgba(90, 209, 255, 0.8);
    box-shadow: 0 0 0 2px rgba(90, 209, 255, 0.22);
  }

  .file-row {
    display: flex;
    gap: 10px;
    align-items: center;
    flex-wrap: wrap;
  }

  .path-label {
    color: #cfd9e9;
    opacity: 0.85;
    word-break: break-all;
    font-size: 13px;
  }

  .form-error {
    margin: 10px 0 0;
    color: #ffb4a2;
    font-weight: 600;
  }

  @media (max-width: 720px) {
    main.app-shell {
      padding: 22px 16px 36px;
    }
  }
</style>
