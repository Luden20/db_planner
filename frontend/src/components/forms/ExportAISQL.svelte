<script lang="ts">
  import { fade } from "svelte/transition";
  import { onDestroy } from "svelte";
  import ButtonIcon from "../ButtonIcon.svelte";
  import Modal from "../ui/Modal.svelte";
  import { GenerateSQLFromEntities, GetAISettings, SaveOpenAIAPIKey } from "../../../wailsjs/go/main/App";
  import { utils } from "../../../wailsjs/go/models";
  import { showToast } from "../../lib/toast";
  import { 
    Sparkle, 
    Database as DatabaseIcon, 
    Table, 
    FileCode, 
    Copy, 
    ArrowLeft, 
    Check, 
    Trash, 
    CircleNotch,
    X,
    WarningCircle
  } from "phosphor-svelte";
  import cn from "clsx";

  let { 
    entities = [], 
    intersectionEntities = [], 
    isOpen = $bindable(false) 
  } = $props<{
    entities?: utils.Entity[];
    intersectionEntities?: utils.IntersectionEntity[];
    isOpen?: boolean;
  }>();

  type ViewState = "select" | "loading" | "result" | "error";
  type ResultView = "generated" | "ai";
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
    {id: "schema", title: "Empaquetando esquema", detail: "Consolidando tablas, cruces y relaciones."},
    {id: "request", title: "Generando script", detail: "Construyendo DDL, relaciones y documentación."},
    {id: "assembly", title: "Completando salida", detail: "Enriqueciendo resultado con SQL de IA."},
  ];

  const baseAISettings = () => new utils.AISettings({HasAPIKey: false, OpenAIModel: "gpt-5-mini"});

  let viewState: ViewState = $state("select");
  let aiSettings: utils.AISettings = $state(baseAISettings());
  let generatedResult: utils.SQLGenerationResult | null = $state(null);
  let resultView: ResultView = $state("generated");
  let selectedIds = $state(new Set<number>());
  let selectedIntersectionIds = $state(new Set<number>());
  let database = $state(databaseOptions[0].value);
  let apiKeyDraft = $state("");
  let settingsExpanded = $state(false);
  let settingsBusy = $state(false);
  let generateBusy = $state(false);
  let inlineErrorMessage = $state("");
  let pipelineErrorMessage = $state("");
  let loadingStageIndex = $state(0);
  let loadingProgress = $state(0);
  let loadingPulse = $state("Preparando consulta");
  let stageTimer: ReturnType<typeof setInterval> | null = null;
  let progressTimer: ReturnType<typeof setInterval> | null = null;
  let pulseTimer: ReturnType<typeof setInterval> | null = null;

  const pulseMessages = ["Preparando", "Leyendo", "Generando", "Puliendo"];

  const totalSelectable = $derived(entities.length + intersectionEntities.length);
  const selectedCount = $derived(selectedIds.size + selectedIntersectionIds.size);
  const canGenerate = $derived(selectedCount > 0 && database.trim().length > 0 && !generateBusy);
  const hasAISQL = $derived(!!generatedResult?.SQL?.trim());
  const activeStage = $derived(pipelineStages[Math.min(loadingStageIndex, pipelineStages.length - 1)]);

  onDestroy(() => stopLoadingPresentation());

  const clearMessages = () => {
    inlineErrorMessage = "";
    pipelineErrorMessage = "";
  };

  const loadSettings = async () => {
    try {
      aiSettings = await GetAISettings();
    } catch (err) {
      aiSettings = baseAISettings();
    }
  };

  const resetState = async () => {
    clearMessages();
    generatedResult = null;
    resultView = "generated";
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
    if (settingsBusy || generateBusy) return;
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
    selectedIds = new Set(entities.map(e => e.Id));
    selectedIntersectionIds = new Set(intersectionEntities.map(e => e.Entity.Id));
  };

  const toggleEntity = (id: number) => {
    if (selectedIds.has(id)) selectedIds.delete(id); else selectedIds.add(id);
    selectedIds = new Set(selectedIds);
  };

  const toggleIntersection = (id: number) => {
    if (selectedIntersectionIds.has(id)) selectedIntersectionIds.delete(id); else selectedIntersectionIds.add(id);
    selectedIntersectionIds = new Set(selectedIntersectionIds);
  };

  const handleGenerate = async () => {
    if (!canGenerate) return;
    generateBusy = true;
    viewState = "loading";
    startLoadingPresentation();
    try {
      generatedResult = await GenerateSQLFromEntities(Array.from(selectedIds), Array.from(selectedIntersectionIds), database);
      resultView = "generated";
      finishLoadingPresentation();
      await new Promise(r => setTimeout(r, 400));
      viewState = "result";
    } catch (err: any) {
      stopLoadingPresentation();
      pipelineErrorMessage = err?.message || err || "Error desconocido";
      viewState = "error";
    } finally {
      generateBusy = false;
    }
  };

  const saveApiKey = async () => {
    if (!apiKeyDraft.trim()) return;
    settingsBusy = true;
    try {
      aiSettings = await SaveOpenAIAPIKey(apiKeyDraft.trim());
      apiKeyDraft = "";
      settingsExpanded = false;
      showToast("Key guardada", "success");
    } catch (err: any) {
      inlineErrorMessage = err?.message || "Error al guardar key";
    } finally { settingsBusy = false; }
  };

  const startLoadingPresentation = () => {
    stopLoadingPresentation();
    loadingStageIndex = 0;
    loadingProgress = 10;
    progressTimer = setInterval(() => { loadingProgress = Math.min(95, loadingProgress + Math.random() * 3); }, 200);
    stageTimer = setInterval(() => { loadingStageIndex = Math.min(pipelineStages.length - 1, loadingStageIndex + 1); }, 1800);
  };

  const stopLoadingPresentation = () => {
    if (progressTimer) clearInterval(progressTimer);
    if (stageTimer) clearInterval(stageTimer);
  };

  const finishLoadingPresentation = () => {
    stopLoadingPresentation();
    loadingProgress = 100;
  };

  const copyText = async (text: string, msg: string) => {
    await navigator.clipboard.writeText(text);
    showToast(msg, "success");
  };

  const getTitle = () => {
    if (viewState === "select") return "Exportar Scripts SQL";
    if (viewState === "loading") return "Generando Pipeline";
    if (viewState === "result") return "Resultado Final";
    return "Error de Generación";
  };
