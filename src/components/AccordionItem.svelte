<script lang="ts">
  import { slide } from 'svelte/transition';

  let open = false;

  function toggle() {
    open = !open;
  }
</script>

<button
  on:click={toggle}
  class="rounded-xl w-full border-4 border-b-8 border-black p-3 sm:px-10 lg:hover:scale-105 transition ease-in-out"
>
  <div class="flex justify-between items-center w-full sm:h-20">
    <slot name="header" />
    {#if open}
      <i class="bi bi-chevron-up ml-4 text-3xl" />
    {:else}
      <i class="bi bi-chevron-down ml-4 text-3xl" />
    {/if}
  </div>
  {#if open}
    <div class="mt-3 sm:mb-5" transition:slide>
      <slot name="content" />
    </div>
  {/if}
</button>
