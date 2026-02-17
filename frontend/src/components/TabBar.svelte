<script lang="ts">
  export let activeTab:'entities' | 'relations' | 'tertiary';
  export let onSelect:(tab:'entities' | 'relations' | 'tertiary') => void;
</script>

<div class="tabs">
  <button class:active={activeTab === 'entities'} on:click={() => onSelect('entities')}>Entidades</button>
  <button class:active={activeTab === 'relations'} on:click={() => onSelect('relations')}>Relaciones</button>
  <button class:active={activeTab === 'tertiary'} on:click={() => onSelect('tertiary')}>Tab 3</button>
</div>

<style>
  .tabs {
    display: inline-flex;
    gap: 6px;
    margin-top: 22px;
    padding: 6px;
    border-radius: 14px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.07);
  }

  .tabs button {
    padding: 10px 18px;
    border-radius: 10px;
    border: 1px solid transparent;
    background: transparent;
    color: #d9e4f5;
    cursor: pointer;
    transition: all 150ms ease;
  }

  .tabs button.active {
    background: rgba(255, 255, 255, 0.08);
    border-color: rgba(255, 255, 255, 0.12);
    color: #6dd8ff;
    box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.04);
  }

  @media (max-width: 720px) {
    .tabs {
      width: 100%;
    }

    .tabs button {
      flex: 1;
      text-align: center;
    }
  }
</style>
