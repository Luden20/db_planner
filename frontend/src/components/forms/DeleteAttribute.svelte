<script lang="ts">
  import ModalLauncher from "../ModalLauncher.svelte";
  import { RemoveAttribute, RemoveIntersectionAttribute, Save } from "../../../wailsjs/go/main/App";
  import { showToast } from "../../lib/toast";

  let { 
    entityId = null, 
    relationId = null, 
    attributeId, 
    onSaved = async () => {}, 
    triggerClass = "" 
  } = $props<{
    entityId?: number | null;
    relationId?: number | null;
    attributeId: number;
    onSaved?: () => Promise<void>;
    triggerClass?: string;
  }>();

  const handleRemove = async () => {
    try {
      if (relationId !== null) {
        await RemoveIntersectionAttribute(relationId, attributeId);
      } else if (entityId !== null) {
        await RemoveAttribute(entityId, attributeId);
      } else throw new Error("No se encontró el contenedor del atributo.");
      
      await Save();
      await onSaved();
      showToast("Atributo eliminado.", "success");
    } catch (err: any) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      showToast(`Error al borrar: ${message}`, "error");
      throw err;
    }
  };
</script>

<ModalLauncher
  triggerLabel="Borrar"
  title="Eliminar atributo"
  confirmLabel="Eliminar"
  triggerVariant="danger"
  confirmVariant="danger"
  {triggerClass}
  onSuccess={handleRemove}
/>