</script>

<Modal
  bind:open={isOpen}
  title={getTitle()}
  description={viewState === 'select' ? "Elige el motor, las tablas y opcionalmente habilita IA para enriquecer el script." : ""}
  size="form"
>
  <div class="flex flex-col gap-6">
    {#if viewState === "select"}
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4" in:fade>
        <!-- Layout similar to the original grid but cleaner -->
        <div class="flex flex-col gap-4">
          <div class="rounded-card-sm border border-border-card bg-muted/20 p-4">
            <div class="flex items-center justify-between mb-4">
              <span class="text-[10px] font-black uppercase tracking-widest text-accent">Configuración</span>
              <span class={cn("text-[10px] px-2 py-0.5 rounded-full font-bold", aiSettings.HasAPIKey ? "bg-green-500/10 text-green-500" : "bg-orange-500/10 text-orange-500")}>
                {aiSettings.HasAPIKey ? "IA Disponible" : "Solo local"}
              </span>
            </div>
            
            <div class="flex flex-col gap-1.5 mb-4">
              <label class="text-xs font-bold text-foreground-alt">Motor de DB</label>
              <select bind:value={database} class="h-10 w-full rounded-card-sm border border-border-input bg-background px-3 text-sm focus:ring-2 focus:ring-accent outline-none">
                {#each databaseOptions as opt}
                  <option value={opt.value}>{opt.label}</option>
                {/each}
              </select>
              <p class="text-[10px] text-muted-foreground mt-1 italic">
                 {databaseOptions.find(o => o.value === database)?.hint}
              </p>
            </div>

            <button 
              class="w-full h-9 rounded-card-sm border border-dashed border-border-card text-[11px] font-bold hover:bg-muted/50 transition-colors flex items-center justify-center gap-2"
              onclick={() => settingsExpanded = !settingsExpanded}
            >
              <Sparkle class="size-3" />
              {settingsExpanded ? "Ocultar ajustes IA" : "Configurar API Key IA"}
            </button>

            {#if settingsExpanded}
              <div class="mt-4 flex flex-col gap-3 p-3 bg-background rounded-card-sm border border-border-card" in:fade>
                <input 
                  type="password" 
                  bind:value={apiKeyDraft} 
                  placeholder="sk-..." 
                  class="h-9 w-full rounded-card-sm border border-border-input bg-muted/30 px-3 text-xs outline-none"
                />
                <button class="h-8 rounded-input bg-dark text-background text-[11px] font-bold" onclick={saveApiKey} disabled={settingsBusy}>
                  {settingsBusy ? "Guardando..." : "Guardar API Key"}
                </button>
              </div>
            {/if}
          </div>

          <div class="rounded-card-sm border border-border-card p-4 bg-muted/10">
             <span class="text-[10px] font-black uppercase tracking-widest text-accent mb-3 block">Estadísticas</span>
             <div class="grid grid-cols-3 gap-2 text-center">
               <div class="flex flex-col">
                 <span class="text-lg font-bold leading-none">{selectedCount}</span>
                 <span class="text-[9px] text-muted-foreground uppercase font-bold">Marcadas</span>
               </div>
               <div class="flex flex-col">
                 <span class="text-lg font-bold leading-none">{entities.length}</span>
                 <span class="text-[9px] text-muted-foreground uppercase font-bold">Fuertes</span>
               </div>
               <div class="flex flex-col border-none">
                 <span class="text-lg font-bold leading-none">{intersectionEntities.length}</span>
                 <span class="text-[9px] text-muted-foreground uppercase font-bold">Cruces</span>
               </div>
             </div>
          </div>
        </div>

        <div class="rounded-card-sm border border-border-card p-4 flex flex-col gap-3 max-h-[400px]">
           <div class="flex items-center justify-between">
              <span class="text-[10px] font-black uppercase tracking-widest text-accent">Selección de tablas</span>
              <button class="text-[10px] font-bold hover:underline" onclick={toggleAll}>Alternar todo</button>
           </div>
           
           <div class="flex-1 overflow-y-auto pr-2 custom-scrollbar flex flex-col gap-1">
              {#each entities as entity}
                <label class={cn("flex items-center gap-3 p-2 rounded-lg border border-transparent hover:bg-muted/30 cursor-pointer", selectedIds.has(entity.Id) && "bg-accent/5 border-accent/20")}>
                  <input type="checkbox" checked={selectedIds.has(entity.Id)} onchange={() => toggleEntity(entity.Id)} class="size-3.5" />
                  <span class="text-xs font-bold leading-none truncate">{entity.Name}</span>
                </label>
              {/each}
              {#each intersectionEntities as ie}
                <label class={cn("flex items-center gap-3 p-2 rounded-lg border border-transparent hover:bg-muted/30 cursor-pointer", selectedIntersectionIds.has(ie.Entity.Id) && "bg-accent/5 border-accent/20")}>
                  <input type="checkbox" checked={selectedIntersectionIds.has(ie.Entity.Id)} onchange={() => toggleIntersection(ie.Entity.Id)} class="size-3.5" />
                  <span class="text-xs font-bold leading-none italic truncate">{ie.Entity.Name}</span>
                </label>
              {/each}
           </div>
        </div>
      </div>
    {:else if viewState === "loading"}
      <div class="flex flex-col items-center py-12 gap-8" in:fade>
        <div class="relative size-24">
          <CircleNotch class="size-24 text-accent animate-spin" />
          <div class="absolute inset-0 flex items-center justify-center">
            <span class="text-xs font-bold">{Math.round(loadingProgress)}%</span>
          </div>
        </div>
        <div class="text-center">
           <h3 class="text-lg font-bold">{activeStage.title}</h3>
           <p class="text-sm text-foreground-alt">{activeStage.detail}</p>
        </div>
        <div class="w-64 h-1.5 bg-muted rounded-full overflow-hidden">
           <div class="h-full bg-accent transition-all duration-300" style="width: {loadingProgress}%"></div>
        </div>
      </div>
    {:else if viewState === "result" && generatedResult}
      <div class="flex flex-col gap-4" in:fade>
        <div class="flex items-center justify-between">
          <div class="flex gap-2 p-1 bg-muted rounded-lg">
            <button 
              class={cn("px-4 py-1.5 text-xs font-bold rounded-md transition-all", resultView === 'generated' ? "bg-background shadow-mini text-accent" : "text-muted-foreground")}
              onclick={() => resultView = "generated"}
            >Generado</button>
            <button 
              class={cn("px-4 py-1.5 text-xs font-bold rounded-md transition-all", resultView === 'ai' ? "bg-background shadow-mini text-accent" : "text-muted-foreground")}
              onclick={() => resultView = "ai"}
              disabled={!hasAISQL}
            >IA SQL</button>
          </div>
          <div class="flex gap-2">
            <button class="flex items-center gap-1.5 text-[10px] font-bold px-3 py-1.5 border border-border-card rounded-md hover:bg-muted transition-colors" onclick={() => copyText(generatedResult.ExportJSON, "JSON copiado")}>
              <DatabaseIcon class="size-3" /> JSON
            </button>
            <button class="flex items-center gap-1.5 text-[10px] font-bold px-4 py-1.5 bg-dark text-background rounded-md hover:bg-dark/90" onclick={() => copyText(resultView === 'generated' ? generatedResult.GeneratedScript : generatedResult.SQL, "Script copiado")}>
              <Copy class="size-3" /> Copiar Script
            </button>
          </div>
        </div>

        <div class="bg-dark rounded-card-sm p-4 h-[400px] overflow-hidden">
          <pre class="text-[12px] font-mono text-[#4fc1ff] h-full overflow-auto custom-scrollbar"><code>{resultView === 'generated' ? generatedResult.GeneratedScript : generatedResult.SQL}</code></pre>
        </div>
      </div>
    {:else if viewState === "error"}
       <div class="flex flex-col items-center justify-center py-16 gap-4 text-center">
          <div class="size-16 rounded-full bg-destructive/10 flex items-center justify-center text-destructive mb-2">
             <WarningCircle class="size-10" />
          </div>
          <h3 class="text-xl font-bold">Error en la exportación</h3>
          <p class="text-sm text-foreground-alt max-w-[300px]">{pipelineErrorMessage}</p>
       </div>
    {/if}
  </div>

  {#snippet footer()}
    <div class="flex w-full justify-between items-center mt-4">
      {#if viewState === "select"}
        <button class="h-11 px-6 text-sm font-semibold text-foreground-alt hover:bg-muted rounded-input" onclick={closeDialog}>Cerrar</button>
        <button class="h-11 px-10 bg-dark text-background text-sm font-bold rounded-input shadow-mini hover:bg-dark/90 active:scale-[0.98] disabled:opacity-50" onclick={handleGenerate} disabled={!canGenerate}>
           <Sparkle class="size-4 mr-2" /> Generar Scripts
        </button>
      {:else if viewState === "loading"}
         <div class="flex-1 text-center py-2 text-[10px] font-black uppercase text-muted-foreground tracking-tighter animate-pulse">
            El servidor wails está procesando el esquema...
         </div>
      {:else if viewState === "result"}
        <button class="h-11 px-6 text-sm font-semibold text-foreground-alt hover:bg-muted rounded-input flex items-center gap-2" onclick={returnToSelection}>
          <ArrowLeft class="size-4" /> Nueva corrida
        </button>
        <button class="h-11 px-10 bg-dark text-background text-sm font-bold rounded-input shadow-mini" onclick={closeDialog}>Listo</button>
      {:else}
        <button class="h-11 px-6 text-sm font-semibold text-foreground-alt hover:bg-muted rounded-input" onclick={returnToSelection}>Atrás</button>
        <button class="h-11 px-10 bg-destructive text-white text-sm font-bold rounded-input" onclick={handleGenerate}>Reintentar</button>
      {/if}
    </div>
  {/snippet}
</Modal>

<style>
  .custom-scrollbar::-webkit-scrollbar { width: 4px; height: 4px; }
  .custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(255, 255, 255, 0.1); border-radius: 10px; }
  .custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
</style>
