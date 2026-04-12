<script lang="ts">
  let { message = "Contenido pendiente para esta pestaña." } = $props<{ message?: string }>();
</script>

<div class="empty-panel">
  <p>{message}</p>
</div>
