<script lang="ts">
  import {fade} from "svelte/transition";
  import {onDestroy} from "svelte";
  import ButtonIcon from "../ButtonIcon.svelte";
  import {GenerateSQLFromEntities, GetAISettings, SaveOpenAIAPIKey} from "../../../wailsjs/go/main/App";
  import {utils} from "../../../wailsjs/go/models";
  import {showToast} from "../../lib/toast";

  export let entities: utils.Entity[] = [];
  export let intersectionEntities: utils.IntersectionEntity[] = [];

  type ViewState = "select" | "loading" | "result" | "error";
  type DatabaseOption = {
    value: string;
    label: string;
    hint: string;
  };
  type PipelineStage = {
    id: string;
    title: string;
    detail: string;
  };

  const databaseOptions: DatabaseOption[] = [
    {value: "PostgreSQL", label: "PostgreSQL", hint: "DDL con claves foraneas y tipos comunes"},
    {value: "MySQL", label: "MySQL", hint: "Enfoque compatible con InnoDB"},
    {value: "SQL Server", label: "SQL Server", hint: "Sintaxis T-SQL para tablas y constraints"},
    {value: "SQLite", label: "SQLite", hint: "Esquema compacto para prototipos locales"},
  ];

  const pipelineStages: PipelineStage[] = [
    {id: "schema", title: "Empaquetando esquema", detail: "Consolidando tablas, cruces y relaciones visibles."},
    {id: "request", title: "Consultando modelo", detail: "Enviando el subconjunto del proyecto al motor configurado."},
    {id: "assembly", title: "Ensamblando SQL", detail: "Normalizando salida, limpiando bloques y verificando contenido util."},
  ];

  const baseAISettings = () => new utils.AISettings({HasAPIKey: false, OpenAIModel: "gpt-5-mini"});

  let isOpen = false;
  let viewState: ViewState = "select";
  let aiSettings: utils.AISettings = baseAISettings();
  let generatedResult: utils.SQLGenerationResult | null = null;
  let selectedIds = new Set<number>();
  let selectedIntersectionIds = new Set<number>();
  let database = databaseOptions[0].value;
  let apiKeyDraft = "";
  let settingsExpanded = false;
  let settingsBusy = false;
  let generateBusy = false;
  let inlineErrorMessage = "";
  let pipelineErrorMessage = "";
  let loadingStageIndex = 0;
  let loadingProgress = 0;
  let loadingPulse = "Preparando consulta";
  let stageTimer: ReturnType<typeof setInterval> | null = null;
  let progressTimer: ReturnType<typeof setInterval> | null = null;
  let pulseTimer: ReturnType<typeof setInterval> | null = null;

  const pulseMessages = [
    "Preparando consulta",
    "Leyendo definiciones",
    "Esperando SQL",
    "Puliendo salida",
  ];

  $: totalSelectable = entities.length + intersectionEntities.length;
  $: selectedCount = selectedIds.size + selectedIntersectionIds.size;
  $: canGenerate = aiSettings.HasAPIKey && selectedCount > 0 && database.trim().length > 0 && !generateBusy;
  $: activeStage = pipelineStages[Math.min(loadingStageIndex, pipelineStages.length - 1)];

  onDestroy(() => {
    stopLoadingPresentation();
  });

  const clearMessages = () => {
    inlineErrorMessage = "";
    pipelineErrorMessage = "";
  };

  const loadSettings = async () => {
    try {
      aiSettings = await GetAISettings();
    } catch (err) {
      const message = err?.error ?? err?.message ?? err ?? "No se pudo leer la configuracion de IA.";
      aiSettings = baseAISettings();
      inlineErrorMessage = `${message}`;
    }
  };

  const resetState = async () => {
    clearMessages();
    generatedResult = null;
    viewState = "select";
    apiKeyDraft = "";
    settingsExpanded = false;
    database = databaseOptions[0].value;
    selectedIds = new Set();
    selectedIntersectionIds = new Set();
    stopLoadingPresentation();
    await loadSettings();
  };

  export const openDialog = async () => {
    isOpen = true;
    await resetState();
  };

  export const closeDialog = () => {
    if (settingsBusy || generateBusy) {
      return;
    }
    isOpen = false;
  };

  const returnToSelection = () => {
    clearMessages();
    generatedResult = null;
    viewState = "select";
  };

  const toggleAll = () => {
    if (selectedCount === totalSelectable) {
      selectedIds = new Set();
      selectedIntersectionIds = new Set();
      return;
    }

    selectedIds = new Set(entities.map((entity) => entity.Id));
    selectedIntersectionIds = new Set(intersectionEntities.map((item) => item.Entity.Id));
  };

  const toggleEntity = (id: number) => {
    if (selectedIds.has(id)) {
      selectedIds.delete(id);
    } else {
      selectedIds.add(id);
    }
    selectedIds = new Set(selectedIds);
  };

  const toggleIntersection = (id: number) => {
    if (selectedIntersectionIds.has(id)) {
      selectedIntersectionIds.delete(id);
    } else {
      selectedIntersectionIds.add(id);
    }
    selectedIntersectionIds = new Set(selectedIntersectionIds);
  };

  const saveApiKey = async () => {
    if (!apiKeyDraft.trim()) {
      inlineErrorMessage = "Ingresa una API key para guardarla.";
      return;
    }

    settingsBusy = true;
    inlineErrorMessage = "";
    try {
      aiSettings = await SaveOpenAIAPIKey(apiKeyDraft.trim());
      apiKeyDraft = "";
      settingsExpanded = false;
      showToast("API key guardada.", "success");
    } catch (err) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      inlineErrorMessage = `${message}`;
    } finally {
      settingsBusy = false;
    }
  };

  const clearApiKey = async () => {
    settingsBusy = true;
    inlineErrorMessage = "";
    try {
      aiSettings = await SaveOpenAIAPIKey("");
      apiKeyDraft = "";
      settingsExpanded = false;
      showToast("API key eliminada.", "success");
    } catch (err) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      inlineErrorMessage = `${message}`;
    } finally {
      settingsBusy = false;
    }
  };

  const startLoadingPresentation = () => {
    stopLoadingPresentation();
    loadingStageIndex = 0;
    loadingProgress = 7;
    loadingPulse = pulseMessages[0];

    progressTimer = setInterval(() => {
      loadingProgress = Math.min(94, loadingProgress + 1 + Math.random() * 4);
    }, 180);

    stageTimer = setInterval(() => {
      loadingStageIndex = Math.min(pipelineStages.length - 1, loadingStageIndex + 1);
    }, 1450);

    let pulseIndex = 0;
    pulseTimer = setInterval(() => {
      pulseIndex = (pulseIndex + 1) % pulseMessages.length;
      loadingPulse = pulseMessages[pulseIndex];
    }, 1100);
  };

  const stopLoadingPresentation = () => {
    if (progressTimer) {
      clearInterval(progressTimer);
      progressTimer = null;
    }
    if (stageTimer) {
      clearInterval(stageTimer);
      stageTimer = null;
    }
    if (pulseTimer) {
      clearInterval(pulseTimer);
      pulseTimer = null;
    }
  };

  const finishLoadingPresentation = () => {
    stopLoadingPresentation();
    loadingStageIndex = pipelineStages.length - 1;
    loadingProgress = 100;
    loadingPulse = "Salida lista";
  };

  const handleGenerate = async () => {
    if (!canGenerate) {
      if (!aiSettings.HasAPIKey) {
        settingsExpanded = true;
      }
      return;
    }

    generateBusy = true;
    clearMessages();
    generatedResult = null;
    viewState = "loading";
    startLoadingPresentation();

    try {
      generatedResult = await GenerateSQLFromEntities(
        Array.from(selectedIds),
        Array.from(selectedIntersectionIds),
        database,
      );

      if (!generatedResult?.SQL?.trim()) {
        generatedResult = null;
        throw new Error("La IA no devolvio SQL.");
      }

      finishLoadingPresentation();
      await new Promise((resolve) => setTimeout(resolve, 320));
      viewState = "result";
    } catch (err) {
      stopLoadingPresentation();
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      pipelineErrorMessage = `${message}`;
      viewState = "error";
    } finally {
      generateBusy = false;
    }
  };

  const retryFromError = async () => {
    viewState = "loading";
    pipelineErrorMessage = "";
    await handleGenerate();
  };

  const copySQL = async () => {
    if (!generatedResult?.SQL?.trim()) {
      return;
    }
    await navigator.clipboard.writeText(generatedResult.SQL);
    showToast("SQL copiado al portapapeles.", "success");
  };

  const copyJSON = async () => {
    if (!generatedResult?.ExportJSON?.trim()) {
      return;
    }
    await navigator.clipboard.writeText(generatedResult.ExportJSON);
    showToast("JSON copiado al portapapeles.", "success");
  };
