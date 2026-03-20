<script lang="ts">
  export let onSave:  () => Promise<void> = async () => {
  };
  export let onExport: () => void = () => {};
  export let onExportWithoutRelations: () => void = () => {};
  export let onExit: () => void = () => {};
</script>

<section class="actions-bar">
  <div>
    <p class="kicker">Proyecto cargado</p>
    <h1 class="bar-title">Acciones rápidas</h1>
    <p class="hint">Guarda o exporta antes de seguir editando.</p>
  </div>
  <div class="action-buttons">
    <button class="btn primary" on:click={onSave}>
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <path d="M5 4a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V8.414a2 2 0 0 0-.586-1.414l-3.414-3.414A2 2 0 0 0 15.586 3H5Zm10 2.5V8H5V6a1 1 0 0 1 1-1h8a1 1 0 0 1 1 1Zm-10 6h14V18a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1v-5.5Z"/>
      </svg>
      <span>Guardar</span>
    </button>
    <div class="export-buttons">
      <button class="btn secondary" on:click={onExport}>
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M12 3.5a.75.75 0 0 0-1.5 0V13a.75.75 0 0 0 1.5 0V3.5Zm-4.22 3.97a.75.75 0 0 0-1.06 1.06l4.5 4.5a.75.75 0 0 0 1.06 0l4.5-4.5a.75.75 0 0 0-1.06-1.06L12 10.94 7.78 7.47ZM4 15.75A2.75 2.75 0 0 1 6.75 13h10.5A2.75 2.75 0 0 1 20 15.75v2.5A2.75 2.75 0 0 1 17.25 21H6.75A2.75 2.75 0 0 1 4 18.25v-2.5Zm2.75-1.25a1.25 1.25 0 0 0-1.25 1.25v2.5c0 .69.56 1.25 1.25 1.25h10.5c.69 0 1.25-.56 1.25-1.25v-2.5c0-.69-.56-1.25-1.25-1.25H6.75Z"/>
        </svg>
        <span>Exportar con relaciones</span>
      </button>
      <button class="btn secondary" on:click={onExportWithoutRelations}>
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M12 3.5a.75.75 0 0 0-1.5 0V13a.75.75 0 0 0 1.5 0V3.5Zm-4.22 3.97a.75.75 0 0 0-1.06 1.06l4.5 4.5a.75.75 0 0 0 1.06 0l4.5-4.5a.75.75 0 0 0-1.06-1.06L12 10.94 7.78 7.47ZM4 15.75A2.75 2.75 0 0 1 6.75 13h10.5A2.75 2.75 0 0 1 20 15.75v2.5A2.75 2.75 0 0 1 17.25 21H6.75A2.75 2.75 0 0 1 4 18.25v-2.5Zm2.75-1.25a1.25 1.25 0 0 0-1.25 1.25v2.5c0 .69.56 1.25 1.25 1.25h10.5c.69 0 1.25-.56 1.25-1.25v-2.5c0-.69-.56-1.25-1.25-1.25H6.75Z"/>
        </svg>
        <span>Exportar sin relaciones</span>
      </button>
    </div>
    <button class="btn danger" on:click={onExit}>
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <path d="M5.25 3A2.25 2.25 0 0 0 3 5.25v13.5A2.25 2.25 0 0 0 5.25 21h8.5A2.25 2.25 0 0 0 16 18.75V15.5a.75.75 0 0 0-1.5 0v3.25c0 .414-.336.75-.75.75h-8a.75.75 0 0 1-.75-.75V5.25c0-.414.336-.75.75-.75h8c.414 0 .75.336.75.75V9.5a.75.75 0 0 0 1.5 0V5.25A2.25 2.25 0 0 0 13.75 3h-8.5Zm11.53 6.22a.75.75 0 0 0-1.06 1.06l1.97 1.97H11.5a.75.75 0 0 0 0 1.5h6.19l-1.97 1.97a.75.75 0 1 0 1.06 1.06l3.25-3.25a.75.75 0 0 0 0-1.06l-3.25-3.25Z"/>
      </svg>
      <span>Salir</span>
    </button>
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
    flex-wrap: wrap;
  }

  .export-buttons {
    display: flex;
    gap: 10px;
    align-items: center;
    flex-wrap: nowrap;
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

    .export-buttons {
      width: 100%;
      flex-wrap: wrap;
    }
  }
</style>
