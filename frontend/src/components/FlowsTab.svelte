<script lang="ts">
  import { flip } from "svelte/animate";
  import { quintOut } from "svelte/easing";
  import { fade, fly, scale } from "svelte/transition";
  import { tick, onMount } from "svelte";
  import {
    AddBigProcess,
    AddProcess,
    AddResource,
    AddStep,
    EditBigProcess,
    EditProcess,
    EditResource,
    EditStep,
    MoveBigProcess,
    MoveProcess,
    MoveStep,
    RemoveBigProcess,
    RemoveProcess,
    RemoveResource,
    RemoveStep,
    Save
  } from "../../wailsjs/go/main/App";
  import type { utils } from "../../wailsjs/go/models";
  import ButtonIcon from "./ButtonIcon.svelte";
  import ModalLauncher from "./ModalLauncher.svelte";
  import { showToast } from "../lib/toast";

  type ResourceDraft = { tableId: number; role: string; };
  type ViewTransitionDocument = Document & { startViewTransition?: (update: () => void | Promise<void>) => { finished: Promise<void>; }; };

  let { 
    project, 
    entities = [], 
    onRefresh = async () => {} 
  } = $props<{
    project: utils.DbProject;
    entities?: utils.Entity[];
    onRefresh?: () => Promise<void>;
  }>();

  const resourceRoles = ["Input", "Output"];

  let selectedBigProcessId = $state<number | null>(null);
  let selectedProcessId = $state<number | null>(null);
  let selectedStepId = $state<number | null>(null);

  let lastBigProcessSyncId: number | null = null;
  let lastProcessSyncId: number | null = null;
  let lastStepSyncId = $state<number | null>(null);
  let lastResourceSignature = $state("");

  let bigProcessDraftName = $state("");
  let bigProcessDraftDescription = $state("");
  let processDraftName = $state("");
  let processDraftDescription = $state("");
  let stepDraftName = $state("");
  let stepDraftDescription = $state("");

  let newBigProcessName = $state("");
  let newBigProcessDescription = $state("");
  let newProcessName = $state("");
  let newProcessDescription = $state("");
  let newStepName = $state("");
  let newStepDescription = $state("");
  let newResourceTableId = $state<number | null>(null);
  let newResourceRole = $state("Input");

  let resourceEdits = $state<Record<number, ResourceDraft>>({});
  let busySection = $state<string | null>(null);
  let draggingBigProcessId = $state<number | null>(null);
  let hoverBigProcessId = $state<number | null>(null);
  let draggingProcessId = $state<number | null>(null);
  let hoverProcessId = $state<number | null>(null);
  let draggingStepId = $state<number | null>(null);
  let hoverStepId = $state<number | null>(null);
  let autoScrollDirection = $state<-1 | 0 | 1>(0);

  const AUTO_SCROLL_EDGE_PX = 96;
  const AUTO_SCROLL_STEP = 18;

  const bigProcesses = $derived(project?.BigProcesses ?? []);
  const currentBigProcess = $derived(bigProcesses.find((item) => item.Id === selectedBigProcessId) ?? null);
  const currentProcesses = $derived(currentBigProcess?.Processes ?? []);
  const currentProcess = $derived(currentProcesses.find((item) => item.Id === selectedProcessId) ?? null);
  const currentSteps = $derived(currentProcess?.Steps ?? []);
  const currentStep = $derived(currentSteps.find((item) => item.Id === selectedStepId) ?? null);

  const prefersReducedMotion = () => typeof window !== "undefined" && typeof window.matchMedia === "function" && window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  const runStageTransition = async (update: () => void | Promise<void>) => {
    const doc = typeof document !== "undefined" ? (document as ViewTransitionDocument) : null;
    if (doc?.startViewTransition && !prefersReducedMotion()) {
      try {
        const transition = doc.startViewTransition(update);
        await transition.finished;
        return;
      } catch (err) { console.warn("No se pudo aplicar la transición de vista:", err); }
    }
    await update();
  };

  const extractError = (err: unknown) => {
    if (typeof err === "string") return err;
    if (err && typeof err === "object") {
      const maybeRecord = err as Record<string, unknown>;
      return String(maybeRecord.error ?? maybeRecord.message ?? "Error desconocido");
    }
    return "Error desconocido";
  };

  const entityLabel = (entityId: number) => entities.find((e) => e.Id === entityId)?.Name ?? `Tabla ${entityId}`;
  const entityDescription = (entityId: number) => entities.find((e) => e.Id === entityId)?.Description ?? "Sin detalle de la tabla.";
  const countProcessResources = (p: utils.Process) => (p.Steps ?? []).reduce((sum, s) => sum + (s.Resources?.length ?? 0), 0);
  const countBigProcessSteps = (bp: utils.BigProcess) => (bp.Processes ?? []).reduce((sum, p) => sum + (p.Steps?.length ?? 0), 0);
  const countBigProcessResources = (bp: utils.BigProcess) => (bp.Processes ?? []).reduce((sum, p) => sum + countProcessResources(p), 0);
  const getStepResourcesByRole = (s: utils.Step | null, r: string) => (s?.Resources ?? []).filter((res) => res.Role === r);

  const updateResourceDraft = (resourceId: number, patch: Partial<ResourceDraft>) => {
    resourceEdits[resourceId] = { ...(resourceEdits[resourceId] ?? { tableId: entities[0]?.Id ?? 0, role: "Input" }), ...patch };
  };

  const handleResourceTableChange = (resourceId: number, event: Event) => {
    const target = event.currentTarget as HTMLSelectElement;
    updateResourceDraft(resourceId, { tableId: Number(target.value) });
  };

  const handleResourceRoleChange = (resourceId: number, event: Event) => {
    const target = event.currentTarget as HTMLSelectElement;
    updateResourceDraft(resourceId, { role: target.value });
  };

  const persistFlowChange = async (action: () => Promise<void>, successMessage: string, busyKey = "flow", options: { throwOnError?: boolean } = {}) => {
    if (busySection !== null) return false;
    busySection = busyKey;
    try {
      await action();
      await Save();
      await onRefresh();
      await tick();
      showToast(successMessage, "success");
      return true;
    } catch (err) {
      const message = extractError(err);
      showToast(`No se pudo actualizar el flujo: ${message}`, "error");
      if (options.throwOnError) throw new Error(message);
      return false;
    } finally { busySection = null; }
  };

  const updateAutoScroll = (event: DragEvent) => {
    if (typeof window === "undefined") return;
    if (event.clientY <= AUTO_SCROLL_EDGE_PX) { autoScrollDirection = -1; }
    else if (event.clientY >= window.innerHeight - AUTO_SCROLL_EDGE_PX) { autoScrollDirection = 1; }
    else { autoScrollDirection = 0; }
  };

  const stopAutoScroll = () => { autoScrollDirection = 0; };

  const reorderBigProcesses = async (fromId: number, toId: number) => {
    if (fromId === toId) return;
    const fromIndex = bigProcesses.findIndex((item) => item.Id === fromId);
    const toIndex = bigProcesses.findIndex((item) => item.Id === toId);
    if (fromIndex < 0 || toIndex < 0) return;
    const direction: "up" | "down" = toIndex < fromIndex ? "up" : "down";
    const steps = Math.abs(toIndex - fromIndex);
    selectedBigProcessId = fromId;
    selectedProcessId = null;
    selectedStepId = null;
    await persistFlowChange(async () => { for (let idx = 0; idx < steps; idx++) await MoveBigProcess(fromId, direction); }, `Macroflujo movido hacia ${direction === "up" ? "arriba" : "abajo"}.`, "drag-big-process");
  };

  const reorderProcesses = async (fromId: number, toId: number) => {
    if (!currentBigProcess || fromId === toId) return;
    const fromIndex = currentProcesses.findIndex((item) => item.Id === fromId);
    const toIndex = currentProcesses.findIndex((item) => item.Id === toId);
    if (fromIndex < 0 || toIndex < 0) return;
    const direction: "up" | "down" = toIndex < fromIndex ? "up" : "down";
    const steps = Math.abs(toIndex - fromIndex);
    selectedProcessId = fromId;
    selectedStepId = null;
    await persistFlowChange(async () => { for (let idx = 0; idx < steps; idx++) await MoveProcess(currentBigProcess!.Id, fromId, direction); }, `Proceso movido hacia ${direction === "up" ? "arriba" : "abajo"}.`, "drag-process");
  };

  const reorderSteps = async (fromId: number, toId: number) => {
    if (!currentBigProcess || !currentProcess || fromId === toId) return;
    const fromIndex = currentSteps.findIndex((item) => item.Id === fromId);
    const toIndex = currentSteps.findIndex((item) => item.Id === toId);
    if (fromIndex < 0 || toIndex < 0) return;
    const direction: "up" | "down" = toIndex < fromIndex ? "up" : "down";
    const steps = Math.abs(toIndex - fromIndex);
    selectedStepId = fromId;
    await persistFlowChange(async () => { for (let idx = 0; idx < steps; idx++) await MoveStep(currentBigProcess!.Id, currentProcess!.Id, fromId, direction); }, `Paso movido hacia ${direction === "up" ? "arriba" : "abajo"}.`, "drag-step");
  };

  const startBigProcessDrag = (id: number, event: DragEvent) => {
    if (busySection !== null) { event.preventDefault(); return; }
    draggingBigProcessId = id;
    hoverBigProcessId = id;
    event.dataTransfer?.setData("text/plain", `${id}`);
    if (event.dataTransfer) event.dataTransfer.effectAllowed = "move";
  };

  const handleBigProcessDragOver = (id: number, event: DragEvent) => { event.preventDefault(); hoverBigProcessId = id; updateAutoScroll(event); };
  const handleBigProcessDrop = async (id: number, event: DragEvent) => {
    event.preventDefault();
    if (draggingBigProcessId === null) return;
    const draggingId = draggingBigProcessId;
    draggingBigProcessId = null;
    hoverBigProcessId = null;
    stopAutoScroll();
    await reorderBigProcesses(draggingId, id);
  };
  const clearBigProcessDrag = () => { draggingBigProcessId = null; hoverBigProcessId = null; stopAutoScroll(); };

  const startProcessDrag = (id: number, event: DragEvent) => {
    if (busySection !== null) { event.preventDefault(); return; }
    draggingProcessId = id;
    hoverProcessId = id;
    event.dataTransfer?.setData("text/plain", `${id}`);
    if (event.dataTransfer) event.dataTransfer.effectAllowed = "move";
  };
  const handleProcessDragOver = (id: number, event: DragEvent) => { event.preventDefault(); hoverProcessId = id; updateAutoScroll(event); };
  const handleProcessDrop = async (id: number, event: DragEvent) => {
    event.preventDefault();
    if (draggingProcessId === null) return;
    const draggingId = draggingProcessId;
    draggingProcessId = null;
    hoverProcessId = null;
    stopAutoScroll();
    await reorderProcesses(draggingId, id);
  };
  const clearProcessDrag = () => { draggingProcessId = null; hoverProcessId = null; stopAutoScroll(); };

  const startStepDrag = (id: number, event: DragEvent) => {
    if (busySection !== null) { event.preventDefault(); return; }
    draggingStepId = id;
    hoverStepId = id;
    event.dataTransfer?.setData("text/plain", `${id}`);
    if (event.dataTransfer) event.dataTransfer.effectAllowed = "move";
  };
  const handleStepDragOver = (id: number, event: DragEvent) => { event.preventDefault(); hoverStepId = id; updateAutoScroll(event); };
  const handleStepDrop = async (id: number, event: DragEvent) => {
    event.preventDefault();
    if (draggingStepId === null) return;
    const draggingId = draggingStepId;
    draggingStepId = null;
    hoverStepId = null;
    stopAutoScroll();
    await reorderSteps(draggingId, id);
  };
  const clearStepDrag = () => { draggingStepId = null; hoverStepId = null; stopAutoScroll(); };

  const selectBigProcess = async (bigProcessId: number) => { await runStageTransition(async () => { selectedBigProcessId = bigProcessId; selectedProcessId = null; selectedStepId = null; await tick(); }); };
  const selectProcess = async (processId: number) => { await runStageTransition(async () => { selectedProcessId = processId; selectedStepId = null; await tick(); }); };

  const prepareBigProcessCreate = () => { newBigProcessName = ""; newBigProcessDescription = ""; };
  const prepareProcessCreate = (bigProcessId = currentBigProcess?.Id ?? null) => {
    if (bigProcessId === null) throw new Error("Primero crea o selecciona un macroflujo.");
    selectedBigProcessId = bigProcessId;
    selectedProcessId = null;
    selectedStepId = null;
    newProcessName = "";
    newProcessDescription = "";
  };
  const prepareStepCreate = (processId = currentProcess?.Id ?? null) => {
    if (!currentBigProcess || processId === null) throw new Error("Primero activa un proceso para poder colgarle pasos.");
    selectedBigProcessId = currentBigProcess.Id;
    selectedProcessId = processId;
    selectedStepId = null;
    newStepName = "";
    newStepDescription = "";
  };
  const prepareResourceCreate = (processId = currentProcess?.Id ?? null, stepId = currentStep?.Id ?? null) => {
    if (!currentBigProcess || processId === null || stepId === null) throw new Error("Activa un paso antes de vincular tablas.");
    if (!entities.length) throw new Error("Necesitas entidades para vincular recursos.");
    selectedBigProcessId = currentBigProcess.Id;
    selectedProcessId = processId;
    selectedStepId = stepId;
    newResourceTableId = entities[0]?.Id ?? null;
    newResourceRole = "Input";
  };

  const handleAddBigProcess = async () => {
    const name = newBigProcessName.trim();
    if (!name) throw new Error("Ingresa un nombre para el macroflujo.");
    const nextBigProcessId = (project?.BigProcessLastMax ?? 0) + 1;
    selectedBigProcessId = nextBigProcessId;
    selectedProcessId = null;
    selectedStepId = null;
    await persistFlowChange(() => AddBigProcess(name, newBigProcessDescription.trim()), "Macroflujo creado.", "add-big-process", { throwOnError: true });
    newBigProcessName = ""; newBigProcessDescription = "";
  };

  const prepareBigProcessEdit = (bigProcess: utils.BigProcess) => {
    selectedBigProcessId = bigProcess.Id; selectedProcessId = null; selectedStepId = null;
    bigProcessDraftName = bigProcess.Name ?? ""; bigProcessDraftDescription = bigProcess.Description ?? "";
  };

  const handleSaveBigProcess = async (bigProcessId = currentBigProcess?.Id ?? null) => {
    if (bigProcessId === null) throw new Error("Selecciona un macroflujo para editarlo.");
    const name = bigProcessDraftName.trim();
    if (!name) throw new Error("El macroflujo necesita un nombre.");
    selectedBigProcessId = bigProcessId; selectedProcessId = null; selectedStepId = null;
    await persistFlowChange(() => EditBigProcess(bigProcessId, name, bigProcessDraftDescription.trim()), "Macroflujo actualizado.", "edit-big-process", { throwOnError: true });
  };

  const handleRemoveBigProcess = async (bigProcessId = currentBigProcess?.Id ?? null) => {
    if (bigProcessId === null) throw new Error("Selecciona un macroflujo para eliminarlo.");
    const currentIndex = bigProcesses.findIndex((item) => item.Id === bigProcessId);
    selectedBigProcessId = bigProcesses[currentIndex + 1]?.Id ?? bigProcesses[currentIndex - 1]?.Id ?? null;
    selectedProcessId = null; selectedStepId = null;
    await persistFlowChange(() => RemoveBigProcess(bigProcessId), "Macroflujo eliminado.", "remove-big-process", { throwOnError: true });
  };

  const handleAddProcess = async () => {
    if (!currentBigProcess) throw new Error("Primero crea o selecciona un macroflujo.");
    const name = newProcessName.trim();
    if (!name) throw new Error("Ingresa un nombre para el proceso.");
    selectedBigProcessId = currentBigProcess.Id;
    selectedProcessId = (project?.ProcessLastMax ?? 0) + 1;
    selectedStepId = null;
    await persistFlowChange(() => AddProcess(currentBigProcess!.Id, name, newProcessDescription.trim()), "Proceso agregado al macroflujo.", "add-process", { throwOnError: true });
    newProcessName = ""; newProcessDescription = "";
  };

  const prepareProcessEdit = (p: utils.Process) => {
    if (!currentBigProcess) throw new Error("Primero selecciona un macroflujo.");
    selectedBigProcessId = currentBigProcess.Id; selectedProcessId = p.Id; selectedStepId = null;
    processDraftName = p.Name ?? ""; processDraftDescription = p.Description ?? "";
  };

  const handleSaveProcess = async (processId = currentProcess?.Id ?? null) => {
    if (!currentBigProcess || processId === null) throw new Error("Selecciona un proceso para editarlo.");
    const name = processDraftName.trim();
    if (!name) throw new Error("El proceso necesita un nombre.");
    selectedProcessId = processId; selectedStepId = null;
    await persistFlowChange(() => EditProcess(currentBigProcess!.Id, processId, name, processDraftDescription.trim()), "Proceso actualizado.", "edit-process", { throwOnError: true });
  };

  const handleRemoveProcess = async (processId = currentProcess?.Id ?? null) => {
    if (!currentBigProcess || processId === null) throw new Error("Selecciona un proceso para eliminarlo.");
    const currentIndex = currentProcesses.findIndex((item) => item.Id === processId);
    selectedProcessId = currentProcesses[currentIndex + 1]?.Id ?? currentProcesses[currentIndex - 1]?.Id ?? null;
    selectedStepId = null;
    await persistFlowChange(() => RemoveProcess(currentBigProcess!.Id, processId), "Proceso eliminado.", "remove-process", { throwOnError: true });
  };

  const handleAddStep = async () => {
    if (!currentBigProcess || !currentProcess) throw new Error("Selecciona un proceso antes de crear pasos.");
    const name = newStepName.trim();
    if (!name) throw new Error("Ingresa un nombre para el paso.");
    selectedStepId = (project?.StepsLastMax ?? 0) + 1;
    await persistFlowChange(() => AddStep(currentBigProcess!.Id, currentProcess!.Id, name, newStepDescription.trim()), "Paso agregado al proceso.", "add-step", { throwOnError: true });
    newStepName = ""; newStepDescription = "";
  };

  const prepareStepEdit = (processId: number, s: utils.Step) => {
    if (!currentBigProcess) throw new Error("Primero selecciona un macroflujo.");
    selectedBigProcessId = currentBigProcess.Id; selectedProcessId = processId; selectedStepId = s.Id;
    stepDraftName = s.Name ?? ""; stepDraftDescription = s.Description ?? "";
  };

  const prepareStepDetail = (processId: number, stepId: number) => {
    if (!currentBigProcess) throw new Error("Primero selecciona un macroflujo.");
    selectedBigProcessId = currentBigProcess.Id; selectedProcessId = processId; selectedStepId = stepId;
  };

  const handleSaveStep = async (processId = currentProcess?.Id ?? null, stepId = currentStep?.Id ?? null) => {
    if (!currentBigProcess || processId === null || stepId === null) throw new Error("Selecciona un paso para editarlo.");
    const name = stepDraftName.trim();
    if (!name) throw new Error("El paso necesita un nombre.");
    selectedProcessId = processId; selectedStepId = stepId;
    await persistFlowChange(() => EditStep(currentBigProcess!.Id, processId, stepId, name, stepDraftDescription.trim()), "Paso actualizado.", "edit-step", { throwOnError: true });
  };

  const handleRemoveStep = async (processId = currentProcess?.Id ?? null, stepId = currentStep?.Id ?? null) => {
    if (!currentBigProcess || processId === null || stepId === null) throw new Error("Selecciona un paso para eliminarlo.");
    const steps = currentProcess?.Steps ?? [];
    const currentIndex = steps.findIndex((item) => item.Id === stepId);
    selectedProcessId = processId;
    selectedStepId = steps[currentIndex + 1]?.Id ?? steps[currentIndex - 1]?.Id ?? null;
    await persistFlowChange(() => RemoveStep(currentBigProcess!.Id, processId, stepId), "Paso eliminado.", "remove-step", { throwOnError: true });
  };

  const handleAddResource = async () => {
    if (!currentBigProcess || !currentProcess || !currentStep) throw new Error("Selecciona un paso antes de asociar tablas.");
    if (newResourceTableId === null) throw new Error("Necesitas entidades para vincular recursos.");
    await persistFlowChange(() => AddResource(currentBigProcess!.Id, currentProcess!.Id, currentStep!.Id, newResourceTableId!, newResourceRole), "Tabla vinculada al paso.", "add-resource", { throwOnError: true });
  };

  const prepareResourceEdit = (resource: utils.Resource) => { updateResourceDraft(resource.Id, { tableId: resource.TableId, role: resource.Role }); };

  const handleSaveResource = async (resourceId: number) => {
    if (!currentBigProcess || !currentProcess || !currentStep) throw new Error("Selecciona un paso antes de editar recursos.");
    const draft = resourceEdits[resourceId];
    if (!draft) throw new Error("No se encontró el recurso a editar.");
    await persistFlowChange(() => EditResource(currentBigProcess!.Id, currentProcess!.Id, currentStep!.Id, resourceId, draft.tableId, draft.role), "Recurso actualizado.", "edit-resource", { throwOnError: true });
  };

  const handleRemoveResource = async (resourceId: number) => {
    if (!currentBigProcess || !currentProcess || !currentStep) throw new Error("Selecciona un paso antes de eliminar recursos.");
    await persistFlowChange(() => RemoveResource(currentBigProcess!.Id, currentProcess!.Id, currentStep!.Id, resourceId), "Recurso eliminado del paso.", "remove-resource", { throwOnError: true });
  };

  $effect(() => {
    if (selectedBigProcessId === null && bigProcesses.length > 0) selectedBigProcessId = bigProcesses[0].Id;
    if (selectedBigProcessId !== null && !bigProcesses.some((item) => item.Id === selectedBigProcessId)) selectedBigProcessId = bigProcesses[0]?.Id ?? null;
  });

  $effect(() => {
    if (selectedProcessId === null && currentProcesses.length > 0) selectedProcessId = currentProcesses[0].Id;
    if (selectedProcessId !== null && !currentProcesses.some((item) => item.Id === selectedProcessId)) selectedProcessId = currentProcesses[0]?.Id ?? null;
  });

  $effect(() => {
    if (selectedStepId === null && currentSteps.length > 0) selectedStepId = currentSteps[0].Id;
    if (selectedStepId !== null && !currentSteps.some((item) => item.Id === selectedStepId)) selectedStepId = currentSteps[0]?.Id ?? null;
  });

  $effect(() => {
    if ((currentBigProcess?.Id ?? null) !== lastBigProcessSyncId) {
      bigProcessDraftName = currentBigProcess?.Name ?? "";
      bigProcessDraftDescription = currentBigProcess?.Description ?? "";
      lastBigProcessSyncId = currentBigProcess?.Id ?? null;
    }
  });

  $effect(() => {
    if ((currentProcess?.Id ?? null) !== lastProcessSyncId) {
      processDraftName = currentProcess?.Name ?? "";
      processDraftDescription = currentProcess?.Description ?? "";
      lastProcessSyncId = currentProcess?.Id ?? null;
    }
  });

  $effect(() => {
    if ((currentStep?.Id ?? null) !== lastStepSyncId) {
      stepDraftName = currentStep?.Name ?? "";
      stepDraftDescription = currentStep?.Description ?? "";
      lastStepSyncId = currentStep?.Id ?? null;
    }
  });

  $effect(() => {
    if (newResourceTableId === null && entities.length > 0) newResourceTableId = entities[0].Id;
  });

  $effect(() => {
    const signature = currentStep ? `${currentStep.Id}:${(currentStep.Resources ?? []).map((r) => `${r.Id}-${r.TableId}-${r.Role}`).join("|")}` : "";
    if (signature !== lastResourceSignature) {
      if (currentStep) {
        const nextEdits: Record<number, ResourceDraft> = {};
        for (const resource of currentStep.Resources ?? []) {
          nextEdits[resource.Id] = { tableId: resource.TableId, role: resource.Role };
        }
        resourceEdits = nextEdits;
      } else { resourceEdits = {}; }
      lastResourceSignature = signature;
    }
  });

  $effect(() => {
    if (autoScrollDirection === 0) return;
    const interval = setInterval(() => { window.scrollBy({top: autoScrollDirection * AUTO_SCROLL_STEP, behavior: "auto"}); }, 16);
    return () => clearInterval(interval);
  });
