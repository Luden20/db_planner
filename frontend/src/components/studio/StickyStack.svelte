<script lang="ts">
  import { onMount, type Snippet } from "svelte";

  let { 
    height = $bindable(0),
    children
  } = $props<{
    height?: number;
    children?: Snippet;
  }>();
  
  let sentinelEl = $state<HTMLDivElement | null>(null);
  let stackEl = $state<HTMLDivElement | null>(null);
  let pinned = $state(false);

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

  $effect(() => { if (sentinelEl) sync(); });
  $effect(() => { if (stackEl) syncHeight(); });
</script>

<svelte:window onscroll={sync} onresize={sync} />

<div class="studio-sticky-sentinel" bind:this={sentinelEl} aria-hidden="true"></div>
<div class="studio-sticky-stack" class:studio-sticky-stack--pinned={pinned} bind:this={stackEl}>
  {#if children}
    {@render children()}
  {/if}
</div>
