<script lang="ts">
  import { Button as BitsButton } from "bits-ui";
  import ButtonIcon from "../ButtonIcon.svelte";

  let { 
    variant = "secondary", // primary, secondary, accent, edit, success, danger, ghost
    size = "default", // default, sm, icon
    icon = null, 
    disabled = false,
    class: cssClass = "",
    children,
    onclick,
    type = "button",
    ...restProps
  } = $props();

  const getVariantClass = () => {
    if (size === "sm") {
       return `control control--sm control--${variant}`;
    }
    if (size === "icon") {
       return `control control--sm control--icon control--${variant}`;
    }
    return `btn ${variant}`;
  }
</script>

<BitsButton.Root
  {disabled}
  {type}
  class="{getVariantClass()} {cssClass}".trim()
  {onclick}
  {...restProps}
>
  {#if icon}
    <ButtonIcon name={icon}/>
  {/if}
  {#if children}
    <span>{@render children()}</span>
  {/if}
</BitsButton.Root>
