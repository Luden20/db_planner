<script lang="ts">
  import {onMount} from "svelte";
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
  import {showToast, toastState} from "./lib/toast";

  type TabKey = 'home' | 'entities' | 'relations' | 'roles' | 'flows' | 'tertiary';
  type ThemeMode = "light" | "dark";
  const THEME_STORAGE_KEY = "db-planner-theme";

  let data:utils.DbProject | null = null;
  let activeTab:TabKey = 'entities';
  let themeMode: ThemeMode = "light";
  let focusEntityId: number | null = null;
  let ribbonHeight = 0;
  let exportAISQLModal: ExportAISQL | null = null;
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
    } catch (e) {
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
    } catch (e) {
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
    } catch (e) {
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
      showToast("Exportado a Excel sin tipos de relación.", "success");
    } catch (e) {
      const message = e?.error ?? e?.message ?? e;
      showToast(`Error al exportar sin relaciones: ${message}`, "error");
    }
  };
  const handleExportScripts = () => {
    exportAISQLModal?.openDialog();
  };
  const handleRefresh = async () => {
    try {
      data = await GetActualProject();
    } catch (err) {
      const message = err?.error ?? err?.message ?? err;
      showToast(`No se pudo recargar el proyecto: ${message}`, "error");
    }
  };
  const handleTabSelect = (tab:TabKey) => {
    activeTab = tab;
  };

  const applyTheme = (nextTheme: ThemeMode) => {
    themeMode = nextTheme;
    if (typeof document !== "undefined") {
      document.documentElement.dataset.theme = nextTheme;
    }
    try {
      window.localStorage.setItem(THEME_STORAGE_KEY, nextTheme);
    } catch (err) {
      console.warn("No se pudo persistir el tema:", err);
    }
  };

  const toggleTheme = () => {
    applyTheme(themeMode === "dark" ? "light" : "dark");
  };

  const handleJumpToEntityTab = (tab: TabKey, entityId: number | null = null) => {
    focusEntityId = entityId;
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
      showToast(`No se pudo salir del proyecto: ${message}`, "error");
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
      showToast(`No se pudo guardar y salir: ${message}`, "error");
    } finally {
      exitInProgress = false;
      showExitDialog = false;
    }
  };

  const cancelExit = () => {
    showExitDialog = false;
  };

  onMount(() => {
    let nextTheme: ThemeMode = "light";
    try {
      const storedTheme = window.localStorage.getItem(THEME_STORAGE_KEY);
      const prefersDark = typeof window.matchMedia === "function"
        && window.matchMedia("(prefers-color-scheme: dark)").matches;
      if (storedTheme === "light" || storedTheme === "dark") {
        nextTheme = storedTheme;
      } else if (prefersDark) {
        nextTheme = "dark";
      }
    } catch (err) {
      console.warn("No se pudo leer el tema guardado:", err);
    }

    applyTheme(nextTheme);
  });
</script>

