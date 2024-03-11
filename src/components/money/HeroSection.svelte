<script lang="ts">
  import { slide } from 'svelte/transition';
  import { onMount } from 'svelte';
  import SymbolOverviewWidget from '../base/SymbolOverviewWidget.svelte';

  export let headlineFirst: string;
  export let headlineSecond: string;
  export let dollarInBitcoin: string;
  export let bitcoinInDollar: string;

  const assetsBitcoinInDollar = [
    {
      name: 'Bitcoin in Dollar',
      symbol: 'BTCUSD|ALL',
    },
  ];

  const assetsDollarInBitcoin = [
    {
      name: 'Dollar in Bitcoin',
      symbol: 'USD/BTCUSD|ALL',
    },
  ];

  let dollarInBitcoinChartIsVisible = false;
  let chartToggleButtonDisabled = false;

  function toggleChart() {
    dollarInBitcoinChartIsVisible = !dollarInBitcoinChartIsVisible;
  }

  onMount(() => {
    window.addEventListener('scroll', handleScroll);

    return () => {
      window.removeEventListener('scroll', handleScroll);
    };
  });

  function handleScroll() {
    dollarInBitcoinChartIsVisible = window.scrollY > 0;
    chartToggleButtonDisabled = window.scrollY > 0;
  }
</script>

---

<div class="relative h-[calc(100vh-5rem)] xl:h-screen">
  <div
    class="slide absolute xl:fixed bottom-0 w-full h-1/2 xl:h-full xl:pt-32 -z-10"
    class:show={!dollarInBitcoinChartIsVisible}
  >
    <SymbolOverviewWidget
      chartOnly={true}
      noTimeScale={true}
      assets={assetsBitcoinInDollar}
      lineType={2}
      backgroundColor="rgba(255, 255, 255, 0)"
      copyright={false}
      hideDateRanges={true}
      scalePosition={'no'}
      chartType={'area'}
      reloadable={false}
    />
  </div>
  <div
    class="slide absolute xl:fixed bottom-0 w-full h-1/2 xl:h-full xl:pt-32 -z-10"
    class:show={dollarInBitcoinChartIsVisible}
  >
    <SymbolOverviewWidget
      chartOnly={true}
      noTimeScale={true}
      assets={assetsDollarInBitcoin}
      lineType={2}
      backgroundColor="rgba(255, 255, 255, 0)"
      copyright={false}
      hideDateRanges={true}
      scalePosition={'no'}
      chartType={'area'}
      reloadable={false}
    />
  </div>
  <div
    class="container mx-auto max-w-4xl w-full h-3/4 md:h-5/6 flex flex-col justify-center items-center text-3xl md:text-6xl"
  >
    <h1
      class="font-bold mb-4 text-center transition-all"
      class:font-normal={dollarInBitcoinChartIsVisible}
      class:mb-0={dollarInBitcoinChartIsVisible}
    >
      {headlineFirst}
      {#if !dollarInBitcoinChartIsVisible}
        <span class="-m-2">?</span>
      {/if}

      {#if !dollarInBitcoinChartIsVisible}
        <span>🚀</span>
      {/if}
      <br />
    </h1>
    {#if dollarInBitcoinChartIsVisible}
      <h1 class="font-bold mb-4 text-center" transition:slide>{headlineSecond}?</h1>
    {/if}

    <span class="relative inline-flex">
      <button
        disabled={chartToggleButtonDisabled}
        on:click={toggleChart}
        class:animate-wiggle={!dollarInBitcoinChartIsVisible}
        class="flex justify-center items-center transition-colors border-black border-4 rounded-xl group hover:text-white hover:bg-black disabled:bg-white disabled:text-black disabled:border-white"
      >
        <span class="flex items-center border-r-0 px-4 py-2 font-bold h-full text-base md:text-xl">
          {#if dollarInBitcoinChartIsVisible}
            <span>{dollarInBitcoin}</span>
          {:else}
            <span>{bitcoinInDollar}</span>
          {/if}
        </span>
        <span
          class="flex items-center border-black border-s-4 px-4 py-2 group-hover:border-white group-disabled:border-white"
        >
          <i class="bi bi-arrow-repeat text-xl transition-transform" class:rotate-90={dollarInBitcoinChartIsVisible} />
        </span>
      </button>
    </span>
  </div>
</div>

<!--
<div class="container mx-auto max-w-4xl xl:mb-10">
  <div class="text-3xl h-[calc(100vh-5rem)] xl:h-screen flex flex-col items-center justify-center md:text-6xl">
    <h1
      class="font-bold mb-4 text-center transition-all"
      class:font-normal={dollarInBitcoinChartIsVisible}
      class:mb-0={dollarInBitcoinChartIsVisible}
    >
      {headlineFirst}
      {#if !dollarInBitcoinChartIsVisible}
        <span class="-m-2">?</span>
      {/if}

      {#if !dollarInBitcoinChartIsVisible}
        <span>🚀</span>
      {/if}
      <br />
    </h1>
    {#if dollarInBitcoinChartIsVisible}
      <h1 class="font-bold mb-4 text-center" transition:slide>{headlineSecond}?</h1>
    {/if}

    <span class="relative inline-flex">
      <button
        disabled={chartToggleButtonDisabled}
        on:click={toggleChart}
        class:animate-wiggle={!dollarInBitcoinChartIsVisible}
        class="flex justify-center items-center transition-colors border-black border-4 rounded-xl group hover:text-white hover:bg-black disabled:bg-white disabled:text-black disabled:border-white"
      >
        <span class="flex items-center border-r-0 px-4 py-2 font-bold h-full text-base md:text-xl">
          {#if dollarInBitcoinChartIsVisible}
            <span>{dollarInBitcoin}</span>
          {:else}
            <span>{bitcoinInDollar}</span>
          {/if}
        </span>
        <span
          class="flex items-center border-black border-s-4 px-4 py-2 group-hover:border-white group-disabled:border-white"
        >
          <i class="bi bi-arrow-repeat text-xl transition-transform" class:rotate-90={dollarInBitcoinChartIsVisible} />
        </span>
      </button>
    </span>
  </div>
</div>
-->

<style>
  .slide {
    transition: all 0.3s ease-in-out;
    transform: translateY(100%);
    opacity: 0;
  }

  .slide.show {
    transform: translateY(0%);
    opacity: 1;
  }
</style>
