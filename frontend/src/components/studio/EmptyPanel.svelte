<script lang="ts">
  export let message: string;
  export let description: string = "";
</script>

<div class="studio-empty-panel">
  <strong>{message}</strong>
  {#if description}
    <p>{description}</p>
  {/if}
</div>
