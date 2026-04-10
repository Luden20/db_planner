<script lang="ts">
  import ButtonIcon from "./ButtonIcon.svelte";

  export let onOpen: () => void;
  export let onCreate: () => void;

  const features = [
    { icon: "database", label: "Entorno Local" },
    { icon: "relations", label: "Flujo Técnico" },
    { icon: "save", label: "Auto-guardado" },
    { icon: "check", label: "Validación Lógica" }
  ];
</script>

<section class="hero">
  <div class="hero-content">
    <header class="hero-header">
      <div class="hero-badge">
        <ButtonIcon name="spark" />
        <span>DB Planner v2.0</span>
      </div>
      <h1>Modela tu base de datos <span class="text-accent">sin fricciones</span></h1>
      <p class="hero-subtitle">
        Diseña esquemas profesionales, gestiona relaciones complejas y visualiza el flujo de datos en tiempo real.
      </p>
    </header>

    <div class="hero-features">
      {#each features as feature}
        <div class="feature-tag">
          <ButtonIcon name={feature.icon} />
          <span>{feature.label}</span>
        </div>
      {/each}
    </div>

    <footer class="hero-footer">
      <div class="hero-actions">
        <button class="control control--primary control--stack" on:click={onCreate}>
          <div class="control__label">
            <ButtonIcon name="plus" />
            <span>Crear Proyecto</span>
          </div>
          <span class="control__meta">Inicia un nuevo esquema desde cero</span>
        </button>

        <button class="control control--soft control--stack" on:click={onOpen}>
          <div class="control__label">
            <ButtonIcon name="folder" />
            <span>Cargar Existente</span>
          </div>
          <span class="control__meta">Continúa trabajando en tus archivos .json</span>
        </button>
      </div>

      <div class="hero-credits">
        <p class="hint">Diseñado por <strong>Alfonso Chafla</strong></p>
        <span class="copyright-tag">Propiedad intelectual</span>
      </div>
    </footer>
  </div>

  <div class="hero-visual">
    <div class="visual-orb visual-orb--1"></div>
    <div class="visual-orb visual-orb--2"></div>
    <div class="visual-grid"></div>
  </div>
</section>

<style>
  .hero {
    position: relative;
    display: flex;
    flex-direction: column;
    min-height: clamp(400px, 70vh, 800px);
    padding: clamp(2rem, 5vw, 4rem);
    border-radius: var(--radius-lg);
    border: 1px solid var(--border-strong);
    background: var(--panel-surface-strong);
    box-shadow: var(--shadow-lg);
    overflow: hidden;
  }

  .hero-content {
    position: relative;
    z-index: 2;
    max-width: 800px;
    display: flex;
    flex-direction: column;
    gap: 2.5rem;
  }

  .hero-header {
    display: flex;
    flex-direction: column;
    gap: 1.2rem;
  }

  .hero-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.6rem;
    padding: 0.4rem 0.8rem;
    padding-right: 1rem;
    background: var(--accent-soft);
    border: 1px solid var(--accent);
    color: var(--accent);
    border-radius: 999px;
    font-size: 0.75rem;
    font-weight: 800;
    letter-spacing: 0.05em;
    width: fit-content;
    text-transform: uppercase;
  }

  .hero-badge :global(svg) {
    width: 1rem;
    height: 1rem;
  }

  .hero h1 {
    margin: 0;
    font-size: clamp(2.5rem, 8vw, 4.2rem);
    line-height: 1.05;
    letter-spacing: -0.04em;
    color: var(--ink);
  }

  .text-accent {
    color: var(--accent);
    position: relative;
    display: inline-block;
  }

  .text-accent::after {
    content: "";
    position: absolute;
    bottom: 0.15em;
    left: 0;
    right: 0;
    height: 0.2em;
    background: var(--accent);
    opacity: 0.15;
    z-index: -1;
    border-radius: 2px;
  }

  .hero-subtitle {
    margin: 0;
    max-width: 42ch;
    font-size: 1.25rem;
    line-height: 1.5;
    color: var(--ink-soft);
  }

  .hero-features {
    display: flex;
    flex-wrap: wrap;
    gap: 0.8rem;
  }

  .feature-tag {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    padding: 0.5rem 1rem;
    background: var(--surface);
    border: 1px solid var(--line-soft);
    border-radius: var(--radius-sm);
    font-size: 0.88rem;
    font-weight: 600;
    color: var(--ink-soft);
    backdrop-filter: var(--glass);
    box-shadow: var(--shadow-inner);
    transition: all 0.2s ease;
  }

  .feature-tag:hover {
    border-color: var(--accent);
    color: var(--ink);
    transform: translateY(-2px);
  }

  .feature-tag :global(svg) {
    width: 1rem;
    height: 1rem;
    opacity: 0.7;
  }

  .hero-footer {
    display: flex;
    flex-direction: column;
    gap: 2rem;
  }

  .hero-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 1.2rem;
  }

  .hero-actions .control {
    flex: 1;
    min-width: 240px;
  }

  .control__label {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    font-size: 1.1rem;
  }

  .control__label :global(svg) {
    width: 1.2rem;
    height: 1.2rem;
  }

  .hero-credits {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding-top: 1rem;
    border-top: 1px solid var(--line-faint);
  }

  .hero-credits .hint {
    margin: 0;
    font-size: 0.9rem;
  }

  .copyright-tag {
    font-size: 0.72rem;
    padding: 0.2rem 0.5rem;
    background: var(--line-faint);
    color: var(--ink-faint);
    border-radius: 4px;
    text-transform: uppercase;
    font-weight: 700;
    letter-spacing: 0.02em;
  }

  /* Visual Elements */
  .hero-visual {
    position: absolute;
    inset: 0;
    z-index: 1;
    pointer-events: none;
    overflow: hidden;
  }

  .visual-grid {
    position: absolute;
    inset: 0;
    background-image:
      linear-gradient(var(--grid-line) 1px, transparent 1px),
      linear-gradient(90deg, var(--grid-line) 1px, transparent 1px);
    background-size: 30px 30px;
    mask-image: radial-gradient(circle at 70% 40%, black, transparent 80%);
    opacity: 0.4;
  }

  .visual-orb {
    position: absolute;
    border-radius: 50%;
    filter: blur(80px);
    opacity: 0.15;
    mix-blend-mode: multiply;
  }

  :global([data-theme="dark"]) .visual-orb {
    mix-blend-mode: screen;
    opacity: 0.1;
  }

  .visual-orb--1 {
    top: 10%;
    right: 5%;
    width: 40%;
    height: 60%;
    background: var(--accent);
  }

  .visual-orb--2 {
    bottom: -10%;
    left: 20%;
    width: 30%;
    height: 40%;
    background: var(--edit);
  }

  @media (max-width: 640px) {
    .hero {
      padding: 2rem;
    }
    .hero-actions .control {
      flex: 1 1 100%;
    }
    .hero h1 {
      font-size: 2.8rem;
    }
    .hero-credits {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.5rem;
    }
  }
</style>
