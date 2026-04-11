<script lang="ts">
  import type { utils } from "../../../wailsjs/go/models";
  import ButtonIcon from "../ButtonIcon.svelte";
  import ModalLauncher from "../ModalLauncher.svelte";
  import { AnalyzeDependencies } from "../../../wailsjs/go/main/App.js";

  let report: utils.AnalysisReport | null = null;
  let loading = false;
  let modal: ModalLauncher;

  export const openDialog = async () => {
    loading = true;
    modal.openDialog();
    try {
      const res = await AnalyzeDependencies();
      report = res;
    } catch (err) {
      console.error("Error al analizar dependencias:", err);
    } finally {
      loading = false;
    }
  };

  const getCircular = (issues: utils.AnalysisIssue[]) => issues.filter(i => i.type === "Circular");
  const getAmbiguity = (issues: utils.AnalysisIssue[]) => issues.filter(i => i.type === "Ambiguity");

  let showEdges = false;
  let expandedIssues = new Set<string>();

  const toggleExpand = (u: string, v: string) => {
    const key = `${u}-${v}`;
    if (expandedIssues.has(key)) {
      expandedIssues.delete(key);
    } else {
      expandedIssues.add(key);
    }
    expandedIssues = expandedIssues;
  };
</script>

<ModalLauncher
  title="Análisis de Estructura Relacional"
  confirmLabel="Cerrar"
  confirmVariant="primary"
  showTrigger={false}
  size="form"
  bind:this={modal}
