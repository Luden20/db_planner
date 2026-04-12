<script lang="ts">
  import { NavigationMenu } from "bits-ui";
  import { 
    CaretDown, 
    Cardholder, 
    Database, 
    FileArrowUp, 
    FileCode, 
    SignOut, 
    FloppyDisk, 
    Table, 
    TreeStructure, 
    Graph, 
    IdentificationCard,
    Sparkle
  } from "phosphor-svelte";
  import cn from "clsx";

  let { 
    projectName,
    activeTab, 
    onSelect, 
    onSave, 
    onExport, 
    onExportWithoutRelations, 
    onExportScripts, 
    onExportPowerDesigner, 
    onExit 
  } = $props<{
    projectName: string;
    activeTab: string;
    onSelect: (tab: any) => void;
    onSave: () => void;
    onExport: () => void;
    onExportWithoutRelations: () => void;
    onExportScripts: () => void;
    onExportPowerDesigner: () => void;
    onExit: () => void;
  }>();

  const categories = [
     {
      label: "Archivo",
      value: "file",
      items: [
        { key: "save", title: "Guardar Proyecto", description: "Persiste todos los cambios en el archivo JSON local.", icon: FloppyDisk, onclick: onSave },
        { key: "export-full", title: "Excel (Completo)", description: "Exporta entidades y todas sus relaciones a Excel.", icon: FileArrowUp, onclick: onExport },
        { key: "export-plain", title: "Excel (Plano)", description: "Exporta solo combinaciones de atributos sin relaciones.", icon: Table, onclick: onExportWithoutRelations },
        { key: "export-sql", title: "Script SQL / IA", description: "Genera DDL completo y análisis de inteligencia artificial.", icon: FileCode, onclick: onExportScripts },
        { key: "export-pd", title: "PowerDesigner", description: "Genera script VBS para SAP PowerDesigner LDM.", icon: TreeStructure, onclick: onExportPowerDesigner },
        { key: "exit", title: "Cerrar Proyecto", description: "Libera el proyecto actual y vuelve a la pantalla de inicio.", icon: SignOut, danger: true, onclick: onExit }
      ]
    },
    {
      label: "Modelo",
      value: "model",
      items: [
        { key: "entities", title: "Entidades", description: "Define las tablas base y su estructura fundamental.", icon: Database, onclick: () => onSelect('entities') },
        { key: "tertiary", title: "Atributos", description: "Detalle profundo de campos, tipos y validaciones.", icon: IdentificationCard, onclick: () => onSelect('tertiary') },
        { key: "relations", title: "Relaciones", description: "Configura cruces, cardinalidades y claves foráneas.", icon: TreeStructure, onclick: () => onSelect('relations') }
      ]
    },
    {
      label: "Herramientas",
      value: "tools",
      items: [
        { key: "flows", title: "Flujos", description: "Visualiza el deck operativo y los procesos de datos.", icon: Graph, onclick: () => onSelect('flows') },
        { key: "roles", title: "Roles", description: "Gestiona la matriz de acceso y permisos del sistema.", icon: Cardholder, onclick: () => onSelect('roles') }
      ]
    }
  ];

  type ListItemProps = {
    className?: string;
    title: string;
    active?: boolean;
    onclick: () => void;
    content: string;
    icon?: any;
    danger?: boolean;
  };
</script>

