<script lang="ts">
  export let onOpen: () => void;
  export let onCreate: () => void;
</script>

<section class="hero">
  <div class="hero-text">
    <p class="kicker">DB Planner</p>
    <h1>Panel principal</h1>
    <p class="hint">Carga un proyecto para navegarlo en pestañas.</p>
  </div>
  <button class="btn secondary" on:click={onOpen}>
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M4.75 6.5A1.75 1.75 0 0 0 3 8.25v8.5A2.25 2.25 0 0 0 5.25 19h13.5A2.25 2.25 0 0 0 21 16.75v-8A2.25 2.25 0 0 0 18.75 6.5h-6.19c-.3 0-.59-.13-.8-.36l-.7-.74A1.75 1.75 0 0 0 9.56 5H5.75c-.37 0-.73.12-1 .34v1.16Zm0 1.5V6a.25.25 0 0 1 .25-.25h3.4c.2 0 .4.09.53.24l.7.74c.46.49 1.1.77 1.77.77h7.35a.75.75 0 0 1 .75.75V8h-14.7ZM5.25 17.5a.75.75 0 0 1-.75-.75V9.5h15v7.25a.75.75 0 0 1-.75.75H5.25Z"/>
    </svg>
    <span>Cargar proyecto</span>
  </button>
  <button class="btn primary" on:click={onCreate}>
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path d="M12 4a.75.75 0 0 0-.75.75V11H5.75a.75.75 0 0 0 0 1.5h5.5v6.25a.75.75 0 0 0 1.5 0V12.5h5.5a.75.75 0 0 0 0-1.5h-5.5V4.75A.75.75 0 0 0 12 4Z"/>
    </svg>
    <span>Crear proyecto</span>
  </button>

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
