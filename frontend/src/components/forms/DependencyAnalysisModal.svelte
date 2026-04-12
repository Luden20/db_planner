<script lang="ts">
  import { fade } from "svelte/transition";
  import type { utils } from "../../../wailsjs/go/models";
  import Modal from "../ui/Modal.svelte";
  import { AnalyzeDependencies } from "../../../wailsjs/go/main/App.js";
  import { 
    CircleNotch, 
    CheckCircle, 
    Warning as WarningIcon, 
    GitBranch, 
    CaretRight, 
    CaretDown, 
    CaretUp,
    ArrowsClockwise,
    Bug
  } from "phosphor-svelte";
  import cn from "clsx";

  let { isOpen = $bindable(false) } = $props<{ isOpen?: boolean }>();

  let report: utils.AnalysisReport | null = $state(null);
  let loading = $state(false);
  let showEdges = $state(false);
  let expandedIssues = $state(new Set<string>());

  export const openDialog = async () => {
    loading = true;
    isOpen = true;
    try {
      const res = await AnalyzeDependencies();
      report = res;
    } catch (err) {
      console.error("Error al analizar dependencias:", err);
    } finally {
      loading = false;
    }
  };

  const getCircular = (issues: utils.AnalysisIssue[]) =>
    issues.filter((i) => i.type === "Circular");
  const getAmbiguity = (issues: utils.AnalysisIssue[]) =>
    issues.filter((i) => i.type === "Ambiguity");

  const toggleExpand = (u: string, v: string) => {
    const key = `${u}-${v}`;
    if (expandedIssues.has(key)) {
      expandedIssues.delete(key);
    } else {
      expandedIssues.add(key);
    }
    expandedIssues = new Set(expandedIssues);
  };
</script>

<Modal
  bind:open={isOpen}
  title="Análisis Relacional"
  description="Validación profunda de la jerarquía táctica y detección de redundancias estructurales."
  size="form"
  confirmLabel="Cerrar"
