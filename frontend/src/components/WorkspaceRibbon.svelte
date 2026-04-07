<script lang="ts">
  export let name: string;
  export let entityCount: number;
  export let activeTab: "entities" | "relations" | "roles" | "flows" | "tertiary";
  export let themeMode: "light" | "dark" = "light";
  export let onSelect: (tab: "entities" | "relations" | "roles" | "flows" | "tertiary") => void = () => {};
  export let onToggleTheme: () => void = () => {};
  export let onSave: () => Promise<void> = async () => {};
  export let onExport: () => void = () => {};
  export let onExportWithoutRelations: () => void = () => {};
  export let onExit: () => void = () => {};

  const viewItems = [
    {key: "entities" as const, label: "Entidades", hint: "Base del modelo"},
    {key: "relations" as const, label: "Relaciones", hint: "Cruces y tipos"},
    {key: "roles" as const, label: "Roles", hint: "Matriz de acceso"},
    {key: "flows" as const, label: "Flujos", hint: "Deck operativo"},
    {key: "tertiary" as const, label: "Atributos", hint: "Detalle por entidad"}
  ];
</script>

<section class="workspace-ribbon">
  <div class="ribbon-head">
    <div class="ribbon-copy">
      <p class="ribbon-kicker">Workspace Ribbon</p>
      <div class="ribbon-title-row">
        <h1>{name}</h1>
        <span class="meta-chip">{entityCount} entidades</span>
        <span class="meta-chip meta-chip--quiet">Sesion local</span>
      </div>
      <p class="ribbon-hint">Acciones frecuentes, exportacion y cambio de vistas siempre visibles, sin romper el flujo de trabajo.</p>
    </div>

    <button
      class="theme-toggle control control--soft control--sm"
      type="button"
      on:click={onToggleTheme}
      aria-label={themeMode === "dark" ? "Cambiar a modo claro" : "Cambiar a modo oscuro"}
    >
      {#if themeMode === "dark"}
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M21 12.79A9 9 0 0 1 11.21 3a.75.75 0 0 0-.82.98 7.5 7.5 0 1 0 9.63 9.63.75.75 0 0 0 .98-.82Z"/>
        </svg>
        <span>Usar claro</span>
      {:else}
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M12 3.25a.75.75 0 0 1 .75.75v1.5a.75.75 0 0 1-1.5 0V4a.75.75 0 0 1 .75-.75Zm0 14a5.25 5.25 0 1 0 0-10.5 5.25 5.25 0 0 0 0 10.5Zm8.75-6a.75.75 0 0 1 0 1.5h-1.5a.75.75 0 0 1 0-1.5h1.5ZM5.25 12a.75.75 0 0 1-.75.75H3a.75.75 0 0 1 0-1.5h1.5a.75.75 0 0 1 .75.75Zm11.16 5.41a.75.75 0 0 1 1.06 0l1.06 1.06a.75.75 0 1 1-1.06 1.06l-1.06-1.06a.75.75 0 0 1 0-1.06ZM5.53 5.53a.75.75 0 0 1 1.06 0l1.06 1.06A.75.75 0 0 1 6.59 7.65L5.53 6.59a.75.75 0 0 1 0-1.06Zm12 0a.75.75 0 0 1 0 1.06l-1.06 1.06a.75.75 0 0 1-1.06-1.06l1.06-1.06a.75.75 0 0 1 1.06 0Zm-12 12a.75.75 0 0 1 1.06 0l1.06 1.06a.75.75 0 1 1-1.06 1.06L5.53 18.6a.75.75 0 0 1 0-1.06ZM12 18.5a.75.75 0 0 1 .75.75v1.5a.75.75 0 0 1-1.5 0v-1.5a.75.75 0 0 1 .75-.75Z"/>
        </svg>
        <span>Usar oscuro</span>
      {/if}
    </button>
  </div>

  <div class="ribbon-grid">
    <section class="ribbon-group">
      <p class="group-label">Archivo</p>
      <div class="group-actions">
        <button class="ribbon-btn control control--stack control--primary" type="button" on:click={onSave}>
          <strong>Guardar</strong>
          <span class="control__meta">Persistir cambios</span>
        </button>
        <button class="ribbon-btn control control--stack control--danger" type="button" on:click={onExit}>
          <strong>Salir</strong>
          <span class="control__meta">Cerrar proyecto</span>
        </button>
      </div>
    </section>

    <section class="ribbon-group">
      <p class="group-label">Exportar</p>
      <div class="group-actions group-actions--wide">
        <button class="ribbon-btn control control--stack control--soft" type="button" on:click={onExport}>
          <strong>Con relaciones</strong>
          <span class="control__meta">Excel completo</span>
        </button>
        <button class="ribbon-btn control control--stack control--soft" type="button" on:click={onExportWithoutRelations}>
          <strong>Sin relaciones</strong>
          <span class="control__meta">Combinaciones</span>
        </button>
      </div>
    </section>

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
    z-index: 30;
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
    overflow: hidden;
  }

  .workspace-ribbon::before {
    content: "";
    position: absolute;
    inset: 0;
    background:
      linear-gradient(90deg, color-mix(in srgb, var(--ink) 6%, transparent) 1px, transparent 1px),
      linear-gradient(color-mix(in srgb, var(--ink) 6%, transparent) 1px, transparent 1px);
    background-size: 22px 22px;
    mask-image: linear-gradient(180deg, black, transparent 72%);
    pointer-events: none;
    opacity: 0.55;
  }

  .ribbon-head,
  .ribbon-grid {
    position: relative;
    z-index: 1;
  }

  .ribbon-head {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: start;
    gap: 0.9rem 1rem;
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
    grid-template-columns: minmax(15rem, 0.92fr) minmax(18rem, 1.08fr) minmax(22rem, 1.65fr);
    gap: 0.72rem;
    align-items: stretch;
  }

  .ribbon-group {
    display: grid;
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

  .ribbon-btn {
    justify-content: stretch;
    min-height: 4.55rem;
  }

  .ribbon-group--views {
    min-width: 0;
  }

  .view-strip {
    display: grid;
    grid-template-columns: repeat(5, minmax(0, 1fr));
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
      grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
    }

    .ribbon-group--views {
      grid-column: 1 / -1;
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

    .theme-toggle {
      width: 100%;
      justify-self: stretch;
    }

    .group-actions,
    .group-actions--wide,
    .view-strip {
      grid-template-columns: 1fr;
    }
  }
</style>
