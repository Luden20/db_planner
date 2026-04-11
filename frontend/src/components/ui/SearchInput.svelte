<script lang="ts">
  let {
    value = $bindable(""),
    placeholder = "Buscar...",
    onkeydown = null,
    onclear = null,
    scope = "strong"
  } = $props();

  import ButtonIcon from "../ButtonIcon.svelte";

  const clearSearch = () => {
    value = "";
    if (onclear) onclear();
  };
</script>

<div class="search-field">
  <input
    class="search-input"
    type="search"
    bind:value={value}
    {placeholder}
    aria-label="Buscar entidad"
    onkeydown={onkeydown}
  />
  {#if value}
    <button class="control control--sm control--ghost" onclick={clearSearch} aria-label="Limpiar búsqueda">
      <ButtonIcon name="clear"/>
      <span>Limpiar</span>
    </button>
  {/if}
</div>