>
  {#if loading}
    <div class="loading-state">
      <div class="spinner"></div>
      <p>Construyendo grafo estructural (Padre → Hijo)...</p>
    </div>
  {:else if report}
    <div class="analysis-results">
      <div class="results-header">
        <div class="summary-stats">
          <div class="stat-card ambiguity">
            <span class="stat-value">{getAmbiguity(report.issues || []).length}</span>
            <span class="stat-label">Ambigüedades</span>
          </div>
          <div class="stat-card circular">
            <span class="stat-value">{getCircular(report.issues || []).length}</span>
            <span class="stat-label">Ciclos</span>
          </div>
          <div class="stat-card edges">
            <span class="stat-value">{report.edgeList?.length || 0}</span>
            <span class="stat-label">Relaciones</span>
          </div>
        </div>

        <section class="edge-verification">
          <button class="edge-toggle" on:click={() => showEdges = !showEdges}>
            <ButtonIcon name={showEdges ? "chevron-down" : "chevron-right"} />
            <span>Ver Aristas Interpretadas</span>
          </button>
          {#if showEdges}
            <div class="edge-list-container">
              <p class="group-desc">Sentido jerárquico detectado: Padre → Hijo.</p>
              <div class="edge-grid">
                {#each report.edgeList || [] as edge}
                  <div class="edge-item">{edge}</div>
                {/each}
              </div>
            </div>
          {/if}
        </section>
      </div>

      {#if (report.issues?.length || 0) === 0}
        <div class="success-state">
          <div class="success-icon">
            <ButtonIcon name="check" />
          </div>
          <h3>¡Estructura Correcta!</h3>
          <p>No se detectaron ambigüedades ni dependencias circulares en el flujo actual.</p>
        </div>
      {:else}
        {#if getAmbiguity(report.issues).length > 0}
          <section class="issue-group">
            <h4 class="group-title path">
              <ButtonIcon name="relations" />
              Ambigüedad Estructural
            </h4>
            <div class="issue-list">
              {#each getAmbiguity(report.issues) as issue}
                {@const isExpanded = expandedIssues.has(`${issue.entities[0]}-${issue.entities[1]}`)}
                <div class="issue-card path-card" class:expanded={isExpanded}>
                  <button class="card-header-btn" on:click={() => toggleExpand(issue.entities[0], issue.entities[1])}>
                    <div class="header-main">
                      <span class="entity-name">{issue.entities[0]}</span>
                      <span class="arrow-icon">→</span>
                      <span class="entity-name">{issue.entities[1]}</span>
                    </div>
                    <div class="header-meta">
                      <span class="path-count-badge">{issue.pathCount} rutas</span>
                      <ButtonIcon name={isExpanded ? "chevron-up" : "chevron-down"} />
                    </div>
                  </button>
                  
                  {#if isExpanded}
                    <div class="card-expanded-content">
                      <div class="paths-container">
                        <p class="paths-label">Rutas Dirigidas:</p>
                        <div class="paths-list">
                          {#each issue.paths.slice(0, 10) as path, i}
                            <div class="path-item">
                              <span class="path-index">{i + 1}.</span>
                              <span class="path-text">{path.join(" → ")}</span>
                            </div>
                          {/each}
                          {#if issue.paths.length > 10}
                            <div class="path-more">... y {issue.paths.length - 10} caminos más</div>
                          {/if}
                        </div>
                      </div>
                      <div class="status-tag">Ambigüedad estructural</div>
                    </div>
                  {/if}
                </div>
              {/each}
            </div>
          </section>
        {/if}

        {#if getCircular(report.issues).length > 0}
          <section class="issue-group">
            <h4 class="group-title cycle">
              <ButtonIcon name="spark" />
              Dependencia Circular
            </h4>
            <div class="issue-list">
              {#each getCircular(report.issues) as issue}
                <div class="issue-card cycle-card">
                  <p class="issue-path">{issue.entities.join(" → ")} → {issue.entities[0]}</p>
                  <p class="issue-desc">Ciclo de llaves foráneas detectado.</p>
                </div>
              {/each}
            </div>
          </section>
        {/if}
      {/if}
    </div>
  {:else}
    <div class="empty-state">
      <p>No hay datos de análisis disponibles.</p>
    </div>
  {/if}
</ModalLauncher>

<style>
  .edge-verification {
    margin-bottom: 2rem;
    background: var(--surface-strong);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  .edge-toggle {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.85rem 1.25rem;
    border: none;
    background: none;
    cursor: pointer;
    font-weight: 700;
    color: var(--ink);
    font-size: 0.95rem;
    text-align: left;
    transition: background 150ms ease;
  }

  .edge-toggle:hover {
    background: var(--surface-hover);
  }

  .edge-list-container {
    padding: 0 1.25rem 1.25rem;
    border-top: 1px solid var(--border);
    background: var(--field-surface);
  }

  .edge-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 0.5rem;
    font-family: var(--font-mono);
    font-size: 0.8rem;
  }

  .edge-item {
    background: var(--surface-strong);
    padding: 0.4rem 0.6rem;
    border-radius: 0.3rem;
    border: 1px solid var(--border);
    color: var(--ink-soft);
  }

  .results-header {
    margin-bottom: 2rem;
  }

  .summary-stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .stat-card {
    padding: 1rem 1.25rem;
    border-radius: var(--radius-md);
    background: var(--surface-strong);
    border: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .stat-value {
    font-size: 1.75rem;
    font-weight: 800;
    line-height: 1;
    font-family: var(--font-display);
  }

  .stat-label {
    font-size: 0.75rem;
    font-weight: 700;
    text-transform: uppercase;
    color: var(--ink-faint);
    letter-spacing: 0.05em;
  }

  .stat-card.ambiguity .stat-value { color: var(--accent); }
  .stat-card.circular .stat-value { color: var(--danger); }
  .stat-card.edges .stat-value { color: var(--ink-soft); }

  .edge-verification {
    background: var(--surface-strong);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  .edge-toggle {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1.15rem;
    border: none;
    background: none;
    cursor: pointer;
    font-weight: 700;
    color: var(--ink-soft);
    font-size: 0.85rem;
    text-align: left;
    transition: background 150ms ease;
  }

  .edge-toggle:hover {
    background: var(--surface-hover);
  }

  .edge-list-container {
    padding: 0 1.15rem 1.15rem;
    border-top: 1px solid var(--border);
    background: var(--field-surface);
  }

  .edge-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: 0.4rem;
    font-family: var(--font-mono);
    font-size: 0.75rem;
  }

  .edge-item {
    background: var(--surface-strong);
    padding: 0.3rem 0.6rem;
    border-radius: 0.3rem;
    border: 1px solid var(--border);
    color: var(--ink-soft);
  }

  .card-header-btn {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    padding: 1.15rem 1.25rem;
    border: none;
    background: none;
    cursor: pointer;
    text-align: left;
    transition: background 150ms ease;
    border-radius: var(--radius-md);
  }

  .card-header-btn:hover {
    background: var(--surface-hover);
  }

  .header-main {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex: 1;
    min-width: 0;
  }

  .entity-name {
    font-weight: 700;
    font-size: 1rem;
    color: var(--ink);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .arrow-icon {
    color: var(--ink-faint);
    font-weight: 700;
  }

  .header-meta {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .path-count-badge {
    background: var(--accent);
    color: white;
    font-size: 0.75rem;
    font-weight: 800;
    padding: 0.2rem 0.6rem;
    border-radius: 1rem;
    white-space: nowrap;
  }

  .card-expanded-content {
    padding: 0 1.25rem 1.25rem;
    border-top: 1px dashed var(--border);
    margin-top: -1px;
    animation: slideDown 200ms ease-out;
  }

  @keyframes slideDown {
    from { opacity: 0; transform: translateY(-5px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .loading-state, .success-state {
    text-align: center;
    padding: 3rem 1rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1.5rem;
  }

  .success-icon {
    width: 3.5rem;
    height: 3.5rem;
    background: var(--success);
    color: white;
    border-radius: 50%;
    display: grid;
    place-items: center;
    font-size: 1.75rem;
    box-shadow: 0 0 20px color-mix(in srgb, var(--success) 30%, transparent);
  }

  .success-state h3 {
    margin: 0;
    font-size: 1.5rem;
    color: var(--ink);
  }

  .success-state p {
    margin: 0;
    color: var(--ink-soft);
  }

  .spinner {
    width: 2.5rem;
    height: 2.5rem;
    border: 3px solid var(--border);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .issue-group {
    margin-bottom: 2.5rem;
  }

  .group-title {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    font-size: 1.25rem;
    font-weight: 800;
    margin: 0 0 1rem;
    padding-bottom: 0.75rem;
    border-bottom: 1px solid var(--border);
  }

  .paths-container {
    margin: 1.25rem 0;
  }

  .paths-label {
    font-weight: 700;
    font-size: 0.8rem;
    color: var(--ink-soft);
    margin: 0 0 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .paths-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .path-item {
    font-family: var(--font-mono);
    font-size: 0.8rem;
    background: var(--surface-strong);
    padding: 0.6rem 0.85rem;
    border-radius: 0.4rem;
    border: 1px solid var(--border);
    display: flex;
    gap: 0.6rem;
    line-height: 1.5;
  }

  .path-index {
    color: var(--accent);
    font-weight: 800;
  }

  .path-text {
    word-break: break-all;
    color: var(--ink);
  }

  .path-more {
    font-size: 0.8rem;
    color: var(--ink-faint);
    font-style: italic;
    padding-top: 0.5rem;
  }

  .status-tag {
    display: inline-block;
    font-size: 0.7rem;
    font-weight: 800;
    text-transform: uppercase;
    color: var(--accent);
    background: color-mix(in srgb, var(--accent) 10%, transparent);
    padding: 0.2rem 0.6rem;
    border-radius: 1rem;
  }

  .group-desc {
    font-size: 0.85rem;
    color: var(--ink-faint);
    margin: 0 0 1rem;
  }

  .group-title.cycle { color: var(--danger); }
  .group-title.path { color: var(--accent); }

  .issue-list {
    display: grid;
    gap: 0.75rem;
  }

  .issue-card {
    border-radius: var(--radius-md);
    background: var(--field-surface);
    border: 1px solid var(--border);
    overflow: hidden;
    transition: box-shadow 150ms ease, border-color 150ms ease;
  }

  .issue-card:hover {
    border-color: var(--focus-border);
    box-shadow: var(--shadow-sm);
  }

  .issue-card.expanded {
    border-color: var(--focus-border);
    box-shadow: var(--shadow-md);
  }

  .cycle-card {
    border-left: 4px solid var(--danger);
    padding: 1.25rem;
  }

  .path-card {
    border-left: 4px solid var(--accent);
  }

  .issue-path {
    font-family: var(--font-mono);
    font-size: 1rem;
    font-weight: 700;
    margin: 0 0 0.5rem;
  }

  .issue-desc {
    font-size: 0.9rem;
    color: var(--ink-soft);
    margin: 0;
  }

  .entity-name {
    background: var(--surface-strong);
    padding: 0.2rem 0.5rem;
    border-radius: 0.4rem;
    border: 1px solid var(--border);
  }
</style>
