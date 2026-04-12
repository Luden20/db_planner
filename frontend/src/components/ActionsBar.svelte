<script lang="ts">
  import ButtonIcon from "./ButtonIcon.svelte";

  let { 
    onSave = async () => {}, 
    onExport = () => {}, 
    onExportWithoutRelations = () => {}, 
    onExit = () => {} 
  } = $props<{
    onSave?: () => Promise<void>;
    onExport?: () => void;
    onExportWithoutRelations?: () => void;
    onExit?: () => void;
  }>();
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
    <button class="btn primary" onclick={onSave}>
      <ButtonIcon name="save"/>
      <span>Guardar</span>
    </button>
    <div class="export-buttons">
      <button class="btn secondary" onclick={onExport}>
        <ButtonIcon name="download"/>
        <span>Exportar con relaciones</span>
      </button>
      <button class="btn secondary" onclick={onExportWithoutRelations}>
        <ButtonIcon name="download"/>
        <span>Exportar sin relaciones</span>
      </button>
    </div>
    <button class="btn danger" onclick={onExit}>
      <ButtonIcon name="exit"/>
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