</script>

{#if isOpen}
  <div class="modal-backdrop" role="presentation" on:click={closeDialog}>
    <div class="modal-shell" role="dialog" aria-modal="true" aria-labelledby="ai-export-title" on:click|stopPropagation>
      <header class="modal-head">
        <div>
          <p class="modal-kicker">@ Exportar con IA</p>
          <h2 id="ai-export-title">
            {#if viewState === "select"}
              Selecciona tablas y genera SQL
            {:else if viewState === "loading"}
              Generando SQL
            {:else if viewState === "result"}
              Resultado listo
            {:else}
              No se pudo generar SQL
            {/if}
          </h2>
          <p class="modal-hint">
            {#if viewState === "select"}
              Elige el subconjunto del modelo, define el motor y lanza una corrida limpia.
            {:else if viewState === "loading"}
              El pipeline esta procesando tu esquema. La salida aparecera en una vista separada.
            {:else if viewState === "result"}
              Revisa, copia y vuelve a correr si quieres afinar otro subconjunto del proyecto.
            {:else}
              La generacion se interrumpio. Puedes volver a la seleccion o reintentar sin perder contexto.
            {/if}
          </p>
        </div>
        <button class="control control--icon control--soft" type="button" on:click={closeDialog} aria-label="Cerrar" disabled={generateBusy || settingsBusy}>
          <ButtonIcon name="close"/>
        </button>
      </header>

      <div class="modal-body">
        {#if viewState === "select"}
          <section class="state-view state-view--select" in:fade={{duration: 180}}>
            <section class="top-grid">
              <div class="config-card">
                <div class="config-card__head">
                  <div>
                    <p class="label">Destino SQL</p>
                    <p class="muted">El motor cambia el dialecto y el estilo del script generado.</p>
                  </div>
                  <span class={`status-chip ${aiSettings.HasAPIKey ? 'status-chip--ok' : 'status-chip--warn'}`}>
                    {#if aiSettings.HasAPIKey}
                      API key lista
                    {:else}
                      Falta API key
                    {/if}
                  </span>
                </div>

                <label class="field">
                  <span>Base de datos</span>
                  <select bind:value={database}>
                    {#each databaseOptions as option}
                      <option value={option.value}>{option.label}</option>
                    {/each}
                  </select>
                </label>

                <p class="selection-hint">
                  {databaseOptions.find((option) => option.value === database)?.hint}
                </p>

                <div class="config-actions">
                  <button class="control control--sm control--ghost" type="button" on:click={() => settingsExpanded = !settingsExpanded}>
                    <ButtonIcon name="spark"/>
                    <span>{settingsExpanded ? "Ocultar configuracion" : "Configurar API key"}</span>
                  </button>
                  <span class="model-note">Modelo: {aiSettings.OpenAIModel}</span>
                </div>

                {#if settingsExpanded}
                  <div class="settings-panel">
                    <label class="field">
                      <span>OpenAI API key</span>
                      <input
                        type="password"
                        bind:value={apiKeyDraft}
                        placeholder={aiSettings.HasAPIKey ? "Pega una nueva key o usa quitar key" : "sk-..."}
                        autocomplete="off"
                        spellcheck="false"
                        disabled={settingsBusy}
                      />
                    </label>
                    <p class="settings-note">Se guarda en la configuracion local del usuario, fuera del archivo del proyecto.</p>
                    <div class="settings-actions">
                      <button class="control control--sm control--accent" type="button" on:click={saveApiKey} disabled={settingsBusy}>
                        <ButtonIcon name="save"/>
                        <span>{settingsBusy ? "Guardando..." : "Guardar key"}</span>
                      </button>
                      {#if aiSettings.HasAPIKey}
                        <button class="control control--sm control--ghost" type="button" on:click={clearApiKey} disabled={settingsBusy}>
                          <ButtonIcon name="trash"/>
                          <span>Quitar key</span>
                        </button>
                      {/if}
                    </div>
                  </div>
                {/if}
              </div>

              <div class="summary-card">
                <p class="label">Pipeline</p>
                <div class="summary-stats">
                  <div>
                    <strong>{selectedCount}</strong>
                    <span>tablas marcadas</span>
                  </div>
                  <div>
                    <strong>{entities.length}</strong>
                    <span>fuertes</span>
                  </div>
                  <div>
                    <strong>{intersectionEntities.length}</strong>
                    <span>intersecciones</span>
                  </div>
                </div>
                <div class="summary-callout">
                  <span class="summary-callout__step">01</span>
                  <div>
                    <strong>Seleccion</strong>
                    <p>Elige solo lo necesario. Menos ruido produce SQL mas consistente.</p>
                  </div>
                </div>
              </div>
            </section>

            <section class="selection-card">
              <div class="selection-card__head">
                <div>
                  <p class="label">Tablas</p>
                  <p class="muted">Marca solo las necesarias para que el prompt sea mas preciso.</p>
                </div>
                <button class="control control--sm control--ghost" type="button" on:click={toggleAll}>
                  <ButtonIcon name="check"/>
                  <span>{selectedCount === totalSelectable ? "Desmarcar todo" : "Marcar todo"}</span>
                </button>
              </div>

              <div class="selection-list">
                {#if entities.length > 0}
                  <div class="list-section">
                    <p class="list-title">Entidades fuertes</p>
                    {#each entities as entity}
                      <label class={`entity-row ${selectedIds.has(entity.Id) ? 'entity-row--selected' : ''}`}>
                        <input
                          type="checkbox"
                          checked={selectedIds.has(entity.Id)}
                          on:change={() => toggleEntity(entity.Id)}
                        />
                        <div>
                          <strong>{entity.Name}</strong>
                          <span>{entity.Description || "Sin descripcion."}</span>
                        </div>
                      </label>
                    {/each}
                  </div>
                {/if}

                {#if intersectionEntities.length > 0}
                  <div class="list-section">
                    <p class="list-title">Tablas de interseccion</p>
                    {#each intersectionEntities as item}
                      <label class={`entity-row ${selectedIntersectionIds.has(item.Entity.Id) ? 'entity-row--selected' : ''}`}>
                        <input
                          type="checkbox"
                          checked={selectedIntersectionIds.has(item.Entity.Id)}
                          on:change={() => toggleIntersection(item.Entity.Id)}
                        />
                        <div>
                          <strong>{item.Entity.Name}</strong>
                          <span>{item.Entity.Description || "Sin descripcion."}</span>
                        </div>
                      </label>
                    {/each}
                  </div>
                {/if}
              </div>
            </section>

            {#if inlineErrorMessage}
              <div class="notice notice--error">
                <strong>Configuracion incompleta</strong>
                <p>{inlineErrorMessage}</p>
              </div>
            {/if}
          </section>
        {:else if viewState === "loading"}
          <section class="state-view state-view--loading" in:fade={{duration: 180}}>
            <div class="loading-hero">
              <div class="loader-orbit" aria-hidden="true">
                <span class="loader-orbit__ring"></span>
                <span class="loader-orbit__ring loader-orbit__ring--delayed"></span>
                <span class="loader-orbit__core"></span>
              </div>
              <div class="loading-copy">
                <p class="label">Pipeline en curso</p>
                <h3>{activeStage.title}</h3>
                <p>{activeStage.detail}</p>
                <span class="loading-pulse">{loadingPulse}</span>
              </div>
            </div>

            <div class="progress-shell" aria-hidden="true">
              <div class="progress-bar">
                <span class="progress-bar__fill" style={`transform: scaleX(${loadingProgress / 100});`}></span>
              </div>
              <span class="progress-value">{Math.round(loadingProgress)}%</span>
            </div>

            <div class="pipeline-list">
              {#each pipelineStages as stage, index}
                <article class={`pipeline-item ${index < loadingStageIndex ? 'pipeline-item--done' : ''} ${index === loadingStageIndex ? 'pipeline-item--active' : ''}`}>
                  <span class="pipeline-item__index">0{index + 1}</span>
                  <div>
                    <strong>{stage.title}</strong>
                    <p>{stage.detail}</p>
                  </div>
                </article>
              {/each}
            </div>
          </section>
        {:else if viewState === "result" && generatedResult}
          <section class="state-view state-view--result" in:fade={{duration: 180}}>
            <div class="result-banner">
              <div>
                <p class="label">Salida final</p>
                <h3>SQL generado para {generatedResult.Database}</h3>
                <p class="muted">Subconjunto: {selectedCount} tablas | Modelo: {generatedResult.Model}</p>
              </div>
              <div class="result-banner__actions">
                <button class="control control--sm control--ghost" type="button" on:click={copyJSON}>
                  <ButtonIcon name="copy"/>
                  <span>Copiar JSON</span>
                </button>
                <button class="control control--sm control--accent" type="button" on:click={copySQL}>
                  <ButtonIcon name="copy"/>
                  <span>Copiar SQL</span>
                </button>
              </div>
            </div>

            <div class="result-grid">
              <aside class="result-sidecard">
                <p class="label">Resumen</p>
                <div class="result-metrics">
                  <div>
                    <strong>{selectedCount}</strong>
                    <span>tablas</span>
                  </div>
                  <div>
                    <strong>{generatedResult.Database}</strong>
                    <span>motor</span>
                  </div>
                </div>
                <div class="summary-callout">
                  <span class="summary-callout__step">03</span>
                  <div>
                    <strong>Resultado</strong>
                    <p>Si quieres comparar variantes, vuelve a la seleccion sin cerrar el flujo.</p>
                  </div>
                </div>
              </aside>

              <section class="result-card">
                <div class="result-card__head">
                  <div>
                    <p class="label">Script</p>
                    <p class="muted">Revisa el DDL antes de llevarlo a produccion.</p>
                  </div>
                </div>
                <pre class="sql-output"><code>{generatedResult.SQL}</code></pre>
              </section>
            </div>
          </section>
        {:else if viewState === "error"}
          <section class="state-view state-view--error" in:fade={{duration: 180}}>
            <div class="error-stage">
              <div class="error-stage__badge">
                <ButtonIcon name="close"/>
              </div>
              <p class="label">Pipeline interrumpido</p>
              <h3>No pudimos cerrar la corrida</h3>
              <p class="error-stage__copy">{pipelineErrorMessage || "La generacion se detuvo antes de producir SQL utilizable."}</p>
              <div class="notice notice--soft">
                <strong>Que puedes hacer ahora</strong>
                <p>Reintentar con la misma seleccion o volver al paso anterior para ajustar tablas, motor o API key.</p>
              </div>
            </div>
          </section>
        {/if}
      </div>

      <footer class="modal-footer">
        {#if viewState === "select"}
          <button class="control control--ghost" type="button" on:click={closeDialog} disabled={settingsBusy}>
            <ButtonIcon name="close"/>
            <span>Cerrar</span>
          </button>
          <button class="control control--accent" type="button" on:click={handleGenerate} disabled={!canGenerate}>
            <ButtonIcon name="spark"/>
            <span>Generar SQL</span>
          </button>
        {:else if viewState === "loading"}
          <span class="footer-status">Consultando OpenAI y preparando una salida ejecutable.</span>
        {:else if viewState === "result"}
          <button class="control control--ghost" type="button" on:click={closeDialog}>
            <ButtonIcon name="close"/>
            <span>Cerrar</span>
          </button>
          <button class="control control--soft" type="button" on:click={returnToSelection}>
            <ButtonIcon name="arrow-left"/>
            <span>Nueva corrida</span>
          </button>
        {:else}
          <button class="control control--ghost" type="button" on:click={returnToSelection}>
            <ButtonIcon name="arrow-left"/>
            <span>Volver a seleccion</span>
          </button>
          <button class="control control--accent" type="button" on:click={retryFromError}>
            <ButtonIcon name="spark"/>
            <span>Reintentar</span>
          </button>
        {/if}
      </footer>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    z-index: var(--layer-modal);
    display: grid;
    place-items: center;
    padding: 1rem;
    background: var(--overlay-scrim);
    backdrop-filter: blur(10px);
  }

  .modal-shell {
    width: min(920px, 100%);
    height: min(82vh, 760px);
    max-height: min(82vh, 760px);
    display: grid;
    grid-template-rows: auto minmax(0, 1fr) auto;
    gap: 1rem;
    padding: 1rem;
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    background:
      radial-gradient(circle at top right, color-mix(in srgb, var(--accent) 10%, transparent), transparent 32%),
      var(--popover-surface);
    box-shadow: var(--shadow-lg);
    overflow: hidden;
  }

  .modal-head,
  .config-card__head,
  .selection-card__head,
  .result-card__head,
  .modal-footer,
  .config-actions,
  .settings-actions,
  .result-banner,
  .result-banner__actions {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .modal-kicker,
  .label,
  .list-title {
    margin: 0;
    color: var(--accent);
    font-size: 0.72rem;
    font-weight: 800;
    letter-spacing: 0.16em;
    text-transform: uppercase;
  }

  .modal-head h2,
  .loading-copy h3,
  .result-banner h3,
  .error-stage h3 {
    margin: 0.35rem 0 0;
    font-size: clamp(1.45rem, 3vw, 2.1rem);
    line-height: 1;
  }

  .modal-hint,
  .muted,
  .selection-hint,
  .settings-note,
  .model-note,
  .entity-row span,
  .loading-copy p,
  .pipeline-item p,
  .summary-callout p,
  .error-stage__copy,
  .notice p {
    margin: 0;
    color: var(--ink-faint);
  }

  .modal-body {
    overflow: auto;
    min-height: 0;
    padding-right: 0.15rem;
  }

  .state-view {
    display: grid;
    gap: 1rem;
    min-height: 100%;
    align-content: start;
  }

  .top-grid,
  .result-grid {
    display: grid;
    grid-template-columns: minmax(0, 1.1fr) minmax(15rem, 0.72fr);
    gap: 1rem;
  }

  .config-card,
  .summary-card,
  .selection-card,
  .result-card,
  .result-sidecard,
  .settings-panel,
  .notice {
    border: 1px solid var(--border);
    border-radius: calc(var(--radius-md) - 4px);
    background: color-mix(in srgb, var(--surface) 88%, transparent);
  }

  .config-card,
  .summary-card,
  .selection-card,
  .result-card,
  .result-sidecard {
    padding: 1rem;
  }

  .field {
    display: grid;
    gap: 0.45rem;
  }

  .field span {
    font-size: 0.85rem;
    font-weight: 700;
    color: var(--ink-soft);
  }

  .field input,
  .field select {
    width: 100%;
    min-height: 2.8rem;
    box-sizing: border-box;
    border-radius: 0.9rem;
    border: 1px solid var(--border);
    background: var(--field-surface);
    color: var(--ink);
    padding: 0.8rem 0.95rem;
    font-size: 0.95rem;
    outline: none;
    transition: border 140ms ease, box-shadow 140ms ease, background 140ms ease;
  }

  .field input:focus,
  .field select:focus {
    border-color: var(--focus-border);
    box-shadow: var(--focus-ring);
    background: var(--field-surface-focus);
  }

  .status-chip {
    display: inline-flex;
    align-items: center;
    min-height: 1.9rem;
    padding: 0 0.7rem;
    border-radius: 999px;
    border: 1px solid var(--line-soft);
    font-size: 0.74rem;
    font-weight: 800;
    letter-spacing: 0.06em;
    text-transform: uppercase;
  }

  .status-chip--ok {
    background: color-mix(in srgb, var(--success) 14%, var(--surface-strong));
    color: color-mix(in srgb, var(--success) 70%, var(--ink));
  }

  .status-chip--warn {
    background: color-mix(in srgb, var(--danger) 10%, var(--surface-strong));
    color: color-mix(in srgb, var(--danger) 80%, var(--ink));
  }

  .settings-panel {
    display: grid;
    gap: 0.8rem;
    margin-top: 0.9rem;
    padding: 0.9rem;
  }

  .summary-stats,
  .result-metrics {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 0.75rem;
    margin-top: 0.8rem;
  }

  .result-metrics {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .summary-stats div,
  .result-metrics div {
    display: grid;
    gap: 0.2rem;
    padding: 0.8rem 0.85rem;
    border-radius: 0.9rem;
    background: color-mix(in srgb, var(--surface-strong) 78%, transparent);
    border: 1px solid var(--line-soft);
  }

  .summary-stats strong,
  .result-metrics strong {
    font-size: 1.25rem;
    line-height: 1;
    color: var(--ink);
  }

  .summary-stats span,
  .result-metrics span {
    color: var(--ink-faint);
    font-size: 0.78rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.06em;
  }

  .summary-callout {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 0.75rem;
    align-items: start;
    margin-top: 0.95rem;
    padding: 0.9rem;
    border-radius: 1rem;
    background: color-mix(in srgb, var(--accent) 8%, var(--surface-strong));
    border: 1px solid color-mix(in srgb, var(--accent) 14%, var(--border));
  }

  .summary-callout__step,
  .pipeline-item__index {
    display: inline-grid;
    place-items: center;
    width: 2.2rem;
    height: 2.2rem;
    border-radius: 999px;
    background: color-mix(in srgb, var(--accent) 12%, var(--surface-strong));
    color: var(--accent-strong);
    font-weight: 800;
    letter-spacing: 0.06em;
  }

  .summary-callout strong,
  .pipeline-item strong,
  .entity-row strong,
  .model-note,
  .notice strong,
  .loading-pulse {
    color: var(--ink-soft);
    font-size: 0.86rem;
    font-weight: 800;
  }

  .selection-list {
    display: grid;
    gap: 1rem;
    margin-top: 0.9rem;
  }

  .list-section {
    display: grid;
    gap: 0.55rem;
  }

  .entity-row {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr);
    gap: 0.75rem;
    align-items: flex-start;
    padding: 0.72rem 0.8rem;
    border-radius: 0.9rem;
    border: 1px solid var(--line-soft);
    background: color-mix(in srgb, var(--surface-strong) 84%, transparent);
    transition: transform 180ms ease, border-color 180ms ease, background 180ms ease, box-shadow 180ms ease;
  }

  .entity-row:hover {
    transform: translateY(-1px);
    border-color: color-mix(in srgb, var(--accent) 18%, var(--border));
  }

  .entity-row--selected {
    background: color-mix(in srgb, var(--accent) 11%, var(--surface-strong));
    border-color: color-mix(in srgb, var(--accent) 20%, var(--border));
    box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--accent) 12%, transparent);
  }

  .entity-row input {
    margin-top: 0.25rem;
  }

  .entity-row div {
    display: grid;
    gap: 0.16rem;
  }

  .notice {
    display: grid;
    gap: 0.3rem;
    padding: 0.9rem 1rem;
  }

  .notice--error {
    border-color: color-mix(in srgb, var(--danger) 20%, var(--border));
    background: color-mix(in srgb, var(--danger) 10%, var(--surface-strong));
  }

  .notice--soft {
    border-color: var(--line-soft);
    background: color-mix(in srgb, var(--surface-strong) 88%, transparent);
  }

  .state-view--loading,
  .state-view--error {
    align-content: center;
    min-height: 100%;
    padding: 0.2rem 0.05rem 0.35rem;
  }

  .loading-hero {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr);
    gap: 1.25rem;
    align-items: center;
    padding: 1.15rem;
    border: 1px solid var(--border);
    border-radius: calc(var(--radius-lg) - 6px);
    background:
      radial-gradient(circle at center, color-mix(in srgb, var(--accent) 12%, transparent), transparent 58%),
      color-mix(in srgb, var(--surface-strong) 92%, transparent);
  }

  .loader-orbit {
    position: relative;
    width: 7.4rem;
    height: 7.4rem;
    display: grid;
    place-items: center;
  }

  .loader-orbit__ring,
  .loader-orbit__ring--delayed {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    border: 1px solid color-mix(in srgb, var(--accent) 24%, transparent);
    animation: ripple 1.8s cubic-bezier(.19, 1, .22, 1) infinite;
  }

  .loader-orbit__ring--delayed {
    animation-delay: 0.4s;
  }

  .loader-orbit__core {
    width: 2.3rem;
    height: 2.3rem;
    border-radius: 50%;
    background:
      radial-gradient(circle at 35% 35%, color-mix(in srgb, white 42%, var(--accent) 8%), transparent 45%),
      linear-gradient(135deg, color-mix(in srgb, var(--accent) 18%, var(--surface-strong)), color-mix(in srgb, var(--accent) 30%, var(--surface-strong)));
    box-shadow:
      0 0 0 0.65rem color-mix(in srgb, var(--accent) 8%, transparent),
      0 18px 36px color-mix(in srgb, var(--accent) 16%, transparent);
    animation: pulseCore 1.35s ease-in-out infinite;
  }

  .loading-copy {
    display: grid;
    gap: 0.45rem;
  }

  .loading-pulse {
    display: inline-flex;
    align-items: center;
    gap: 0.45rem;
    margin-top: 0.2rem;
  }

  .loading-pulse::before {
    content: "";
    width: 0.55rem;
    height: 0.55rem;
    border-radius: 50%;
    background: color-mix(in srgb, var(--accent) 62%, var(--surface-strong));
    box-shadow: 0 0 0 0.24rem color-mix(in srgb, var(--accent) 14%, transparent);
    animation: blink 1.1s ease-in-out infinite;
  }

  .progress-shell {
    display: grid;
    gap: 0.45rem;
  }

  .progress-bar {
    position: relative;
    overflow: hidden;
    height: 0.72rem;
    border-radius: 999px;
    background: color-mix(in srgb, var(--surface-strong) 78%, var(--ink) 4%);
    border: 1px solid var(--line-soft);
  }

  .progress-bar__fill {
    position: absolute;
    inset: 0;
    transform-origin: left center;
    background:
      linear-gradient(90deg, color-mix(in srgb, var(--accent) 42%, var(--surface-strong)), color-mix(in srgb, var(--accent) 70%, var(--surface-strong)));
    border-radius: inherit;
    transition: transform 220ms ease-out;
  }

  .progress-bar__fill::after {
    content: "";
    position: absolute;
    inset: 0;
    background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.58), transparent);
    transform: translateX(-100%);
    animation: shimmer 1.4s linear infinite;
  }

  .progress-value {
    justify-self: end;
    color: var(--ink-soft);
    font-size: 0.82rem;
    font-weight: 800;
  }

  .pipeline-list {
    display: grid;
    gap: 0.75rem;
  }

  .pipeline-item {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 0.85rem;
    align-items: start;
    padding: 0.85rem 0.95rem;
    border-radius: 1rem;
    border: 1px solid var(--line-soft);
    background: color-mix(in srgb, var(--surface-strong) 86%, transparent);
    transition: transform 220ms ease, border-color 220ms ease, background 220ms ease, box-shadow 220ms ease;
  }

  .pipeline-item--active {
    transform: translateY(-1px);
    border-color: color-mix(in srgb, var(--accent) 22%, var(--border));
    background: color-mix(in srgb, var(--accent) 10%, var(--surface-strong));
    box-shadow: 0 18px 32px color-mix(in srgb, var(--accent) 10%, transparent);
  }

  .pipeline-item--done {
    border-color: color-mix(in srgb, var(--success) 20%, var(--border));
  }

  .pipeline-item--done .pipeline-item__index {
    background: color-mix(in srgb, var(--success) 18%, var(--surface-strong));
    color: color-mix(in srgb, var(--success) 74%, var(--ink));
  }

  .result-banner {
    padding: 1rem 1.05rem;
    border: 1px solid var(--border);
    border-radius: calc(var(--radius-lg) - 6px);
    background:
      linear-gradient(135deg, color-mix(in srgb, var(--accent) 10%, var(--surface-strong)), color-mix(in srgb, var(--surface-strong) 92%, transparent));
  }

  .sql-output {
    margin: 0.9rem 0 0;
    max-height: min(25rem, 42vh);
    overflow: auto;
    padding: 1rem;
    border-radius: 1rem;
    background: color-mix(in srgb, var(--surface-strong) 94%, transparent);
    border: 1px solid var(--line-soft);
    color: var(--ink);
    white-space: pre;
    box-shadow: inset 0 1px 0 color-mix(in srgb, white 28%, transparent);
  }

  .error-stage {
    display: grid;
    gap: 0.8rem;
    justify-items: center;
    text-align: center;
    padding: 1.4rem 1rem;
  }

  .error-stage__badge {
    display: grid;
    place-items: center;
    width: 4rem;
    height: 4rem;
    border-radius: 1.4rem;
    border: 1px solid color-mix(in srgb, var(--danger) 18%, var(--border));
    background: color-mix(in srgb, var(--danger) 10%, var(--surface-strong));
    color: color-mix(in srgb, var(--danger) 84%, var(--ink));
  }

  .error-stage__copy {
    max-width: 42rem;
    font-size: 1rem;
    line-height: 1.6;
  }

  .modal-footer {
    padding-top: 0.2rem;
    border-top: 1px solid var(--line-soft);
    min-height: 3.2rem;
  }

  .footer-status {
    color: var(--ink-faint);
    font-size: 0.9rem;
    font-weight: 700;
  }

  @keyframes ripple {
    0% {
      opacity: 0;
      transform: scale(0.76);
    }
    25% {
      opacity: 1;
    }
    100% {
      opacity: 0;
      transform: scale(1.08);
    }
  }

  @keyframes pulseCore {
    0%, 100% {
      transform: scale(0.96);
    }
    50% {
      transform: scale(1.04);
    }
  }

  @keyframes shimmer {
    to {
      transform: translateX(100%);
    }
  }

  @keyframes blink {
    0%, 100% {
      opacity: 0.55;
    }
    50% {
      opacity: 1;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .loader-orbit__ring,
    .loader-orbit__ring--delayed,
    .loader-orbit__core,
    .loading-pulse::before,
    .progress-bar__fill::after {
      animation: none;
    }

    .entity-row,
    .pipeline-item,
    .progress-bar__fill {
      transition: none;
    }
  }

  @media (max-width: 860px) {
    .top-grid,
    .result-grid,
    .summary-stats,
    .result-metrics,
    .loading-hero {
      grid-template-columns: 1fr;
    }

    .loader-orbit {
      margin: 0 auto;
    }
  }

  @media (max-width: 720px) {
    .modal-shell {
      padding: 0.9rem;
      border-radius: var(--radius-md);
      width: min(100%, 100%);
      height: min(88vh, 100%);
      max-height: min(88vh, 100%);
    }

    .modal-footer {
      flex-direction: column-reverse;
      align-items: stretch;
    }

    .result-banner__actions {
      width: 100%;
    }
  }
</style>
