<script lang="ts">
  import {utils} from "../../wailsjs/go/models";
  import CreateEntity from "./forms/CreateEntity.svelte";
  import DeleteEntity from "./forms/DeleteEntity.svelte";

  export let onSave: () => Promise<void> = async () => {};
  export let entities: utils.Entity[] = [];
</script>

<div class="tab-toolbar">
  <div>
    <p class="label">Entidades</p>
    <p class="muted">Vista general del proyecto. Agrega nuevas entidades desde aquí.</p>
  </div>
  <CreateEntity onSave={onSave}/>
</div>

<div class="table-wrapper">
  <table class="entities-table">
    <thead>
    <tr>
      <th style="width: 80px;">ID</th>
      <th>Nombre</th>
      <th>Descripción</th>
      <th style="width: 140px;">Acciones</th>
    </tr>
    </thead>

    <tbody>
    {#each entities as entity}
      <tr>
        <td>{entity.Id}</td>
        <td>{entity.Name}</td>
        <td>{entity.Description}</td>
        <td>
          <div class="row-actions">
            <CreateEntity onSave={onSave} id={entity.Id}/>
            <DeleteEntity onSave={onSave} id={entity.Id}/>
          </div>
        </td>
      </tr>
    {/each}
    </tbody>
  </table>
</div>

<style>
  .tab-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
    margin-bottom: 14px;
  }

  .label {
    margin: 0;
    color: #9ab5e4;
    font-size: 12px;
    letter-spacing: 0.6px;
    text-transform: uppercase;
  }

  .muted {
    margin: 6px 0 0;
    color: #cfd9e9;
    opacity: 0.75;
  }

  .table-wrapper {
    overflow: auto;
  }

  .entities-table {
    width: 100%;
    border-collapse: collapse;
    color: #e8edf7;
  }

  .entities-table th,
  .entities-table td {
    text-align: left;
    padding: 12px 10px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.07);
    font-size: 14px;
  }

  .entities-table thead th {
    font-size: 13px;
    color: #9ab5e4;
    letter-spacing: 0.3px;
    text-transform: uppercase;
  }

  .entities-table tbody tr:hover {
    background: rgba(255, 255, 255, 0.03);
  }

  .row-actions {
    display: inline-flex;
    gap: 8px;
    align-items: center;
  }
</style>
