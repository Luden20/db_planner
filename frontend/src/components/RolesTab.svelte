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

  let { 
    project, 
    entities = [], 
    onRefresh = async () => {} 
  } = $props<{
    project: utils.DbProject;
    entities?: utils.Entity[];
    onRefresh?: () => Promise<void>;
  }>();

  const permissionColumns: Array<{key: PermissionKey; label: string; short: string; hint: string}> = [
    {key: "InsertPermission", label: "Create", short: "C", hint: "Crear registros"},
    {key: "ViewPermission", label: "Read", short: "R", hint: "Consultar registros"},
    {key: "UpdatePermission", label: "Update", short: "U", hint: "Modificar registros"},
    {key: "DeletePermission", label: "Delete", short: "D", hint: "Eliminar registros"}
  ];

  let selectedRoleId = $state<number | null>(null);
  let busySection = $state<string | null>(null);
  let newRoleName = $state("");
  let newRoleDescription = $state("");
  let roleDraftName = $state("");
  let roleDraftDescription = $state("");
  let processPermissionDraft = $state<Record<number, boolean>>({});

  const roles = $derived(project?.Roles ?? []);
  const currentRole = $derived(roles.find((role) => role.Id === selectedRoleId) ?? null);
  const processGroups = $derived((project?.BigProcesses ?? [])
    .map((bigProcess) => ({
      id: bigProcess.Id,
      name: bigProcess.Name,
      description: bigProcess.Description ?? "",
      processes: bigProcess.Processes ?? []
    }))
    .filter((group) => group.processes.length > 0));

  const processCatalog = $derived(processGroups.flatMap((group) =>
    group.processes.map((process) => ({
      id: process.Id,
      bigProcessId: group.id,
      bigProcessName: group.name,
      name: process.Name,
      description: process.Description ?? ""
    }))
  ));

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
    if (busySection !== null) return false;
    busySection = options.busyKey ?? "roles";
    try {
      await action();
      await Save();
      await onRefresh();
      await tick();
      if (options.successMessage) showToast(options.successMessage, "success");
      return true;
    } catch (err) {
      showToast(`No se pudieron actualizar los roles: ${getErrorMessage(err)}`, "error");
      return false;
    } finally {
      busySection = null;
    }
  };

  const getTablePermission = (role: utils.Role | null, tableId: number) =>
    (role?.TablePermissions ?? []).find((permission) => permission.TableId === tableId) ?? null;

  const hasAnyTablePermission = (permission: { ViewPermission: boolean; InsertPermission: boolean; UpdatePermission: boolean; DeletePermission: boolean; }) =>
    permission.ViewPermission || permission.InsertPermission || permission.UpdatePermission || permission.DeletePermission;

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
    if (!trimmedName) throw new Error("Ingresa un nombre para el rol.");
    selectedRoleId = (project?.RoleLastMax ?? 0) + 1;
    const ok = await persistRoleChange(() => AddRole(trimmedName, newRoleDescription.trim()), {successMessage: "Rol creado.", busyKey: "add-role"});
    if (!ok) throw new Error("No se pudo crear el rol.");
  };

  const prepareRoleEdit = (role: utils.Role) => {
    selectedRoleId = role.Id;
    roleDraftName = role.Name ?? "";
    roleDraftDescription = role.Description ?? "";
  };

  const handleSaveRole = async (roleId = currentRole?.Id ?? null) => {
    if (roleId === null) throw new Error("Selecciona un rol para editarlo.");
    const trimmedName = roleDraftName.trim();
    if (!trimmedName) throw new Error("El rol necesita un nombre.");
    selectedRoleId = roleId;
    const ok = await persistRoleChange(() => EditRole(roleId, trimmedName, roleDraftDescription.trim()), {successMessage: "Rol actualizado.", busyKey: "edit-role"});
    if (!ok) throw new Error("No se pudo guardar el rol.");
  };

  const handleRemoveRole = async (roleId = currentRole?.Id ?? null) => {
    if (roleId === null) throw new Error("Selecciona un rol para eliminarlo.");
    const currentIndex = roles.findIndex((role) => role.Id === roleId);
    selectedRoleId = roles[currentIndex + 1]?.Id ?? roles[currentIndex - 1]?.Id ?? null;
    const ok = await persistRoleChange(() => RemoveRole(roleId), {successMessage: "Rol eliminado.", busyKey: "remove-role"});
    if (!ok) throw new Error("No se pudo eliminar el rol.");
  };

  const prepareProcessPermissions = (role = currentRole) => {
    if (!role) throw new Error("Selecciona un rol para gestionar procesos.");
    const nextDraft: Record<number, boolean> = {};
    const granted = new Set((role.ProcessPermissions ?? []).map((p) => p.ProcessId));
    for (const process of processCatalog) {
      nextDraft[process.id] = granted.has(process.id);
    }
    processPermissionDraft = nextDraft;
  };

  const setProcessGroupDraft = (group: ProcessGroup, checked: boolean) => {
    const nextDraft = {...processPermissionDraft};
    for (const process of group.processes) nextDraft[process.Id] = checked;
    processPermissionDraft = nextDraft;
  };

  const handleProcessDraftToggle = (processId: number, checked: boolean) => {
    processPermissionDraft = { ...processPermissionDraft, [processId]: checked };
  };

  const handleProcessDraftChange = (processId: number, event: Event) => {
    const target = event.currentTarget as HTMLInputElement;
    handleProcessDraftToggle(processId, target.checked);
  };

  const handleSaveProcessPermissions = async () => {
    if (!currentRole) throw new Error("Selecciona un rol para gestionar procesos.");
    const existingByProcess = new Map((currentRole.ProcessPermissions ?? []).map((p) => [p.ProcessId, p]));
    const ok = await persistRoleChange(async () => {
      for (const process of processCatalog) {
        const shouldHaveAccess = Boolean(processPermissionDraft[process.id]);
        const existingPermission = existingByProcess.get(process.id) ?? null;
        if (shouldHaveAccess && !existingPermission) await AddRoleProcessPermission(currentRole.Id, process.id);
        if (!shouldHaveAccess && existingPermission) await RemoveRoleProcessPermission(currentRole.Id, existingPermission.Id);
      }
    }, { successMessage: "Permisos de procesos actualizados.", busyKey: "edit-role-processes" });
    if (!ok) throw new Error("No se pudieron guardar los permisos de procesos.");
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
    if (!existingPermission && !hasAnyTablePermission(nextPermission)) return;
    if (!existingPermission) {
      await persistRoleChange(() => AddRoleTablePermission(currentRole.Id, tableId, nextPermission.InsertPermission, nextPermission.DeletePermission, nextPermission.UpdatePermission, nextPermission.ViewPermission), { busyKey: `table-${tableId}-${key}` });
      return;
    }
    if (!hasAnyTablePermission(nextPermission)) {
      await persistRoleChange(() => RemoveRoleTablePermission(currentRole.Id, existingPermission.Id), { busyKey: `table-${tableId}-${key}` });
      return;
    }
    await persistRoleChange(() => EditRoleTablePermission(currentRole.Id, existingPermission.Id, tableId, nextPermission.InsertPermission, nextPermission.DeletePermission, nextPermission.UpdatePermission, nextPermission.ViewPermission), { busyKey: `table-${tableId}-${key}` });
  };

  const handleTablePermissionChange = async (tableId: number, key: PermissionKey, event: Event) => {
    const target = event.currentTarget as HTMLInputElement;
    await handleTablePermissionToggle(tableId, key, target.checked);
  };

  const prevRole = async () => {
    if (!roles.length) return;
    const currentIndex = roles.findIndex((role) => role.Id === currentRole?.Id);
    const nextIndex = currentIndex <= 0 ? roles.length - 1 : currentIndex - 1;
    await prefersRoleSelection(roles[nextIndex].Id);
  };

  const nextRole = async () => {
    if (!roles.length) return;
    const currentIndex = roles.findIndex((role) => role.Id === currentRole?.Id);
    const nextIndex = currentIndex === -1 || currentIndex === roles.length - 1 ? 0 : currentIndex + 1;
    await prefersRoleSelection(roles[nextIndex].Id);
  };

  $effect(() => {
    if (selectedRoleId === null && roles.length > 0) selectedRoleId = roles[0].Id;
    if (selectedRoleId !== null && !roles.some((role) => role.Id === selectedRoleId)) selectedRoleId = roles[0]?.Id ?? null;
  });
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
            {#snippet children()}
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
            {/snippet}
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
            {#snippet children()}
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
                                onchange={(event) => handleProcessDraftChange(process.Id, event)}
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
            {/snippet}
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
            {#snippet children()}
              <p class="modal-hint local-modal-hint">Se eliminara <strong>{currentRole.Name}</strong> con todos sus permisos asociados.</p>
            {/snippet}
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
          {#snippet children()}
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
          {/snippet}
        </ModalLauncher>
      </div>

      {#if roles.length}
        <div class="role-list">
          {#each roles as role, index (role.Id)}
            <button
              type="button"
              class:role-card={true}
              class:role-card--active={role.Id === currentRole?.Id}
              onclick={() => prefersRoleSelection(role.Id)}
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
                              onchange={(event) => handleTablePermissionChange(entity.Id, column.key, event)}
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

<style>
  .roles-tab {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .roles-shell {
    display: grid;
    grid-template-columns: 340px 1fr;
    gap: 2rem;
    align-items: start;
  }

  .roles-rail {
    display: flex;
    flex-direction: column;
    gap: 2rem;
    position: sticky;
    top: calc(var(--sticky-stack-total-height, 0px) + 2rem);
    background: var(--surface);
    padding: 1.5rem;
    border: 1px solid var(--border-card);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-sm);
  }

  .roles-rail__head {
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
  }

  .roles-rail__hint {
    font-size: 0.75rem;
    color: var(--ink-soft);
    line-height: 1.4;
  }

  .role-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .role-card {
    position: relative;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    padding: 1.25rem;
    background: var(--background);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-md);
    text-align: left;
    transition: all 0.2s cubic-bezier(0.19, 1, 0.22, 1);
    cursor: pointer;
    overflow: hidden;
  }

  .role-card:hover {
    border-color: var(--accent);
    transform: translateX(4px);
    box-shadow: var(--shadow-mini);
  }

  .role-card--active {
    background: var(--accent-ghost);
    border-color: var(--accent);
    box-shadow: inset 4px 0 0 var(--accent);
  }

  .role-card__index {
    position: absolute;
    top: 1rem;
    right: 1.25rem;
    font-size: 0.65rem;
    font-weight: 800;
    color: var(--accent-soft-ink);
    opacity: 0.5;
  }

  .role-card strong {
    font-size: 0.95rem;
    font-weight: 800;
    color: var(--ink);
    margin-top: 0.5rem;
  }

  .role-card__meta {
    font-size: 0.65rem;
    font-weight: 700;
    color: var(--accent);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .role-card__hint {
    font-size: 0.7rem;
    color: var(--ink-soft);
    line-height: 1.4;
    margin-top: 0.25rem;
  }

  .roles-stage {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 2rem;
  }

  .permission-panel {
    background: var(--background);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-sm);
    overflow: hidden;
  }

  .permission-panel__head {
    padding: 2rem;
    border-bottom: 1px solid var(--border-card);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .permission-panel__hint {
    font-size: 0.65rem;
    font-weight: 800;
    text-transform: uppercase;
    color: var(--success);
    letter-spacing: 0.05em;
    padding: 0.25rem 0.6rem;
    background: var(--success-ghost);
    border-radius: 99px;
  }

  .table-cell {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
  }

  .table-cell strong {
    font-size: 0.85rem;
    font-weight: 800;
  }

  .table-cell span {
    font-size: 0.7rem;
    color: var(--ink-soft);
  }

  .permission-toggle {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .permission-toggle input {
    position: absolute;
    opacity: 0;
    cursor: pointer;
    height: 100%;
    width: 100%;
    z-index: 2;
  }

  .permission-toggle__box {
    width: 1.75rem;
    height: 1.75rem;
    background: var(--muted-soft);
    border: 1px solid var(--border-card);
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s ease;
  }

  .permission-toggle__box svg {
    width: 1rem;
    height: 1rem;
    fill: white;
    transform: scale(0.5);
    opacity: 0;
    transition: all 0.2s cubic-bezier(0.19, 1, 0.22, 1);
  }

  .permission-toggle input:checked + .permission-toggle__box {
    background: var(--accent);
    border-color: var(--accent);
    box-shadow: 0 2px 8px color-mix(in srgb, var(--accent) 30%, transparent);
  }

  .permission-toggle input:checked + .permission-toggle__box svg {
    transform: scale(1);
    opacity: 1;
  }

  .permission-toggle--c input:checked + .permission-toggle__box { background: #3b82f6; border-color: #2563eb; }
  .permission-toggle--r input:checked + .permission-toggle__box { background: #10b981; border-color: #059669; }
  .permission-toggle--u input:checked + .permission-toggle__box { background: #f59e0b; border-color: #d97706; }
  .permission-toggle--d input:checked + .permission-toggle__box { background: #ef4444; border-color: #dc2626; }

  .process-modal-groups {
    display: flex;
    flex-direction: column;
    gap: 2rem;
    padding: 1rem 0;
  }

  .process-modal-group {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    border: 1px solid var(--border-card);
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  .process-modal-group__head {
    padding: 1.25rem;
    background: var(--muted-soft);
    border-bottom: 1px solid var(--border-card);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .process-modal-group__head strong {
    font-size: 0.95rem;
    color: var(--ink);
  }

  .process-modal-group__head p {
    font-size: 0.7rem;
    color: var(--ink-soft);
  }

  .process-modal-group__actions {
    display: flex;
    gap: 0.5rem;
  }

  .process-modal-list {
    display: flex;
    flex-direction: column;
  }

  .process-permission-row {
    padding: 1rem 1.25rem;
    border-bottom: 1px solid var(--border-card);
    display: flex;
    justify-content: space-between;
    align-items: center;
    cursor: pointer;
    transition: background 0.15s ease;
  }

  .process-permission-row:last-child {
    border-bottom: none;
  }

  .process-permission-row:hover {
    background: var(--muted-ghost);
  }

  .process-permission-row__copy {
    display: flex;
    flex-direction: column;
    gap: 0.1rem;
  }

  .process-permission-row__copy strong {
    font-size: 0.85rem;
  }

  .process-permission-row__copy span {
    font-size: 0.7rem;
    color: var(--ink-soft);
  }

  .process-permission-row__toggle {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    font-size: 0.7rem;
    font-weight: 700;
    color: var(--ink-soft);
  }

  .process-permission-row__toggle input {
    width: 1rem;
    height: 1rem;
    cursor: pointer;
  }

  .entity-nav {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  @media (max-width: 1100px) {
    .roles-shell {
      grid-template-columns: 1fr;
    }
    .roles-rail {
      position: static;
    }
  }
</style>
