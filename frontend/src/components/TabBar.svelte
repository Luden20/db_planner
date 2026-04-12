<script lang="ts">
  import ButtonIcon from "./ButtonIcon.svelte";

  let { 
    activeTab, 
    onSelect 
  } = $props<{
    activeTab: 'entities' | 'relations' | 'tertiary';
    onSelect: (tab: 'entities' | 'relations' | 'tertiary') => void;
  }>();
</script>

<nav class="tabs" aria-label="Vistas del proyecto">
  <button
    class:active={activeTab === 'entities'}
    class={`control control--stack ${activeTab === 'entities' ? 'control--accent control--active' : 'control--ghost'}`}
    onclick={() => onSelect('entities')}
  >
    <ButtonIcon name="database"/>
    <span class="tab-title">Entidades</span>
    <span class="tab-hint">Base estructural</span>
  </button>
  <button
    class:active={activeTab === 'relations'}
    class={`control control--stack ${activeTab === 'relations' ? 'control--accent control--active' : 'control--ghost'}`}
    onclick={() => onSelect('relations')}
  >
    <ButtonIcon name="relations"/>
    <span class="tab-title">Relaciones</span>
    <span class="tab-hint">Mapa combinatorio</span>
  </button>
  <button
    class:active={activeTab === 'tertiary'}
    class={`control control--stack ${activeTab === 'tertiary' ? 'control--accent control--active' : 'control--ghost'}`}
    onclick={() => onSelect('tertiary')}
  >
    <ButtonIcon name="attributes"/>
    <span class="tab-title">Atributos</span>
    <span class="tab-hint">Detalle por entidad</span>
  </button>
</nav>

<style>
  .tabs {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 0.55rem;
    margin-top: 1.2rem;
    padding: 0.55rem;
    border-radius: calc(var(--radius-md) - 2px);
    background: rgba(255, 252, 247, 0.72);
    border: 1px solid var(--border);
    box-shadow: var(--shadow-sm);
  }

  .tabs :global(.control) {
    justify-content: stretch;
  }

  .tab-title {
    font-size: 1rem;
    font-weight: 800;
  }

  .tab-hint {
    font-size: 0.77rem;
    color: var(--ink-faint);
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  @media (max-width: 720px) {
    .tabs {
      width: 100%;
      grid-template-columns: 1fr;
    }
  }
</style>
