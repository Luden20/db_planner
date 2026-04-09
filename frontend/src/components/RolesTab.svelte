<script lang="ts">
  import { flip } from "svelte/animate";
  import { quintOut } from "svelte/easing";
  import { fade, fly } from "svelte/transition";
  import { onMount, tick } from "svelte";
  import {
    AddRole,
    AddRoleProcessPermission,
    AddRoleTablePermission,
    EditRole,
    EditRoleTablePermission,
    RemoveRole,
    RemoveRoleProcessPermission,
    RemoveRoleTablePermission,
    Save
  } from "../../wailsjs/go/main/App";
  import type { utils } from "../../wailsjs/go/models";
  import ButtonIcon from "./ButtonIcon.svelte";
  import ModalLauncher from "./ModalLauncher.svelte";
  import { showToast } from "../lib/toast";

  type PermissionKey = "ViewPermission" | "InsertPermission" | "UpdatePermission" | "DeletePermission";
  type ProcessGroup = {
    id: number;
    name: string;
    description: string;
    processes: utils.Process[];
  };
  type ViewTransitionDocument = Document & {
    startViewTransition?: (update: () => void | Promise<void>) => {
      finished: Promise<void>;
    };
  };

  export let project: utils.DbProject;
  export let entities: utils.Entity[] = [];
  export let onRefresh: () => Promise<void> = async () => {};

  const permissionColumns: Array<{key: PermissionKey; label: string; short: string; hint: string}> = [
    {key: "InsertPermission", label: "Create", short: "C", hint: "Crear registros"},
    {key: "ViewPermission", label: "Read", short: "R", hint: "Consultar registros"},
    {key: "UpdatePermission", label: "Update", short: "U", hint: "Modificar registros"},
    {key: "DeletePermission", label: "Delete", short: "D", hint: "Eliminar registros"}
  ];

  let selectedRoleId: number | null = null;
  let stickySentinel: HTMLDivElement | null = null;
  let stickyStack: HTMLDivElement | null = null;
  let stickyStackHeight = 0;
  let stickyStackPinned = false;
  let busySection: string | null = null;
  let newRoleName = "";
  let newRoleDescription = "";
  let roleDraftName = "";
  let roleDraftDescription = "";
  let processPermissionDraft: Record<number, boolean> = {};

  const prefersReducedMotion = () =>
    typeof window !== "undefined"
    && typeof window.matchMedia === "function"
    && window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  const runRoleTransition = async (update: () => void | Promise<void>) => {
    const doc = typeof document !== "undefined" ? (document as ViewTransitionDocument) : null;
    if (doc?.startViewTransition && !prefersReducedMotion()) {
      try {
        const transition = doc.startViewTransition(update);
        await transition.finished;
        return;
      } catch (err) {
        console.warn("No se pudo aplicar la transicion de roles:", err);
      }
    }
    await update();
  };

  const extractError = (err: unknown) => {
    if (typeof err === "string") {
      return err;
    }
    if (err && typeof err === "object") {
      const record = err as Record<string, unknown>;
      return String(record.error ?? record.message ?? "Error desconocido");
    }
    return "Error desconocido";
  };

  const syncStickyState = () => {
    if (!stickySentinel) {
      stickyStackPinned = false;
      return;
    }

    stickyStackPinned = stickySentinel.getBoundingClientRect().top <= 0;
  };

  const syncStickyStackHeight = () => {
    stickyStackHeight = stickyStack?.offsetHeight ?? 0;
  };

  const entityDescription = (entityId: number) =>
    entities.find((entity) => entity.Id === entityId)?.Description ?? "Sin detalle para esta tabla.";

  const prefersRoleSelection = async (roleId: number) => {
    await runRoleTransition(async () => {
      selectedRoleId = roleId;
      await tick();
    });
  };

  const persistRoleChange = async (
    action: () => Promise<void>,
    options: { successMessage?: string; busyKey?: string } = {}
  ) => {
    if (busySection !== null) {
      return false;
    }

    busySection = options.busyKey ?? "roles";
    try {
      await action();
      await Save();
      await onRefresh();
      await tick();
      if (options.successMessage) {
        showToast(options.successMessage, "success");
      }
      return true;
    } catch (err) {
      const message = extractError(err);
      showToast(`No se pudieron actualizar los roles: ${message}`, "error");
      return false;
    } finally {
      busySection = null;
    }
  };

  const getTablePermission = (role: utils.Role | null, tableId: number) =>
    (role?.TablePermissions ?? []).find((permission) => permission.TableId === tableId) ?? null;

  const hasAnyTablePermission = (permission: {
    ViewPermission: boolean;
    InsertPermission: boolean;
    UpdatePermission: boolean;
    DeletePermission: boolean;
  }) =>
    permission.ViewPermission
    || permission.InsertPermission
    || permission.UpdatePermission
    || permission.DeletePermission;

  const getTablePermissionValue = (role: utils.Role | null, tableId: number, key: PermissionKey) =>
    getTablePermission(role, tableId)?.[key] ?? false;

  const countGrantedTables = (role: utils.Role | null) =>
    (role?.TablePermissions ?? []).filter((permission) => hasAnyTablePermission(permission)).length;

  const countGrantedChecks = (role: utils.Role | null) =>
    (role?.TablePermissions ?? []).reduce((total, permission) => total
      + Number(permission.ViewPermission)
      + Number(permission.InsertPermission)
      + Number(permission.UpdatePermission)
      + Number(permission.DeletePermission), 0);

  const countGrantedProcesses = (role: utils.Role | null) => role?.ProcessPermissions?.length ?? 0;

  const prepareRoleCreate = () => {
    newRoleName = "";
    newRoleDescription = "";
  };

  const handleAddRole = async () => {
    const trimmedName = newRoleName.trim();
    if (!trimmedName) {
      throw new Error("Ingresa un nombre para el rol.");
    }

    selectedRoleId = (project?.RoleLastMax ?? 0) + 1;
    const ok = await persistRoleChange(
      () => AddRole(trimmedName, newRoleDescription.trim()),
      {successMessage: "Rol creado.", busyKey: "add-role"}
    );
    if (!ok) {
      throw new Error("No se pudo crear el rol.");
    }
  };

  const prepareRoleEdit = (role: utils.Role) => {
    selectedRoleId = role.Id;
    roleDraftName = role.Name ?? "";
    roleDraftDescription = role.Description ?? "";
  };

  const handleSaveRole = async (roleId = currentRole?.Id ?? null) => {
    if (roleId === null) {
      throw new Error("Selecciona un rol para editarlo.");
    }
    const trimmedName = roleDraftName.trim();
    if (!trimmedName) {
      throw new Error("El rol necesita un nombre.");
    }

    selectedRoleId = roleId;
    const ok = await persistRoleChange(
      () => EditRole(roleId, trimmedName, roleDraftDescription.trim()),
      {successMessage: "Rol actualizado.", busyKey: "edit-role"}
    );
    if (!ok) {
      throw new Error("No se pudo guardar el rol.");
    }
  };

  const handleRemoveRole = async (roleId = currentRole?.Id ?? null) => {
    if (roleId === null) {
      throw new Error("Selecciona un rol para eliminarlo.");
    }

    const currentIndex = roles.findIndex((role) => role.Id === roleId);
    const fallbackId = roles[currentIndex + 1]?.Id ?? roles[currentIndex - 1]?.Id ?? null;
    selectedRoleId = fallbackId;

    const ok = await persistRoleChange(
      () => RemoveRole(roleId),
      {successMessage: "Rol eliminado.", busyKey: "remove-role"}
    );
    if (!ok) {
      throw new Error("No se pudo eliminar el rol.");
    }
  };

  const prepareProcessPermissions = (role = currentRole) => {
    if (!role) {
      throw new Error("Selecciona un rol para gestionar procesos.");
    }

    const nextDraft: Record<number, boolean> = {};
    const granted = new Set((role.ProcessPermissions ?? []).map((permission) => permission.ProcessId));
    for (const process of processCatalog) {
      nextDraft[process.id] = granted.has(process.id);
    }
    processPermissionDraft = nextDraft;
  };

  const setProcessGroupDraft = (group: ProcessGroup, checked: boolean) => {
    const nextDraft = {...processPermissionDraft};
    for (const process of group.processes) {
      nextDraft[process.Id] = checked;
    }
    processPermissionDraft = nextDraft;
  };

  const handleProcessDraftToggle = (processId: number, checked: boolean) => {
    processPermissionDraft = {
      ...processPermissionDraft,
      [processId]: checked
    };
  };

  const handleProcessDraftChange = (processId: number, event: Event) => {
    const target = event.currentTarget as HTMLInputElement;
    handleProcessDraftToggle(processId, target.checked);
  };

  const handleSaveProcessPermissions = async () => {
    if (!currentRole) {
      throw new Error("Selecciona un rol para gestionar procesos.");
    }

    const existingByProcess = new Map(
      (currentRole.ProcessPermissions ?? []).map((permission) => [permission.ProcessId, permission])
    );

    const ok = await persistRoleChange(async () => {
      for (const process of processCatalog) {
        const shouldHaveAccess = Boolean(processPermissionDraft[process.id]);
        const existingPermission = existingByProcess.get(process.id) ?? null;

        if (shouldHaveAccess && !existingPermission) {
          await AddRoleProcessPermission(currentRole.Id, process.id);
        }

        if (!shouldHaveAccess && existingPermission) {
          await RemoveRoleProcessPermission(currentRole.Id, existingPermission.Id);
        }
      }
    }, {
      successMessage: "Permisos de procesos actualizados.",
      busyKey: "edit-role-processes"
    });

    if (!ok) {
      throw new Error("No se pudieron guardar los permisos de procesos.");
    }
  };

  const handleTablePermissionToggle = async (tableId: number, key: PermissionKey, checked: boolean) => {
    if (!currentRole) {
      showToast("Selecciona un rol para editar permisos.", "error");
      return;
    }

    const existingPermission = getTablePermission(currentRole, tableId);
    const nextPermission = {
      ViewPermission: existingPermission?.ViewPermission ?? false,
      InsertPermission: existingPermission?.InsertPermission ?? false,
      UpdatePermission: existingPermission?.UpdatePermission ?? false,
      DeletePermission: existingPermission?.DeletePermission ?? false,
      [key]: checked
    };

    if (!existingPermission && !hasAnyTablePermission(nextPermission)) {
      return;
    }

    if (!existingPermission) {
      await persistRoleChange(
        () => AddRoleTablePermission(
          currentRole.Id,
          tableId,
          nextPermission.InsertPermission,
          nextPermission.DeletePermission,
          nextPermission.UpdatePermission,
          nextPermission.ViewPermission
        ),
        {busyKey: `table-${tableId}-${key}`}
      );
      return;
    }

    if (!hasAnyTablePermission(nextPermission)) {
      await persistRoleChange(
        () => RemoveRoleTablePermission(currentRole.Id, existingPermission.Id),
        {busyKey: `table-${tableId}-${key}`}
      );
      return;
    }

    await persistRoleChange(
      () => EditRoleTablePermission(
        currentRole.Id,
        existingPermission.Id,
        tableId,
        nextPermission.InsertPermission,
        nextPermission.DeletePermission,
        nextPermission.UpdatePermission,
        nextPermission.ViewPermission
      ),
      {busyKey: `table-${tableId}-${key}`}
    );
  };

  const handleTablePermissionChange = async (tableId: number, key: PermissionKey, event: Event) => {
    const target = event.currentTarget as HTMLInputElement;
    await handleTablePermissionToggle(tableId, key, target.checked);
  };

  const prevRole = async () => {
    if (!roles.length) {
      return;
    }
    const currentIndex = roles.findIndex((role) => role.Id === currentRole?.Id);
    const nextIndex = currentIndex <= 0 ? roles.length - 1 : currentIndex - 1;
    await prefersRoleSelection(roles[nextIndex].Id);
  };

  const nextRole = async () => {
    if (!roles.length) {
      return;
    }
    const currentIndex = roles.findIndex((role) => role.Id === currentRole?.Id);
    const nextIndex = currentIndex === -1 || currentIndex === roles.length - 1 ? 0 : currentIndex + 1;
    await prefersRoleSelection(roles[nextIndex].Id);
  };

  let roles: utils.Role[] = [];
  let currentRole: utils.Role | null = null;
  let processGroups: ProcessGroup[] = [];
  let processCatalog: Array<{id: number; bigProcessId: number; bigProcessName: string; name: string; description: string}> = [];

  $: roles = project?.Roles ?? [];

  $: if (selectedRoleId === null && roles.length > 0) {
    selectedRoleId = roles[0].Id;
  }

  $: if (selectedRoleId !== null && !roles.some((role) => role.Id === selectedRoleId)) {
    selectedRoleId = roles[0]?.Id ?? null;
  }

  $: currentRole = roles.find((role) => role.Id === selectedRoleId) ?? null;

  $: processGroups = (project?.BigProcesses ?? [])
    .map((bigProcess) => ({
      id: bigProcess.Id,
      name: bigProcess.Name,
      description: bigProcess.Description ?? "",
      processes: bigProcess.Processes ?? []
    }))
    .filter((group) => group.processes.length > 0);

  $: processCatalog = processGroups.flatMap((group) =>
    group.processes.map((process) => ({
      id: process.Id,
      bigProcessId: group.id,
      bigProcessName: group.name,
      name: process.Name,
      description: process.Description ?? ""
    }))
  );

  onMount(() => {
    syncStickyStackHeight();
    syncStickyState();

    if (typeof ResizeObserver === "undefined" || !stickyStack) {
      return;
    }

    const observer = new ResizeObserver(() => {
      syncStickyStackHeight();
    });
    observer.observe(stickyStack);

    return () => {
      observer.disconnect();
    };
  });

  $: if (stickySentinel) {
    syncStickyState();
  }
  $: if (stickyStack) {
    syncStickyStackHeight();
  }
