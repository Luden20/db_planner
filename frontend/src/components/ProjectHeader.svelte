<script lang="ts">
  export let name: string;
  export let entityCount: number;
</script>

<section class="project-header">
  <div class="project-copy">
    <p class="label">Proyecto activo</p>
    <div class="title-row">
      <h2>{name}</h2>
      <span class="project-tag">Modelo en edicion</span>
    </div>
  </div>
  <div class="project-meta">
    <p class="count">{entityCount}</p>
    <p class="muted">Entidades</p>
  </div>
</section>

<style>
  .project-header {
    margin-top: 1.6rem;
    padding: 1.2rem 1.25rem;
    border-radius: var(--radius-md);
    border: 1px solid var(--border);
    background: linear-gradient(180deg, rgba(255, 252, 247, 0.78), rgba(247, 241, 232, 0.82));
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    box-shadow: var(--shadow-sm);
  }

  .project-header h2 {
    margin: 0;
    font-size: clamp(1.9rem, 3.2vw, 2.8rem);
    line-height: 0.96;
  }

  .label {
    margin: 0;
    color: var(--accent);
    font-size: 0.75rem;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    font-weight: 800;
  }

  .project-copy {
    min-width: 0;
  }

  .title-row {
    display: flex;
    align-items: center;
    gap: 0.8rem;
    flex-wrap: wrap;
    margin-top: 0.35rem;
  }

  .project-tag {
    display: inline-flex;
    align-items: center;
    min-height: 2rem;
    padding: 0 0.85rem;
    border-radius: 999px;
    background: var(--accent-soft);
    border: 1px solid rgba(30, 95, 102, 0.18);
    color: var(--accent-strong);
    font-size: 0.84rem;
    font-weight: 800;
  }

  .project-meta {
    text-align: right;
    min-width: 8rem;
  }

  .project-meta .count {
    margin: 0;
    font-size: clamp(2.2rem, 4vw, 3.2rem);
    font-weight: 800;
    color: var(--ink);
    line-height: 1;
  }

  .project-meta .muted {
    margin: 0.35rem 0 0;
    color: var(--ink-faint);
    font-size: 0.84rem;
    text-transform: uppercase;
    letter-spacing: 0.12em;
  }

  @media (max-width: 720px) {
    .project-header {
      align-items: flex-start;
      flex-direction: column;
    }

    .project-meta {
      text-align: left;
    }
  }
</style>
