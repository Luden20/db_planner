<script lang="ts">
  export let onSave:  () => Promise<void> = async () => {
  };
  export let onExport: () => void = () => {};
  export let onExportWithoutRelations: () => void = () => {};
  export let onExit: () => void = () => {};
</script>

<section class="actions-bar">
  <div class="bar-copy">
    <p class="kicker">Proyecto cargado</p>
    <h1 class="bar-title">Sesion de trabajo activa</h1>
    <p class="hint">Guarda, exporta o cierra la sesion sin perder de vista el estado actual del modelo.</p>
    <div class="bar-status">
      <span class="status-led"></span>
      <span>Modo edicion local</span>
    </div>
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
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 1.25rem 1.5rem;
    padding: 1.25rem 1.35rem;
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: linear-gradient(135deg, rgba(255, 252, 247, 0.9), rgba(243, 236, 226, 0.8));
    box-shadow: var(--shadow-sm);
  }

  .action-buttons {
    display: flex;
    gap: 0.75rem;
    align-items: center;
    flex-wrap: wrap;
    justify-content: flex-end;
  }

  .export-buttons {
    display: flex;
    gap: 0.75rem;
    align-items: center;
    flex-wrap: nowrap;
  }

  .bar-copy {
    min-width: 0;
  }

  .bar-title {
    margin: 0.2rem 0 0.45rem;
    font-size: clamp(1.8rem, 3vw, 2.4rem);
    line-height: 0.95;
  }

  .kicker {
    margin: 0;
    color: var(--accent);
    text-transform: uppercase;
    font-size: 0.74rem;
    letter-spacing: 0.16em;
    font-weight: 800;
  }

  .hint {
    margin: 0;
    color: var(--ink-soft);
    line-height: 1.55;
  }

  .bar-status {
    display: inline-flex;
    align-items: center;
    gap: 0.55rem;
    margin-top: 0.9rem;
    color: var(--ink-faint);
    font-size: 0.9rem;
    font-weight: 700;
  }

  .status-led {
    width: 0.72rem;
    height: 0.72rem;
    border-radius: 999px;
    background: linear-gradient(135deg, #6b9777, #3d7251);
    box-shadow: 0 0 0 0.22rem rgba(61, 114, 81, 0.14);
  }

  @media (max-width: 720px) {
    .actions-bar {
      grid-template-columns: 1fr;
    }

    .export-buttons {
      width: 100%;
      flex-wrap: wrap;
    }

    .action-buttons {
      justify-content: stretch;
    }
  }
</style>
