<script lang="ts">
  export let activeScope: "strong" | "intersection" = "strong";
  export let onSwitch: (scope: "strong" | "intersection") => void;
</script>

<div class="scope-switch" role="tablist" aria-label="Tipo de vista">
  <button
    class={`scope-switch__item ${activeScope === 'strong' ? 'scope-switch__item--active' : ''}`}
    type="button"
    role="tab"
    aria-selected={activeScope === "strong"}
    on:click={() => onSwitch("strong")}
  >
    Fuertes
  </button>
  <button
    class={`scope-switch__item ${activeScope === 'intersection' ? 'scope-switch__item--active' : ''}`}
    type="button"
    role="tab"
    aria-selected={activeScope === "intersection"}
    on:click={() => onSwitch("intersection")}
  >
    Intersección
  </button>
</div>
