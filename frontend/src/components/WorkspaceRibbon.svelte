<script lang="ts">
  import ButtonIcon from "./ButtonIcon.svelte";

  export let name: string;
  export let entityCount: number;
  export let activeTab: "home" | "entities" | "relations" | "roles" | "flows" | "tertiary";
  export let themeMode: "light" | "dark" = "light";
  export let onSelect: (tab: "home" | "entities" | "relations" | "roles" | "flows" | "tertiary") => void = () => {};
  export let onToggleTheme: () => void = () => {};
  export let onSave: () => Promise<void> = async () => {};
  export let onExport: () => void = () => {};
  export let onExportWithoutRelations: () => void = () => {};
  export let onExportScripts: () => void = () => {};
  export let onExit: () => void = () => {};
  let exportMenuOpen = false;
  let exportMenuEl: HTMLDivElement | null = null;
  let exportTriggerEl: HTMLButtonElement | null = null;

  const viewItems = [
    {key: "home" as const, label: "Home", hint: "Sandbox libre", icon: "home" as const},
    {key: "entities" as const, label: "Entidades", hint: "Base del modelo", icon: "database" as const},
    {key: "relations" as const, label: "Relaciones", hint: "Cruces y tipos", icon: "relations" as const},
    {key: "tertiary" as const, label: "Atributos", hint: "Detalle por entidad", icon: "attributes" as const},
    {key: "flows" as const, label: "Flujos", hint: "Deck operativo", icon: "flows" as const},
    {key: "roles" as const, label: "Roles", hint: "Matriz de acceso", icon: "roles" as const}
  ];

  const toggleExportMenu = () => {
    exportMenuOpen = !exportMenuOpen;
  };

  const closeExportMenu = () => {
    exportMenuOpen = false;
  };

  const handleExportAction = (mode: "full" | "plain" | "scripts") => {
    closeExportMenu();
    if (mode === "full") {
      onExport();
      return;
    }
    if (mode === "scripts") {
      onExportScripts();
      return;
    }
    onExportWithoutRelations();
  };

  const handleWindowClick = (event: MouseEvent) => {
    const target = event.target;
    if (!(target instanceof Node)) return;
    if (exportMenuEl?.contains(target) || exportTriggerEl?.contains(target)) return;
    closeExportMenu();
  };

  const handleWindowKeydown = (event: KeyboardEvent) => {
    if (event.key === "Escape") {
      closeExportMenu();
    }
  };
</script>

<svelte:window on:click={handleWindowClick} on:keydown={handleWindowKeydown}/>

