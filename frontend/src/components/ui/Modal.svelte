<script lang="ts">
  import { Dialog, Separator } from "bits-ui";
  import { X, Sparkle } from "phosphor-svelte";
  import cn from "clsx";

  let { 
    title = "", 
    description = "",
    open = $bindable(false),
    triggerLabel = "Abrir",
    triggerIcon = null,
    triggerVariant = "primary",
    triggerSize = "default", // default, sm, icon
    triggerClass = "",
    showTrigger = false,
    size = "default", // default, form
    busy = false,
    errorMessage = "",
    confirmLabel = "Confirmar",
    confirmIcon = null,
    confirmVariant = "primary",
    cancelLabel = "Cancelar",
    children, // main body
    footer, // optional custom footer
    onConfirm = null,
    onCancel = null,
    triggerDisabled = false,
    trigger,
  } = $props();

  const handleConfirm = async () => {
     if (onConfirm) {
         await onConfirm();
     }
  }

  const handleCancel = () => {
    if (onCancel) {
      onCancel();
    }
    open = false;
  }
</script>

<Dialog.Root bind:open>
  {#if showTrigger}
    {#if trigger}
      {@render trigger()}
    {:else}
      <Dialog.Trigger
        disabled={triggerDisabled}
        class={cn(
          "rounded-input inline-flex items-center justify-center whitespace-nowrap font-bold transition-all active:scale-[0.98] disabled:opacity-50",
          triggerVariant === "primary" ? "bg-accent text-white hover:bg-accent/90 shadow-mini" : "bg-muted text-foreground hover:bg-muted/80",
          triggerSize === "sm" ? "h-9 px-4 text-sm" : "h-12 px-6 text-[15px]",
          triggerClass
        )}
      >
        {#if triggerIcon}
          {@const Icon = triggerIcon}
          <Icon class="size-4 mr-2" />
        {/if}
        <span>{triggerLabel}</span>
      </Dialog.Trigger>
    {/if}
  {/if}

  <Dialog.Portal>
    <Dialog.Overlay
      class="fixed inset-0 z-[100] bg-black/80 backdrop-blur-sm data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0"
    />
    <Dialog.Content
      class={cn(
        "fixed inset-0 m-auto z-[110] h-fit outline-hidden",
        "rounded-card-lg bg-background shadow-popover border p-6 flex flex-col",
        "data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95",
        size === "form" ? "w-[95vw] sm:max-w-[700px]" : "w-[95vw] sm:max-w-[490px]"
      )}
    >
        <Dialog.Title
          class="flex w-full items-center justify-center text-lg font-bold tracking-tight uppercase"
        >
          {title}
        </Dialog.Title>
        
        <Separator.Root class="bg-border-card -mx-6 mb-6 mt-5 block h-px" />
        
        {#if description}
          <Dialog.Description class="text-foreground-alt text-sm text-center mb-4">
            {description}
          </Dialog.Description>
        {/if}

        <div class="flex flex-col gap-4 overflow-y-auto max-h-[70vh]">
          {#if children}
            {@render children()}
          {/if}
          
          {#if errorMessage}
            <p class="text-destructive text-sm font-semibold mt-2">{errorMessage}</p>
          {/if}
        </div>

        <div class="flex w-full justify-end gap-3 mt-8">
          {#if footer}
            {@render footer()}
          {:else}
            <button
              class="h-11 rounded-input bg-muted text-foreground-alt hover:bg-muted/80 inline-flex items-center justify-center px-6 text-[14px] font-semibold transition-all active:scale-[0.98]"
              onclick={handleCancel}
              disabled={busy}
            >
              {cancelLabel}
            </button>
            
            <button
              class={cn(
                "h-11 rounded-input shadow-mini inline-flex items-center justify-center px-8 text-[14px] font-bold transition-all active:scale-[0.98] disabled:opacity-50",
                confirmVariant === "danger" ? "bg-destructive text-white hover:bg-destructive/90" : "bg-accent text-white hover:bg-accent/90"
              )}
              onclick={handleConfirm}
              disabled={busy}
            >
              {#if busy}
                <Sparkle class="size-4 mr-2 animate-spin" weight="fill" />
                <span>Procesando...</span>
              {:else}
                {#if confirmIcon}
                  {@const Icon = confirmIcon}
                  <Icon class="size-4 mr-2" />
                {/if}
                <span>{confirmLabel}</span>
              {/if}
            </button>
          {/if}
        </div>

        <Dialog.Close
          class="focus-visible:ring-foreground focus-visible:ring-offset-background focus-visible:outline-hidden absolute right-5 top-5 rounded-md focus-visible:ring-2 focus-visible:ring-offset-2 active:scale-[0.98] text-foreground-alt hover:text-foreground"
        >
          <X class="size-5" />
          <span class="sr-only">Cerrar</span>
        </Dialog.Close>
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>
