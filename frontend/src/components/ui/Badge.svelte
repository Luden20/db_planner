<script lang="ts">
  let { children, variant = "default" } = $props();
</script>
<span class="studio-chip {variant === 'quiet' ? 'studio-chip--quiet' : ''}">
  {@render children()}
</span>