{#snippet ListItem({ className, title, content, active, onclick, icon: Icon, danger }: ListItemProps)}
  <li>
    <NavigationMenu.Link
      class={cn(
        "hover:bg-muted focus:bg-muted outline-hidden block select-none space-y-1 rounded-md p-3 leading-none no-underline transition-colors",
        active ? "bg-accent/10 text-accent font-medium" : "text-foreground-alt hover:text-foreground",
        danger && "hover:bg-destructive/10 hover:text-destructive",
        className
      )}
      {onclick}
    >
      <div class="flex items-center gap-2 text-sm font-medium leading-none">
        {#if Icon}<Icon class="size-4" />{/if}
        {title}
      </div>
      <p class="text-muted-foreground line-clamp-2 text-sm leading-snug">
        {content}
      </p>
    </NavigationMenu.Link>
  </li>
{/snippet}

<NavigationMenu.Root class="relative z-10 flex w-full justify-center">
  <NavigationMenu.List
    class="group flex list-none items-center justify-center p-1 rounded-xl border border-border bg-background/50 backdrop-blur-xl shadow-2xl"
  >
    <!-- Project Name Side -->
    <div class="flex items-center px-4 py-1 mr-2 border-r border-border/50 select-none">
      <Sparkle class="size-3.5 text-accent mr-2" weight="fill" />
      <span class="text-sm font-bold tracking-tight text-foreground uppercase">{projectName}</span>
    </div>

    <!-- Home Item -->
    <NavigationMenu.Item value="home">
      <NavigationMenu.Link
        class={cn(
          "hover:text-accent-foreground focus:bg-muted focus:text-accent-foreground data-[state=open]:shadow-mini dark:hover:bg-muted dark:data-[state=open]:bg-muted focus:outline-hidden group inline-flex h-8 w-max items-center justify-center rounded-[8px] bg-transparent px-4 py-2 text-sm font-medium transition-colors hover:bg-white/10 disabled:pointer-events-none disabled:opacity-50",
          activeTab === 'home' && "text-accent bg-accent/10"
        )}
        onclick={() => onSelect('home')}
      >
        Inicio
      </NavigationMenu.Link>
    </NavigationMenu.Item>

    {#each categories as cat}
      <NavigationMenu.Item value={cat.value}>
        <NavigationMenu.Trigger
          class="hover:text-accent-foreground focus-visible:bg-muted focus-visible:text-accent-foreground data-[state=open]:shadow-mini dark:hover:bg-muted dark:data-[state=open]:bg-muted focus-visible:outline-hidden group inline-flex h-8 w-max items-center justify-center rounded-[8px] bg-transparent px-4 py-2 text-sm font-medium transition-colors hover:bg-white/10 disabled:pointer-events-none disabled:opacity-50 data-[state=open]:bg-white/10"
        >
          {cat.label}
          <CaretDown
            class="relative top-[1px] ml-1 size-3 transition-transform duration-200 group-data-[state=open]:rotate-180"
            aria-hidden="true"
          />
        </NavigationMenu.Trigger>
        <NavigationMenu.Content
          class="data-[motion=from-end]:animate-enter-from-right data-[motion=from-start]:animate-enter-from-left data-[motion=to-end]:animate-exit-to-right data-[motion=to-start]:animate-exit-to-left absolute left-0 top-0 w-full sm:w-auto"
        >
          <ul class={cn(
             "grid gap-2 p-3 sm:p-4",
             cat.value === 'file' ? "sm:w-[500px] sm:grid-cols-2" : "sm:w-[360px] sm:grid-cols-1"
          )}>
            {#each cat.items as item}
              {@render ListItem({
                title: item.title,
                content: item.description,
                icon: item.icon,
                danger: item.danger,
                active: activeTab === item.key,
                onclick: item.onclick
              })}
            {/each}
          </ul>
        </NavigationMenu.Content>
      </NavigationMenu.Item>
    {/each}

    <!-- Credit Side -->
    <div class="flex items-center px-4 py-1 ml-2 border-l border-border/50 select-none">
      <span class="text-[10px] font-black uppercase tracking-[0.2em] text-muted-foreground/60">by Luden20</span>
    </div>

    <!-- Indicator -->
    <NavigationMenu.Indicator
      class="data-[state=hidden]:animate-fade-out data-[state=visible]:animate-fade-in top-full z-10 flex h-2.5 items-end justify-center overflow-hidden opacity-100 transition-[all,transform_250ms_ease] duration-200 data-[state=hidden]:opacity-0"
    >
      <div
        class="bg-border relative top-[70%] size-2.5 rotate-[45deg] rounded-tl-[2px]"
      ></div>
    </NavigationMenu.Indicator>
  </NavigationMenu.List>

  <div
    class="perspective-[2000px] absolute left-0 top-full flex w-full justify-center"
  >
    <NavigationMenu.Viewport
      class="text-popover-foreground bg-background data-[state=closed]:animate-scale-out data-[state=open]:animate-scale-in relative mt-2.5 h-[var(--bits-navigation-menu-viewport-height)] w-full origin-[top_center] overflow-hidden rounded-xl border border-border shadow-2xl transition-[width,_height] duration-200 sm:w-[var(--bits-navigation-menu-viewport-width)]"
    />
  </div>
</NavigationMenu.Root>

<style>
  :global(.dark) [data-bits-navigation-menu-viewport] {
    background-color: var(--background);
    border-color: var(--border-card);
  }
</style>
