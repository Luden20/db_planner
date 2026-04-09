<script lang="ts">
  export let kicker = "Entidad";
  export let name = "";
  export let description = "";
  export let approved = false;
  export let pendingLabel = "Pendiente";
  export let transitionName = "";
</script>

<article
  class:entity-focus-card={true}
  class:entity-focus-card--approved={approved}
  style={transitionName ? `view-transition-name: ${transitionName};` : undefined}
>
  <div class="entity-focus-card__copy">
    <p class="entity-focus-card__kicker">{kicker}</p>
    <div class="entity-focus-card__title-row">
      <h3>{name}</h3>
      {#if approved}
        <span class="entity-focus-card__status entity-focus-card__status--approved">Aprobada</span>
      {:else}
        <span class="entity-focus-card__status">{pendingLabel}</span>
      {/if}
    </div>
    <p class="entity-focus-card__description">{description || "Sin definición."}</p>
  </div>

  <div class="entity-focus-card__actions">
    <slot name="actions" />
  </div>
</article>

<style>
  .entity-focus-card {
    position: relative;
    display: grid;
    gap: 0.95rem;
    padding: 1rem 1.05rem;
    border-radius: calc(var(--radius-md) - 4px);
    border: 1px solid var(--border);
    background:
      linear-gradient(180deg, color-mix(in srgb, var(--surface-strong) 98%, var(--surface)), color-mix(in srgb, var(--surface) 96%, var(--surface-strong))),
      linear-gradient(90deg, color-mix(in srgb, var(--accent) 8%, var(--surface-strong)), transparent 42%);
    box-shadow: var(--shadow-sm);
    overflow: clip;
  }

  .entity-focus-card::before {
    content: "";
    position: absolute;
    inset: 0 auto auto 0;
    width: min(220px, 48%);
    height: 1px;
    background: linear-gradient(90deg, color-mix(in srgb, var(--accent) 34%, transparent), transparent);
    pointer-events: none;
  }

  .entity-focus-card--approved {
    border-color: color-mix(in srgb, var(--success) 24%, var(--border));
    background:
      linear-gradient(180deg, color-mix(in srgb, var(--surface-strong) 98%, var(--surface)), color-mix(in srgb, var(--surface) 96%, var(--surface-strong))),
      linear-gradient(90deg, color-mix(in srgb, var(--success) 10%, var(--surface-strong)), transparent 42%);
  }

  .entity-focus-card__copy {
    display: grid;
    gap: 0.5rem;
    min-width: 0;
  }

  .entity-focus-card__kicker {
    margin: 0;
    color: var(--accent);
    font-size: 0.74rem;
    letter-spacing: 0.16em;
    font-weight: 800;
    text-transform: uppercase;
  }

  .entity-focus-card__title-row {
    display: flex;
    align-items: center;
    gap: 0.65rem;
    flex-wrap: wrap;
  }

  .entity-focus-card__title-row h3 {
    margin: 0;
    color: var(--ink);
    font-size: 1.22rem;
    line-height: 1.08;
  }

  .entity-focus-card__description {
    margin: 0;
    color: var(--ink-soft);
    line-height: 1.55;
  }

  .entity-focus-card__status {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-height: 1.85rem;
    padding: 0.2rem 0.72rem;
    border-radius: 999px;
    border: 1px solid var(--line-soft);
    background: var(--chip-surface);
    color: var(--ink-soft);
    font-size: 0.74rem;
    font-weight: 800;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .entity-focus-card__status--approved {
    border-color: color-mix(in srgb, var(--success) 24%, var(--border));
    background: var(--chip-success-surface);
    color: var(--success);
  }

  .entity-focus-card__actions {
    display: flex;
    flex-wrap: nowrap;
    gap: 0.55rem;
    align-items: center;
    justify-content: flex-start;
    min-width: 0;
  }

  .entity-focus-card__actions :global(.toolbar-actions) {
    display: contents;
  }

  .entity-focus-card__actions :global(.entity-focus-actions) {
    display: inline-flex;
    align-items: center;
    gap: 0.55rem;
    flex-wrap: nowrap;
    min-width: 0;
  }

  .entity-focus-card__actions :global(.entity-focus-actions .btn),
  .entity-focus-card__actions :global(.entity-focus-actions .control) {
    min-height: 2.45rem;
    padding: 0.55rem 0.82rem;
    border-radius: 0.95rem;
    font-size: 0.88rem;
    gap: 0.42rem;
    white-space: nowrap;
  }

  @media (max-width: 720px) {
    .entity-focus-card__actions {
      flex-wrap: wrap;
      align-items: stretch;
    }

    .entity-focus-card__actions :global(.entity-focus-actions) {
      width: 100%;
      flex-wrap: wrap;
    }

    .entity-focus-card__actions :global(.btn) {
      width: 100%;
      justify-content: center;
    }

    .entity-focus-card__actions :global(.entity-focus-actions .control) {
      width: 100%;
      justify-content: center;
    }
  }
</style>
