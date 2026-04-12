<script lang="ts">
  let { 
    activeScope = "strong", 
    onSwitch 
  } = $props<{
    activeScope?: "strong" | "intersection";
    onSwitch: (scope: "strong" | "intersection") => void;
  }>();
</script>

<div class="scope-switch" role="tablist" aria-label="Tipo de vista">
  <button
    class={`scope-switch__item ${activeScope === 'strong' ? 'scope-switch__item--active' : ''}`}
    type="button"
    role="tab"
    aria-selected={activeScope === "strong"}
    onclick={() => onSwitch("strong")}
  >
    Fuertes
  </button>
  <button
    class={`scope-switch__item ${activeScope === 'intersection' ? 'scope-switch__item--active' : ''}`}
    type="button"
    role="tab"
    aria-selected={activeScope === "intersection"}
    onclick={() => onSwitch("intersection")}
  >
    Intersección
  </button>
</div>
