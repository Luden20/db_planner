  <script lang="ts">
  import ModalLauncher from "../ModalLauncher.svelte";
  import {RemoveAttribute, RemoveIntersectionAttribute, Save} from "../../../wailsjs/go/main/App";
  import {showToast} from "../../lib/toast";

  export let entityId: number | null = null;
  export let relationId: number | null = null;
  export let attributeId: number;
  export let onSaved: () => Promise<void> = async () => {};

  const handleRemove = async () => {
    try {
      if (relationId !== null) {
        await RemoveIntersectionAttribute(relationId, attributeId);
      } else if (entityId !== null) {
        await RemoveAttribute(entityId, attributeId);
      } else {
        throw new Error("No se encontro el contenedor del atributo.");
      }
      await Save();
      await onSaved();
      showToast("Atributo eliminado.", "success");
    } catch (err) {
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
  onSuccess={handleRemove}
/>