<main class="app-shell">
  {#if data === null}
    <div class="shell-utility">
      <button
        class="theme-utility control control--soft control--sm"
        type="button"
        on:click={toggleTheme}
        aria-label={themeMode === "dark" ? "Cambiar a modo claro" : "Cambiar a modo oscuro"}
      >
        <ButtonIcon name={themeMode === "dark" ? "theme-light" : "theme-dark"}/>
        <span>{themeMode === "dark" ? "Cambiar a claro" : "Cambiar a oscuro"}</span>
      </button>
    </div>
    <Hero onOpen={openProject} onCreate={openCreateDialog}/>
  {:else}
    <div class="workspace-ribbon-slot" bind:clientHeight={ribbonHeight}>
      <WorkspaceRibbon
        name={data.Name}
        entityCount={data.Entities.length}
        activeTab={activeTab}
        themeMode={themeMode}
        onSelect={handleTabSelect}
        onToggleTheme={toggleTheme}
        onSave={handleSave}
        onExport={handleExport}
        onExportWithoutRelations={handleExportWithoutRelations}
        onExportScripts={handleExportScripts}
        onExit={handleExitRequest}
      />
    </div>
  {/if}

  {#if data != null}
    <ExportAISQL bind:this={exportAISQLModal} entities={data.Entities} intersectionEntities={data.IntersectionEntities}/>
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

  {#if showExitDialog}
    <div class="modal-backdrop">
      <div class="modal">
        <p class="modal-kicker">Salir del proyecto</p>
        <h2>¿Deseas salir sin guardar?</h2>
        <p class="modal-hint">Puedes salir sin guardar, guardar y salir o cancelar para continuar editando.</p>
        <div class="modal-actions">
          <button class="btn danger" on:click={exitWithoutSave} disabled={exitInProgress}>
            <ButtonIcon name="exit"/>
            <span>Salir sin guardar</span>
          </button>
          <button class="btn primary" on:click={saveAndExit} disabled={exitInProgress}>
            <ButtonIcon name="save"/>
            <span>Guardar y salir</span>
          </button>
          <button class="btn secondary" on:click={cancelExit} disabled={exitInProgress}>
            <ButtonIcon name="close"/>
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
              <ButtonIcon name="folder"/>
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
          <button class="btn secondary" on:click={closeCreateDialog} disabled={createBusy}>
            <ButtonIcon name="close"/>
            <span>Cancelar</span>
          </button>
          <button class="btn primary" on:click={submitCreateProject} disabled={createBusy}>
            <ButtonIcon name="plus"/>
            <span>{createBusy ? "Creando..." : "Crear proyecto"}</span>
          </button>
        </div>
      </div>
    </div>
  {/if}
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

  .shell-utility {
    display: flex;
    justify-content: flex-end;
    margin-bottom: 0.85rem;
  }

  .theme-utility {
    min-width: 10.5rem;
  }

  .workspace-ribbon-slot {
    margin-bottom: 0.85rem;
  }

  .tab-panel {
    margin-top: 0.85rem;
    padding: clamp(0.9rem, 1.6vw, 1.3rem);
    border-radius: var(--radius-lg);
    border: 1px solid var(--border);
    background: var(--panel-surface);
    box-shadow: var(--shadow-lg);
    min-height: 18rem;
  }

  .tab-panel--workspace {
    margin-top: 0;
  }

  .tab-panel--workspace :global(.relations-tab) {
    --relations-sticky-top: var(--workspace-ribbon-offset);
  }

  :global(.btn) {
    --btn-surface: var(--action-neutral-surface);
    --btn-surface-strong: var(--action-neutral-surface-strong);
    --btn-border: var(--action-neutral-border);
    --btn-ink: var(--action-neutral-ink);
    --btn-shadow: 0 14px 24px color-mix(in srgb, var(--ink) 10%, transparent);
    --btn-glow: color-mix(in srgb, var(--ink) 12%, transparent);
    --btn-highlight: inset 0 1px 0 color-mix(in srgb, white 30%, transparent);
    position: relative;
    border: 1px solid var(--btn-border);
    border-radius: 1rem;
    min-height: 2.95rem;
    padding: 0.72rem 1.08rem;
    background: linear-gradient(
      180deg,
      var(--btn-surface),
      var(--btn-surface-strong)
    );
    color: var(--btn-ink);
    font-weight: 760;
    letter-spacing: -0.01em;
    line-height: 1;
    box-shadow: var(--btn-shadow), var(--btn-highlight);
    cursor: pointer;
    transition:
      transform 150ms cubic-bezier(.19, 1, .22, 1),
      box-shadow 180ms ease,
      background 180ms ease,
      border-color 180ms ease,
      color 180ms ease;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
  }

  :global(.btn .button-glyph) {
    opacity: 0.92;
  }

  :global(.btn.primary) {
    --btn-surface: var(--action-strong-surface);
    --btn-surface-strong: var(--action-strong-surface-strong);
    --btn-border: var(--action-strong-border);
    --btn-ink: var(--action-strong-ink);
    --btn-shadow: 0 18px 30px color-mix(in srgb, var(--ink) 18%, transparent);
    --btn-glow: color-mix(in srgb, var(--accent) 18%, transparent);
  }

  :global(.btn.accent) {
    --btn-surface: color-mix(in srgb, var(--accent) 26%, var(--action-neutral-surface));
    --btn-surface-strong: color-mix(in srgb, var(--accent) 38%, var(--action-neutral-surface-strong));
    --btn-border: color-mix(in srgb, var(--accent) 28%, var(--border));
    --btn-ink: var(--accent-strong);
    --btn-shadow: 0 14px 24px color-mix(in srgb, var(--accent) 12%, transparent);
    --btn-glow: color-mix(in srgb, var(--accent) 22%, transparent);
  }

  :global(.btn.edit) {
    --btn-surface: var(--action-edit-surface);
    --btn-surface-strong: var(--action-edit-surface-strong);
    --btn-border: var(--action-edit-border);
    --btn-ink: var(--action-edit-ink);
    --btn-shadow: 0 16px 28px color-mix(in srgb, var(--edit) 22%, transparent);
    --btn-glow: color-mix(in srgb, var(--edit) 26%, transparent);
  }

  :global(.btn.success) {
    --btn-surface: var(--action-success-surface);
    --btn-surface-strong: var(--action-success-surface-strong);
    --btn-border: var(--action-success-border);
    --btn-ink: var(--action-success-ink);
    --btn-shadow: 0 14px 24px color-mix(in srgb, var(--success) 14%, transparent);
    --btn-glow: color-mix(in srgb, var(--success) 22%, transparent);
  }

  :global(.btn:disabled) {
    opacity: 0.6;
    cursor: not-allowed;
    transform: none;
    box-shadow: none;
  }

  :global(.btn.danger) {
    --btn-surface: var(--action-danger-surface);
    --btn-surface-strong: var(--action-danger-surface-strong);
    --btn-border: var(--action-danger-border);
    --btn-ink: var(--action-danger-ink);
    --btn-shadow: 0 14px 26px color-mix(in srgb, var(--danger) 18%, transparent);
    --btn-glow: color-mix(in srgb, var(--danger) 26%, transparent);
  }

  :global(.btn:hover) {
    transform: translateY(-2px);
    border-color: color-mix(in srgb, var(--btn-border) 78%, white 22%);
    box-shadow: 0 18px 28px var(--btn-glow), inset 0 1px 0 color-mix(in srgb, white 34%, transparent);
  }

  :global(.btn:hover .button-glyph),
  :global(.btn:focus-visible .button-glyph) {
    transform: translateY(-1px) scale(1.06);
  }

  :global(.btn:active) {
    transform: translateY(1px) scale(0.985);
    box-shadow: 0 9px 16px color-mix(in srgb, var(--btn-glow) 88%, transparent), inset 0 1px 0 color-mix(in srgb, white 22%, transparent);
  }

  :global(.btn:active .button-glyph) {
    transform: scale(0.96);
  }

  :global(.btn.secondary) {
    --btn-surface: var(--action-neutral-surface);
    --btn-surface-strong: var(--action-neutral-surface-strong);
    --btn-border: var(--action-neutral-border);
    --btn-ink: var(--action-neutral-ink);
  }

  :global(.empty-panel) {
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--ink-soft);
    font-size: 0.96rem;
    min-height: 11rem;
    border: 1px dashed var(--line-soft);
    border-radius: calc(var(--radius-md) - 2px);
    background: color-mix(in srgb, var(--surface) 78%, transparent);
  }

  :global(.empty-panel.ghost) {
    margin-top: 1rem;
  }

  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: var(--overlay-scrim);
    display: grid;
    place-items: center;
    padding: 1.2rem;
    z-index: var(--layer-modal);
    backdrop-filter: blur(10px);
  }

  .modal {
    width: min(520px, 100%);
    background: var(--popover-surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    padding: 1.35rem 1.45rem;
    box-shadow: var(--shadow-lg);
  }

  .modal-kicker {
    margin: 0;
    text-transform: uppercase;
    letter-spacing: 0.16em;
    color: var(--accent);
    font-size: 0.74rem;
    font-weight: 800;
  }

  .modal h2 {
    margin: 0.5rem 0 0.4rem;
    font-size: 2rem;
    line-height: 0.98;
  }

  .modal-hint {
    margin: 0 0 1rem;
    color: var(--ink-soft);
    line-height: 1.55;
  }

  .modal-actions {
    display: flex;
    gap: 0.7rem;
    flex-wrap: wrap;
  }

  .modal-actions .btn {
    min-width: 8.75rem;
  }

  .field {
    display: grid;
    gap: 0.5rem;
    color: var(--ink-soft);
    font-size: 0.92rem;
    margin: 0.9rem 0 0;
  }

  .field input {
    width: 100%;
    border-radius: 1rem;
    border: 1px solid var(--border);
    background: var(--field-surface);
    color: var(--ink);
    padding: 0.85rem 0.95rem;
    font-size: 0.95rem;
    outline: none;
    transition: border 140ms ease, box-shadow 140ms ease, background 140ms ease;
  }

  .field input:focus {
    border-color: var(--focus-border);
    box-shadow: var(--focus-ring);
    background: var(--field-surface-focus);
  }

  .file-row {
    display: flex;
    gap: 0.75rem;
    align-items: center;
    flex-wrap: wrap;
  }

  .path-label {
    color: var(--ink-faint);
    word-break: break-all;
    font-size: 0.84rem;
  }

  .form-error {
    margin: 0.75rem 0 0;
    color: var(--danger);
    font-weight: 600;
  }

  .toast {
    position: fixed;
    top: 1.2rem;
    right: 1.2rem;
    z-index: var(--layer-toast);
    min-width: 15rem;
    max-width: min(22rem, calc(100vw - 2rem));
    padding: 0.95rem 1rem;
    border-radius: calc(var(--radius-sm) - 2px);
    border: 1px solid var(--border);
    color: var(--ink);
    background: var(--popover-surface);
    box-shadow: var(--shadow-sm);
    backdrop-filter: blur(12px);
    animation: toast-in 180ms cubic-bezier(.19,1,.22,1);
  }

  .toast--info {
    background: linear-gradient(
      135deg,
      color-mix(in srgb, var(--accent) 14%, var(--surface-strong)),
      color-mix(in srgb, var(--accent) 8%, var(--surface))
    );
    border-color: color-mix(in srgb, var(--accent) 22%, var(--border));
  }

  .toast--success {
    background: linear-gradient(
      135deg,
      color-mix(in srgb, var(--success) 18%, var(--surface-strong)),
      color-mix(in srgb, var(--success) 10%, var(--surface))
    );
    border-color: color-mix(in srgb, var(--success) 24%, var(--border));
  }

  .toast--error {
    background: linear-gradient(
      135deg,
      color-mix(in srgb, var(--danger) 18%, var(--surface-strong)),
      color-mix(in srgb, var(--danger) 10%, var(--surface))
    );
    border-color: color-mix(in srgb, var(--danger) 24%, var(--border));
  }

  @keyframes toast-in {
    from {
      opacity: 0;
      transform: translate3d(1rem, -0.5rem, 0);
    }
    to {
      opacity: 1;
      transform: translate3d(0, 0, 0);
    }
  }

  @media (max-width: 720px) {
    main.app-shell {
      padding: 0.9rem 0.75rem 1.6rem;
    }

    .toast {
      top: 1rem;
      right: 1rem;
      left: 1rem;
      min-width: 0;
      max-width: none;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    :global(.btn) {
      transition: none;
    }
  }
</style>
