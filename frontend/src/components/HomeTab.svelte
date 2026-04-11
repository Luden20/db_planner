<script lang="ts">
  import type { utils } from "../../wailsjs/go/models";
  import ErdCanvas from "./ErdCanvas.svelte";

  let { project, onRefresh } = $props<{
    project: utils.DbProject;
    onRefresh: () => Promise<void>;
  }>();
</script>

<section class="home-sandbox">
  <header class="home-header">
    <div class="home-copy">
      <h1>Home (ERD Sandbox)</h1>
      <p class="muted">Visualiza tu esquema y analiza la integridad de las dependencias.</p>
    </div>
  </header>

  <ErdCanvas {project} {onRefresh} />
</section>

<style>
  .home-sandbox {
    height: 100%;
    padding: 1rem;
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .home-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .home-copy h1 {
    margin: 0;
    font-size: 2.5rem;
    letter-spacing: -0.04em;
    line-height: 1;
  }

  .muted {
    margin: 0.5rem 0 0;
    color: var(--ink-soft);
    font-size: 1.1rem;
  }
</style>
