<script lang="ts">
  import Modal from "./ui/Modal.svelte";
  import Button from "./ui/Button.svelte";
  import { 
    FloppyDisk, 
    SignOut, 
    Trash, 
    PencilLine, 
    X, 
    Folder, 
    Eye, 
    IdentificationCard, 
    TreeStructure, 
    Cardholder, 
    Graph, 
    Database, 
    Link, 
    Plus, 
    Check, 
    Sparkle 
  } from "phosphor-svelte";

  let {
    triggerLabel = "Abrir modal",
    title = "Acción requerida",
    confirmLabel = "Confirmar",
    triggerVariant = "primary",
    confirmVariant = "primary",
    triggerIcon = null,
    confirmIcon = null,
    cancelIcon = "Cerrar",
    size = "default",
    triggerSize = "default",
    showTrigger = true,
    triggerClass = "",
    triggerDisabled = false,
    onSuccess = () => {},
    onOpen = () => {},
    children,
    trigger,
    isOpen = $bindable(false)
  } = $props<{
    triggerLabel?: string;
    title?: string;
    confirmLabel?: string;
    triggerVariant?: any;
    confirmVariant?: any;
    triggerIcon?: any;
    confirmIcon?: any;
    cancelIcon?: string;
    size?: any;
    triggerSize?: any;
    showTrigger?: boolean;
    triggerClass?: string;
    triggerDisabled?: boolean;
    onSuccess?: () => void | Promise<void>;
    onOpen?: () => void | Promise<void>;
    children?: import("svelte").Snippet;
    trigger?: import("svelte").Snippet;
    isOpen?: boolean;
  }>();

  let busy = $state(false);
  let errorMessage = $state("");

  const iconMap: Record<string, any> = {
    save: FloppyDisk,
    exit: SignOut,
    trash: Trash,
    edit: PencilLine,
    close: X,
    folder: Folder,
    eye: Eye,
    attributes: IdentificationCard,
    relations: TreeStructure,
    roles: Cardholder,
    flows: Graph,
    database: Database,
    link: Link,
    plus: Plus,
    check: Check,
    spark: Sparkle
  };

  const normalize = (value: string) => value.trim().toLowerCase();

  const inferIcon = (label: string, variant: string, fallback: string): any => {
    const normalized = normalize(label);
    if (normalized.includes("guardar")) return FloppyDisk;
    if (normalized.includes("salir")) return SignOut;
    if (normalized.includes("eliminar") || normalized.includes("borrar") || normalized.includes("quitar")) return Trash;
    if (normalized.includes("editar")) return PencilLine;
    if (normalized.includes("cerrar") || normalized.includes("cancelar")) return X;
    if (normalized.includes("cargar") || normalized.includes("abrir")) return Folder;
    if (normalized.includes("detalle") || normalized.includes("ver")) return Eye;
    if (normalized.includes("atribut")) return IdentificationCard;
    if (normalized.includes("relacion")) return TreeStructure;
    if (normalized.includes("rol")) return Cardholder;
    if (normalized.includes("flujo") || normalized.includes("proceso")) return Graph;
    if (normalized.includes("tabla") || normalized.includes("entidad")) return Database;
    if (normalized.includes("vincular")) return Link;
    if (normalized.includes("crear") || normalized.includes("nuevo")) return Plus;

    if (variant === "danger") return Trash;
    if (variant === "edit") return PencilLine;
    if (variant === "success") return Check;
    
    // In case string icon name was passed
    if (typeof fallback === "string" && iconMap[fallback]) return iconMap[fallback];
    return Sparkle;
  };

  const resolvedTriggerIcon = $derived(
    typeof triggerIcon === 'string' ? iconMap[triggerIcon] : (triggerIcon ?? inferIcon(`${triggerLabel} ${title}`, triggerVariant, triggerVariant === "primary" ? "spark" : "plus"))
  );
  
  const resolvedConfirmIcon = $derived(
    typeof confirmIcon === 'string' ? iconMap[confirmIcon] : (confirmIcon ?? inferIcon(confirmLabel, confirmVariant, confirmVariant === "primary" ? "save" : "check"))
  );

  const handleOpen = () => {
    errorMessage = "";
    try {
      const result = onOpen();
      if (result instanceof Promise) {
        result.catch((err) => {
          errorMessage = err?.message || "Error al abrir";
        });
      }
    } catch (err: any) {
      errorMessage = err?.message || "Error al abrir";
    }
  };

  const handleSuccess = async () => {
    busy = true;
    errorMessage = "";
    try {
      await onSuccess();
      isOpen = false;
    } catch (err: any) {
      errorMessage = err?.message || "Error en la acción";
    } finally {
      busy = false;
    }
  };

  export const openDialog = () => {
    handleOpen();
    isOpen = true;
  };

  export const closeDialog = () => {
    isOpen = false;
  };
</script>

<Modal
  bind:open={isOpen}
  {title}
  {size}
  {showTrigger}
  {confirmLabel}
  {confirmVariant}
  confirmIcon={resolvedConfirmIcon}
  cancelLabel={cancelIcon}
  {busy}
  {errorMessage}
  onConfirm={handleSuccess}
>
  {#snippet trigger()}
    {#if showTrigger}
      <Button
        variant={triggerVariant}
        size={triggerSize}
        icon={resolvedTriggerIcon}
        disabled={triggerDisabled}
        class={triggerClass}
        onclick={openDialog}
      >
        {triggerLabel}
      </Button>
    {/if}
  {/snippet}

  {#if children}
    {@render children()}
  {/if}
</Modal>
