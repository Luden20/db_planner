<script lang="ts">
  import { Dialog } from "bits-ui";
  import ModalLauncher from "../ModalLauncher.svelte";
  import ButtonIcon from "../ButtonIcon.svelte";
  import { ExportEntitiesToJSON } from "../../../wailsjs/go/main/App";
  import type { utils } from "../../../wailsjs/go/models";
  import { showToast } from "../../lib/toast";

  let { 
    entities = [], 
    intersectionEntities = [] 
  } = $props<{
    entities?: utils.Entity[];
    intersectionEntities?: utils.IntersectionEntity[];
  }>();
  
  let selectedIds = $state<Set<number>>(new Set());
  let selectedIntersectionIds = $state<Set<number>>(new Set());
  let jsonResult = $state("");
  let step = $state<"select" | "result">("select");

  const toggleAll = () => {
    const allSelected = selectedIds.size === entities.length && selectedIntersectionIds.size === intersectionEntities.length;
    if (allSelected) {
      selectedIds = new Set();
      selectedIntersectionIds = new Set();
    } else {
      selectedIds = new Set(entities.map(e => e.Id));
      selectedIntersectionIds = new Set(intersectionEntities.map(e => e.Entity.Id));
    }
  };

  const toggleEntity = (id: number) => {
    if (selectedIds.has(id)) selectedIds.delete(id);
    else selectedIds.add(id);
    selectedIds = new Set(selectedIds);
  };

  const toggleIntersection = (id: number) => {
    if (selectedIntersectionIds.has(id)) selectedIntersectionIds.delete(id);
    else selectedIntersectionIds.add(id);
    selectedIntersectionIds = new Set(selectedIntersectionIds);
  };

  const handleExport = async () => {
    if (selectedIds.size === 0 && selectedIntersectionIds.size === 0) {
      showToast("Selecciona al menos una entidad para exportar", "warning");
      return;
    }

    try {
      jsonResult = await ExportEntitiesToJSON(Array.from(selectedIds), Array.from(selectedIntersectionIds));
      step = "result";
    } catch (err) {
      showToast("Error al exportar: " + err, "error");
    }
  };

  const copyToClipboard = () => {
    navigator.clipboard.writeText(jsonResult);
    showToast("Copiado al portapapeles", "success");
  };

  const downloadJSON = () => {
    const blob = new Blob([jsonResult], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url; a.download = 'export_db_planner.json';
    document.body.appendChild(a); a.click(); document.body.removeChild(a);
    URL.revokeObjectURL(url);
    showToast("Descarga iniciada", "success");
  };

  const reset = () => { step = "select"; jsonResult = ""; };
</script>

{#snippet trigger()}
  <Dialog.Trigger class="control control--sm control--ghost">
    <ButtonIcon name="plus"/>
    <span>Exportar JSON (IA)</span>
  </Dialog.Trigger>
{/snippet}

<ModalLauncher {trigger} onOpen={reset}>
  <div class="export-json-modal">
    <div class="modal-header">
      <h2 class="title">{step === 'select' ? 'Seleccionar tablas para exportar' : 'JSON Generado'}</h2>
      <p class="muted">
        {step === 'select' 
          ? 'Elige las entidades que quieres incluir en el JSON para la IA.' 
          : 'Copia este JSON y pégalo en tu IA favorita para generar SQL.'}
      </p>
    </div>

    {#if step === 'select'}
      <div class="export-selection">
        <div class="selection-actions">
          <button class="control control--xs control--ghost" onclick={toggleAll}>
            {selectedIds.size === entities.length && selectedIntersectionIds.size === intersectionEntities.length ? 'Desmarcar todo' : 'Marcar todo'}
          </button>
          <span class="selection-count">{selectedIds.size + selectedIntersectionIds.size} seleccionadas</span>
        </div>
        <div class="entities-list">
          {#if entities.length > 0}
            <div class="list-section-header">Tablas Fuertes</div>
            {#each entities as entity}
              <label class="entity-item">
                <input 
                  type="checkbox" 
                  checked={selectedIds.has(entity.Id)} 
                  onchange={() => toggleEntity(entity.Id)}
                />
                <span class="entity-name">{entity.Name}</span>
                {#if entity.Description}
                  <span class="entity-desc" title={entity.Description}>- {entity.Description}</span>
                {/if}
              </label>
            {/each}
          {/if}

          {#if intersectionEntities.length > 0}
            <div class="list-section-header">Tablas de Intersección</div>
            {#each intersectionEntities as item}
              <label class="entity-item">
                <input 
                  type="checkbox" 
                  checked={selectedIntersectionIds.has(item.Entity.Id)} 
                  onchange={() => toggleIntersection(item.Entity.Id)}
                />
                <span class="entity-name">{item.Entity.Name}</span>
                {#if item.Entity.Description}
                  <span class="entity-desc" title={item.Entity.Description}>- {item.Entity.Description}</span>
                {/if}
              </label>
            {/each}
          {/if}
        </div>
      </div>
      <div class="modal-actions">
        <button class="btn btn--primary" onclick={handleExport} disabled={selectedIds.size === 0 && selectedIntersectionIds.size === 0}>
          Generar JSON
        </button>
      </div>
    {:else}
      <div class="export-result">
        <pre class="json-preview"><code>{jsonResult}</code></pre>
      </div>
      <div class="modal-actions">
        <button class="btn btn--secondary" onclick={() => step = 'select'}>
          Volver
        </button>
        <div class="button-group">
          <button class="btn btn--secondary" onclick={copyToClipboard}>
            Copiar
          </button>
          <button class="btn btn--primary" onclick={downloadJSON}>
            Descargar .json
          </button>
        </div>
      </div>
    {/if}
  </div>
</ModalLauncher>

<style>
  .export-json-modal {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    min-width: 500px;
    max-width: 800px;
    max-height: 80vh;
  }

  .modal-header {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .title {
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--text);
    margin: 0;
  }

  .muted {
    font-size: 0.875rem;
    color: var(--text-muted);
    margin: 0;
  }

  .export-selection {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--surface-low);
    overflow: hidden;
  }

  .selection-actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem 1rem;
    background: var(--surface-strong);
    border-bottom: 1px solid var(--border);
  }

  .selection-count {
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--text-muted);
    text-transform: uppercase;
  }

  .list-section-header {
    padding: 0.75rem 0.75rem 0.25rem;
    font-size: 0.7rem;
    font-weight: 700;
    color: var(--primary);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .entities-list {
    display: flex;
    flex-direction: column;
    max-height: 400px;
    overflow-y: auto;
    padding: 0.5rem;
  }

  .entity-item {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.5rem 0.75rem;
    border-radius: var(--radius-sm);
    cursor: pointer;
    transition: background 0.2s;
  }

  .entity-item:hover {
    background: var(--surface-medium);
  }

  .entity-name {
    font-weight: 500;
    font-size: 0.9375rem;
  }

  .entity-desc {
    font-size: 0.8125rem;
    color: var(--text-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .export-result {
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: #1e1e1e; /* Dark theme for code */
    padding: 1rem;
    overflow: hidden;
  }

  .json-preview {
    margin: 0;
    max-height: 400px;
    overflow: auto;
    font-family: 'JetBrains Mono', 'Fira Code', monospace;
    font-size: 0.8125rem;
    color: #d4d4d4;
    line-height: 1.5;
  }

  .modal-actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    margin-top: 0.5rem;
  }

  .button-group {
    display: flex;
    gap: 0.75rem;
  }

  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 0.5rem 1rem;
    font-size: 0.875rem;
    font-weight: 600;
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: all 0.2s;
    border: 1px solid transparent;
  }

  .btn--primary {
    background: var(--primary);
    color: white;
  }

  .btn--primary:hover:not(:disabled) {
    background: var(--primary-hover);
  }

  .btn--primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn--secondary {
    background: var(--surface-medium);
    color: var(--text);
    border-color: var(--border);
  }

  .btn--secondary:hover {
    background: var(--surface-strong);
    border-color: var(--text-muted);
  }
</style>
