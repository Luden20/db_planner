<script lang="ts">
  export let name: string;
  export let entityCount: number;
</script>

<section class="project-header">
  <div>
    <p class="label">Proyecto</p>
    <h2>{name}</h2>
  </div>
  <div class="project-meta">
    <p class="count">{entityCount}</p>
    <p class="muted">Entidades</p>
  </div>
</section>

<style>
  .project-header {
    margin-top: 24px;
    padding: 16px 18px;
    border-radius: 12px;
    border: 1px solid rgba(255, 255, 255, 0.08);
    background: rgba(255, 255, 255, 0.02);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .project-header h2 {
    margin: 4px 0 0;
    font-size: 22px;
  }

  .label {
    margin: 0;
    color: #9ab5e4;
    font-size: 12px;
    letter-spacing: 0.6px;
    text-transform: uppercase;
  }

  .project-meta {
    text-align: right;
  }

  .project-meta .count {
    margin: 0;
    font-size: 28px;
    font-weight: 800;
    color: #6dd8ff;
    line-height: 1;
  }

  .project-meta .muted {
    margin: 4px 0 0;
    color: #cfd9e9;
    opacity: 0.7;
    font-size: 13px;
  }
</style>