>
  {#if loading}
    <div class="flex flex-col items-center justify-center py-16 gap-4">
      <CircleNotch class="size-10 text-accent animate-spin" />
      <p class="text-sm font-medium text-foreground-alt">Construyendo grafo estructural...</p>
    </div>
  {:else if report}
    <div class="flex flex-col gap-6">
      <!-- Summary Chips -->
      <div class="grid grid-cols-3 gap-3">
        <div class="bg-accent/5 border border-accent/20 rounded-xl p-4 flex flex-col items-center text-center">
          <span class="text-2xl font-black text-accent">{getAmbiguity(report.issues || []).length}</span>
          <span class="text-[9px] font-black uppercase tracking-widest text-accent/60">Ambigüedades</span>
        </div>
        <div class="bg-destructive/5 border border-destructive/20 rounded-xl p-4 flex flex-col items-center text-center">
          <span class="text-2xl font-black text-destructive">{getCircular(report.issues || []).length}</span>
          <span class="text-[9px] font-black uppercase tracking-widest text-destructive/60">Ciclos</span>
        </div>
        <div class="bg-muted/30 border border-border-card rounded-xl p-4 flex flex-col items-center text-center">
          <span class="text-2xl font-black text-foreground-alt">{report.edgeList?.length || 0}</span>
          <span class="text-[9px] font-black uppercase tracking-widest text-muted-foreground">Relaciones</span>
        </div>
      </div>

      <!-- Edge List -->
      <div class="border border-border-card rounded-xl overflow-hidden bg-muted/10">
        <button 
          class="w-full flex items-center justify-between p-4 hover:bg-muted/20 transition-colors"
          onclick={() => showEdges = !showEdges}
        >
          <div class="flex items-center gap-2">
            <GitBranch class="size-4 text-accent" />
            <span class="text-xs font-bold uppercase tracking-tight">Aristas Interpretadas (Padre → Hijo)</span>
          </div>
          {#if showEdges}<CaretDown class="size-4" />{:else}<CaretRight class="size-4" />{/if}
        </button>
        {#if showEdges}
          <div class="p-4 border-t border-border-card bg-background grid grid-cols-2 sm:grid-cols-3 gap-2" in:fade>
            {#each report.edgeList || [] as edge}
              <div class="text-[10px] font-mono bg-muted/30 px-2 py-1 rounded border border-border-card truncate">{edge}</div>
            {/each}
          </div>
        {/if}
      </div>

      {#if (report.issues?.length || 0) === 0}
        <div class="flex flex-col items-center justify-center py-8 gap-3 bg-green-500/5 border border-green-500/20 rounded-xl">
          <CheckCircle class="size-10 text-green-500" weight="fill" />
          <div class="text-center">
            <h3 class="text-lg font-bold text-green-500 leading-none">¡Estructura Correcta!</h3>
            <p class="text-xs text-green-500/70 mt-1">No se detectaron redundancias ni ciclos en el flujo actual.</p>
          </div>
        </div>
      {:else}
        <div class="flex flex-col gap-4">
          {#if getAmbiguity(report.issues).length > 0}
            <div class="flex flex-col gap-2">
              <div class="flex items-center gap-2 text-accent">
                 <ArrowsClockwise class="size-4" weight="bold" />
                 <span class="text-xs font-black uppercase tracking-widest">Ambigüedad Estructural</span>
              </div>
              <div class="flex flex-col gap-2">
                {#each getAmbiguity(report.issues) as issue}
                  {@const key = `${issue.entities[0]}-${issue.entities[1]}`}
                  {@const isExpanded = expandedIssues.has(key)}
                  <div class="border border-border-card rounded-xl overflow-hidden bg-background">
                    <button 
                      class="w-full flex items-center justify-between p-4 hover:bg-muted/20 transition-colors"
                      onclick={() => toggleExpand(issue.entities[0], issue.entities[1])}
                    >
                      <div class="flex items-center gap-3">
                         <span class="text-xs font-bold px-2 py-1 bg-muted rounded border border-border-card">{issue.entities[0]}</span>
                         <CaretRight class="size-3 text-muted-foreground" />
                         <span class="text-xs font-bold px-2 py-1 bg-muted rounded border border-border-card">{issue.entities[1]}</span>
                      </div>
                      <div class="flex items-center gap-3">
                        <span class="text-[10px] font-bold text-accent bg-accent/10 px-2 py-0.5 rounded-full">{issue.pathCount} rutas</span>
                        {#if isExpanded}<CaretUp class="size-4" />{:else}<CaretDown class="size-4" />{/if}
                      </div>
                    </button>
                    {#if isExpanded}
                      <div class="p-4 border-t border-border-card bg-muted/5 flex flex-col gap-3" in:fade>
                        <p class="text-[10px] font-black uppercase text-muted-foreground tracking-widest">Rutas Dirigidas:</p>
                        <div class="flex flex-col gap-1.5">
                          {#each issue.paths.slice(0, 5) as path, i}
                            <div class="text-[11px] font-mono p-2 bg-background border border-border-card rounded-lg flex gap-2">
                               <span class="text-accent font-bold">{i+1}.</span>
                               <span>{path.join(" → ")}</span>
                            </div>
                          {/each}
                        </div>
                      </div>
                    {/if}
                  </div>
                {/each}
              </div>
            </div>
          {/if}

          {#if getCircular(report.issues).length > 0}
            <div class="flex flex-col gap-2 pt-4 border-t border-border-card mt-2">
               <div class="flex items-center gap-2 text-destructive">
                 <Bug class="size-4" weight="bold" />
                 <span class="text-xs font-black uppercase tracking-widest">Dependencias Circulares</span>
              </div>
              {#each getCircular(report.issues) as issue}
                <div class="p-4 bg-destructive/5 border border-destructive/20 rounded-xl flex items-center justify-between">
                   <p class="text-[11px] font-mono font-bold">{issue.entities.join(" → ")} → {issue.entities[0]}</p>
                   <span class="text-[10px] font-black uppercase text-destructive">Ciclo crítico</span>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      {/if}
    </div>
  {:else}
    <div class="text-center py-12 text-muted-foreground text-sm font-medium">
      No hay datos de análisis disponibles.
    </div>
  {/if}
</Modal>
