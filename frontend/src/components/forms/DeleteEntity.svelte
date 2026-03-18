<script lang="ts">
  import ModalLauncher from "../ModalLauncher.svelte";
  import {RemoveEntity} from "../../../wailsjs/go/main/App";
  import {showToast} from "../../lib/toast";

  export let id: number;
  export let onSave: () => Promise<void> = async () => {};

  const handleRemove = async () => {
    try {
      showToast("Eliminando entidad...", "info", 1200);
      await RemoveEntity(id);
      showToast("Entidad eliminada.", "success");
      Promise.resolve(onSave()).catch((err) => {
        const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
        showToast(`Se borró, pero falló la actualización: ${message}`, "error");
        console.error("onSave failed:", err);
      });
    } catch (err) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      showToast(`Error al borrar: ${message}`, "error");
      throw err;
    }
  };
</script>

<div class="toolbar-actions">
  <ModalLauncher
    triggerLabel="Borrar"
    title="Seguro de borrar?"
    confirmLabel="Eliminar"
    triggerVariant="danger"
    confirmVariant="danger"
    onSuccess={handleRemove}
  />
</div>

<style>
  .toolbar-actions {
    display: inline-flex;
    gap: 8px;
  }
</style>
