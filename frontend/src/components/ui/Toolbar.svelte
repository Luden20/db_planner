<script lang="ts">
  let { 
    title, 
    description = "", 
    actions = null,
    meta = null,
    class: cssClass = ""
  } = $props();
</script>

<div class="tab-toolbar {cssClass}">
  <div class="toolbar-left">
    <p class="label">{title}</p>
    {#if description}
      <p class="muted">{description}</p>
    {/if}
  </div>
  <div class="toolbar-right">
    {#if meta}
      <div class="toolbar-meta">
        {@render meta()}
      </div>
    {/if}
    {#if actions}
      <div class="toolbar-actions">
        {@render actions()}
      </div>
    {/if}
  </div>
</div>
