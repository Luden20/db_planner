  <script lang="ts">
  import ModalLauncher from "../ModalLauncher.svelte";
  import {RemoveAttribute, Save} from "../../../wailsjs/go/main/App";
  import {showToast} from "../../lib/toast";

  export let entityId: number;
  export let attributeId: number;
  export let onSaved: () => Promise<void> = async () => {};

  const handleRemove = async () => {
    try {
      await RemoveAttribute(entityId, attributeId);
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