</script>

<svelte:window on:scroll={syncStickyState} on:resize={syncStickyState}/>

<section class="roles-tab" style={`--roles-sticky-total-height: ${stickyStackHeight}px;`}>
  <div class="roles-sticky-sentinel" bind:this={stickySentinel} aria-hidden="true"></div>
  <div
    class:roles-sticky-stack={true}
    class:roles-sticky-stack--pinned={stickyStackPinned}
    bind:this={stickyStack}
  >

    {#if currentRole}
      <header class="role-hero" style={`view-transition-name: role-hero-${currentRole.Id};`}>
        <div class="role-hero__copy">
          <div class="role-hero__title-row">
            <h2>{currentRole.Name}</h2>
            <div class="role-pager">
              <button class="control control--icon control--soft" on:click={prevRole} aria-label="Rol anterior" disabled={roles.length <= 1 || busySection !== null}>
                <ButtonIcon name="chevron-left"/>
              </button>
              <span>{roles.findIndex((role) => role.Id === currentRole.Id) + 1} / {roles.length}</span>
              <button class="control control--icon control--soft" on:click={nextRole} aria-label="Rol siguiente" disabled={roles.length <= 1 || busySection !== null}>
                <ButtonIcon name="chevron-right"/>
              </button>
            </div>
          </div>
          <p>{currentRole.Description || "Usa este perfil para decidir qué tablas toca y qué procesos puede activar."}</p>
        </div>

        <div class="role-hero__side">
          <div class="role-stats">
            <div class="role-stat">
              <span>Tablas con acceso</span>
              <strong>{countGrantedTables(currentRole)}</strong>
            </div>
            <div class="role-stat">
              <span>Checks activos</span>
              <strong>{countGrantedChecks(currentRole)}</strong>
            </div>
            <div class="role-stat">
              <span>Procesos habilitados</span>
              <strong>{countGrantedProcesses(currentRole)}</strong>
            </div>
          </div>

          <div class="role-actions">
            <ModalLauncher
              triggerLabel="Editar rol"
              title="Editar rol"
              confirmLabel="Guardar"
              triggerVariant="edit"
              confirmVariant="primary"
              size="form"
              triggerClass="roles-modal-trigger roles-modal-trigger--hero"
              triggerDisabled={busySection !== null}
              onOpen={() => prepareRoleEdit(currentRole)}
              onSuccess={() => handleSaveRole(currentRole.Id)}
            >
              <div class="modal-intro">
                <p class="modal-kicker local-modal-kicker">Editor de rol</p>
                <p class="modal-hint local-modal-hint">Ajusta el nombre operativo y la descripcion del perfil.</p>
              </div>
              <label class="field">
                <span>Nombre</span>
                <input type="text" bind:value={roleDraftName} placeholder="Nombre del rol" />
              </label>
              <label class="field">
                <span>Descripcion</span>
                <textarea rows="3" bind:value={roleDraftDescription} placeholder="Describe para quién sirve este rol." />
              </label>
            </ModalLauncher>

            <ModalLauncher
              triggerLabel="Procesos"
              title="Permisos de procesos"
              confirmLabel="Guardar permisos"
              triggerVariant="primary"
              confirmVariant="primary"
              size="form"
              triggerClass="roles-modal-trigger roles-modal-trigger--hero"
              triggerDisabled={busySection !== null || !processCatalog.length}
              onOpen={() => prepareProcessPermissions(currentRole)}
              onSuccess={handleSaveProcessPermissions}
            >
              <div class="modal-intro">
                <p class="modal-kicker local-modal-kicker">Acceso por proceso</p>
                <p class="modal-hint local-modal-hint">Marca qué procesos puede ejecutar este rol. Aqui solo existe acceso o no acceso.</p>
              </div>

              {#if processGroups.length}
                <div class="process-modal-groups">
                  {#each processGroups as group (group.id)}
                    <section class="process-modal-group">
                      <div class="process-modal-group__head">
                        <div>
                          <strong>{group.name}</strong>
                          <p>{group.description || "Macroproceso sin descripcion."}</p>
                        </div>
                        <div class="process-modal-group__actions">
                          <button type="button" class="control control--soft control--sm" on:click={() => setProcessGroupDraft(group, true)}>
                            <ButtonIcon name="check"/>
                            <span>Todo</span>
                          </button>
                          <button type="button" class="control control--ghost control--sm" on:click={() => setProcessGroupDraft(group, false)}>
                            <ButtonIcon name="clear"/>
                            <span>Ninguno</span>
                          </button>
                        </div>
                      </div>
                      <div class="process-modal-list">
                        {#each group.processes as process (process.Id)}
                          <label class="process-permission-row">
                            <span class="process-permission-row__copy">
                              <strong>{process.Name}</strong>
                              <span>{process.Description || "Sin descripcion todavia."}</span>
                            </span>
                            <span class="process-permission-row__toggle">
                              <input
                                type="checkbox"
                                checked={Boolean(processPermissionDraft[process.Id])}
                                on:change={(event) => handleProcessDraftChange(process.Id, event)}
                              />
                              <span>Permitido</span>
                            </span>
                          </label>
                        {/each}
                      </div>
                    </section>
                  {/each}
                </div>
              {:else}
                <div class="roles-empty roles-empty--modal">
                  <strong>Aun no hay procesos</strong>
                  <p>Crea macroprocesos y procesos para poder asignarlos a roles.</p>
                </div>
              {/if}
            </ModalLauncher>

            <ModalLauncher
              triggerLabel="Eliminar rol"
              title="Eliminar rol"
              confirmLabel="Eliminar"
              triggerVariant="danger"
              confirmVariant="danger"
              size="default"
              triggerClass="roles-modal-trigger roles-modal-trigger--hero"
              triggerDisabled={busySection !== null}
              onSuccess={() => handleRemoveRole(currentRole.Id)}
            >
              <p class="modal-hint local-modal-hint">Se eliminara <strong>{currentRole.Name}</strong> con todos sus permisos asociados.</p>
            </ModalLauncher>
          </div>
        </div>
      </header>
    {/if}
  </div>

  <div class="roles-shell">
    <aside class="roles-rail">
      <div class="roles-rail__head">
        <div>
          <p class="section-kicker">Control Deck</p>
          <h3>Roles</h3>
          <p class="roles-rail__hint">Crea perfiles y entra directo a su matriz de acceso.</p>
        </div>
        <ModalLauncher
          triggerLabel="Crear rol"
          title="Crear rol"
          confirmLabel="Crear"
          triggerVariant="primary"
          confirmVariant="primary"
          size="form"
          triggerClass="roles-modal-trigger roles-modal-trigger--rail"
          triggerDisabled={busySection !== null}
          onOpen={prepareRoleCreate}
          onSuccess={handleAddRole}
        >
          <div class="modal-intro">
            <p class="modal-kicker local-modal-kicker">Nuevo rol</p>
            <p class="modal-hint local-modal-hint">Define un perfil de acceso para modelar permisos por tabla y por proceso.</p>
          </div>
          <label class="field">
            <span>Nombre</span>
            <input type="text" bind:value={newRoleName} placeholder="Analista, auditor, operador..." />
          </label>
          <label class="field">
            <span>Descripcion</span>
            <textarea rows="3" bind:value={newRoleDescription} placeholder="Explica el alcance funcional del rol." />
          </label>
        </ModalLauncher>
      </div>

      {#if roles.length}
        <div class="role-list">
          {#each roles as role, index (role.Id)}
            <button
              type="button"
              class:role-card={true}
              class:role-card--active={role.Id === currentRole?.Id}
              on:click={() => prefersRoleSelection(role.Id)}
              in:fly={{y: 16, delay: index * 34, duration: 320, easing: quintOut}}
              animate:flip={{duration: 360, easing: quintOut}}
              style={`view-transition-name: role-card-${role.Id};`}
            >
              <span class="role-card__index">{String(index + 1).padStart(2, "0")}</span>
              <ButtonIcon name="roles"/>
              <strong>{role.Name}</strong>
              <span class="role-card__meta">{countGrantedTables(role)} tablas · {countGrantedProcesses(role)} procesos</span>
              <span class="role-card__hint">{role.Description || "Sin descripcion todavia."}</span>
            </button>
          {/each}
        </div>
      {:else}
        <div class="roles-empty roles-empty--rail">
          <strong>No hay roles creados</strong>
          <p>Empieza por uno y luego define qué puede ver, modificar o ejecutar.</p>
        </div>
      {/if}
    </aside>

    <section class="roles-stage">
      {#if currentRole}
        <section class="permission-panel">
          <div class="permission-panel__head">
            <div>
              <p class="section-kicker">Matriz de tablas</p>
              <h3>Acceso por entidad</h3>
              <p>Activa permisos puntuales por tabla. La fila se limpia sola cuando todos los checks quedan apagados.</p>
            </div>
            <span class="permission-panel__hint">Toggle directo con guardado inmediato</span>
          </div>

          {#if entities.length}
            <div class="permission-matrix-wrap">
              <table class="permission-matrix">
                <thead>
                  <tr>
                    <th>Tabla</th>
                    {#each permissionColumns as column}
                      <th title={column.hint}>{column.label}</th>
                    {/each}
                  </tr>
                </thead>
                <tbody>
                  {#each entities as entity, index (entity.Id)}
                    <tr
                      in:fade={{duration: 180, delay: index * 12}}
                      animate:flip={{duration: 280, easing: quintOut}}
                    >
                      <td>
                        <div class="table-cell">
                          <strong>{entity.Name}</strong>
                          <span>{entityDescription(entity.Id)}</span>
                        </div>
                      </td>
                      {#each permissionColumns as column}
                        <td>
                          <label class={`permission-toggle permission-toggle--${column.short.toLowerCase()}`}>
                            <input
                              type="checkbox"
                              checked={getTablePermissionValue(currentRole, entity.Id, column.key)}
                              disabled={busySection !== null}
                              on:change={(event) => handleTablePermissionChange(entity.Id, column.key, event)}
                            />
                            <span class="permission-toggle__box" aria-hidden="true">
                              <svg viewBox="0 0 24 24">
                                <path d="M9.2 16.4 4.8 12a1 1 0 0 1 1.4-1.4l3 3 8.6-8.6a1 1 0 1 1 1.4 1.4l-10 10a1 1 0 0 1-1.4 0Z"/>
                              </svg>
                            </span>
                            <span class="sr-only">{column.label} en {entity.Name}</span>
                          </label>
                        </td>
                      {/each}
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          {:else}
            <div class="roles-empty">
              <strong>No hay tablas para mapear</strong>
              <p>Crea entidades primero y luego decide qué puede hacer cada rol con ellas.</p>
            </div>
          {/if}
        </section>
      {:else}
        <div class="roles-empty roles-empty--stage">
          <strong>Crea el primer rol para abrir la matriz</strong>
          <p>La pestaña esta lista para modelar permisos por tabla y por proceso, pero necesita al menos un perfil.</p>
        </div>
      {/if}
    </section>
  </div>
</section>

<style>
  .roles-tab {
    --roles-sticky-total-height: 0px;
    display: grid;
    gap: 1rem;
  }

  .roles-sticky-stack {
    display: grid;
    gap: 1rem;
    margin-bottom: 18px;
  }

  .roles-sticky-stack--pinned {
    position: sticky;
    top: 0;
    z-index: calc(var(--layer-ribbon) - 2);
  }

  .roles-sticky-sentinel {
    height: 1px;
    margin-top: -1px;
  }

  .roles-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: end;
    gap: 1rem;
    padding: 1rem 1.1rem;
    border: 1px solid var(--border);
    border-radius: calc(var(--radius-lg) - 8px);
    background:
      linear-gradient(180deg, color-mix(in srgb, var(--surface-strong) 98%, var(--surface)), color-mix(in srgb, var(--surface) 100%, var(--surface-strong))),
      linear-gradient(90deg, color-mix(in srgb, var(--accent) 10%, var(--surface-strong)), transparent 36%);
    box-shadow: var(--shadow-sm);
    backdrop-filter: blur(18px);
    position: relative;
    overflow: clip;
  }

  .roles-toolbar .label,
  .roles-toolbar .muted {
    margin: 0;
  }

  .roles-toolbar::before,
  .role-hero::before {
    content: "";
    position: absolute;
    inset: 0 auto auto 0;
    width: min(220px, 42%);
    height: 1px;
    background: linear-gradient(90deg, color-mix(in srgb, var(--accent) 34%, transparent), transparent);
    pointer-events: none;
  }

  .roles-toolbar__meta {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 0.55rem;
  }

  .roles-chip {
    display: inline-flex;
    align-items: center;
    min-height: 2rem;
    padding: 0 0.82rem;
    border-radius: 999px;
    border: 1px solid color-mix(in srgb, var(--accent) 16%, var(--border));
    background: color-mix(in srgb, var(--accent) 10%, var(--surface-strong));
    color: var(--accent-strong);
    font-size: 0.8rem;
    font-weight: 800;
  }

  .roles-chip--quiet {
    border-color: var(--border);
    background: color-mix(in srgb, var(--surface-strong) 92%, transparent);
    color: var(--ink-soft);
  }

  .roles-shell {
    display: grid;
    grid-template-columns: minmax(17rem, 0.82fr) minmax(0, 1.42fr);
    gap: 1rem;
    align-items: start;
  }

  .roles-rail,
  .roles-stage {
    position: relative;
    overflow: hidden;
    border: 1px solid var(--border);
    border-radius: calc(var(--radius-lg) - 8px);
    background: color-mix(in srgb, var(--surface-strong) 95%, transparent);
    box-shadow: var(--shadow-sm);
  }

  .roles-rail {
    position: sticky;
    top: calc(var(--roles-sticky-total-height) + 1rem);
    align-self: start;
    display: grid;
    gap: 0.95rem;
    padding: 1rem;
    background:
      linear-gradient(180deg, color-mix(in srgb, var(--surface-strong) 96%, transparent), color-mix(in srgb, var(--surface) 90%, transparent)),
      radial-gradient(circle at top, color-mix(in srgb, var(--accent) 12%, transparent), transparent 44%);
  }

  .roles-stage {
    display: grid;
    gap: 1rem;
    padding: 1.05rem;
    background:
      radial-gradient(circle at top right, color-mix(in srgb, var(--accent) 14%, transparent), transparent 32%),
      linear-gradient(180deg, color-mix(in srgb, var(--surface-strong) 97%, transparent), color-mix(in srgb, var(--surface) 89%, transparent));
  }

  .roles-rail__head {
    display: grid;
    gap: 0.82rem;
  }

  .roles-rail__head h3,
  .permission-panel__head h3,
  .role-hero__copy h2 {
    margin: 0.18rem 0 0;
    line-height: 0.98;
  }

  .roles-rail__hint,
  .permission-panel__head p:last-child,
  .role-hero__copy p:last-child {
    margin: 0.38rem 0 0;
    color: var(--ink-soft);
    line-height: 1.5;
  }

  .role-list {
    display: grid;
    gap: 0.7rem;
  }

  .role-card {
    position: relative;
    display: grid;
    gap: 0.26rem;
    padding: 0.9rem 0.95rem 0.96rem 1.05rem;
    border-radius: calc(var(--radius-md) - 10px);
    border: 1px solid var(--border);
    background:
      linear-gradient(180deg, color-mix(in srgb, var(--surface-strong) 98%, transparent), color-mix(in srgb, var(--surface) 86%, transparent));
    text-align: left;
    transition:
      transform 200ms cubic-bezier(0.22, 1, 0.36, 1),
      border-color 180ms ease,
      box-shadow 180ms ease;
  }

  .role-card::before {
    content: "";
    position: absolute;
    inset: 0 auto 0 0;
    width: 4px;
    border-radius: 999px;
    background: linear-gradient(180deg, color-mix(in srgb, var(--accent) 88%, white 12%), color-mix(in srgb, var(--accent-strong) 78%, transparent));
    opacity: 0.28;
  }

  .role-card:hover,
  .role-card--active {
    transform: translateX(6px);
    border-color: color-mix(in srgb, var(--accent) 26%, var(--border));
    box-shadow:
      0 20px 32px color-mix(in srgb, var(--ink) 10%, transparent),
      0 0 0 0.16rem color-mix(in srgb, var(--accent) 10%, transparent);
  }

  .role-card__index {
    color: var(--ink-faint);
    font-size: 0.72rem;
    font-weight: 800;
    letter-spacing: 0.15em;
  }

  .role-card :global(.button-glyph) {
    color: var(--accent-strong);
    opacity: 0.86;
  }

  .role-card strong {
    font-size: 1rem;
    line-height: 1.18;
  }

  .role-card__meta,
  .role-card__hint {
    color: var(--ink-soft);
    font-size: 0.82rem;
    line-height: 1.42;
  }

  .role-hero {
    display: grid;
    grid-template-columns: minmax(0, 1.05fr) auto;
    gap: 1rem;
    padding: 1.1rem 1.15rem;
    border-radius: calc(var(--radius-lg) - 10px);
    border: 1px solid color-mix(in srgb, var(--accent) 18%, var(--border));
    background:
      linear-gradient(130deg, color-mix(in srgb, var(--surface-strong) 99%, var(--surface)), color-mix(in srgb, var(--surface) 96%, var(--surface-strong))),
      linear-gradient(90deg, color-mix(in srgb, var(--accent) 8%, var(--surface-strong)), transparent 48%);
    backdrop-filter: blur(18px);
  }

  .role-hero__title-row {
    display: flex;
    align-items: center;
    gap: 0.9rem;
    flex-wrap: wrap;
  }

  .role-hero__side {
    display: grid;
    gap: 0.85rem;
    align-content: start;
  }

  .role-stats {
    display: grid;
    grid-template-columns: repeat(3, minmax(6.6rem, 1fr));
    gap: 0.65rem;
  }

  .role-stat {
    display: grid;
    gap: 0.28rem;
    padding: 0.82rem 0.88rem;
    border-radius: 1rem;
    border: 1px solid color-mix(in srgb, var(--accent) 12%, var(--border));
    background: color-mix(in srgb, var(--surface-strong) 86%, transparent);
    box-shadow: inset 0 1px 0 color-mix(in srgb, white 20%, transparent);
  }

  .role-stat span {
    color: var(--ink-faint);
    font-size: 0.73rem;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .role-stat strong {
    font-size: 1.28rem;
    line-height: 1;
  }

  .role-pager {
    display: inline-flex;
    align-items: center;
    gap: 0.42rem;
    padding: 0.3rem 0.38rem;
    border-radius: 999px;
    border: 1px solid color-mix(in srgb, var(--accent) 16%, var(--border));
    background: color-mix(in srgb, var(--surface-strong) 90%, transparent);
  }

  .role-pager span {
    min-width: 3.7rem;
    text-align: center;
    color: var(--ink-soft);
    font-size: 0.8rem;
    font-weight: 800;
  }

  .role-actions {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 0.55rem;
  }

  :global(.roles-modal-trigger) {
    width: 100%;
    justify-content: center;
  }

  :global(.roles-modal-trigger--rail),
  :global(.roles-modal-trigger--hero) {
    width: auto;
  }

  .permission-panel {
    display: grid;
    gap: 0.9rem;
  }

  .permission-panel__head {
    display: flex;
    justify-content: space-between;
    gap: 1rem;
    align-items: end;
  }

  .permission-panel__hint {
    display: inline-flex;
    align-items: center;
    min-height: 2rem;
    padding: 0 0.8rem;
    border-radius: 999px;
    border: 1px solid color-mix(in srgb, var(--accent) 16%, var(--border));
    background: color-mix(in srgb, var(--accent) 10%, var(--surface-strong));
    color: var(--accent-strong);
    font-size: 0.8rem;
    font-weight: 800;
    white-space: nowrap;
  }

  .permission-matrix-wrap {
    overflow: auto;
    border-radius: calc(var(--radius-lg) - 12px);
    border: 1px solid var(--border);
    background: color-mix(in srgb, var(--surface-strong) 94%, transparent);
  }

  .permission-matrix {
    width: 100%;
    border-collapse: collapse;
    min-width: 42rem;
  }

  .permission-matrix th,
  .permission-matrix td {
    padding: 0.92rem 0.88rem;
    border-bottom: 1px solid color-mix(in srgb, var(--ink) 9%, transparent);
    vertical-align: middle;
  }

  .permission-matrix th {
    position: sticky;
    top: 0;
    z-index: 1;
    background: color-mix(in srgb, var(--surface-strong) 98%, transparent);
    color: var(--ink-soft);
    font-size: 0.78rem;
    font-weight: 800;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    text-align: center;
  }

  .permission-matrix th:first-child,
  .permission-matrix td:first-child {
    text-align: left;
    width: 15.5rem;
    min-width: 15.5rem;
  }

  .permission-matrix tbody tr:hover {
    background: color-mix(in srgb, var(--accent) 5%, transparent);
  }

  .table-cell {
    display: grid;
    gap: 0.18rem;
  }

  .table-cell strong {
    font-size: 0.95rem;
    line-height: 1.2;
  }

  .table-cell span {
    color: var(--ink-soft);
    font-size: 0.8rem;
    line-height: 1.45;
    max-width: 26ch;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .permission-toggle {
    position: relative;
    display: inline-grid;
    place-items: center;
    width: 100%;
    cursor: pointer;
  }

  .permission-toggle input {
    position: absolute;
    opacity: 0;
    pointer-events: none;
  }

  .permission-toggle__box {
    width: 2.15rem;
    height: 2.15rem;
    display: inline-grid;
    place-items: center;
    border-radius: 0.85rem;
    border: 1px solid var(--border);
    background: color-mix(in srgb, var(--surface-strong) 94%, transparent);
    box-shadow: inset 0 1px 0 color-mix(in srgb, white 24%, transparent);
    transition:
      transform 170ms cubic-bezier(0.22, 1, 0.36, 1),
      border-color 170ms ease,
      background 170ms ease,
      box-shadow 170ms ease;
  }

  .permission-toggle__box svg {
    width: 1rem;
    height: 1rem;
    fill: currentColor;
    opacity: 0;
    transform: scale(0.76);
    transition: transform 150ms ease, opacity 150ms ease;
  }

  .permission-toggle--r input:checked + .permission-toggle__box {
    border-color: color-mix(in srgb, var(--accent) 24%, var(--border));
    background: color-mix(in srgb, var(--accent) 13%, var(--surface-strong));
    color: var(--accent-strong);
  }

  .permission-toggle--c input:checked + .permission-toggle__box {
    border-color: color-mix(in srgb, var(--success) 26%, var(--border));
    background: color-mix(in srgb, var(--success) 14%, var(--surface-strong));
    color: var(--success);
  }

  .permission-toggle--u input:checked + .permission-toggle__box {
    border-color: color-mix(in srgb, var(--accent) 32%, var(--border));
    background: color-mix(in srgb, var(--accent) 18%, var(--surface-strong));
    color: var(--accent-strong);
  }

  .permission-toggle--d input:checked + .permission-toggle__box {
    border-color: color-mix(in srgb, var(--danger) 26%, var(--border));
    background: color-mix(in srgb, var(--danger) 14%, var(--surface-strong));
    color: var(--danger);
  }

  .permission-toggle input:checked + .permission-toggle__box {
    transform: translateY(-1px) scale(1.02);
    box-shadow:
      0 10px 18px color-mix(in srgb, var(--ink) 8%, transparent),
      inset 0 1px 0 color-mix(in srgb, white 34%, transparent);
  }

  .permission-toggle input:checked + .permission-toggle__box svg {
    opacity: 1;
    transform: scale(1);
  }

  .permission-toggle input:focus-visible + .permission-toggle__box {
    outline: none;
    box-shadow:
      0 0 0 0.2rem color-mix(in srgb, var(--accent) 14%, transparent),
      inset 0 1px 0 color-mix(in srgb, white 24%, transparent);
  }

  .process-modal-groups {
    display: grid;
    gap: 0.9rem;
  }

  .process-modal-group {
    display: grid;
    gap: 0.7rem;
    padding: 0.95rem;
    border-radius: 1.1rem;
    border: 1px solid color-mix(in srgb, var(--accent) 12%, var(--border));
    background: color-mix(in srgb, var(--surface-strong) 94%, transparent);
  }

  .process-modal-group__head {
    display: flex;
    justify-content: space-between;
    align-items: start;
    gap: 0.8rem;
  }

  .process-modal-group__head strong {
    display: block;
    font-size: 0.96rem;
  }

  .process-modal-group__head p {
    margin: 0.32rem 0 0;
    color: var(--ink-soft);
    font-size: 0.82rem;
    line-height: 1.45;
  }

  .process-modal-group__actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.45rem;
  }

  .process-modal-list {
    display: grid;
    gap: 0.58rem;
  }

  .process-permission-row {
    display: flex;
    justify-content: space-between;
    gap: 0.8rem;
    align-items: center;
    padding: 0.82rem 0.88rem;
    border-radius: 0.95rem;
    border: 1px solid var(--border);
    background: color-mix(in srgb, var(--surface) 88%, transparent);
  }

  .process-permission-row__copy {
    display: grid;
    gap: 0.16rem;
  }

  .process-permission-row__copy strong {
    font-size: 0.92rem;
  }

  .process-permission-row__copy span {
    color: var(--ink-soft);
    font-size: 0.8rem;
    line-height: 1.42;
  }

  .process-permission-row__toggle {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    color: var(--ink-soft);
    font-size: 0.82rem;
    font-weight: 700;
    white-space: nowrap;
  }

  .process-permission-row__toggle input {
    width: 1.05rem;
    height: 1.05rem;
    accent-color: var(--accent);
  }

  .roles-empty {
    display: grid;
    place-items: center;
    gap: 0.35rem;
    min-height: 12rem;
    padding: 1.2rem;
    text-align: center;
    border-radius: calc(var(--radius-md) - 8px);
    border: 1px dashed color-mix(in srgb, var(--accent) 16%, var(--border));
    background: color-mix(in srgb, var(--surface-strong) 78%, transparent);
    color: var(--ink-soft);
  }

  .roles-empty strong {
    color: var(--ink);
    font-size: 1rem;
  }

  .roles-empty p {
    margin: 0;
    line-height: 1.55;
  }

  .roles-empty--rail {
    min-height: 10rem;
  }

  .roles-empty--stage {
    min-height: 30rem;
  }

  .roles-empty--modal {
    min-height: 8rem;
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
  .field textarea {
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
  .field textarea:focus {
    outline: none;
    border-color: color-mix(in srgb, var(--accent) 34%, var(--border));
    box-shadow: 0 0 0 0.22rem color-mix(in srgb, var(--accent) 14%, transparent);
    transform: translateY(-1px);
    background: color-mix(in srgb, var(--surface-strong) 98%, transparent);
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

  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }

  @media (max-width: 1240px) {
    .roles-shell {
      grid-template-columns: 1fr;
    }

    .roles-rail {
      position: static;
    }

    .role-hero {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 920px) {
    .roles-toolbar,
    .permission-panel__head,
    .process-modal-group__head,
    .process-permission-row {
      flex-direction: column;
      align-items: stretch;
    }

    .roles-toolbar__meta,
    .role-actions {
      justify-content: flex-start;
    }

    .role-stats {
      grid-template-columns: 1fr;
    }

    :global(.roles-modal-trigger--rail),
    :global(.roles-modal-trigger--hero) {
      width: 100%;
    }
  }

  @media (max-width: 720px) {
    .role-hero__title-row {
      align-items: stretch;
    }

    .role-pager {
      justify-content: space-between;
    }

    .permission-matrix {
      min-width: 44rem;
    }

    .process-modal-group__actions,
    .role-actions {
      flex-direction: column;
      align-items: stretch;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .role-card,
    .permission-toggle__box {
      transition: none;
    }
  }
</style>
