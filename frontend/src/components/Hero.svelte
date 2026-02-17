<script lang="ts">
  export let onOpen: () => void;
</script>

<section class="hero">
  <div class="hero-text">
    <p class="kicker">DB Planner</p>
    <h1>Panel principal</h1>
    <p class="hint">Carga un proyecto para navegarlo en pestañas.</p>
  </div>
  <button class="btn primary" on:click={onOpen}>Cargar proyecto</button>
</section>

<style>
  .hero {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 20px;
    padding: 18px 20px;
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 14px;
    background: linear-gradient(135deg, rgba(255,255,255,0.08), rgba(255,255,255,0.02));
    backdrop-filter: blur(4px);
    box-shadow: 0 18px 50px rgba(0,0,0,0.28);
  }

  .hero-text h1 {
    margin: 0;
    font-size: 26px;
    letter-spacing: 0.4px;
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
    .hero {
      flex-direction: column;
      align-items: flex-start;
    }
  }
</style>