<section class="workspace-ribbon">
  <div class="ribbon-head">
    <div class="ribbon-copy">
      <p class="ribbon-kicker">By Luden20 github.com/Luden20</p>
      <div class="ribbon-title-row">
        <h1>{name}</h1>
        <span class="meta-chip">{entityCount} entidades</span>
      </div>
      <p class="ribbon-hint">Acciones frecuentes, exportacion y cambio de vistas siempre visibles, sin romper el flujo de trabajo.</p>
    </div>

    <div class="ribbon-head-side">
      <button
        class="theme-toggle control control--soft control--sm"
        type="button"
        on:click={onToggleTheme}
        aria-label={themeMode === "dark" ? "Cambiar a modo claro" : "Cambiar a modo oscuro"}
      >
        {#if themeMode === "dark"}
          <ButtonIcon name="theme-light"/>
          <span>Usar claro</span>
        {:else}
          <ButtonIcon name="theme-dark"/>
          <span>Usar oscuro</span>
        {/if}
      </button>

      <section class="ribbon-group ribbon-group--file ribbon-group--file-inline">
        <p class="group-label">Archivo</p>
        <div class="group-actions group-actions--file">
          <button class="control control--primary" type="button" on:click={onSave}>
            <ButtonIcon name="save"/>
            <strong>Guardar</strong>
          </button>
          <button class="control control--danger" type="button" on:click={onExit}>
            <ButtonIcon name="exit"/>
            <strong>Salir</strong>
          </button>
          <div class="export-menu" bind:this={exportMenuEl}>
            <button
              bind:this={exportTriggerEl}
              class="control control--soft control--sm export-trigger"
              class:export-trigger--open={exportMenuOpen}
              type="button"
              aria-haspopup="menu"
              aria-expanded={exportMenuOpen}
              on:click={toggleExportMenu}
            >
              <ButtonIcon name="download"/>
              <span>Exportar</span>
              <ButtonIcon name={exportMenuOpen ? "arrow-up" : "arrow-down"}/>
            </button>

            {#if exportMenuOpen}
              <div class="export-popover" role="menu" aria-label="Opciones de exportacion">
                <button class="export-option" type="button" role="menuitem" on:click={() => handleExportAction("full")}>
                  <strong>Con relaciones</strong>
                  <span>Excel completo</span>
                </button>
                <button class="export-option" type="button" role="menuitem" on:click={() => handleExportAction("plain")}>
                  <strong>Sin relaciones</strong>
                  <span>Combinaciones</span>
                </button>
                <button class="export-option export-option--ai" type="button" role="menuitem" on:click={() => handleExportAction("scripts")}>
                  <strong>Script SQL completo</strong>
                  <span>DDL, relaciones, docs, inserts y salida IA opcional</span>
                </button>
              </div>
            {/if}
          </div>
        </div>
      </section>
    </div>
  </div>

  <div class="ribbon-grid">
    <section class="ribbon-group ribbon-group--views">
      <p class="group-label">Vistas</p>
      <div class="view-strip">
        {#each viewItems as item}
          <button
            class:active={activeTab === item.key}
            class={`view-tab control control--stack ${activeTab === item.key ? 'control--accent control--active' : 'control--ghost'}`}
            type="button"
            on:click={() => onSelect(item.key)}
          >
            <ButtonIcon name={item.icon}/>
            <span class="view-tab__label">{item.label}</span>
            <span class="view-tab__hint control__meta">{item.hint}</span>
          </button>
        {/each}
      </div>
    </section>
  </div>
</section>

<style>
  .workspace-ribbon {
    position: sticky;
    top: 0.7rem;
    z-index: var(--layer-ribbon);
    display: grid;
    gap: 0.8rem;
    padding: clamp(0.9rem, 1.2vw, 1.05rem) clamp(0.95rem, 1.4vw, 1.2rem) clamp(0.95rem, 1.3vw, 1.1rem);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    background:
      linear-gradient(180deg, color-mix(in srgb, var(--surface-strong) 86%, transparent), color-mix(in srgb, var(--surface) 94%, transparent)),
      linear-gradient(90deg, color-mix(in srgb, var(--accent) 12%, transparent), transparent 36%);
    box-shadow: var(--shadow-lg);
    backdrop-filter: blur(18px);
    overflow: visible;
    isolation: isolate;
  }

  .workspace-ribbon::before {
    content: "";
    position: absolute;
    inset: 0;
    background:
      linear-gradient(90deg, color-mix(in srgb, var(--ink) 6%, transparent) 1px, transparent 1px),
      linear-gradient(color-mix(in srgb, var(--ink) 6%, transparent) 1px, transparent 1px);
    background-size: 22px 22px;
    border-radius: inherit;
    mask-image: linear-gradient(180deg, black, transparent 72%);
    pointer-events: none;
    opacity: 0.55;
  }

  .ribbon-head {
    position: relative;
    z-index: 2;
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: start;
    gap: 0.9rem 1rem;
  }

  .ribbon-grid {
    position: relative;
    z-index: 1;
  }

  .ribbon-head-side {
    display: grid;
    justify-items: end;
    gap: 0.6rem;
    min-width: min(100%, 28rem);
  }

  .ribbon-copy {
    min-width: 0;
  }

  .ribbon-kicker,
  .group-label {
    margin: 0;
    color: var(--accent);
    font-size: 0.72rem;
    font-weight: 800;
    letter-spacing: 0.18em;
    text-transform: uppercase;
  }

  .ribbon-title-row {
    display: flex;
    align-items: center;
    gap: 0.7rem;
    flex-wrap: wrap;
    margin-top: 0.45rem;
  }

  .ribbon-title-row h1 {
    margin: 0;
    font-size: clamp(1.8rem, 4vw, 3rem);
    line-height: 0.94;
  }

  .meta-chip {
    display: inline-flex;
    align-items: center;
    min-height: 2rem;
    padding: 0 0.8rem;
    border-radius: 999px;
    background: color-mix(in srgb, var(--accent) 10%, var(--surface-strong));
    border: 1px solid color-mix(in srgb, var(--accent) 16%, var(--border));
    color: var(--accent-strong);
    font-size: 0.82rem;
    font-weight: 800;
  }

  .meta-chip--quiet {
    background: color-mix(in srgb, var(--surface-strong) 88%, var(--ink) 3%);
    color: var(--ink-soft);
    border-color: var(--border);
  }

  .ribbon-hint {
    margin: 0.55rem 0 0;
    max-width: 82ch;
    color: var(--ink-soft);
    line-height: 1.55;
  }

  .theme-toggle {
    min-width: 9.75rem;
    justify-self: end;
  }

  .ribbon-grid {
    display: grid;
    grid-template-columns: minmax(0, 1fr);
    gap: 0.6rem;
    align-items: start;
  }

  .ribbon-group {
    display: grid;
    min-width: 0;
    gap: 0.65rem;
    padding: 0.78rem 0.82rem;
    border-radius: calc(var(--radius-md) - 6px);
    border: 1px solid var(--border);
    background: color-mix(in srgb, var(--surface) 92%, transparent);
    box-shadow: inset 0 1px 0 color-mix(in srgb, white 38%, transparent);
  }

  .group-actions {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.65rem;
  }

  .group-actions--wide {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .group-actions--file {
    grid-template-columns: auto auto auto;
    align-items: center;
  }

  .ribbon-group--file {
    align-self: start;
    justify-self: start;
    width: fit-content;
    max-width: 100%;
  }

  .ribbon-group--file-inline {
    justify-self: end;
  }

  .export-menu {
    position: relative;
    align-self: center;
  }

  .export-trigger {
    min-width: 9rem;
    justify-content: space-between;
    gap: 0.45rem;
  }

  .export-trigger--open {
    border-color: color-mix(in srgb, var(--accent) 28%, var(--border-strong));
    box-shadow:
      0 14px 22px color-mix(in srgb, var(--ink) 10%, transparent),
      0 0 0 0.18rem color-mix(in srgb, var(--accent) 10%, transparent),
      inset 0 1px 0 color-mix(in srgb, white 30%, transparent);
  }

  .export-popover {
    position: absolute;
    top: calc(100% + 0.45rem);
    left: 0;
    z-index: calc(var(--layer-ribbon) + 2);
    display: grid;
    gap: 0.35rem;
    min-width: 12.5rem;
    padding: 0.45rem;
    border: 1px solid var(--border);
    border-radius: 1rem;
    background: var(--popover-surface);
    box-shadow: var(--shadow-sm);
    white-space: nowrap;
  }

  .export-option {
    display: grid;
    gap: 0.12rem;
    width: 100%;
    padding: 0.68rem 0.78rem;
    border: 0;
    border-radius: 0.78rem;
    background: transparent;
    color: var(--ink);
    text-align: left;
    cursor: pointer;
    transition: background 160ms ease, color 160ms ease, transform 160ms ease;
  }

  .export-option strong {
    font-size: 0.9rem;
    font-weight: 800;
  }

  .export-option span {
    color: var(--ink-faint);
    font-size: 0.76rem;
    font-weight: 700;
  }

  .export-option:hover {
    background: color-mix(in srgb, var(--accent) 10%, var(--surface-strong));
    transform: translateY(-1px);
  }

  .export-option--ai {
    border: 1px solid color-mix(in srgb, var(--accent) 14%, var(--border));
    background: color-mix(in srgb, var(--accent) 8%, var(--surface-strong));
  }

  .ribbon-group--views {
    min-width: 0;
  }

  .view-strip {
    display: grid;
    grid-template-columns: repeat(6, minmax(0, 1fr));
    gap: 0.6rem;
  }

  .view-tab {
    justify-content: stretch;
  }

  .view-tab__label {
    font-size: 0.96rem;
    font-weight: 800;
  }

  .view-tab__hint {
    color: var(--ink-faint);
    font-size: 0.76rem;
    line-height: 1.38;
  }

  @media (max-width: 1320px) {
    .ribbon-grid {
      grid-template-columns: minmax(0, 1fr);
    }
  }

  @media (max-width: 980px) {
    .workspace-ribbon {
      top: 0.5rem;
    }

    .ribbon-grid {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 720px) {
    .workspace-ribbon {
      padding: 0.82rem;
    }

    .ribbon-head {
      grid-template-columns: 1fr;
    }

    .ribbon-head-side {
      min-width: 0;
      justify-items: stretch;
    }

    .theme-toggle {
      width: 100%;
      justify-self: stretch;
    }

    .group-actions,
    .group-actions--file,
    .group-actions--wide,
    .view-strip {
      grid-template-columns: 1fr;
    }

    .export-menu,
    .export-trigger {
      width: 100%;
    }

    .ribbon-group--file-inline {
      justify-self: stretch;
      width: 100%;
    }

    .export-popover {
      right: 0;
      left: 0;
      min-width: 0;
    }
  }
</style>
