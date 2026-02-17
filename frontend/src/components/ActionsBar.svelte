<script lang="ts">
  export let onSave: () => void = () => {};
  export let onExport: () => void = () => {};
</script>

<section class="actions-bar">
  <div>
    <p class="kicker">Proyecto cargado</p>
    <h1 class="bar-title">Acciones rápidas</h1>
    <p class="hint">Guarda o exporta antes de seguir editando.</p>
  </div>
  <div class="action-buttons">
    <button class="btn secondary" on:click={onSave}>Guardar</button>
    <button class="btn secondary" on:click={onExport}>Exportar</button>
  </div>
 </section>

<style>
  .actions-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 20px;
    padding: 16px 18px;
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 14px;
    background: rgba(255, 255, 255, 0.03);
    box-shadow: 0 14px 38px rgba(0,0,0,0.24);
  }

  .action-buttons {
    display: flex;
    gap: 10px;
    align-items: center;
  }

  .bar-title {
    margin: 4px 0 6px;
    font-size: 22px;
    letter-spacing: 0.3px;
  }

  .kicker {
    margin: 0 0 4px;
    color: #9ab5e4;
    text-transform: uppercase;
    font-size: 12px;
    letter-spacing: 1px;
  }

  .hint {
    margin: 4px 0 0;
    color: #cfd9e9;
    opacity: 0.8;
  }

  @media (max-width: 720px) {
    .actions-bar {
      flex-direction: column;
      align-items: flex-start;
    }
  }
</style>
