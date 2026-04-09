<script lang="ts">
  import { flip } from "svelte/animate";
  import { quintOut } from "svelte/easing";
  import { fade, fly, scale } from "svelte/transition";
  import { tick } from "svelte";
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

  type ResourceDraft = {
    tableId: number;
    role: string;
  };

  type ViewTransitionDocument = Document & {
    startViewTransition?: (update: () => void | Promise<void>) => {
      finished: Promise<void>;
    };
  };

  export let project: utils.DbProject;
  export let entities: utils.Entity[] = [];
  export let onRefresh: () => Promise<void> = async () => {};

  const resourceRoles = ["Input", "Output"];

  let selectedBigProcessId: number | null = null;
  let selectedProcessId: number | null = null;
  let selectedStepId: number | null = null;

  let lastBigProcessSyncId: number | null = null;
  let lastProcessSyncId: number | null = null;
  let lastStepSyncId: number | null = null;
  let lastResourceSignature = "";

  let bigProcessDraftName = "";
  let bigProcessDraftDescription = "";
  let processDraftName = "";
  let processDraftDescription = "";
  let stepDraftName = "";
  let stepDraftDescription = "";

  let newBigProcessName = "";
  let newBigProcessDescription = "";
  let newProcessName = "";
  let newProcessDescription = "";
  let newStepName = "";
  let newStepDescription = "";
  let newResourceTableId: number | null = null;
  let newResourceRole = "Input";

  let resourceEdits: Record<number, ResourceDraft> = {};
  let busySection: string | null = null;
  let draggingBigProcessId: number | null = null;
  let hoverBigProcessId: number | null = null;
  let draggingProcessId: number | null = null;
  let hoverProcessId: number | null = null;
  let draggingStepId: number | null = null;
  let hoverStepId: number | null = null;
  let autoScrollFrame: number | null = null;
  let autoScrollDirection: -1 | 0 | 1 = 0;

  const AUTO_SCROLL_EDGE_PX = 96;
  const AUTO_SCROLL_STEP = 18;

  const prefersReducedMotion = () =>
    typeof window !== "undefined"
    && typeof window.matchMedia === "function"
    && window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  const runStageTransition = async (update: () => void | Promise<void>) => {
    const doc = typeof document !== "undefined" ? (document as ViewTransitionDocument) : null;
    if (doc?.startViewTransition && !prefersReducedMotion()) {
      try {
        const transition = doc.startViewTransition(update);
        await transition.finished;
        return;
      } catch (err) {
        console.warn("No se pudo aplicar la transición de vista:", err);
      }
    }
    await update();
  };

  const extractError = (err: unknown) => {
    if (typeof err === "string") {
      return err;
    }
    if (err && typeof err === "object") {
      const maybeRecord = err as Record<string, unknown>;
      return String(maybeRecord.error ?? maybeRecord.message ?? "Error desconocido");
    }
    return "Error desconocido";
  };

  const entityLabel = (entityId: number) =>
    entities.find((entity) => entity.Id === entityId)?.Name ?? `Tabla ${entityId}`;

  const entityDescription = (entityId: number) =>
    entities.find((entity) => entity.Id === entityId)?.Description ?? "Sin detalle de la tabla.";

  const countProcessResources = (process: utils.Process) =>
    (process.Steps ?? []).reduce((sum, step) => sum + (step.Resources?.length ?? 0), 0);

  const countBigProcessSteps = (bigProcess: utils.BigProcess) =>
    (bigProcess.Processes ?? []).reduce((sum, process) => sum + (process.Steps?.length ?? 0), 0);

  const countBigProcessResources = (bigProcess: utils.BigProcess) =>
    (bigProcess.Processes ?? []).reduce((sum, process) => sum + countProcessResources(process), 0);

  const getStepResourcesByRole = (step: utils.Step | null, role: string) =>
    (step?.Resources ?? []).filter((resource) => resource.Role === role);

  const updateResourceDraft = (resourceId: number, patch: Partial<ResourceDraft>) => {
    resourceEdits = {
      ...resourceEdits,
      [resourceId]: {
        ...(resourceEdits[resourceId] ?? {
          tableId: entities[0]?.Id ?? 0,
          role: "Input"
        }),
        ...patch
      }
    };
  };

  const handleResourceTableChange = (resourceId: number, event: Event) => {
    const target = event.currentTarget as HTMLSelectElement;
    updateResourceDraft(resourceId, {tableId: Number(target.value)});
  };

  const handleResourceRoleChange = (resourceId: number, event: Event) => {
    const target = event.currentTarget as HTMLSelectElement;
    updateResourceDraft(resourceId, {role: target.value});
  };

  const persistFlowChange = async (
    action: () => Promise<void>,
    successMessage: string,
    busyKey = "flow",
    options: { throwOnError?: boolean } = {}
  ) => {
    if (busySection !== null) {
      return false;
    }

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
      if (options.throwOnError) {
        throw new Error(message);
      }
      return false;
    } finally {
      busySection = null;
    }
  };

  const runAutoScroll = () => {
    if (autoScrollDirection === 0) {
      autoScrollFrame = null;
      return;
    }
    window.scrollBy({top: autoScrollDirection * AUTO_SCROLL_STEP, behavior: "auto"});
    autoScrollFrame = window.requestAnimationFrame(runAutoScroll);
  };

  const startAutoScroll = (direction: -1 | 1) => {
    if (autoScrollDirection === direction && autoScrollFrame !== null) {
      return;
    }
    autoScrollDirection = direction;
    if (autoScrollFrame === null) {
      autoScrollFrame = window.requestAnimationFrame(runAutoScroll);
    }
  };

  const stopAutoScroll = () => {
    autoScrollDirection = 0;
    if (autoScrollFrame !== null) {
      window.cancelAnimationFrame(autoScrollFrame);
      autoScrollFrame = null;
    }
  };

  const updateAutoScroll = (event: DragEvent) => {
    if (typeof window === "undefined") {
      return;
    }
    if (event.clientY <= AUTO_SCROLL_EDGE_PX) {
      startAutoScroll(-1);
      return;
    }
    if (event.clientY >= window.innerHeight - AUTO_SCROLL_EDGE_PX) {
      startAutoScroll(1);
      return;
    }
    stopAutoScroll();
  };

  const reorderBigProcesses = async (fromId: number, toId: number) => {
    if (fromId === toId) {
      return;
    }
    const fromIndex = bigProcesses.findIndex((item) => item.Id === fromId);
    const toIndex = bigProcesses.findIndex((item) => item.Id === toId);
    if (fromIndex < 0 || toIndex < 0) {
      return;
    }
    const direction: "up" | "down" = toIndex < fromIndex ? "up" : "down";
    const steps = Math.abs(toIndex - fromIndex);
    selectedBigProcessId = fromId;
    selectedProcessId = null;
    selectedStepId = null;
    await persistFlowChange(
      async () => {
        for (let idx = 0; idx < steps; idx++) {
          await MoveBigProcess(fromId, direction);
        }
      },
      `Macroflujo movido hacia ${direction === "up" ? "arriba" : "abajo"}.`,
      "drag-big-process"
    );
  };

  const reorderProcesses = async (fromId: number, toId: number) => {
    if (!currentBigProcess || fromId === toId) {
      return;
    }
    const fromIndex = currentProcesses.findIndex((item) => item.Id === fromId);
    const toIndex = currentProcesses.findIndex((item) => item.Id === toId);
    if (fromIndex < 0 || toIndex < 0) {
      return;
    }
    const direction: "up" | "down" = toIndex < fromIndex ? "up" : "down";
    const steps = Math.abs(toIndex - fromIndex);
    selectedProcessId = fromId;
    selectedStepId = null;
    await persistFlowChange(
      async () => {
        for (let idx = 0; idx < steps; idx++) {
          await MoveProcess(currentBigProcess.Id, fromId, direction);
        }
      },
      `Proceso movido hacia ${direction === "up" ? "arriba" : "abajo"}.`,
      "drag-process"
    );
  };

  const reorderSteps = async (fromId: number, toId: number) => {
    if (!currentBigProcess || !currentProcess || fromId === toId) {
      return;
    }
    const fromIndex = currentSteps.findIndex((item) => item.Id === fromId);
    const toIndex = currentSteps.findIndex((item) => item.Id === toId);
    if (fromIndex < 0 || toIndex < 0) {
      return;
    }
    const direction: "up" | "down" = toIndex < fromIndex ? "up" : "down";
    const steps = Math.abs(toIndex - fromIndex);
    selectedStepId = fromId;
    await persistFlowChange(
      async () => {
        for (let idx = 0; idx < steps; idx++) {
          await MoveStep(currentBigProcess.Id, currentProcess.Id, fromId, direction);
        }
      },
      `Paso movido hacia ${direction === "up" ? "arriba" : "abajo"}.`,
      "drag-step"
    );
  };

  const startBigProcessDrag = (id: number, event: DragEvent) => {
    if (busySection !== null) {
      event.preventDefault();
      return;
    }
    draggingBigProcessId = id;
    hoverBigProcessId = id;
    event.dataTransfer?.setData("text/plain", `${id}`);
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = "move";
    }
  };

  const handleBigProcessDragOver = (id: number, event: DragEvent) => {
    event.preventDefault();
    hoverBigProcessId = id;
    updateAutoScroll(event);
  };

  const handleBigProcessDrop = async (id: number, event: DragEvent) => {
    event.preventDefault();
    if (draggingBigProcessId === null) {
      return;
    }
    const draggingId = draggingBigProcessId;
    draggingBigProcessId = null;
    hoverBigProcessId = null;
    stopAutoScroll();
    await reorderBigProcesses(draggingId, id);
  };

  const clearBigProcessDrag = () => {
    draggingBigProcessId = null;
    hoverBigProcessId = null;
    stopAutoScroll();
  };

  const startProcessDrag = (id: number, event: DragEvent) => {
    if (busySection !== null) {
      event.preventDefault();
      return;
    }
    draggingProcessId = id;
    hoverProcessId = id;
    event.dataTransfer?.setData("text/plain", `${id}`);
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = "move";
    }
  };

  const handleProcessDragOver = (id: number, event: DragEvent) => {
    event.preventDefault();
    hoverProcessId = id;
    updateAutoScroll(event);
  };

  const handleProcessDrop = async (id: number, event: DragEvent) => {
    event.preventDefault();
    if (draggingProcessId === null) {
      return;
    }
    const draggingId = draggingProcessId;
    draggingProcessId = null;
    hoverProcessId = null;
    stopAutoScroll();
    await reorderProcesses(draggingId, id);
  };

  const clearProcessDrag = () => {
    draggingProcessId = null;
    hoverProcessId = null;
    stopAutoScroll();
  };

  const startStepDrag = (id: number, event: DragEvent) => {
    if (busySection !== null) {
      event.preventDefault();
      return;
    }
    draggingStepId = id;
    hoverStepId = id;
    event.dataTransfer?.setData("text/plain", `${id}`);
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = "move";
    }
  };

  const handleStepDragOver = (id: number, event: DragEvent) => {
    event.preventDefault();
    hoverStepId = id;
    updateAutoScroll(event);
  };

  const handleStepDrop = async (id: number, event: DragEvent) => {
    event.preventDefault();
    if (draggingStepId === null) {
      return;
    }
    const draggingId = draggingStepId;
    draggingStepId = null;
    hoverStepId = null;
    stopAutoScroll();
    await reorderSteps(draggingId, id);
  };

  const clearStepDrag = () => {
    draggingStepId = null;
    hoverStepId = null;
    stopAutoScroll();
  };

  const selectBigProcess = async (bigProcessId: number) => {
    await runStageTransition(async () => {
      selectedBigProcessId = bigProcessId;
      selectedProcessId = null;
      selectedStepId = null;
      await tick();
    });
  };

  const selectProcess = async (processId: number) => {
    await runStageTransition(async () => {
      selectedProcessId = processId;
      selectedStepId = null;
      await tick();
    });
  };

  const prepareBigProcessCreate = () => {
    newBigProcessName = "";
    newBigProcessDescription = "";
  };

  const prepareProcessCreate = (bigProcessId = currentBigProcess?.Id ?? null) => {
    if (bigProcessId === null) {
      throw new Error("Primero crea o selecciona un macroflujo.");
    }
    selectedBigProcessId = bigProcessId;
    selectedProcessId = null;
    selectedStepId = null;
    newProcessName = "";
    newProcessDescription = "";
  };

  const prepareStepCreate = (processId = currentProcess?.Id ?? null) => {
    if (!currentBigProcess || processId === null) {
      throw new Error("Primero activa un proceso para poder colgarle pasos.");
    }
    selectedBigProcessId = currentBigProcess.Id;
    selectedProcessId = processId;
    selectedStepId = null;
    newStepName = "";
    newStepDescription = "";
  };

  const prepareResourceCreate = (processId = currentProcess?.Id ?? null, stepId = currentStep?.Id ?? null) => {
    if (!currentBigProcess || processId === null || stepId === null) {
      throw new Error("Activa un paso antes de vincular tablas.");
    }
    if (!entities.length) {
      throw new Error("Necesitas entidades para vincular recursos.");
    }
    selectedBigProcessId = currentBigProcess.Id;
    selectedProcessId = processId;
    selectedStepId = stepId;
    newResourceTableId = entities[0]?.Id ?? null;
    newResourceRole = "Input";
  };

  const handleAddBigProcess = async () => {
    const name = newBigProcessName.trim();
    if (!name) {
      throw new Error("Ingresa un nombre para el macroflujo.");
    }

    const nextBigProcessId = (project?.BigProcessLastMax ?? 0) + 1;
    selectedBigProcessId = nextBigProcessId;
    selectedProcessId = null;
    selectedStepId = null;
    await persistFlowChange(
      () => AddBigProcess(name, newBigProcessDescription.trim()),
      "Macroflujo creado.",
      "add-big-process",
      { throwOnError: true }
    );
    newBigProcessName = "";
    newBigProcessDescription = "";
  };

  const prepareBigProcessEdit = (bigProcess: utils.BigProcess) => {
    selectedBigProcessId = bigProcess.Id;
    selectedProcessId = null;
    selectedStepId = null;
    bigProcessDraftName = bigProcess.Name ?? "";
    bigProcessDraftDescription = bigProcess.Description ?? "";
  };

  const handleSaveBigProcess = async (bigProcessId = currentBigProcess?.Id ?? null) => {
    if (bigProcessId === null) {
      throw new Error("Selecciona un macroflujo para editarlo.");
    }
    const name = bigProcessDraftName.trim();
    if (!name) {
      throw new Error("El macroflujo necesita un nombre.");
    }
    selectedBigProcessId = bigProcessId;
    selectedProcessId = null;
    selectedStepId = null;
    await persistFlowChange(
      () => EditBigProcess(bigProcessId, name, bigProcessDraftDescription.trim()),
      "Macroflujo actualizado.",
      "edit-big-process",
      { throwOnError: true }
    );
  };

  const handleRemoveBigProcess = async (bigProcessId = currentBigProcess?.Id ?? null) => {
    if (bigProcessId === null) {
      throw new Error("Selecciona un macroflujo para eliminarlo.");
    }

    const currentIndex = bigProcesses.findIndex((item) => item.Id === bigProcessId);
    if (currentIndex === -1) {
      throw new Error("No se encontró el macroflujo.");
    }
    const fallbackId = bigProcesses[currentIndex + 1]?.Id ?? bigProcesses[currentIndex - 1]?.Id ?? null;
    selectedBigProcessId = fallbackId;
    selectedProcessId = null;
    selectedStepId = null;

    await persistFlowChange(
      () => RemoveBigProcess(bigProcessId),
      "Macroflujo eliminado.",
      "remove-big-process",
      { throwOnError: true }
    );
  };

  const handleAddProcess = async () => {
    if (!currentBigProcess) {
      throw new Error("Primero crea o selecciona un macroflujo.");
    }
    const name = newProcessName.trim();
    if (!name) {
      throw new Error("Ingresa un nombre para el proceso.");
    }

    selectedBigProcessId = currentBigProcess.Id;
    selectedProcessId = (project?.ProcessLastMax ?? 0) + 1;
    selectedStepId = null;
    await persistFlowChange(
      () => AddProcess(currentBigProcess.Id, name, newProcessDescription.trim()),
      "Proceso agregado al macroflujo.",
      "add-process",
      { throwOnError: true }
    );
    newProcessName = "";
    newProcessDescription = "";
  };

  const prepareProcessEdit = (process: utils.Process) => {
    if (!currentBigProcess) {
      throw new Error("Primero selecciona un macroflujo.");
    }
    selectedBigProcessId = currentBigProcess.Id;
    selectedProcessId = process.Id;
    selectedStepId = null;
    processDraftName = process.Name ?? "";
    processDraftDescription = process.Description ?? "";
  };

  const handleSaveProcess = async (processId = currentProcess?.Id ?? null) => {
    if (!currentBigProcess || processId === null) {
      throw new Error("Selecciona un proceso para editarlo.");
    }
    const name = processDraftName.trim();
    if (!name) {
      throw new Error("El proceso necesita un nombre.");
    }

    selectedProcessId = processId;
    selectedStepId = null;
    await persistFlowChange(
      () => EditProcess(currentBigProcess.Id, processId, name, processDraftDescription.trim()),
      "Proceso actualizado.",
      "edit-process",
      { throwOnError: true }
    );
  };

  const handleRemoveProcess = async (processId = currentProcess?.Id ?? null) => {
    if (!currentBigProcess || processId === null) {
      throw new Error("Selecciona un proceso para eliminarlo.");
    }

    const currentIndex = currentProcesses.findIndex((item) => item.Id === processId);
    if (currentIndex === -1) {
      throw new Error("No se encontró el proceso.");
    }
    const fallbackId = currentProcesses[currentIndex + 1]?.Id ?? currentProcesses[currentIndex - 1]?.Id ?? null;
    selectedProcessId = fallbackId;
    selectedStepId = null;

    await persistFlowChange(
      () => RemoveProcess(currentBigProcess.Id, processId),
      "Proceso eliminado.",
      "remove-process",
      { throwOnError: true }
    );
  };

  const handleAddStep = async () => {
    if (!currentBigProcess || !currentProcess) {
      throw new Error("Selecciona un proceso antes de crear pasos.");
    }
    const name = newStepName.trim();
    if (!name) {
      throw new Error("Ingresa un nombre para el paso.");
    }

    selectedStepId = (project?.StepsLastMax ?? 0) + 1;
    await persistFlowChange(
      () => AddStep(currentBigProcess.Id, currentProcess.Id, name, newStepDescription.trim()),
      "Paso agregado al proceso.",
      "add-step",
      { throwOnError: true }
    );
    newStepName = "";
    newStepDescription = "";
  };

  const prepareStepEdit = (processId: number, step: utils.Step) => {
    if (!currentBigProcess) {
      throw new Error("Primero selecciona un macroflujo.");
    }
    selectedBigProcessId = currentBigProcess.Id;
    selectedProcessId = processId;
    selectedStepId = step.Id;
    stepDraftName = step.Name ?? "";
    stepDraftDescription = step.Description ?? "";
  };

  const prepareStepDetail = (processId: number, stepId: number) => {
    if (!currentBigProcess) {
      throw new Error("Primero selecciona un macroflujo.");
    }
    selectedBigProcessId = currentBigProcess.Id;
    selectedProcessId = processId;
    selectedStepId = stepId;
  };

  const handleSaveStep = async (
    processId = currentProcess?.Id ?? null,
    stepId = currentStep?.Id ?? null
  ) => {
    if (!currentBigProcess || processId === null || stepId === null) {
      throw new Error("Selecciona un paso para editarlo.");
    }
    const name = stepDraftName.trim();
    if (!name) {
      throw new Error("El paso necesita un nombre.");
    }

    selectedProcessId = processId;
    selectedStepId = stepId;
    await persistFlowChange(
      () => EditStep(currentBigProcess.Id, processId, stepId, name, stepDraftDescription.trim()),
      "Paso actualizado.",
      "edit-step",
      { throwOnError: true }
    );
  };

  const handleRemoveStep = async (
    processId = currentProcess?.Id ?? null,
    stepId = currentStep?.Id ?? null
  ) => {
    if (!currentBigProcess || processId === null || stepId === null) {
      throw new Error("Selecciona un paso para eliminarlo.");
    }

    const process = currentProcesses.find((item) => item.Id === processId);
    const steps = process?.Steps ?? [];
    const currentIndex = steps.findIndex((item) => item.Id === stepId);
    if (currentIndex === -1) {
      throw new Error("No se encontró el paso.");
    }
    const fallbackId = steps[currentIndex + 1]?.Id ?? steps[currentIndex - 1]?.Id ?? null;
    selectedProcessId = processId;
    selectedStepId = fallbackId;

    await persistFlowChange(
      () => RemoveStep(currentBigProcess.Id, processId, stepId),
      "Paso eliminado.",
      "remove-step",
      { throwOnError: true }
    );
  };

  const handleAddResource = async () => {
    if (!currentBigProcess || !currentProcess || !currentStep) {
      throw new Error("Selecciona un paso antes de asociar tablas.");
    }
    if (!entities.length || newResourceTableId === null) {
      throw new Error("Necesitas entidades para vincular recursos.");
    }

    await persistFlowChange(
      () => AddResource(currentBigProcess.Id, currentProcess.Id, currentStep.Id, newResourceTableId, newResourceRole),
      "Tabla vinculada al paso.",
      "add-resource",
      { throwOnError: true }
    );
  };

  const prepareResourceEdit = (resource: utils.Resource) => {
    updateResourceDraft(resource.Id, {
      tableId: resource.TableId,
      role: resource.Role
    });
  };

  const handleSaveResource = async (resourceId: number) => {
    if (!currentBigProcess || !currentProcess || !currentStep) {
      throw new Error("Selecciona un paso antes de editar recursos.");
    }
    const draft = resourceEdits[resourceId];
    if (!draft) {
      throw new Error("No se encontró el recurso a editar.");
    }

    await persistFlowChange(
      () => EditResource(currentBigProcess.Id, currentProcess.Id, currentStep.Id, resourceId, draft.tableId, draft.role),
      "Recurso actualizado.",
      "edit-resource",
      { throwOnError: true }
    );
  };

  const handleRemoveResource = async (resourceId: number) => {
    if (!currentBigProcess || !currentProcess || !currentStep) {
      throw new Error("Selecciona un paso antes de eliminar recursos.");
    }
    await persistFlowChange(
      () => RemoveResource(currentBigProcess.Id, currentProcess.Id, currentStep.Id, resourceId),
      "Recurso eliminado del paso.",
      "remove-resource",
      { throwOnError: true }
    );
  };

  let bigProcesses: utils.BigProcess[] = [];
  let currentBigProcess: utils.BigProcess | null = null;
  let currentProcesses: utils.Process[] = [];
  let currentProcess: utils.Process | null = null;
  let currentSteps: utils.Step[] = [];
  let currentStep: utils.Step | null = null;

  $: bigProcesses = project?.BigProcesses ?? [];

  $: if (selectedBigProcessId === null && bigProcesses.length > 0) {
    selectedBigProcessId = bigProcesses[0].Id;
  }

  $: if (selectedBigProcessId !== null && !bigProcesses.some((item) => item.Id === selectedBigProcessId)) {
    selectedBigProcessId = bigProcesses[0]?.Id ?? null;
  }

  $: currentBigProcess = bigProcesses.find((item) => item.Id === selectedBigProcessId) ?? null;

  $: currentProcesses = currentBigProcess?.Processes ?? [];

  $: if (selectedProcessId === null && currentProcesses.length > 0) {
    selectedProcessId = currentProcesses[0].Id;
  }

  $: if (selectedProcessId !== null && !currentProcesses.some((item) => item.Id === selectedProcessId)) {
    selectedProcessId = currentProcesses[0]?.Id ?? null;
  }

  $: currentProcess = currentProcesses.find((item) => item.Id === selectedProcessId) ?? null;

  $: currentSteps = currentProcess?.Steps ?? [];

  $: if (selectedStepId === null && currentSteps.length > 0) {
    selectedStepId = currentSteps[0].Id;
  }

  $: if (selectedStepId !== null && !currentSteps.some((item) => item.Id === selectedStepId)) {
    selectedStepId = currentSteps[0]?.Id ?? null;
  }

  $: currentStep = currentSteps.find((item) => item.Id === selectedStepId) ?? null;

  $: if ((currentBigProcess?.Id ?? null) !== lastBigProcessSyncId) {
    bigProcessDraftName = currentBigProcess?.Name ?? "";
    bigProcessDraftDescription = currentBigProcess?.Description ?? "";
    lastBigProcessSyncId = currentBigProcess?.Id ?? null;
  }

  $: if ((currentProcess?.Id ?? null) !== lastProcessSyncId) {
    processDraftName = currentProcess?.Name ?? "";
    processDraftDescription = currentProcess?.Description ?? "";
    lastProcessSyncId = currentProcess?.Id ?? null;
  }

  $: if ((currentStep?.Id ?? null) !== lastStepSyncId) {
    stepDraftName = currentStep?.Name ?? "";
    stepDraftDescription = currentStep?.Description ?? "";
    lastStepSyncId = currentStep?.Id ?? null;
  }

  $: if (newResourceTableId === null && entities.length > 0) {
    newResourceTableId = entities[0].Id;
  }

  $: {
    const signature = currentStep
      ? `${currentStep.Id}:${(currentStep.Resources ?? []).map((resource) => `${resource.Id}-${resource.TableId}-${resource.Role}`).join("|")}`
      : "";
    if (signature !== lastResourceSignature) {
      if (currentStep) {
        const nextEdits: Record<number, ResourceDraft> = {};
        for (const resource of currentStep.Resources ?? []) {
          nextEdits[resource.Id] = {
            tableId: resource.TableId,
            role: resource.Role
          };
        }
        resourceEdits = nextEdits;
      } else {
        resourceEdits = {};
      }
      lastResourceSignature = signature;
    }
  }
