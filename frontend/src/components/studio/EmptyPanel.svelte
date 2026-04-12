<script lang="ts">
  let { message, description = "" } = $props<{ message: string; description?: string }>();
</script>

<div class="studio-empty-panel">
  <strong>{message}</strong>
  {#if description}
    <p>{description}</p>
  {/if}
</div>
