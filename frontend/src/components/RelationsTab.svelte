<script lang="ts">
  import {onMount} from "svelte";
  import {utils} from "../../wailsjs/go/models";
  import {AddRelation, GetCombinatory, RemoveRelation} from "../../wailsjs/go/main/App";

  export let onRefresh: () => Promise<void> = async () => {};
  let comb: utils.RelationView[] = [];
  let activeIndex = 0;
  const relationOptions = ["", "N:1", "1:N", "N:N"];
  let updating = false;
  let relationOverrides: Record<string, string> = {};

  const applyOverrides = () => {
    comb = comb.map(view => ({
      ...view,
      Relations: view.Relations.map(rel => {
        const key = `${view.IdPrincipalEntity}-${rel.IdEntity2}`;
        const override = relationOverrides[key];
        return override !== undefined ? {...rel, Relation: override} : rel;
      })
    }));
  };

  async function load(keepIndex = false) {
    const prevIndex = activeIndex;
    const data = await GetCombinatory();
    comb = (data || []).map(view => ({
      ...view,
      Relations: (view.Relations || []).map(rel => ({...rel, Relation: rel.Relation ?? ""}))
    }));
    applyOverrides();
    activeIndex = keepIndex
      ? Math.min(prevIndex, Math.max(comb.length - 1, 0))
      : 0;
  }

  onMount(load);

  $: if (activeIndex >= comb.length) {
    activeIndex = comb.length ? comb.length - 1 : 0;
  }

  const nextSlide = () => {
    if (!comb.length) return;
    activeIndex = (activeIndex + 1) % comb.length;
  };

  const prevSlide = () => {
    if (!comb.length) return;
    activeIndex = activeIndex === 0 ? comb.length - 1 : activeIndex - 1;
  };

  const relationIdentifiers = (relation: utils.RelationViewItem) => ({
    principalId: comb[activeIndex]?.IdPrincipalEntity ?? null,
    // Usamos siempre el Id si existe, incluso cuando Relation está vacío
    relationId: relation?.Id ?? null,
    targetId: relation?.IdEntity2 ?? null
  });

  const handleRelationChange = async (relation: utils.RelationViewItem) => {
    const ids = relationIdentifiers(relation);
    const selection = relation.Relation || "";
    if (ids.principalId == null || ids.targetId == null) {
      alert("Faltan IDs de entidad para registrar la relación.");
      return;
    }

    const key = `${ids.principalId}-${ids.targetId}`;
    relationOverrides[key] = selection; // Guardamos incluso vacío para respetar la selección
    applyOverrides(); // Optimista: refleja en UI mientras persiste

    updating = true;
    try {
      if (!selection) {
        // Borrar relación existente solo si hay ID persistido
        if (ids.relationId != null) {
          alert("Borrando relacion "+ids.relationId);
          await RemoveRelation(ids.relationId);
        }
      } else {
        await AddRelation(ids.principalId, ids.targetId, selection);
      }
      await load(true);
      await onRefresh();
      applyOverrides();
    } catch (err) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      alert(`No se pudo actualizar la relación: ${message}`);
    } finally {
      updating = false;
    }
  };
</script>

<div class="tab-toolbar">
  <div>
    <p class="label">Relaciones</p>
    <p class="muted">Explora relaciones por entidad principal. Cada slide se recorre con las flechas.</p>
  </div>
  <div class="toolbar-actions">
    <div class="action-slot">Acciones</div>
  </div>
</div>

