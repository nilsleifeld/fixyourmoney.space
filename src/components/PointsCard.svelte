<script lang="ts">
  import { onMount } from 'svelte';

  export let headline: string;
  export let icon: string;
  export let points: string[] = [];
  export let openModalText: string;

  var modalOpen = false;

  function toggleModal() {
    modalOpen = !modalOpen;
  }

  function handleKeyPress(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      modalOpen = false;
    }
  }

  onMount(() => {
    window.addEventListener('keydown', handleKeyPress);
    return () => {
      window.removeEventListener('keydown', handleKeyPress);
    };
  });
</script>

<div class="px-3 py-4 border-4 border-black rounded-xl border-b-8">
  <div class="flex justify-between items-center mb-3">
    <h4 class="text-lg font-bold">{headline}</h4>
    <i class="text-4xl bi {icon}" />
  </div>
  <ul class="mb-3">
    {#each points as point}
      <li class="flex justify-start my-2">
        <i class="bi bi-check-circle-fill" />
        <span class="ml-2">{point}</span>
      </li>
    {/each}
  </ul>
  <button on:click={toggleModal} class="group flex">
    <span class="mr-1">{openModalText}</span>
    <i class="block bi bi-arrow-right transition-transform transform group-hover:translate-x-1" />
  </button>
</div>

{#if modalOpen}
  <div class="fixed inset-0 flex items-center justify-center z-50">
    <div class="fixed inset-0 bg-gray-800 bg-opacity-50 transition-opacity" />
    <div
      class="bg-white p-4 md:p-7 max-w-xl mx-auto rounded-xl z-50 border-black border-4 border-b-8 max-h-full overflow-y-scroll"
    >
      <div class="flex justify-between items-center mb-4">
        <h1 class="text-xl font-semibold">{headline}</h1>
        <button on:click={toggleModal}>
          <i class="bi bi-x text-4xl" />
        </button>
      </div>
      <slot />
    </div>
  </div>
{/if}
