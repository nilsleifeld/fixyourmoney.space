<script lang="ts">
  import SymbolOverviewWidget from './base/SymbolOverviewWidget.svelte';
  import Tabs from './base/Tabs.svelte';
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

  const assets = [
    {
      name: 'Apple',
      symbol: 'AAPL',
    },
    {
      name: 'Tesla',
      symbol: 'TSLA',
    },
    {
      name: 'Google',
      symbol: 'GOOG',
    },
    {
      name: 'Amazon',
      symbol: 'AMZN',
    },
    {
      name: 'CocaCola',
      symbol: 'KO',
    },
    {
      name: 'Johnson & Johnson',
      symbol: 'JNJ',
    },
    {
      name: 'Bayer AG',
      symbol: 'BAYRY',
    },
    {
      name: 'Adidas',
      symbol: 'ADDYY',
    },
    {
      name: 'Meta',
      symbol: 'FB',
    },
    {
      name: 'Alphabet',
      symbol: 'GOOGL',
    },
    {
      name: 'MSCI World',
      symbol: 'IRRRF',
    },
    {
      name: 'S&P 500',
      symbol: 'SPY',
    },
    {
      name: 'RWE',
      symbol: 'FWB:RWE',
    },
    {
      name: 'Bayer AG',
      symbol: 'XETR:BAYN',
    },
    {
      name: 'Adidas',
      symbol: 'OTC:ADDYY',
    },
    {
      name: 'Ford',
      symbol: 'NYSE:F',
    },
    {
      name: 'Mastercard',
      symbol: 'NYSE:MA',
    },
    {
      name: 'McDonalds',
      symbol: 'NYSE:MCD',
    },
    {
      name: 'Microsoft',
      symbol: 'NASDAQ:MSFT',
    },
    {
      name: 'Netflix',
      symbol: 'NASDAQ:NFLX',
    },
    {
      name: 'SAP',
      symbol: 'NYSE:SAP',
    },
    {
      name: 'Block',
      symbol: 'NYSE:SQ',
    },
    {
      name: 'Nvidia',
      symbol: 'NASDAQ:NVDA',
    },
    {
      name: 'Microstrategy',
      symbol: 'NASDAQ:MSTR',
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