</script>

<svelte:window
  ondragover={updateAutoScroll}
  ondrop={stopAutoScroll}
  ondragend={stopAutoScroll}
/>

<section class="flows-tab">
  <div class="tab-toolbar flow-toolbar">
    <div>
      <p class="label">Flujos</p>
      <p class="muted">Controla macroflujos, procesos y pasos en un deck operativo pensado para diagramar trabajo real.</p>
    </div>
    <div class="flow-toolbar__meta">
      <span class="signal-chip">{bigProcesses.length} macroflujos</span>
      <span class="signal-chip signal-chip--quiet">{currentProcesses.length} procesos activos</span>
      <span class="signal-chip signal-chip--quiet">{currentSteps.length} pasos en este proceso</span>
    </div>
  </div>

  <div class="studio-shell">
    <aside class="rail-panel">
      <div class="panel-head">
        <div>
          <p class="section-kicker">Deck</p>
          <h3>Macroflujos</h3>
          <p class="drag-hint">Arrastra para reordenar</p>
        </div>
        <ModalLauncher
          triggerLabel="Crear macroflujo"
          title="Crear macroflujo"
          confirmLabel="Crear"
          triggerVariant="primary"
          confirmVariant="primary"
          size="form"
          triggerClass="flow-modal-trigger flow-modal-trigger--rail"
          triggerDisabled={busySection !== null}
          onOpen={prepareBigProcessCreate}
          onSuccess={handleAddBigProcess}
        >
          {#snippet children()}
            <div class="modal-intro">
              <p class="modal-kicker local-modal-kicker">Nuevo macroflujo</p>
              <p class="modal-hint local-modal-hint">Abre un frente completo para agrupar procesos hermanos y darle un nombre operativo claro.</p>
            </div>
            <label class="field">
              <span>Nombre</span>
              <input type="text" bind:value={newBigProcessName} placeholder="Produccion, compras, despacho..." />
            </label>
            <label class="field">
              <span>Descripción</span>
              <textarea rows="3" bind:value={newBigProcessDescription} placeholder="Enmarca el objetivo general del bloque." />
            </label>
          {/snippet}
        </ModalLauncher>
      </div>

      {#if bigProcesses.length}
        <div class="macro-list">
          {#each bigProcesses as bigProcess, index (bigProcess.Id)}
            <button
              type="button"
              class:macro-card={true}
              class:macro-card--active={bigProcess.Id === currentBigProcess?.Id}
              class:macro-card--dragging={bigProcess.Id === draggingBigProcessId}
              class:macro-card--drop-target={bigProcess.Id === hoverBigProcessId && bigProcess.Id !== draggingBigProcessId}
              draggable="true"
              onclick={() => selectBigProcess(bigProcess.Id)}
              ondragstart={(event) => startBigProcessDrag(bigProcess.Id, event)}
              ondragover={(event) => handleBigProcessDragOver(bigProcess.Id, event)}
              ondrop={(event) => handleBigProcessDrop(bigProcess.Id, event)}
              ondragend={clearBigProcessDrag}
              in:fly={{y: 18, delay: index * 40, duration: 360, easing: quintOut}}
              animate:flip={{duration: 420, easing: quintOut}}
              style={`view-transition-name: macro-flow-${bigProcess.Id};`}
            >
              <span class="macro-card__index">{String(index + 1).padStart(2, "0")}</span>
              <ButtonIcon name="flows"/>
              <strong>{bigProcess.Name}</strong>
              <span class="macro-card__meta">{(bigProcess.Processes ?? []).length} procesos · {countBigProcessSteps(bigProcess)} pasos</span>
              <span class="macro-card__hint">{bigProcess.Description || "Sin descripción todavía."}</span>
            </button>
          {/each}
        </div>
      {:else}
        <div class="empty-state empty-state--rail">
          <strong>Aquí nacen los grandes bloques</strong>
          <p>Crea un macroflujo para agrupar etapas como producción, ventas o logística.</p>
        </div>
      {/if}

      {#if currentBigProcess}
        <section class="rail-subpanel">
          <div class="rail-subpanel__head">
            <div>
              <p class="section-kicker">Proceso activo</p>
              <h4>{currentBigProcess.Name}</h4>
              <p>{currentProcesses.length} procesos dentro de este macro.</p>
            </div>
            <ModalLauncher
              triggerLabel="Agregar proceso"
              title="Crear proceso"
              confirmLabel="Crear"
              triggerVariant="primary"
              confirmVariant="primary"
              size="form"
              triggerClass="flow-modal-trigger flow-modal-trigger--subrail"
              triggerDisabled={busySection !== null}
              onOpen={() => prepareProcessCreate(currentBigProcess.Id)}
              onSuccess={handleAddProcess}
            >
              {#snippet children()}
                <div class="modal-intro">
                  <p class="modal-kicker local-modal-kicker">Nuevo proceso</p>
                  <p class="modal-hint local-modal-hint">Se agregará dentro de <strong>{currentBigProcess.Name}</strong>.</p>
                </div>
                <div class="modal-form-grid">
                  <label class="field">
                    <span>Nombre del proceso</span>
                    <input type="text" bind:value={newProcessName} placeholder="Aprobar orden, despachar, validar..." />
                  </label>
                  <label class="field">
                    <span>Descripción</span>
                    <textarea rows="3" bind:value={newProcessDescription} placeholder="Qué resuelve este proceso dentro del macroflujo." />
                  </label>
                </div>
              {/snippet}
            </ModalLauncher>
          </div>

          {#if currentProcesses.length}
            <div class="process-rail-list">
              {#each currentProcesses as process, index (process.Id)}
                <button
                  type="button"
                  class:process-rail-card={true}
                  class:process-rail-card--active={process.Id === currentProcess?.Id}
                  onclick={() => selectProcess(process.Id)}
                >
                  <span class="process-rail-card__index">{String(index + 1).padStart(2, "0")}</span>
                  <ButtonIcon name="stack"/>
                  <strong>{process.Name}</strong>
                  <span>{(process.Steps ?? []).length} pasos</span>
                </button>
              {/each}
            </div>
          {:else}
            <div class="empty-stage-fragment empty-stage-fragment--resource">
              <strong>Este macro aún no tiene procesos</strong>
              <p>Agrega el primero para empezar a navegar el detalle sin salir del rail.</p>
            </div>
          {/if}
        </section>
      {/if}
    </aside>

    <section class="stage-panel">
      {#if currentBigProcess}
        <header class="stage-hero" style={`view-transition-name: stage-big-process-${currentBigProcess.Id};`}>
          <div class="stage-hero__copy">
            <p class="section-kicker">Control Deck</p>
            <h2>{currentBigProcess.Name}</h2>
            <p>{currentBigProcess.Description || "Usa este contenedor para ordenar procesos hermanos y mantener visible el recorrido completo."}</p>
          </div>
          <div class="stage-hero__side">
            <div class="hero-stats">
              <div class="hero-stat">
                <span>Procesos</span>
                <strong>{currentProcesses.length}</strong>
              </div>
              <div class="hero-stat">
                <span>Pasos</span>
                <strong>{countBigProcessSteps(currentBigProcess)}</strong>
              </div>
              <div class="hero-stat">
                <span>Tablas ligadas</span>
                <strong>{countBigProcessResources(currentBigProcess)}</strong>
              </div>
            </div>
            <div class="hero-actions">
              <ModalLauncher
                triggerLabel="Editar macroflujo"
                title="Editar macroflujo"
                confirmLabel="Guardar"
                triggerVariant="edit"
                confirmVariant="primary"
                size="form"
                triggerClass="flow-modal-trigger flow-modal-trigger--hero"
                triggerDisabled={busySection !== null}
                onOpen={() => prepareBigProcessEdit(currentBigProcess)}
                onSuccess={() => handleSaveBigProcess(currentBigProcess.Id)}
              >
                {#snippet children()}
                  <div class="modal-intro">
                    <p class="modal-kicker local-modal-kicker">Editor macroflujo</p>
                    <p class="modal-hint local-modal-hint">Actualiza el nombre y el alcance operativo del bloque principal.</p>
                  </div>
                  <label class="field">
                    <span>Nombre</span>
                    <input type="text" bind:value={bigProcessDraftName} placeholder="Macroflujo principal" />
                  </label>
                  <label class="field">
                    <span>Descripción</span>
                    <textarea rows="3" bind:value={bigProcessDraftDescription} placeholder="Describe el frente operativo que contiene." />
                  </label>
                {/snippet}
              </ModalLauncher>
              <ModalLauncher
                triggerLabel="Eliminar macroflujo"
                title="Eliminar macroflujo"
                confirmLabel="Eliminar"
                triggerVariant="danger"
                confirmVariant="danger"
                size="default"
                triggerClass="flow-modal-trigger flow-modal-trigger--hero"
                triggerDisabled={busySection !== null}
                onSuccess={() => handleRemoveBigProcess(currentBigProcess.Id)}
              >
                {#snippet children()}
                  <p class="modal-hint local-modal-hint">Se eliminará <strong>{currentBigProcess.Name}</strong> con todos sus procesos y pasos.</p>
                {/snippet}
              </ModalLauncher>
            </div>
          </div>
        </header>

        {#if currentProcess}
          <article
            class:process-lane={true}
            class:process-lane--active={true}
            class:process-lane--dragging={currentProcess.Id === draggingProcessId}
            class:process-lane--drop-target={currentProcess.Id === hoverProcessId && currentProcess.Id !== draggingProcessId}
            draggable="true"
            ondragstart={(event) => startProcessDrag(currentProcess.Id, event)}
            ondragover={(event) => handleProcessDragOver(currentProcess.Id, event)}
            ondrop={(event) => handleProcessDrop(currentProcess.Id, event)}
            ondragend={clearProcessDrag}
            style={`view-transition-name: process-lane-${currentProcess.Id};`}
          >
            <header class="process-lane__head">
              <div>
                <p class="section-kicker">Detalle del proceso</p>
                <h3>{currentProcess.Name}</h3>
                <p>{currentProcess.Description || "Sin descripción todavía."}</p>
              </div>
              <div class="process-lane__side">
                <div class="process-lane__meta">
                  <span>{(currentProcess.Steps ?? []).length} pasos</span>
                  <span>{countProcessResources(currentProcess)} tablas</span>
                </div>
                <div class="process-lane__actions">
                  <ModalLauncher
                    triggerLabel="Editar"
                    title="Editar proceso"
                    confirmLabel="Guardar"
                    triggerVariant="edit"
                    confirmVariant="primary"
                    size="form"
                    triggerClass="flow-modal-trigger flow-modal-trigger--inline"
                    triggerDisabled={busySection !== null}
                    onOpen={() => prepareProcessEdit(currentProcess)}
                    onSuccess={() => handleSaveProcess(currentProcess.Id)}
                  >
                    {#snippet children()}
                      <div class="modal-intro">
                        <p class="modal-kicker local-modal-kicker">Editor proceso</p>
                        <p class="modal-hint local-modal-hint">Ajusta la línea de trabajo sin salir del macroflujo actual.</p>
                      </div>
                      <label class="field">
                        <span>Nombre</span>
                        <input type="text" bind:value={processDraftName} placeholder="Proceso específico" />
                      </label>
                      <label class="field">
                        <span>Descripción</span>
                        <textarea rows="3" bind:value={processDraftDescription} placeholder="Qué ocurre en esta línea de trabajo." />
                      </label>
                    {/snippet}
                  </ModalLauncher>
                  <ModalLauncher
                    triggerLabel="Eliminar"
                    title="Eliminar proceso"
                    confirmLabel="Eliminar"
                    triggerVariant="danger"
                    confirmVariant="danger"
                    size="default"
                    triggerClass="flow-modal-trigger flow-modal-trigger--inline"
                    triggerDisabled={busySection !== null}
                    onSuccess={() => handleRemoveProcess(currentProcess.Id)}
                  >
                    {#snippet children()}
                      <p class="modal-hint local-modal-hint">Se eliminará <strong>{currentProcess.Name}</strong> con todos sus pasos.</p>
                    {/snippet}
                  </ModalLauncher>
                </div>
              </div>
            </header>

            {#if currentProcess.Steps?.length}
              <div class="step-sequence">
                {#each currentProcess.Steps as step, stepIndex (step.Id)}
                  <article
                    class:step-node={true}
                    class:step-node--dragging={step.Id === draggingStepId}
                    class:step-node--drop-target={step.Id === hoverStepId && step.Id !== draggingStepId}
                    draggable="true"
                    ondragstart={(event) => startStepDrag(step.Id, event)}
                    ondragover={(event) => handleStepDragOver(step.Id, event)}
                    ondrop={(event) => handleStepDrop(step.Id, event)}
                    ondragend={clearStepDrag}
                    in:scale={{duration: 280, delay: stepIndex * 55, start: 0.92, easing: quintOut}}
                    animate:flip={{duration: 320, easing: quintOut}}
                    style={`view-transition-name: step-node-${step.Id};`}
                  >
                    <span class="step-node__order">{String(step.Order).padStart(2, "0")}</span>
                    <span class="step-node__body">
                      <strong>{step.Name}</strong>
                      <span>{step.Description || "Describe qué pasa en este punto."}</span>
                    </span>
                    <!-- svelte-ignore a11y_no_static_element_interactions -->
                    <!-- svelte-ignore a11y_click_events_have_key_events -->
                    <div class="step-node__meta" onclick={(e) => e.stopPropagation()}>
                      <div class="step-node__meta-top">
                        <span class="step-node__resource-count">{step.Resources?.length ?? 0}</span>
                        <div class="step-node__actions">
                          <ModalLauncher
                            triggerLabel="Editar"
                            title="Editar paso"
                            confirmLabel="Guardar"
                            triggerVariant="edit"
                            confirmVariant="primary"
                            size="form"
                            triggerClass="flow-modal-trigger flow-modal-trigger--inline"
                            triggerDisabled={busySection !== null}
                            onOpen={() => prepareStepEdit(currentProcess.Id, step)}
                            onSuccess={() => handleSaveStep(currentProcess.Id, step.Id)}
                          >
                            {#snippet children()}
                              <div class="modal-intro">
                                <p class="modal-kicker local-modal-kicker">Editor paso</p>
                                <p class="modal-hint local-modal-hint">Refina el nombre y la descripción de este punto del flujo.</p>
                              </div>
                              <label class="field">
                                <span>Nombre del paso</span>
                                <input type="text" bind:value={stepDraftName} placeholder="Nombre del hito" />
                              </label>
                              <label class="field">
                                <span>Descripción</span>
                                <textarea rows="3" bind:value={stepDraftDescription} placeholder="Qué debe pasar exactamente aquí." />
                              </label>
                            {/snippet}
                          </ModalLauncher>
                          <ModalLauncher
                            triggerLabel="Vincular tabla"
                            title="Vincular tabla al paso"
                            confirmLabel="Vincular"
                            triggerVariant="success"
                            confirmVariant="primary"
                            size="form"
                            triggerClass="flow-modal-trigger flow-modal-trigger--inline"
                            triggerDisabled={busySection !== null || !entities.length}
                            onOpen={() => prepareResourceCreate(currentProcess.Id, step.Id)}
                            onSuccess={handleAddResource}
                          >
                            {#snippet children()}
                              <div class="modal-intro">
                                <p class="modal-kicker local-modal-kicker">Nuevo recurso</p>
                                <p class="modal-hint local-modal-hint">Asocia una entidad como input u output para este paso.</p>
                              </div>
                              <div class="modal-form-grid modal-form-grid--compact">
                                <label class="field">
                                  <span>Tabla</span>
                                  <select bind:value={newResourceTableId} disabled={!entities.length}>
                                    {#each entities as entity}
                                      <option value={entity.Id}>{entity.Name}</option>
                                    {/each}
                                  </select>
                                </label>
                                <label class="field">
                                  <span>Rol</span>
                                  <select bind:value={newResourceRole}>
                                    {#each resourceRoles as role}
                                      <option value={role}>{role}</option>
                                    {/each}
                                  </select>
                                </label>
                              </div>
                            {/snippet}
                          </ModalLauncher>
                          <ModalLauncher
                            triggerLabel="Eliminar"
                            title="Eliminar paso"
                            confirmLabel="Eliminar"
                            triggerVariant="danger"
                            confirmVariant="danger"
                            size="default"
                            triggerClass="flow-modal-trigger flow-modal-trigger--inline"
                            triggerDisabled={busySection !== null}
                            onSuccess={() => handleRemoveStep(currentProcess.Id, step.Id)}
                          >
                            {#snippet children()}
                              <p class="modal-hint local-modal-hint">Se eliminará <strong>{step.Name}</strong> de este proceso.</p>
                            {/snippet}
                          </ModalLauncher>
                          <ModalLauncher
                            triggerLabel="Ver detalle"
                            title="Detalle de tablas vinculadas"
                            confirmLabel="Cerrar"
                            triggerVariant="primary"
                            confirmVariant="secondary"
                            size="form"
                            triggerClass="flow-modal-trigger flow-modal-trigger--inline"
                            triggerDisabled={busySection !== null || !(step.Resources?.length)}
                            onOpen={() => prepareStepDetail(currentProcess.Id, step.Id)}
                            onSuccess={async () => {}}
                          >
                            {#snippet children()}
                              <div class="modal-intro">
                                <p class="modal-kicker local-modal-kicker">Recursos del paso</p>
                                <p class="modal-hint local-modal-hint">Revisa la definición completa de cada tabla y ajusta su rol si hace falta.</p>
                              </div>
                              {#if step.Resources?.length}
                                <div class="resource-list">
                                  {#each step.Resources as resource (resource.Id)}
                                    <article
                                      class="resource-row"
                                      transition:scale={{duration: 180, start: 0.96}}
                                      animate:flip={{duration: 280, easing: quintOut}}
                                    >
                                      <div class="resource-row__identity">
                                        <span class="resource-row__index">#{resource.Id} · {resource.Role}</span>
                                        <strong>{entityLabel(resource.TableId)}</strong>
                                        <p>{entityDescription(resource.TableId)}</p>
                                      </div>
                                      <div class="resource-row__actions">
                                        <ModalLauncher
                                          triggerLabel="Editar"
                                          title="Editar recurso"
                                          confirmLabel="Guardar"
                                          triggerVariant="edit"
                                          confirmVariant="primary"
                                          size="form"
                                          triggerClass="flow-modal-trigger flow-modal-trigger--inline"
                                          triggerDisabled={busySection !== null}
                                          onOpen={() => prepareResourceEdit(resource)}
                                          onSuccess={() => handleSaveResource(resource.Id)}
                                        >
                                          {#snippet children()}
                                            <div class="modal-intro">
                                              <p class="modal-kicker local-modal-kicker">Editor recurso</p>
                                              <p class="modal-hint local-modal-hint">Ajusta la tabla relacionada y el rol que cumple en este paso.</p>
                                            </div>
                                            <div class="modal-form-grid modal-form-grid--compact">
                                              <label class="field">
                                                <span>Tabla</span>
                                                <select
                                                  value={resourceEdits[resource.Id]?.tableId ?? resource.TableId}
                                                  onchange={(event) => handleResourceTableChange(resource.Id, event)}
                                                >
                                                  {#each entities as entity}
                                                    <option value={entity.Id}>{entity.Name}</option>
                                                  {/each}
                                                </select>
                                              </label>
                                              <label class="field">
                                                <span>Rol</span>
                                                <select
                                                  value={resourceEdits[resource.Id]?.role ?? resource.Role}
                                                  onchange={(event) => handleResourceRoleChange(resource.Id, event)}
                                                >
                                                  {#each resourceRoles as role}
                                                    <option value={role}>{role}</option>
                                                  {/each}
                                                </select>
                                              </label>
                                            </div>
                                          {/snippet}
                                        </ModalLauncher>
                                        <ModalLauncher
                                          triggerLabel="Quitar"
                                          title="Quitar recurso"
                                          confirmLabel="Quitar"
                                          triggerVariant="danger"
                                          confirmVariant="danger"
                                          size="default"
                                          triggerClass="flow-modal-trigger flow-modal-trigger--inline"
                                          triggerDisabled={busySection !== null}
                                          onSuccess={() => handleRemoveResource(resource.Id)}
                                        >
                                          {#snippet children()}
                                            <p class="modal-hint local-modal-hint">Se quitará <strong>{entityLabel(resource.TableId)}</strong> de este paso.</p>
                                          {/snippet}
                                        </ModalLauncher>
                                      </div>
                                    </article>
                                  {/each}
                                </div>
                              {:else}
                                <div class="empty-stage-fragment empty-stage-fragment--resource">
                                  <strong>Aún no hay tablas conectadas</strong>
                                  <p>Vincula entidades como input u output desde los botones del paso.</p>
                                </div>
                              {/if}
                            {/snippet}
                          </ModalLauncher>
                        </div>
                      </div>
                      <div class="resource-summary">
                        <div class="resource-summary__group">
                          <span class="resource-summary__label resource-summary__label--input">Input</span>
                          <div class="resource-summary__chips">
                            {#if getStepResourcesByRole(step, "Input").length}
                              {#each getStepResourcesByRole(step, "Input") as resource (resource.Id)}
                                <span class="resource-chip resource-chip--input">{entityLabel(resource.TableId)}</span>
                              {/each}
                            {:else}
                              <span class="resource-chip resource-chip--ghost">Sin inputs</span>
                            {/if}
                          </div>
                        </div>
                        <div class="resource-summary__group">
                          <span class="resource-summary__label resource-summary__label--output">Output</span>
                          <div class="resource-summary__chips">
                            {#if getStepResourcesByRole(step, "Output").length}
                              {#each getStepResourcesByRole(step, "Output") as resource (resource.Id)}
                                <span class="resource-chip resource-chip--output">{entityLabel(resource.TableId)}</span>
                              {/each}
                            {:else}
                              <span class="resource-chip resource-chip--ghost">Sin outputs</span>
                            {/if}
                          </div>
                        </div>
                      </div>
                    </div>
                  </article>
                {/each}
              </div>
            {:else}
              <div class="empty-stage-fragment">
                <strong>Este proceso todavía no respira</strong>
                <p>Agrega pasos para convertirlo en una secuencia operable.</p>
              </div>
            {/if}

            <footer class="process-lane__footer">
              <ModalLauncher
                triggerLabel="Agregar paso"
                title="Crear paso"
                confirmLabel="Crear"
                triggerVariant="primary"
                confirmVariant="primary"
                size="form"
                triggerClass="flow-modal-trigger flow-modal-trigger--tail"
                triggerDisabled={busySection !== null}
                onOpen={() => prepareStepCreate(currentProcess.Id)}
                onSuccess={handleAddStep}
              >
                {#snippet children()}
                  <div class="modal-intro">
                    <p class="modal-kicker local-modal-kicker">Nuevo paso</p>
                    <p class="modal-hint local-modal-hint">Agrégalo al final de <strong>{currentProcess.Name}</strong> y luego reordénalo si hace falta.</p>
                  </div>
                  <label class="field">
                    <span>Nombre del paso</span>
                    <input type="text" bind:value={newStepName} placeholder="Validar datos, emitir comprobante..." />
                  </label>
                  <label class="field">
                    <span>Descripción</span>
                    <textarea rows="3" bind:value={newStepDescription} placeholder="Qué debe estar listo o qué debe suceder." />
                  </label>
                {/snippet}
              </ModalLauncher>
            </footer>
          </article>
        {:else if currentProcesses.length}
          <div class="empty-state empty-state--stage">
            <strong>Selecciona un proceso desde el rail</strong>
            <p>Usa la columna izquierda para entrar al detalle del proceso que quieras revisar.</p>
          </div>
        {:else}
          <div class="empty-state empty-state--stage">
            <strong>Sin procesos en este macroflujo</strong>
            <p>Comienza con una línea principal como “aprobar orden” o “despachar” y luego agrega pasos detallados.</p>
          </div>
        {/if}
      {:else}
        <div class="empty-state empty-state--stage">
          <strong>Selecciona o crea un macroflujo</strong>
          <p>La pestaña está lista para modelar recorridos completos, pero primero necesita su bloque principal.</p>
        </div>
      {/if}
    </section>
  </div>
</section>

<style>
  .flows-tab {
    display: grid;
    gap: 1rem;
  }

  .flow-toolbar {
    align-items: end;
  }

  .flow-toolbar__meta {
    display: flex;
    flex-wrap: wrap;
    gap: 0.6rem;
    justify-content: flex-end;
  }

  .label,
  .section-kicker {
    margin: 0;
    color: var(--accent);
    font-size: 0.75rem;
    font-weight: 800;
    letter-spacing: 0.18em;
    text-transform: uppercase;
  }

  .muted {
    margin: 0.35rem 0 0;
    max-width: 74ch;
    color: var(--ink-soft);
    line-height: 1.5;
  }

  .studio-shell {
    display: grid;
    grid-template-columns: minmax(16rem, 0.72fr) minmax(0, 1.7fr);
    gap: 1rem;
    align-items: start;
  }

  .rail-panel,
  .stage-panel {
    position: relative;
    overflow: hidden;
    border: 1px solid var(--border);
    border-radius: calc(var(--radius-lg) - 8px);
    background: color-mix(in srgb, var(--surface-strong) 92%, transparent);
    box-shadow: var(--shadow-sm);
  }

  .rail-panel,
  .stage-panel {
    min-height: 43rem;
  }

  .rail-panel::before,
  .stage-panel::before {
    content: "";
    position: absolute;
    inset: 0;
    background:
      linear-gradient(120deg, color-mix(in srgb, var(--accent) 0%, transparent) 24%, color-mix(in srgb, var(--accent) 10%, transparent) 50%, color-mix(in srgb, var(--accent) 0%, transparent) 76%);
    opacity: 0.7;
    transform: translateX(-100%);
    animation: scan-surface 7.5s linear infinite;
    pointer-events: none;
  }

  .rail-panel {
    display: grid;
    gap: 0.9rem;
    padding: 1rem;
    background:
      linear-gradient(180deg, color-mix(in srgb, var(--surface-strong) 95%, transparent), color-mix(in srgb, var(--surface) 88%, transparent)),
      radial-gradient(circle at top, color-mix(in srgb, var(--accent) 12%, transparent), transparent 42%);
  }

  .panel-head {
    display: flex;
    justify-content: space-between;
    align-items: start;
    gap: 0.8rem;
  }

  .stage-hero__copy {
    display: grid;
  }

  .panel-head h3,
  .stage-hero__copy h2 {
    margin: 0.2rem 0 0;
    font-size: clamp(1.2rem, 2vw, 2.05rem);
    line-height: 0.98;
  }

  .drag-hint {
    margin: 0.38rem 0 0;
    color: var(--ink-faint);
    font-size: 0.8rem;
  }

  .macro-list {
    display: grid;
    gap: 0.72rem;
  }

  .rail-subpanel {
    display: grid;
    gap: 0.8rem;
    padding-top: 0.25rem;
    border-top: 1px solid color-mix(in srgb, var(--ink) 10%, transparent);
  }

  .rail-subpanel__head {
    display: grid;
    gap: 0.65rem;
  }

  .rail-subpanel__head h4 {
    margin: 0.18rem 0 0;
    font-size: 1.02rem;
    line-height: 1.05;
  }

  .rail-subpanel__head p:last-child {
    margin: 0.35rem 0 0;
    color: var(--ink-soft);
    line-height: 1.45;
  }

  .process-rail-list {
    display: grid;
    gap: 0.58rem;
  }

  .process-rail-card {
    display: grid;
    gap: 0.24rem;
    padding: 0.82rem 0.9rem;
    border-radius: calc(var(--radius-md) - 10px);
    border: 1px solid var(--border);
    background: color-mix(in srgb, var(--surface-strong) 94%, transparent);
    text-align: left;
    transition:
      transform 180ms cubic-bezier(0.22, 1, 0.36, 1),
      border-color 180ms ease,
      box-shadow 180ms ease;
  }

  .process-rail-card:hover,
  .process-rail-card--active {
    transform: translateX(4px);
    border-color: color-mix(in srgb, var(--accent) 24%, var(--border));
    box-shadow:
      0 16px 24px color-mix(in srgb, var(--ink) 10%, transparent),
      0 0 0 0.16rem color-mix(in srgb, var(--accent) 10%, transparent);
  }

  .process-rail-card__index {
    color: var(--ink-faint);
    font-size: 0.72rem;
    font-weight: 800;
    letter-spacing: 0.14em;
  }

  .process-rail-card :global(.button-glyph),
  .macro-card :global(.button-glyph) {
    color: var(--accent-strong);
    opacity: 0.86;
  }

  .process-rail-card strong {
    font-size: 0.96rem;
    line-height: 1.2;
  }

  .process-rail-card span:last-child {
    color: var(--ink-soft);
    font-size: 0.78rem;
  }

  :global(.flow-modal-trigger) {
    width: 100%;
    justify-content: center;
  }

  :global(.flow-modal-trigger--rail) {
    min-width: min(100%, 13rem);
    width: auto;
  }

  :global(.flow-modal-trigger--inline) {
    min-height: 2.55rem;
    width: auto;
    padding-inline: 0.9rem;
  }

  :global(.flow-modal-trigger--tail) {
    margin-top: 1rem;
  }

  @keyframes scan-surface {
    to { transform: translateX(100%); }
  }

  .step-sequence {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 1.25rem;
    padding: 1.5rem;
  }

  .step-node {
    background: var(--background);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-lg);
    padding: 1.25rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    cursor: grab;
    transition: all 0.2s ease;
  }

  .step-node:active { cursor: grabbing; }

  .step-node:hover {
    border-color: var(--accent);
    box-shadow: var(--shadow-md);
  }

  .step-node--dragging { opacity: 0.5; }
  .step-node--drop-target { border: 2px dashed var(--accent); }

  .step-node__order {
    font-size: 0.65rem;
    font-weight: 900;
    color: var(--accent);
    opacity: 0.6;
  }

  .step-node__body strong { font-size: 0.95rem; }
  .step-node__body span { font-size: 0.75rem; color: var(--ink-soft); display: block; margin-top: 0.25rem; }

  .step-node__meta {
    margin-top: auto;
    padding-top: 0.75rem;
    border-top: 1px solid var(--border-card);
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .step-node__meta-top {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .step-node__resource-count {
    font-size: 0.7rem;
    font-weight: 800;
    color: var(--accent);
    background: var(--accent-ghost);
    padding: 0.2rem 0.6rem;
    border-radius: 99px;
  }

  .step-node__actions { display: flex; gap: 0.4rem; }

  .resource-summary { display: flex; flex-direction: column; gap: 0.4rem; }
  .resource-summary__label { font-size: 0.6rem; font-weight: 900; text-transform: uppercase; color: var(--ink-faint); margin-bottom: 0.2rem; display: block; }
  .resource-summary__chips { display: flex; flex-wrap: wrap; gap: 0.25rem; }

  .resource-chip { font-size: 0.65rem; font-weight: 700; padding: 0.1rem 0.4rem; border-radius: 4px; }
  .resource-chip--input { background: var(--accent-ghost); color: var(--accent); border: 1px solid var(--accent-soft); }
  .resource-chip--output { background: var(--success-ghost); color: var(--success); border: 1px solid var(--success-soft); }
  .resource-chip--ghost { color: var(--ink-faint); font-style: italic; }

  .stage-hero { padding: 2rem; border-bottom: 1px solid var(--border-card); display: flex; justify-content: space-between; align-items: flex-start; gap: 2rem; }
  .hero-stats { display: flex; gap: 2rem; }
  .hero-stat { display: flex; flex-direction: column; }
  .hero-stat span { font-size: 0.65rem; font-weight: 800; text-transform: uppercase; color: var(--ink-soft); }
  .hero-stat strong { font-size: 1.5rem; line-height: 1; margin-top: 0.25rem; }

  .process-lane { background: var(--surface); border: 1px solid var(--border-card); border-radius: var(--radius-lg); overflow: hidden; margin: 1.5rem; }
  .process-lane__head { padding: 1.5rem; border-bottom: 1px solid var(--border-card); display: flex; justify-content: space-between; align-items: center; }
  .process-lane__side { display: flex; align-items: center; gap: 2rem; }
  .process-lane__meta { display: flex; gap: 1rem; font-size: 0.75rem; font-weight: 700; color: var(--ink-soft); }
  .process-lane__actions { display: flex; gap: 0.5rem; }

  .resource-list { display: flex; flex-direction: column; gap: 1rem; padding: 1rem 0; }
  .resource-row { background: var(--background); border: 1px solid var(--border-card); border-radius: var(--radius-md); padding: 1rem; display: flex; justify-content: space-between; align-items: flex-start; }
  .resource-row__identity strong { display: block; }
  .resource-row__identity p { font-size: 0.75rem; color: var(--ink-soft); margin: 0.25rem 0 0; }
  .resource-row__index { font-size: 0.6rem; font-weight: 900; color: var(--accent); text-transform: uppercase; }
  .resource-row__actions { display: flex; gap: 0.5rem; }

  @media (max-width: 1200px) {
    .studio-shell { grid-template-columns: 1fr; }
    .rail-panel { min-height: auto; position: static; }
  }
</style>
