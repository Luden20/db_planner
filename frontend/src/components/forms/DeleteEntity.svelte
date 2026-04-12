<script lang="ts">
  import ModalLauncher from "../ModalLauncher.svelte";
  import { RemoveEntity } from "../../../wailsjs/go/main/App";
  import { showToast } from "../../lib/toast";

  let { 
    id, 
    onSave = async () => {} 
  } = $props<{
    id: number;
    onSave?: () => Promise<void>;
  }>();

  const handleRemove = async () => {
    try {
      showToast("Eliminando entidad...", "info", 1200);
      await RemoveEntity(id);
      showToast("Entidad eliminada.", "success");
      onSave().catch((err) => {
        const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
        showToast(`Se borró, pero falló la actualización: ${message}`, "error");
        console.error("onSave failed:", err);
      });
    } catch (err: any) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      showToast(`Error al borrar: ${message}`, "error");
      throw err;
    }
  };
</script>

<ModalLauncher
  triggerLabel="Borrar"
  title="Seguro de borrar?"
  confirmLabel="Eliminar"
  triggerVariant="danger"
  confirmVariant="danger"
  triggerSize="sm"
  onSuccess={handleRemove}
/>

<style>
  .toolbar-actions {
    display: inline-flex;
    gap: 8px;
  }
</style>
