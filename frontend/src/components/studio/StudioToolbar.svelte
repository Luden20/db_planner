<script lang="ts">
  let { title, description = "", meta = null, actions = null } = $props();
</script>

<div class="studio-toolbar">
  <div class="studio-toolbar__copy">
    <p class="studio-toolbar__label">{title}</p>
    {#if description}
      <p class="studio-toolbar__muted">{description}</p>
    {/if}
  </div>
  
  <div class="studio-toolbar__meta">
    {#if meta}
      {@render meta()}
    {/if}
  </div>
  
  <div class="studio-toolbar__actions">
    {#if actions}
      {@render actions()}
    {/if}
  </div>
</div>
