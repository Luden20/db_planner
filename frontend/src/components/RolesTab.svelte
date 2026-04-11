<script lang="ts">
  import { flip } from "svelte/animate";
  import { quintOut } from "svelte/easing";
  import { fade, fly } from "svelte/transition";
  import { tick } from "svelte";
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
  import StudioToolbar from "./studio/StudioToolbar.svelte";
  import StickyStack from "./studio/StickyStack.svelte";
  import EmptyPanel from "./studio/EmptyPanel.svelte";
  import Button from "./ui/Button.svelte";
  import Badge from "./ui/Badge.svelte";
  import Table from "./ui/Table.svelte";
  import {getErrorMessage, runViewTransition} from "../lib/ui-helpers";

  type PermissionKey = "ViewPermission" | "InsertPermission" | "UpdatePermission" | "DeletePermission";
  type ProcessGroup = {
    id: number;
    name: string;
    description: string;
    processes: utils.Process[];
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
  let busySection: string | null = null;
  let newRoleName = "";
  let newRoleDescription = "";
  let roleDraftName = "";
  let roleDraftDescription = "";
  let processPermissionDraft: Record<number, boolean> = {};

  const runRoleTransition = (update: () => void | Promise<void>) =>
    runViewTransition(update, "No se pudo aplicar la transicion de roles:");

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
      const message = getErrorMessage(err);
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

</script>

<section class="roles-tab roles-studio">
  <StickyStack>
    {#if currentRole}
      <StudioToolbar 
        title={currentRole.Name} 
        description={currentRole.Description || "Usa este perfil para decidir qué tablas toca y qué procesos puede activar."}
      >
        {#snippet meta()}
          <Badge>{countGrantedTables(currentRole)} tablas con acceso</Badge>
          <Badge variant="quiet">{countGrantedChecks(currentRole)} checks activos</Badge>
          <Badge variant="quiet">{countGrantedProcesses(currentRole)} procesos habilitados</Badge>
        {/snippet}
        {#snippet actions()}
          <div class="entity-nav">
            <Button variant="soft" size="icon" icon="chevron-left" disabled={roles.length <= 1 || busySection !== null} onclick={prevRole} aria-label="Rol anterior" />
            <span style="font-size:12px;font-weight:600;min-width:32px;text-align:center;">{roles.findIndex((role) => role.Id === currentRole?.Id) + 1} / {roles.length}</span>
            <Button variant="soft" size="icon" icon="chevron-right" disabled={roles.length <= 1 || busySection !== null} onclick={nextRole} aria-label="Rol siguiente" />
          </div>


            <ModalLauncher
              triggerLabel="Editar rol"
              title="Editar rol"
              confirmLabel="Guardar"
              triggerVariant="soft"
              confirmVariant="primary"
              size="form"
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
                <textarea rows="3" bind:value={roleDraftDescription} placeholder="Describe para quién sirve este rol."></textarea>
              </label>
            </ModalLauncher>

            <ModalLauncher
              triggerLabel="Procesos"
              title="Permisos de procesos"
              confirmLabel="Guardar permisos"
              triggerVariant="soft"
              confirmVariant="primary"
              size="form"
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
                          <Button variant="soft" size="sm" icon="check" onclick={() => setProcessGroupDraft(group, true)}>
                            Todo
                          </Button>
                          <Button variant="ghost" size="sm" icon="clear" onclick={() => setProcessGroupDraft(group, false)}>
                            Ninguno
                          </Button>
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
              triggerVariant="soft"
              confirmVariant="danger"
              size="default"
              triggerDisabled={busySection !== null}
              onSuccess={() => handleRemoveRole(currentRole.Id)}
            >
              <p class="modal-hint local-modal-hint">Se eliminara <strong>{currentRole.Name}</strong> con todos sus permisos asociados.</p>
            </ModalLauncher>
        {/snippet}
      </StudioToolbar>
    {/if}
  </StickyStack>

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
            <textarea rows="3" bind:value={newRoleDescription} placeholder="Explica el alcance funcional del rol."></textarea>
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
              <h3>Acceso por entidad</h3>
              <p>Permisos por tabla.</p>
            </div>
            <span class="permission-panel__hint">Auto-guardado</span>
          </div>

          {#if entities.length}
            <div class="permission-matrix-wrap">
              <Table class="permission-matrix">
                {#snippet header()}
                  <th>Tabla</th>
                  {#each permissionColumns as column}
                    <th title={column.hint}>{column.label}</th>
                  {/each}
                {/snippet}
                {#snippet body()}
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
                {/snippet}
              </Table>
            </div>
          {:else}
            <EmptyPanel 
              message="No hay tablas para mapear" 
              resolution="Crea entidades primero y luego decide qué puede hacer cada rol con ellas." 
            />
          {/if}
        </section>
      {:else}
        <EmptyPanel 
          message="Crea el primer rol para abrir la matriz" 
          resolution="La pestaña esta lista para modelar permisos por tabla y por proceso, pero necesita al menos un perfil." 
        />
      {/if}
    </section>
  </div>
</section>
