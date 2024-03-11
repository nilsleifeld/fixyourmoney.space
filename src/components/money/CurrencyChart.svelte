<script lang="ts">
  import SymbolOverviewWidget from './../base/SymbolOverviewWidget.svelte';
  import Tabs from './../base/Tabs.svelte';

  var activeTabId = 0;

  const tabs = [
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

  export let text: string;

  function onChangeTab(event: CustomEvent<number>) {
    activeTabId = event.detail;
  }

  let assets = [
    {
      name: 'Dollar',
      symbol: 'FX_IDC:USDEUR/USDEUR',
    },
    {
      name: 'Euro',
      symbol: 'FX:EURUSD',
    },
    {
      name: 'Pound',
      symbol: 'FX:GBPUSD',
    },
    {
      name: 'Yen',
      symbol: 'FX_IDC:JPYUSD',
    },
    {
      name: 'Ruble',
      symbol: 'FX_IDC:RUBUSD',
    },
    {
      name: 'Yuan',
      symbol: 'FX_IDC:CNYUSD',
    },
    {
      name: 'Franken',
      symbol: 'FX_IDC:CHFUSD',
    },
    {
      name: 'Lira',
      symbol: 'FX_IDC:TRYUSD',
    },
  ].sort((a, b) => a.name.localeCompare(b.name));

  const assetsInDollar = assets.map((asset) => {
    return {
      name: asset.name,
      symbol: `${asset.symbol}|ALL`,
    };
  });

  const assetsInGold = assets.map((asset) => {
    return {
      name: asset.name,
      symbol: `${asset.symbol}/XAUUSD|ALL`,
    };
  });

  const assetsInBtc = assets.map((asset) => {
    return {
      name: asset.name,
      symbol: `${asset.symbol}/BTCUSD|ALL`,
    };
  });
</script>

<Tabs {text} {tabs} on:tabChanged={onChangeTab}>
  <div class="h-72 relative lg:h-[26rem]">
    <div class="h-full top-0 left-0 absolute w-full slide" class:show={activeTabId == 0}>
      <SymbolOverviewWidget assets={assetsInDollar} chartOnly={true} />
    </div>
    <div class="h-full top-0 left-0 absolute w-full slide" class:show={activeTabId == 1}>
      <SymbolOverviewWidget assets={assetsInGold} chartOnly={true} />
    </div>
    <div class="h-full top-0 left-0 absolute w-full slide" class:show={activeTabId == 2}>
      <SymbolOverviewWidget assets={assetsInBtc} chartOnly={true} scaleMode="Logarithmic" />
    </div>
  </div>
</Tabs>

<style>
  .slide {
    transition: transform 0.3s ease-in-out;
    transform: translateX(-100%);
  }

  .slide.show {
    transform: translateX(0%);
  }
</style>