{#if comb.length === 0}
  <div class="empty-panel">Sin datos de relaciones.</div>
{:else}
  <div class="slide-shell">
    <button class="nav-btn" on:click={prevSlide} aria-label="Anterior" disabled={comb.length <= 1}>&lt;</button>

    <article class="slide">
      {#if comb[activeIndex]}
        <header class="slide-head">
          <div>
            <p class="label">Entidad principal</p>
            <h3>{comb[activeIndex].PrincipalEntity}</h3>
          </div>
          <div class="head-meta">
            <p class="mini-label">ID</p>
            <p class="id-pill">{comb[activeIndex].IdPrincipalEntity}</p>
          </div>
        </header>


        <div class="table-wrapper">
          <table class="entities-table compact">
            <thead>
            <tr>
              <th>Entidad destino</th>
              <th>Relación</th>
            </tr>
            </thead>
            <tbody>
            {#if comb[activeIndex].Relations && comb[activeIndex].Relations.length > 0}
              {#each comb[activeIndex].Relations as relation}
                <tr>
                  <td>{relation.Entity2}</td>
                  <td>
                    <select
                      class="relation-select"
                      bind:value={relation.Relation}
                      data-principal={relationIdentifiers(relation).principalId}
                      data-relation-id={relationIdentifiers(relation).relationId}
                      data-target={relationIdentifiers(relation).targetId}
                      on:change={() => handleRelationChange(relation)}
                      disabled={updating}
                    >
                      {#each relationOptions as option}
                        <option value={option}>{option || "Sin valor"}</option>
                      {/each}
                    </select>
                  </td>

                </tr>
              {/each}
            {:else}
              <tr class="muted-row">
                <td colspan="5">Sin relaciones para esta entidad.</td>
              </tr>
            {/if}
            </tbody>
          </table>
        </div>
      {/if}
    </article>

    <button class="nav-btn" on:click={nextSlide} aria-label="Siguiente" disabled={comb.length <= 1}>&gt;</button>
  </div>

  <div class="slide-meta">
    <span class="counter">{activeIndex + 1} / {comb.length}</span>
  </div>
{/if}

<style>
  .tab-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
    margin-bottom: 14px;
  }

  .toolbar-actions {
    display: flex;
    gap: 10px;
    align-items: center;
  }

  .action-slot {
    border: 1px dashed rgba(255, 255, 255, 0.18);
    border-radius: 8px;
    min-width: 80px;
    min-height: 32px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: #cfd9e9;
    opacity: 0.8;
    font-size: 12px;
    padding: 6px 10px;
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

  .slide-shell {
    display: grid;
    grid-template-columns: 52px 1fr 52px;
    align-items: center;
    gap: 12px;
  }

  .nav-btn {
    border: 1px solid rgba(255, 255, 255, 0.14);
    background: rgba(255, 255, 255, 0.08);
    color: #d9e4f5;
    border-radius: 12px;
    height: 48px;
    width: 52px;
    cursor: pointer;
    transition: background 150ms ease, transform 120ms ease;
  }

  .nav-btn:hover:enabled {
    background: rgba(255, 255, 255, 0.12);
    transform: translateY(-1px);
  }

  .nav-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .slide {
    background: rgba(255, 255, 255, 0.02);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 14px;
    padding: 14px 14px 16px;
    box-shadow: none;
    display: flex;
    flex-direction: column;
    gap: 12px;
    min-height: 220px;
  }

  .slide-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
  }

  .slide-head h3 {
    margin: 4px 0 0;
    font-size: 18px;
  }

  .head-meta {
    text-align: right;
  }

  .mini-label {
    margin: 0;
    color: #9ab5e4;
    font-size: 11px;
    letter-spacing: 0.6px;
    text-transform: uppercase;
  }

  .id-pill {
    margin: 4px 0 0;
    padding: 6px 10px;
    background: rgba(109, 216, 255, 0.12);
    border: 1px solid rgba(109, 216, 255, 0.35);
    color: #cfeeff;
    border-radius: 10px;
    font-weight: 700;
    font-size: 13px;
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
    padding: 10px 10px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.07);
    font-size: 13px;
  }

  .entities-table thead th {
    font-size: 12px;
    color: #9ab5e4;
    letter-spacing: 0.3px;
    text-transform: uppercase;
  }

  .entities-table tbody tr:hover {
    background: rgba(255, 255, 255, 0.03);
  }

  .entities-table.compact td {
    font-size: 13px;
  }

  .muted-row td {
    color: #cfd9e9;
    opacity: 0.8;
    text-align: center;
  }

  .slide-meta {
    margin-top: 8px;
    text-align: center;
    color: #cfd9e9;
    opacity: 0.8;
    font-size: 13px;
    letter-spacing: 0.6px;
  }

  .counter {
    display: inline-block;
    padding: 6px 10px;
    border: 1px solid rgba(255, 255, 255, 0.14);
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.04);
  }

  .relation-select {
    width: 100%;
    min-width: 120px;
    border-radius: 10px;
    border: 1px solid rgba(255, 255, 255, 0.14);
    background: #0f1726;
    color: #e8edf7;
    padding: 8px 10px;
    font-size: 13px;
    outline: none;
    transition: border 140ms ease, box-shadow 140ms ease;
    appearance: none;
  }

  .relation-select:focus {
    border-color: rgba(90, 209, 255, 0.8);
    box-shadow: 0 0 0 2px rgba(90, 209, 255, 0.22);
  }

  @media (max-width: 720px) {
    .slide-shell {
      grid-template-columns: 40px 1fr 40px;
    }

    .nav-btn {
      width: 40px;
      height: 44px;
    }
  }
</style>
