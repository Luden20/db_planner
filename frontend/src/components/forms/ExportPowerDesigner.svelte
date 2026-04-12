<script lang="ts">
  import { fade } from "svelte/transition";
  import ButtonIcon from "../ButtonIcon.svelte";
  import Modal from "../ui/Modal.svelte";
  import { GeneratePowerDesignerFromEntities, ValidatePowerDesignerExport } from "../../../wailsjs/go/main/App";
  import type { utils } from "../../../wailsjs/go/models";
  import { showToast } from "../../lib/toast";
  import { Check, Copy, FloppyDisk, ArrowLeft, Sparkle, Warning } from "phosphor-svelte";
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

  let selectedIds: Set<number> = new Set();
  let selectedIntersectionIds: Set<number> = new Set();
  let scriptResult = "";
  let step: "select" | "result" | "loading" = "select";
  let validationErrors: any[] = $state([]);

  const totalSelectable = $derived(entities.length + intersectionEntities.length);
  const selectedCount = $derived(selectedIds.size + selectedIntersectionIds.size);
  const canGenerate = $derived(selectedCount > 0 && step !== "loading" && validationErrors.length === 0);

  // Validar selección reactivamente con $effect
  $effect(() => {
    if (isOpen && (selectedIds.size > 0 || selectedIntersectionIds.size > 0)) {
      const ids = Array.from(selectedIds);
      const intersectionIds = Array.from(selectedIntersectionIds);
      ValidatePowerDesignerExport(ids, intersectionIds).then(errs => {
        validationErrors = errs || [];
      });
    } else {
      validationErrors = [];
    }
  });

  const getEntityErrors = (name: string) => {
    return validationErrors.filter(e => e.entity_name === name);
  };

  export const openDialog = () => {
    isOpen = true;
    reset();
  };

  export const closeDialog = () => {
    isOpen = false;
  };

  const toggleAll = () => {
    if (selectedCount === totalSelectable) {
      selectedIds = new Set();
      selectedIntersectionIds = new Set();
    } else {
      selectedIds = new Set(entities.map(e => e.Id));
      selectedIntersectionIds = new Set(intersectionEntities.map(e => e.Entity.Id));
    }
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

  const handleExport = async () => {
    if (selectedIds.size === 0 && selectedIntersectionIds.size === 0) {
      showToast("Selecciona al menos una entidad para cargar el modelo.", "warning");
      return;
    }

    try {
      step = "loading";
      const ids = Array.from(selectedIds);
      const intersectionIds = Array.from(selectedIntersectionIds);
      scriptResult = await GeneratePowerDesignerFromEntities(ids, intersectionIds);
      step = "result";
    } catch (err: any) {
      step = "select";
      const message = err?.error ?? err?.message ?? err;
      showToast("Error al exportar a PowerDesigner: " + message, "error");
    }
  };

  const copyToClipboard = () => {
    navigator.clipboard.writeText(scriptResult);
    showToast("Script PowerDesigner copiado.", "success");
  };

  const downloadVBS = () => {
    const blob = new Blob([scriptResult], { type: 'application/x-vbs' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'export_powerdesigner.vbs';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    showToast("Descarga VBS iniciada", "success");
  };

  const reset = () => {
    step = "select";
    scriptResult = "";
  };

  const getTitle = () => {
    if (step === 'select') return "Exportar modelo lógico (VBS)";
    if (step === 'loading') return "Generando Script...";
    return "Script LDM Generado";
  };

  const getDescription = () => {
    if (step === 'select') return "Elige las entidades que formarán parte del LDM de PowerDesigner.";
    if (step === 'loading') return "Mapeando dominios, claves compuestas y comentarios nativos...";
    return "Copia o descarga el código VBS y ejecútalo en PowerDesigner.";
  };
</script>

<Modal
  bind:open={isOpen}
  title={getTitle()}
  description={getDescription()}
  size="form"
>
  <div class="flex flex-col gap-4">
    {#if step === 'select'}
      <div class="flex flex-col gap-4" in:fade>
        <div class="flex items-center justify-between border-b border-border-card pb-2">
          <span class="text-xs font-bold uppercase tracking-widest text-accent">Entidades e Intersecciones</span>
          <button 
            class="inline-flex items-center gap-2 text-xs font-bold text-foreground-alt hover:text-foreground transition-colors"
            onclick={toggleAll}
          >
            <Check class="size-3" />
            <span>{selectedCount === totalSelectable ? "Desmarcar todo" : "Marcar todo"}</span>
          </button>
        </div>

        {#if validationErrors.length > 0}
          <div class="rounded-card-sm border border-destructive/30 bg-destructive/5 p-4" in:fade>
            <div class="flex items-center gap-2 text-destructive mb-2">
              <Warning class="size-4" weight="bold" />
              <strong class="text-sm">Atributos no compatibles</strong>
            </div>
            <ul class="text-xs text-foreground-alt space-y-1 pl-6 list-disc">
              {#each validationErrors as error}
                <li><strong>{error.entity_name} &gt; {error.attribute_name}:</strong> {error.message}</li>
              {/each}
            </ul>
          </div>
        {/if}

        <div class="grid grid-cols-1 gap-2 max-h-[400px] overflow-y-auto pr-2 custom-scrollbar">
          {#if entities.length > 0}
            <div class="flex flex-col gap-1">
              <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60 mb-1">Fuertes</p>
              {#each entities as entity}
                <label 
                  class={cn(
                    "flex flex-col p-3 rounded-card-sm border border-border-card transition-all cursor-pointer hover:bg-muted/30",
                    selectedIds.has(entity.Id) && "bg-accent/5 border-accent/30",
                    getEntityErrors(entity.Name).length > 0 && "border-destructive/20 opacity-80"
                  )}
                >
                  <div class="flex items-center gap-3">
                    <input
                      type="checkbox"
                      class="size-4 rounded-sm border-border-input"
                      checked={selectedIds.has(entity.Id)}
                      onchange={() => toggleEntity(entity.Id)}
                    />
                    <div class="flex-1">
                      <strong class="text-sm block">{entity.Name}</strong>
                      <span class="text-xs text-muted-foreground line-clamp-1">{entity.Description || "Sin descripción."}</span>
                    </div>
                    {#if getEntityErrors(entity.Name).length > 0}
                      <div class="flex gap-1">
                        {#each getEntityErrors(entity.Name) as err}
                          <span class="text-[9px] font-bold px-1.5 py-0.5 rounded-full bg-destructive/10 text-destructive border border-destructive/20">
                            {err.attribute_name}
                          </span>
                        {/each}
                      </div>
                    {/if}
                  </div>
                </label>
              {/each}
            </div>
          {/if}

          {#if intersectionEntities.length > 0}
            <div class="flex flex-col gap-1 mt-4">
              <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60 mb-1">Intersecciones</p>
              {#each intersectionEntities as item}
                <label 
                  class={cn(
                    "flex flex-col p-3 rounded-card-sm border border-border-card transition-all cursor-pointer hover:bg-muted/30",
                    selectedIntersectionIds.has(item.Entity.Id) && "bg-accent/5 border-accent/30"
                  )}
                >
                  <div class="flex items-center gap-3">
                    <input
                      type="checkbox"
                      class="size-4 rounded-sm border-border-input"
                      checked={selectedIntersectionIds.has(item.Entity.Id)}
                      onchange={() => toggleIntersection(item.Entity.Id)}
                    />
                    <div class="flex-1">
                      <strong class="text-sm block">{item.Entity.Name}</strong>
                      <span class="text-xs text-muted-foreground line-clamp-1">{item.Entity.Description || "Sin descripción."}</span>
                    </div>
                  </div>
                </label>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    {:else if step === 'loading'}
      <div class="flex flex-col items-center justify-center py-12 gap-4" in:fade>
        <Sparkle class="size-12 text-accent animate-spin" weight="fill" />
        <p class="text-foreground-alt font-medium">Traduciendo entidades al dialecto VBScript...</p>
      </div>
    {:else}
      <div class="flex flex-col gap-4" in:fade>
        <div class="bg-dark rounded-card-sm p-4 overflow-hidden">
          <pre class="text-[12px] font-mono text-[#4fc1ff] max-h-[400px] overflow-auto custom-scrollbar"><code>{scriptResult}</code></pre>
        </div>
      </div>
    {/if}
  </div>

  {#snippet footer()}
    <div class="flex w-full justify-between items-center mt-4">
      {#if step === 'select'}
        <button class="h-11 rounded-input px-6 text-sm font-semibold text-foreground-alt hover:bg-muted transition-colors" onclick={closeDialog}>
          Cancelar
        </button>
        <button 
          class="h-11 rounded-input bg-dark text-background px-8 text-sm font-bold shadow-mini hover:bg-dark/90 transition-all active:scale-[0.98] disabled:opacity-50" 
          onclick={handleExport} 
          disabled={!canGenerate}
        >
          Generar Script VBS
        </button>
      {:else if step === 'loading'}
        <div class="w-full text-center py-2 text-xs font-bold text-muted-foreground animate-pulse uppercase tracking-widest">
          Procesando pipeline...
        </div>
      {:else}
        <button class="h-11 rounded-input px-6 text-sm font-semibold text-foreground-alt hover:bg-muted transition-colors flex items-center gap-2" onclick={reset}>
          <ArrowLeft class="size-4" />
          <span>Atrás</span>
        </button>
        <div class="flex gap-3">
          <button class="h-11 rounded-input border border-border-card px-6 text-sm font-semibold hover:bg-muted transition-colors flex items-center gap-2" onclick={copyToClipboard}>
            <Copy class="size-4" />
            <span>Copiar</span>
          </button>
          <button class="h-11 rounded-input bg-dark text-background px-8 text-sm font-bold shadow-mini hover:bg-dark/90 transition-all active:scale-[0.98]" onclick={downloadVBS}>
            <FloppyDisk class="size-4 mr-2" />
            <span>Descargar .vbs</span>
          </button>
        </div>
      {/if}
    </div>
  {/snippet}
</Modal>

<style>
  .custom-scrollbar::-webkit-scrollbar {
    width: 6px;
    height: 6px;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.1);
    border-radius: 10px;
  }
  .custom-scrollbar::-webkit-scrollbar-track {
    background: transparent;
  }
</style>
