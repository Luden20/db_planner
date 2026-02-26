<script lang="ts">
  import ModalLauncher from "../ModalLauncher.svelte";
  import {RemoveEntity} from "../../../wailsjs/go/main/App";

  export let id: number;
  export let onSave: () => Promise<void> = async () => {};

  const handleRemove = async () => {
    try {
      alert("Borrando entidad: " + id + "")
      await RemoveEntity(id);
      //await onSave();
      Promise.resolve(onSave()).catch((err) => {
        const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
        alert(`Se borró, pero falló la actualización: ${message}`);
        console.error("onSave failed:", err);
      });
    } catch (err) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      alert(`Error al borrar: ${message}`);
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
