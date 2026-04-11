<script lang="ts">
  let { 
    header, 
    body, 
    class: cssClass = "",
    tableClass = "",
    ref = $bindable(),
    ...restProps
  } = $props();
</script>

<div class="table-wrapper {cssClass}" bind:this={ref} {...restProps}>
  <table class="entities-table {tableClass}">
    {#if header}
    <thead>
      <tr>
        {@render header()}
      </tr>
    </thead>
    {/if}
    {#if body}
    <tbody class={tableClass.includes('draggable') ? 'draggable-body' : ''}>
      {@render body()}
    </tbody>
    {/if}
  </table>
</div>
