<script lang="ts">
  import { onMount } from "svelte";

  export let height = 0;
  
  let sentinelEl: HTMLDivElement | null = null;
  let stackEl: HTMLDivElement | null = null;
  let pinned = false;

  const sync = () => {
    if (sentinelEl) {
      pinned = sentinelEl.getBoundingClientRect().top <= 0;
    }
  };

  const syncHeight = () => {
    if (stackEl) {
      height = stackEl.offsetHeight;
    }
  };

  onMount(() => {
    syncHeight();
    sync();

    if (typeof ResizeObserver !== "undefined" && stackEl) {
      const observer = new ResizeObserver(syncHeight);
      observer.observe(stackEl);
      return () => observer.disconnect();
    }
  });

  $: if (sentinelEl) sync();
  $: if (stackEl) syncHeight();
</script>

<svelte:window on:scroll={sync} on:resize={sync} />

<div class="studio-sticky-sentinel" bind:this={sentinelEl} aria-hidden="true"></div>
<div class="studio-sticky-stack" class:studio-sticky-stack--pinned={pinned} bind:this={stackEl}>
  <slot></slot>
</div>