</script>

<svelte:window
  on:dragover={updateAutoScroll}
  on:drop={stopAutoScroll}
  on:dragend={stopAutoScroll}
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
              on:click={() => selectBigProcess(bigProcess.Id)}
              on:dragstart={(event) => startBigProcessDrag(bigProcess.Id, event)}
              on:dragover={(event) => handleBigProcessDragOver(bigProcess.Id, event)}
              on:drop={(event) => handleBigProcessDrop(bigProcess.Id, event)}
              on:dragend={clearBigProcessDrag}
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
            </ModalLauncher>
          </div>

          {#if currentProcesses.length}
            <div class="process-rail-list">
              {#each currentProcesses as process, index (process.Id)}
                <button
                  type="button"
                  class:process-rail-card={true}
                  class:process-rail-card--active={process.Id === currentProcess?.Id}
                  on:click={() => selectProcess(process.Id)}
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
                <p class="modal-hint local-modal-hint">Se eliminará <strong>{currentBigProcess.Name}</strong> con todos sus procesos y pasos.</p>
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
            on:dragstart={(event) => startProcessDrag(currentProcess.Id, event)}
            on:dragover={(event) => handleProcessDragOver(currentProcess.Id, event)}
            on:drop={(event) => handleProcessDrop(currentProcess.Id, event)}
            on:dragend={clearProcessDrag}
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
                    <p class="modal-hint local-modal-hint">Se eliminará <strong>{currentProcess.Name}</strong> con todos sus pasos.</p>
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
                    on:dragstart={(event) => startStepDrag(step.Id, event)}
                    on:dragover={(event) => handleStepDragOver(step.Id, event)}
                    on:drop={(event) => handleStepDrop(step.Id, event)}
                    on:dragend={clearStepDrag}
                    in:scale={{duration: 280, delay: stepIndex * 55, start: 0.92, easing: quintOut}}
                    animate:flip={{duration: 320, easing: quintOut}}
                    style={`view-transition-name: step-node-${step.Id};`}
                  >
                    <span class="step-node__order">{String(step.Order).padStart(2, "0")}</span>
                    <span class="step-node__body">
                      <strong>{step.Name}</strong>
                      <span>{step.Description || "Describe qué pasa en este punto."}</span>
                    </span>
                    <div class="step-node__meta" on:click|stopPropagation>
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
                            <p class="modal-hint local-modal-hint">Se eliminará <strong>{step.Name}</strong> de este proceso.</p>
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
                                        <div class="modal-intro">
                                          <p class="modal-kicker local-modal-kicker">Editor recurso</p>
                                          <p class="modal-hint local-modal-hint">Ajusta la tabla relacionada y el rol que cumple en este paso.</p>
                                        </div>
                                        <div class="modal-form-grid modal-form-grid--compact">
                                          <label class="field">
                                            <span>Tabla</span>
                                            <select
                                              value={resourceEdits[resource.Id]?.tableId ?? resource.TableId}
                                              on:change={(event) => handleResourceTableChange(resource.Id, event)}
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
                                              on:change={(event) => handleResourceRoleChange(resource.Id, event)}
                                            >
                                              {#each resourceRoles as role}
                                                <option value={role}>{role}</option>
                                              {/each}
                                            </select>
                                          </label>
                                        </div>
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
                                        <p class="modal-hint local-modal-hint">Se quitará <strong>{entityLabel(resource.TableId)}</strong> de este paso.</p>
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

  :global(.flow-modal-trigger--hero),
  :global(.flow-modal-trigger--tail) {
    width: auto;
  }

  :global(.flow-modal-trigger--subrail) {
    width: 100%;
  }

  .modal-intro {
    display: grid;
    gap: 0.3rem;
    padding-bottom: 0.35rem;
  }

  .local-modal-kicker,
  .local-modal-hint {
    margin: 0;
  }

  .modal-form-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.85rem;
  }

  .modal-form-grid--compact {
    grid-template-columns: minmax(0, 1fr) minmax(9rem, 0.72fr);
  }

  .macro-card {
    position: relative;
    display: grid;
    gap: 0.32rem;
    padding: 0.95rem 1rem 1rem 1.08rem;
    border: 1px solid var(--border);
    border-radius: calc(var(--radius-md) - 10px);
    background:
      linear-gradient(180deg, color-mix(in srgb, var(--surface-strong) 96%, transparent), color-mix(in srgb, var(--surface) 82%, transparent));
    text-align: left;
    box-shadow: 0 14px 28px color-mix(in srgb, var(--ink) 8%, transparent);
    transition:
      transform 220ms cubic-bezier(0.22, 1, 0.36, 1),
      border-color 220ms ease,
      box-shadow 220ms ease;
  }

  .macro-card::after {
    content: "";
    position: absolute;
    inset: 0 auto 0 0;
    width: 4px;
    border-radius: 999px;
    background: linear-gradient(180deg, color-mix(in srgb, var(--accent) 92%, white 8%), color-mix(in srgb, var(--accent-strong) 80%, transparent));
    opacity: 0.24;
    transition: opacity 200ms ease, transform 220ms cubic-bezier(0.22, 1, 0.36, 1);
  }

  .macro-card:hover,
  .macro-card--active {
    transform: translateX(6px) translateY(-2px);
    border-color: color-mix(in srgb, var(--accent) 28%, var(--border));
    box-shadow:
      0 24px 42px color-mix(in srgb, var(--ink) 12%, transparent),
      0 0 0 0.18rem color-mix(in srgb, var(--accent) 10%, transparent);
  }

  .macro-card:hover::after,
  .macro-card--active::after {
    opacity: 1;
    transform: scaleY(1);
  }

  .macro-card--dragging {
    opacity: 0.48;
    transform: scale(0.985);
  }

  .macro-card--drop-target {
    border-color: color-mix(in srgb, var(--accent) 34%, var(--border));
    box-shadow:
      0 24px 42px color-mix(in srgb, var(--ink) 12%, transparent),
      0 0 0 0.22rem color-mix(in srgb, var(--accent) 14%, transparent);
  }

  .macro-card__index {
    color: var(--ink-faint);
    font-size: 0.72rem;
    font-weight: 800;
    letter-spacing: 0.16em;
  }

  .macro-card strong {
    font-size: 1.02rem;
    line-height: 1.2;
  }

  .macro-card__meta,
  .macro-card__hint {
    color: var(--ink-soft);
    font-size: 0.82rem;
    line-height: 1.4;
  }

  .stage-panel {
    display: grid;
    gap: 1rem;
    padding: 1.05rem;
    background:
      radial-gradient(circle at top right, color-mix(in srgb, var(--accent) 16%, transparent), transparent 32%),
      linear-gradient(180deg, color-mix(in srgb, var(--surface-strong) 96%, transparent), color-mix(in srgb, var(--surface) 88%, transparent));
  }

  .stage-hero {
    position: relative;
    display: grid;
    grid-template-columns: minmax(0, 1.1fr) auto;
    gap: 1rem;
    padding: 1.1rem 1.15rem;
    border-radius: calc(var(--radius-lg) - 10px);
    border: 1px solid color-mix(in srgb, var(--accent) 18%, var(--border));
    background:
      linear-gradient(135deg, color-mix(in srgb, var(--surface-strong) 96%, transparent), color-mix(in srgb, var(--surface) 80%, transparent)),
      linear-gradient(90deg, color-mix(in srgb, var(--accent) 10%, transparent), transparent 42%);
    overflow: hidden;
  }

  .stage-hero::after {
    content: "";
    position: absolute;
    inset: auto -10% -18% auto;
    width: 16rem;
    height: 16rem;
    border-radius: 50%;
    background: radial-gradient(circle, color-mix(in srgb, var(--accent) 16%, transparent), transparent 72%);
    animation: hero-orbit 10s linear infinite;
    pointer-events: none;
  }

  .stage-hero__copy p {
    margin: 0.45rem 0 0;
    max-width: 60ch;
    color: var(--ink-soft);
    line-height: 1.55;
  }

  .stage-hero__side {
    display: grid;
    gap: 0.8rem;
    align-content: start;
  }

  .hero-stats {
    display: grid;
    grid-template-columns: repeat(3, minmax(6rem, 1fr));
    gap: 0.7rem;
  }

  .hero-actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 0.6rem;
  }

  .hero-stat {
    display: grid;
    gap: 0.28rem;
    padding: 0.8rem 0.9rem;
    min-width: 7rem;
    border-radius: 1.1rem;
    border: 1px solid color-mix(in srgb, var(--accent) 12%, var(--border));
    background: color-mix(in srgb, var(--surface-strong) 84%, transparent);
    box-shadow: inset 0 1px 0 color-mix(in srgb, white 20%, transparent);
  }

  .hero-stat span {
    color: var(--ink-faint);
    font-size: 0.74rem;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .hero-stat strong {
    font-size: 1.35rem;
    line-height: 1;
  }

  .process-track {
    display: grid;
    gap: 0.95rem;
  }

  .process-lane {
    display: grid;
    gap: 0.9rem;
    padding: 1rem;
    border-radius: calc(var(--radius-lg) - 10px);
    border: 1px solid var(--border);
    background:
      linear-gradient(180deg, color-mix(in srgb, var(--surface-strong) 94%, transparent), color-mix(in srgb, var(--surface) 78%, transparent));
    box-shadow: 0 18px 34px color-mix(in srgb, var(--ink) 8%, transparent);
    cursor: pointer;
    transition:
      transform 220ms cubic-bezier(0.22, 1, 0.36, 1),
      border-color 220ms ease,
      box-shadow 220ms ease;
    outline: none;
  }

  .process-lane:hover,
  .process-lane:focus-visible,
  .process-lane--active {
    transform: translateY(-2px);
    border-color: color-mix(in srgb, var(--accent) 26%, var(--border));
    box-shadow:
      0 26px 46px color-mix(in srgb, var(--ink) 11%, transparent),
      0 0 0 0.18rem color-mix(in srgb, var(--accent) 10%, transparent);
  }

  .process-lane--dragging {
    opacity: 0.46;
    transform: scale(0.988);
  }

  .process-lane--drop-target {
    border-color: color-mix(in srgb, var(--accent) 34%, var(--border));
    box-shadow:
      0 26px 46px color-mix(in srgb, var(--ink) 11%, transparent),
      0 0 0 0.24rem color-mix(in srgb, var(--accent) 13%, transparent);
  }

  .process-lane__head {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 1rem;
  }

  .process-lane__side {
    display: grid;
    gap: 0.72rem;
    align-content: start;
    justify-items: end;
  }

  .process-lane__head h3 {
    margin: 0.16rem 0 0;
    font-size: 1.35rem;
    line-height: 1;
  }

  .process-lane__head p {
    margin: 0.45rem 0 0;
    color: var(--ink-soft);
    line-height: 1.5;
  }

  .process-lane__meta {
    display: grid;
    gap: 0.42rem;
    align-content: start;
    justify-items: end;
    color: var(--ink-faint);
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  .process-lane__actions,
  .process-lane__footer {
    display: flex;
    flex-wrap: wrap;
    gap: 0.55rem;
  }

  .process-lane__footer {
    justify-content: flex-start;
    padding-top: 0.2rem;
  }

  .step-sequence {
    position: relative;
    display: grid;
    gap: 0.8rem;
    padding-left: 1rem;
  }

  .step-sequence::before {
    content: "";
    position: absolute;
    left: 0.28rem;
    top: 0.2rem;
    bottom: 0.2rem;
    width: 2px;
    background: linear-gradient(180deg, color-mix(in srgb, var(--accent) 65%, transparent), color-mix(in srgb, var(--accent) 12%, transparent));
  }

  .step-node {
    position: relative;
    display: grid;
    grid-template-columns: auto minmax(0, 1fr);
    gap: 0.78rem 0.9rem;
    align-items: start;
    padding: 0.82rem 0.95rem;
    border-radius: 1.05rem;
    border: 1px solid var(--border);
    background:
      linear-gradient(180deg, color-mix(in srgb, var(--surface-strong) 96%, transparent), color-mix(in srgb, var(--surface) 82%, transparent));
    text-align: left;
    transition:
      transform 180ms cubic-bezier(0.22, 1, 0.36, 1),
      border-color 180ms ease,
      box-shadow 180ms ease;
  }

  .step-node::before {
    content: "";
    position: absolute;
    top: 50%;
    left: -0.95rem;
    width: 0.95rem;
    height: 2px;
    background: color-mix(in srgb, var(--accent) 40%, transparent);
    transform: translateY(-50%);
  }

  .step-node:hover,
  .step-node:focus-within {
    transform: translateX(6px);
    border-color: color-mix(in srgb, var(--accent) 24%, var(--border));
    box-shadow:
      0 18px 26px color-mix(in srgb, var(--ink) 10%, transparent),
      0 0 0 0.16rem color-mix(in srgb, var(--accent) 9%, transparent);
  }

  .step-node--dragging {
    opacity: 0.42;
    transform: scale(0.985);
  }

  .step-node--drop-target {
    border-color: color-mix(in srgb, var(--accent) 36%, var(--border));
    box-shadow:
      0 18px 26px color-mix(in srgb, var(--ink) 10%, transparent),
      0 0 0 0.2rem color-mix(in srgb, var(--accent) 12%, transparent);
  }

  .step-node__order {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 2.5rem;
    min-width: 2.5rem;
    min-height: 2.5rem;
    border-radius: 999px;
    border: 1px solid color-mix(in srgb, var(--accent) 18%, var(--border));
    background: color-mix(in srgb, var(--accent) 10%, var(--surface-strong));
    color: var(--accent-strong);
    font-size: 0.8rem;
    font-weight: 900;
    letter-spacing: 0.08em;
  }

  .step-node__body {
    display: grid;
    gap: 0.25rem;
    align-self: center;
  }

  .step-node__body strong {
    font-size: 0.96rem;
    line-height: 1.2;
  }

  .step-node__body span {
    color: var(--ink-soft);
    font-size: 0.82rem;
    line-height: 1.45;
  }

  .step-node__meta {
    grid-column: 2;
    display: grid;
    gap: 0.72rem;
    margin-top: 0.1rem;
    padding: 0.78rem 0.82rem 0.85rem;
    border-radius: 0.95rem;
    border: 1px solid color-mix(in srgb, var(--accent) 14%, var(--border));
    background:
      linear-gradient(180deg, color-mix(in srgb, var(--surface-strong) 97%, transparent), color-mix(in srgb, var(--surface) 86%, transparent)),
      linear-gradient(135deg, color-mix(in srgb, var(--accent) 8%, transparent), transparent 58%);
    box-shadow: inset 0 1px 0 color-mix(in srgb, white 24%, transparent);
  }

  .step-node__meta-top {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr);
    gap: 0.65rem;
    align-items: start;
  }

  .step-node__resource-count {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    justify-self: start;
    min-width: 2.25rem;
    min-height: 2.25rem;
    padding: 0 0.65rem;
    border-radius: 999px;
    border: 1px solid color-mix(in srgb, var(--accent) 18%, var(--border));
    background: color-mix(in srgb, var(--accent) 10%, var(--surface-strong));
    color: var(--accent-strong);
    font-size: 0.8rem;
    font-weight: 800;
  }

  .step-node__actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.45rem;
    justify-content: flex-start;
  }

  .resource-summary {
    display: grid;
    gap: 0.55rem;
  }

  .resource-summary__group {
    display: grid;
    gap: 0.35rem;
  }

  .resource-summary__label {
    font-size: 0.73rem;
    font-weight: 900;
    letter-spacing: 0.11em;
    text-transform: uppercase;
  }

  .resource-summary__label--input {
    color: color-mix(in srgb, var(--success) 74%, var(--ink));
  }

  .resource-summary__label--output {
    color: color-mix(in srgb, var(--accent) 78%, var(--ink));
  }

  .resource-summary__chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem;
  }

  .resource-chip {
    display: inline-flex;
    align-items: center;
    min-height: 1.9rem;
    padding: 0 0.72rem;
    border-radius: 999px;
    border: 1px solid var(--border);
    background:
      linear-gradient(180deg, color-mix(in srgb, var(--surface-strong) 97%, transparent), color-mix(in srgb, var(--surface) 90%, transparent));
    font-size: 0.76rem;
    font-weight: 700;
    line-height: 1;
  }

  .resource-chip--input {
    border-color: color-mix(in srgb, var(--success) 24%, var(--border));
    background: color-mix(in srgb, var(--success) 12%, var(--surface-strong));
    color: color-mix(in srgb, var(--success) 84%, var(--ink));
  }

  .resource-chip--output {
    border-color: color-mix(in srgb, var(--accent) 26%, var(--border));
    background: color-mix(in srgb, var(--accent) 12%, var(--surface-strong));
    color: var(--accent-strong);
  }

  .resource-chip--ghost {
    color: var(--ink-faint);
    background: color-mix(in srgb, var(--surface) 72%, transparent);
  }

  .field {
    display: grid;
    gap: 0.38rem;
  }

  .field span {
    font-size: 0.79rem;
    font-weight: 800;
    color: var(--ink-soft);
    letter-spacing: 0.05em;
    text-transform: uppercase;
  }

  .field input,
  .field textarea,
  .field select {
    width: 100%;
    border: 1px solid var(--border);
    border-radius: 1rem;
    background: color-mix(in srgb, var(--surface-strong) 92%, transparent);
    color: var(--ink);
    padding: 0.8rem 0.92rem;
    box-sizing: border-box;
    resize: vertical;
    transition:
      border-color 160ms ease,
      box-shadow 160ms ease,
      transform 160ms ease,
      background 160ms ease;
  }

  .field input:focus,
  .field textarea:focus,
  .field select:focus {
    outline: none;
    border-color: color-mix(in srgb, var(--accent) 34%, var(--border));
    box-shadow: 0 0 0 0.22rem color-mix(in srgb, var(--accent) 14%, transparent);
    transform: translateY(-1px);
    background: color-mix(in srgb, var(--surface-strong) 98%, transparent);
  }

  .field--compact input,
  .field--compact select {
    min-height: 2.8rem;
    padding: 0.72rem 0.85rem;
  }

  .action-row {
    display: flex;
    flex-wrap: wrap;
    gap: 0.55rem;
  }

  .resource-list {
    display: grid;
    gap: 0.72rem;
  }

  .resource-row {
    display: flex;
    justify-content: space-between;
    align-items: start;
    gap: 0.8rem;
    padding: 0.9rem;
    border-radius: 1rem;
    border: 1px solid color-mix(in srgb, var(--accent) 12%, var(--border));
    background: color-mix(in srgb, var(--surface-strong) 92%, transparent);
  }

  .resource-row__identity {
    display: grid;
    gap: 0.2rem;
  }

  .resource-row__index {
    color: var(--ink-faint);
    font-size: 0.74rem;
    letter-spacing: 0.12em;
    text-transform: uppercase;
  }

  .resource-row__identity strong {
    font-size: 0.95rem;
  }

  .resource-row__identity p {
    margin: 0;
    color: var(--ink-soft);
    font-size: 0.82rem;
    line-height: 1.42;
  }

  .resource-row__actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 0.45rem;
  }

  .signal-chip {
    display: inline-flex;
    align-items: center;
    min-height: 2rem;
    padding: 0 0.8rem;
    border-radius: 999px;
    border: 1px solid color-mix(in srgb, var(--accent) 16%, var(--border));
    background: color-mix(in srgb, var(--accent) 11%, var(--surface-strong));
    color: var(--accent-strong);
    font-size: 0.8rem;
    font-weight: 800;
  }

  .signal-chip--quiet {
    border-color: var(--border);
    background: color-mix(in srgb, var(--surface-strong) 90%, transparent);
    color: var(--ink-soft);
  }

  .empty-state,
  .empty-stage-fragment {
    border-radius: calc(var(--radius-md) - 8px);
    border: 1px dashed color-mix(in srgb, var(--accent) 16%, var(--border));
    background: color-mix(in srgb, var(--surface-strong) 76%, transparent);
    color: var(--ink-soft);
  }

  .empty-state,
  .empty-stage-fragment {
    display: grid;
    place-items: center;
    gap: 0.35rem;
    min-height: 13rem;
    padding: 1.2rem;
    text-align: center;
  }

  .empty-state strong,
  .empty-stage-fragment strong {
    color: var(--ink);
    font-size: 1rem;
  }

  .empty-state p,
  .empty-stage-fragment p {
    margin: 0;
    line-height: 1.55;
  }

  .empty-state--rail {
    min-height: 11rem;
  }

  .empty-stage-fragment--resource {
    min-height: 8rem;
  }

  @keyframes scan-surface {
    0% {
      transform: translateX(-120%);
    }
    100% {
      transform: translateX(120%);
    }
  }

  @keyframes hero-orbit {
    0% {
      transform: translate3d(0, 0, 0) scale(0.92);
    }
    50% {
      transform: translate3d(-1.2rem, -0.8rem, 0) scale(1.02);
    }
    100% {
      transform: translate3d(0, 0, 0) scale(0.92);
    }
  }

  @media (max-width: 1380px) {
    .studio-shell {
      grid-template-columns: minmax(15rem, 0.8fr) minmax(0, 1.25fr);
    }
  }

  @media (max-width: 1040px) {
    .studio-shell,
    .modal-form-grid {
      grid-template-columns: 1fr;
    }

    .stage-hero {
      grid-template-columns: 1fr;
    }

    .hero-stats {
      grid-template-columns: repeat(3, minmax(0, 1fr));
    }

    .panel-head {
      flex-direction: column;
      align-items: stretch;
    }

    .rail-subpanel__head,
    .resource-row {
      flex-direction: column;
      align-items: stretch;
    }

    :global(.flow-modal-trigger--rail),
    :global(.flow-modal-trigger--hero),
    :global(.flow-modal-trigger--tail),
    :global(.flow-modal-trigger--subrail) {
      width: 100%;
    }
  }

  @media (max-width: 720px) {
    .flow-toolbar__meta,
    .action-row {
      justify-content: stretch;
    }

    .hero-stats {
      grid-template-columns: 1fr;
    }

    .resource-row__actions,
    .process-lane__actions,
    .hero-actions,
    .process-lane__footer,
    .step-node__actions {
      flex-direction: column;
      align-items: stretch;
    }

    .process-lane__head {
      grid-template-columns: 1fr;
    }

    .process-lane__meta,
    .process-lane__side {
      justify-items: start;
    }

    .step-node__meta-top {
      grid-template-columns: 1fr;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .rail-panel::before,
    .stage-panel::before,
    .stage-hero::after {
      animation: none;
    }

    .macro-card,
    .process-lane,
    .step-node,
    .field input,
    .field textarea,
    .field select {
      transition: none;
    }
  }
</style>
