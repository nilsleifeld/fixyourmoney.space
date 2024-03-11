<script lang="ts">
  import { onMount } from 'svelte';
  import { acceptOnlyRequired, isOnlyRequiredAccepted } from '../cookie';

  interface Asset {
    name: string;
    symbol: string;
  }

  export let assets: Asset[] = [
    {
      name: 'Apple',
      symbol: 'AAPL/BTCUSD|ALL',
    },
    {
      name: 'Tesla',
      symbol: 'TSLA|ALL',
    },
    {
      name: 'Bitcoin',
      symbol: 'BITSTAMP:BTCUSD|ALL',
    },
    {
      name: 'Google',
      symbol: 'GOOG|ALL',
    },
    {
      name: 'Amazon',
      symbol: 'AMZN|ALL',
    },
    {
      name: 'Facebook',
      symbol: 'FB|ALL',
    },
  ];
  export let chartOnly: boolean = false;
  export let width: string = '100%';
  export let height: string = '100%';
  export let locale: 'en' | 'de_DE' = 'en';
  export let colorTheme: 'light' | 'dark' = 'light';
  export let gridLineColor = 'rgba(0, 0, 0, 0)';
  export let autosize: boolean = true;
  export let showVolume: boolean = false;
  export let showMA: boolean = false;
  export let hideDateRanges: boolean = false;
  export let hideMarketStatus: boolean = false;
  export let hideSymbolLogo: boolean = false;
  export let scalePosition: 'right' | 'left' | 'no' = 'right';
  export let scaleMode: 'Normal' | 'Percentage' | 'Logarithmic' = 'Normal';
  export let fontFamily: string = '-apple-system, BlinkMacSystemFont, Trebuchet MS, Roboto, Ubuntu, sans-serif';
  export let fontSize: string = '10';
  export let noTimeScale: boolean = false;
  export let valuesTracking: string = '1';
  export let changeMode: string = 'price-and-percent';
  export let chartType: 'line' | 'area' = 'line';
  export let lineWidth: 1 | 2 | 3 | 4 = 3;
  export let lineType: 1 | 2 = 2;
  export let dateRanges: string[] = ['12m|1D', '60m|1W', '120m|1W', 'all|1M'];
  export let copyright = true;
  export let backgroundColor = 'rgba(255, 255, 255, 1)';
  export let id: string | null = null;
  export let color: string = '#000';
  export let reloadable: boolean = true;

  if (!id) {
    id = `tradingview-widget-${Math.random().toString(36).substr(2, 9)}`;
  }

  onMount(() => {
    isOnlyRequiredAccepted.subscribe((accepted: boolean) => {
      if (accepted) {
        loadWidget();
      } else {
        removeWidget();
      }
    });
  });

  function removeWidget() {
    const container = document.getElementById(id!);

    if (!container) {
      console.error('TradingView Widget: Container not found');
      return;
    }

    let iframe = container.getElementsByTagName('iframe')[0];

    if (iframe) {
      iframe.remove();
    }
  }

  function loadWidget() {
    const container = document.getElementById(id!);

    if (!container) {
      console.error('TradingView Widget: Container not found');
      return;
    }

    let iframe = container.getElementsByTagName('iframe')[0];

    if (iframe) {
      iframe.remove();
    }

    const script = document.createElement('script');
    script.src = 'https://s3.tradingview.com/external-embedding/embed-widget-symbol-overview.js';
    script.type = 'text/javascript';
    script.async = true;
    script.innerHTML = `
        {
          "symbols": [${assets
            .map(
              (asset) =>
                `[
                "${asset.name}",
                "${asset.symbol}"
            ]`
            )
            .join(',')}],
            "chartOnly": ${chartOnly},
            "width": "${width}",
            "height": "${height}",
            "color": "${color}",
            "locale": "${locale}",
            "colorTheme": "${colorTheme}",
            "autosize": ${autosize},
            "showVolume": ${showVolume},
            "showMA": ${showMA},
            "hideDateRanges": ${hideDateRanges},
            "hideMarketStatus": ${hideMarketStatus},
            "hideSymbolLogo": ${hideSymbolLogo},
            "scalePosition": "${scalePosition}",
            "scaleMode": "${scaleMode}",
            "fontFamily": "${fontFamily}",
            "fontSize": "${fontSize}",
            "noTimeScale": ${noTimeScale},
            "valuesTracking": "${valuesTracking}",
            "changeMode": "${changeMode}",
            "chartType": "${chartType}",
            "gridLineColor": "${gridLineColor}",
            "backgroundColor": "${backgroundColor}",
            "lineWidth": ${lineWidth},
            "lineType": ${lineType},
            "dateRanges": [
              ${dateRanges.map((range) => `"${range}"`).join(',')}
            ],
            "lineColor": "rgba(0, 0, 0, 1)",
            "topColor": "rgba(0, 0, 0, 1)",
            "bottomColor": "rgba(0, 0, 0, 1)",
            "downColor": "rgba(0, 0, 0, 1)",
            "upColor": "rgba(0, 0, 0, 1)"
        }`;
    container.appendChild(script);
  }
</script>

{#if !$isOnlyRequiredAccepted && reloadable}
  <div class="flex justify-center items-center flex-col h-full w-full">
    <h3 class="font-bold text-xl mb-4">Zum darstellen der Charts Cookies akzeptieren</h3>
    <button
      on:click={acceptOnlyRequired}
      class="px-4 py-1 font-bold rounded-xl bg-white text-black border-4 transition-colors border-black hover:bg-black hover:text-white"
    >
      Chart erforderliche Cookies akzeptieren
    </button>
  </div>
{/if}

<div class="tradingview-widget-container" {id}>
  <div class="tradingview-widget-container__widget" />
  {#if copyright}
    <div class="tradingview-widget-copyright">
      <a href="https://www.tradingview.com/" rel="noopener" target="_blank"
        ><span class="text-black">by TradingView</span></a
      >
    </div>
  {/if}
</div>
