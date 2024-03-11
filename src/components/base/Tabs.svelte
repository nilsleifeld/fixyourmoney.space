<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  interface Tab {
    name: string;
    id: number;
    hasSignal: boolean;
  }

  export let tabs: Tab[] = [
    {
      name: 'Dollar',
      id: 0,
      hasSignal: false,
    },
    {
      name: 'Gold',
      id: 1,
      hasSignal: true,
    },
    {
      name: 'Bitcoin',
      id: 2,
      hasSignal: true,
    },
  ];

  export let activeTabId = 0;
  export let text = 'Wähle deine Recheneinheit';

  const dispatch = createEventDispatcher();

  function setActiveTab(id: number) {
    activeTabId = id;

    const tab = tabs.find((tab) => tab.id === id);

    if (tab && tab.hasSignal) {
      tab.hasSignal = false;
      tabs = [...tabs];
    }

    dispatch('tabChanged', id);
  }
</script>

<div class="bg-white border-black border-y-4 md:border-4 md:rounded-2xl mb-4 overflow-hidden">
  <div class="flex justify-between border-black border-b-4">
    <h5 class="ms-5 flex items-center text-base font-bold md:text-2xl">{text}</h5>
    <ul class="flex justify-end">
      {#each tabs as tab}
        <li class="m-0 p-0">
          <button
            class="border-black border-s-4 h-full px-2 text-base text-black font-bold md:text-2xl md:px-10 md:py-4"
            class:active={tab.id == activeTabId}
            on:click={() => setActiveTab(tab.id)}
          >
            <span class="inline-block" class:animate-wiggle={tab.hasSignal}>{tab.name}</span>
          </button>
        </li>
      {/each}
    </ul>
  </div>
  <slot class="min-h-screen" />
</div>

<style lang="scss">
  .active {
    @apply bg-black text-white;
  }
</style>
