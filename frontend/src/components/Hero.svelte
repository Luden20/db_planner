<script lang="ts">
  export let onOpen: () => void;
  export let onCreate: () => void;
</script>

<section class="hero">
  <div class="hero-text">
    <p class="kicker">DB Planner </p>
    <h1>Modela tu esquema de base de datos de forma facil</h1>
    <p class="hint">Por Alfonso Chafla (Propiedad intelectual)</p>

  </div>
  <div class="hero-actions">
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
    <p class="hero-meta">Entorno local, flujo tecnico y guardado persistente.</p>
  </div>
</section>

<style>
  .hero {
    position: relative;
    display: grid;
    grid-template-columns: minmax(0, 1.35fr) minmax(280px, 0.85fr);
    gap: 28px;
    padding: clamp(1.4rem, 3vw, 2.4rem);
    border-radius: var(--radius-lg);
    border: 1px solid var(--border);
    background:
      linear-gradient(
        135deg,
        color-mix(in srgb, var(--surface-strong) 94%, transparent),
        color-mix(in srgb, var(--bg-muted) 72%, transparent)
      ),
      linear-gradient(90deg, color-mix(in srgb, var(--accent) 12%, transparent), transparent 45%);
    box-shadow: var(--shadow-lg);
    overflow: hidden;
  }

  .hero::before {
    content: "";
    position: absolute;
    inset: 0;
    background:
      linear-gradient(90deg, var(--grid-line) 1px, transparent 1px),
      linear-gradient(var(--grid-line) 1px, transparent 1px);
    background-size: 22px 22px;
    mask-image: linear-gradient(135deg, black, transparent 72%);
    opacity: 0.5;
    pointer-events: none;
  }

  .hero-text,
  .hero-actions {
    position: relative;
    z-index: 1;
  }

  .hero-text h1 {
    margin: 0;
    max-width: 11ch;
    font-size: clamp(2.5rem, 6vw, 4.8rem);
    line-height: 0.92;
  }

  .kicker {
    margin: 0 0 0.75rem;
    color: var(--accent);
    text-transform: uppercase;
    font-size: 0.76rem;
    letter-spacing: 0.18em;
    font-weight: 800;
  }

  .hint {
    margin: 1rem 0 0;
    max-width: 58ch;
    color: var(--ink-soft);
    font-size: 1.02rem;
    line-height: 1.65;
  }

  .hero-notes {
    display: flex;
    flex-wrap: wrap;
    gap: 0.65rem;
    margin-top: 1.4rem;
  }

  .hero-notes span {
    display: inline-flex;
    align-items: center;
    min-height: 2.2rem;
    padding: 0 0.9rem;
    border-radius: 999px;
    background: color-mix(in srgb, var(--surface-strong) 76%, transparent);
    border: 1px solid var(--line-soft);
    color: var(--ink-soft);
    font-size: 0.88rem;
    font-weight: 700;
  }

  .hero-actions {
    display: grid;
    align-content: start;
    gap: 0.9rem;
    padding: 1.15rem;
    border-radius: calc(var(--radius-md) - 2px);
    border: 1px solid var(--line-soft);
    background: color-mix(in srgb, var(--surface) 84%, transparent);
    box-shadow: var(--surface-highlight);
  }

  .hero-actions :global(.btn) {
    justify-content: center;
    min-height: 3.2rem;
  }

  .hero-meta {
    margin: 0.2rem 0 0;
    color: var(--ink-faint);
    font-size: 0.9rem;
    line-height: 1.55;
  }

  @media (max-width: 720px) {
    .hero {
      grid-template-columns: 1fr;
      gap: 1.25rem;
    }
  }
</style>
