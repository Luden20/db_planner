<script lang="ts">
  import Navigation from "./Navigation.svelte";

  let { 
    name, 
    activeTab, 
    onSelect = () => {}, 
    onSave = async () => {},
    onExport = () => {},
    onExportWithoutRelations = () => {},
    onExportScripts = () => {},
    onExportPowerDesigner = () => {},
    onExit = () => {} 
  } = $props<{
    name: string;
    activeTab: "home" | "entities" | "relations" | "roles" | "flows" | "tertiary";
    onSelect?: (tab: "home" | "entities" | "relations" | "roles" | "flows" | "tertiary") => void;
    onSave?: () => Promise<void>;
    onExport?: () => void;
    onExportWithoutRelations?: () => void;
    onExportScripts?: () => void;
    onExportPowerDesigner?: () => void;
    onExit?: () => void;
  }>();
</script>

<section class="workspace-ribbon">
  <Navigation 
    projectName={name}
    {activeTab} 
    {onSelect} 
    {onSave}
    {onExport}
    {onExportWithoutRelations}
    {onExportScripts}
    {onExportPowerDesigner}
    {onExit}
  />
</section>

<style>
  .workspace-ribbon {
    position: sticky;
    top: 0.7rem;
    z-index: var(--layer-ribbon);
    display: flex;
    justify-content: center;
    padding: 0;
    overflow: visible;
    isolation: isolate;
  }
</style>
