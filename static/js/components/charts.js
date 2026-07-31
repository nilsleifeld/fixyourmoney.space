import {getConsent} from './consent.js';

const WIDGET_SOURCE = 'https://s3.tradingview.com/external-embedding/embed-widget-symbol-overview.js';

const stockAssets = [
  ['Adidas', 'ADDYY'], ['Alphabet', 'GOOGL'], ['Amazon', 'AMZN'], ['Apple', 'AAPL'], ['Bayer AG', 'BAYRY'],
  ['Block', 'NYSE:SQ'], ['CocaCola', 'KO'], ['Ford', 'NYSE:F'], ['Johnson & Johnson', 'JNJ'], ['Mastercard', 'NYSE:MA'],
  ['McDonalds', 'NYSE:MCD'], ['Meta', 'META'], ['Microsoft', 'NASDAQ:MSFT'], ['Microstrategy', 'NASDAQ:MSTR'],
  ['MSCI World', 'IRRRF'], ['Netflix', 'NASDAQ:NFLX'], ['Nvidia', 'NASDAQ:NVDA'], ['RWE', 'FWB:RWE'],
  ['S&P 500', 'SPY'], ['SAP', 'NYSE:SAP'], ['Tesla', 'TSLA'],
];

const dollarUnit = 'FX_IDC:USDEUR/USDEUR';

const currencyAssets = [
  ['Dollar', dollarUnit], ['Euro', 'FX:EURUSD'], ['Pound', 'FX:GBPUSD'], ['Yen', 'FX_IDC:JPYUSD'],
  ['Ruble', 'FX_IDC:RUBUSD'], ['Yuan', 'FX_IDC:CNYUSD'], ['Franc', 'FX_IDC:CHFUSD'], ['Lira', 'FX_IDC:TRYUSD'],
];

const candidateAssets = [
  ['Dollar in Gold', `${dollarUnit}/XAUUSD|ALL`], ['Dollar in Bitcoin', `${dollarUnit}/BTCUSD|ALL`],
  ['Gold in Dollar', 'XAUUSD|ALL'], ['Gold in Bitcoin', 'XAUUSD/BTCUSD|ALL'],
  ['Bitcoin in Dollar', 'BTCUSD|ALL'], ['Bitcoin in Gold', 'BTCUSD/XAUUSD|ALL'],
];

/**
 * Converts base symbols into a selected unit of account.
 * @param {string[][]} assets
 * @param {'dollar'|'gold'|'bitcoin'} unit
 * @returns {string[][]}
 */
function inUnit(assets, unit) {
  const suffixes = {dollar: '|ALL', gold: '/XAUUSD|ALL', bitcoin: '/BTCUSD|ALL'};
  return assets.map(([name, symbol]) => {
    // Special case: TradingView can introduce fluctuations via the regular calculation, but Dollar in Dollar must be a flat line at 1.
    if (symbol === dollarUnit && unit === 'dollar') {
      return [name, 'USD/USD|ALL'];
    }
    return [name, `${symbol}${suffixes[unit]}`];
  });
}

const heroChartOptions = {
  hideDateRanges: true,
  isTransparent: true,
  noTimeScale: true,
  scalePosition: 'no',
};

const chartModels = {
  'hero-bitcoin-in-dollar': {symbols: [['Bitcoin in Dollar', 'BTCUSD|ALL']], options: heroChartOptions},
  'hero-dollar-in-bitcoin': {symbols: [['Dollar in Bitcoin', 'USD/BTCUSD|ALL']], options: heroChartOptions},
  'assets-dollar': {symbols: inUnit(stockAssets, 'dollar')},
  'assets-gold': {symbols: inUnit(stockAssets, 'gold')},
  'assets-bitcoin': {symbols: inUnit(stockAssets, 'bitcoin'), scaleMode: 'Logarithmic'},
  'currencies-dollar': {symbols: inUnit(currencyAssets, 'dollar')},
  'currencies-gold': {symbols: inUnit(currencyAssets, 'gold')},
  'currencies-bitcoin': {symbols: inUnit(currencyAssets, 'bitcoin'), scaleMode: 'Logarithmic'},
  'candidates-linear': {symbols: candidateAssets, dateRanges: ['all|1M']},
  'candidates-log': {symbols: candidateAssets, dateRanges: ['all|1M'], scaleMode: 'Logarithmic'},
};

/**
 * Maps site locales to locales supported by the TradingView embed.
 * @param {string} locale
 * @returns {string}
 */
function tradingViewLocale(locale) {
  const locales = {
    ar: 'ar_AE',
    cs: 'cs_CZ',
    de: 'de_DE',
    en: 'en',
    es: 'es',
    fr: 'fr',
    it: 'it',
    ja: 'ja',
    ko: 'ko',
    pl: 'pl',
    pt: 'pt',
    ru: 'ru',
    tr: 'tr',
  };
  // TradingView does not offer a Hindi widget locale; unknown locales use English.
  return locales[locale] || 'en';
}

/**
 * Reads a chart color from the active CSS design tokens.
 * @param {string} token
 * @returns {string}
 */
function chartToken(token) {
  return window.getComputedStyle(document.documentElement).getPropertyValue(token).trim();
}

/**
 * Creates the TradingView palette from the shared chart design tokens.
 * @returns {{backgroundColor: string, gridLineColor: string, lineColor: string, topColor: string, bottomColor: string, fontColor: string, widgetFontColor: string}}
 */
function chartPalette() {
  const areaColor = chartToken('--chart-widget-area-color');
  const textColor = chartToken('--chart-widget-text-color');

  return {
    backgroundColor: chartToken('--chart-widget-background-color'),
    gridLineColor: chartToken('--chart-widget-grid-color'),
    lineColor: chartToken('--chart-widget-line-color'),
    topColor: areaColor,
    bottomColor: areaColor,
    fontColor: textColor,
    widgetFontColor: textColor,
  };
}

/**
 * Builds one official TradingView widget configuration.
 * @param {{symbols: string[][], scaleMode?: string, dateRanges?: string[], options?: Record<string, unknown>}} model
 * @param {string} locale
 * @returns {Record<string, unknown>}
 */
function widgetConfiguration(model, locale) {
  return {
    symbols: model.symbols,
    chartOnly: true,
    width: '100%',
    height: '100%',
    autosize: true,
    locale: tradingViewLocale(locale),
    colorTheme: 'light',
    scaleMode: model.scaleMode || 'Normal',
    dateRanges: model.dateRanges || ['12m|1D', '60m|1W', '120m|1W', 'all|1M'],
    ...chartPalette(),
    showVolume: false,
    showMA: false,
    hideSymbolLogo: true,
    lineWidth: 3,
    ...model.options,
  };
}

/**
 * Updates a chart's live status text.
 * @param {HTMLElement} panel
 * @param {'optIn'|'loading'|'error'|'noData'|'loaded'} state
 * @returns {void}
 */
function setStatus(panel, state) {
  const status = panel.querySelector('[data-chart-status]');
  if (!(status instanceof HTMLElement)) return;
  const keys = {optIn: 'statusOptIn', loading: 'statusLoading', error: 'statusError', noData: 'statusNoData'};
  status.textContent = state === 'loaded' ? '' : panel.dataset[keys[state]] || '';
  status.hidden = state === 'loaded';
}

/**
 * Removes a loaded widget so revoked consent cannot create more provider activity.
 * @param {HTMLElement} panel
 * @returns {void}
 */
function unloadChart(panel) {
  const mount = panel.querySelector('[data-chart-mount]');
  if (mount instanceof HTMLElement) mount.replaceChildren();
  delete panel.dataset.loaded;
  const consentButton = panel.querySelector('[data-consent-open]');
  if (consentButton instanceof HTMLElement) consentButton.hidden = false;
  setStatus(panel, 'optIn');
}

/**
 * Loads one visible widget at most once and reports load failures accessibly.
 * @param {HTMLElement} panel
 * @param {string} locale
 * @returns {void}
 */
function loadChart(panel, locale) {
  if (panel.hidden || panel.dataset.loaded === 'true') return;
  const model = chartModels[panel.dataset.chartKey || ''];
  if (!model) {
    setStatus(panel, 'noData');
    return;
  }
  const mount = panel.querySelector('[data-chart-mount]');
  if (!(mount instanceof HTMLElement)) return;

  panel.dataset.loaded = 'true';
  const consentButton = panel.querySelector('[data-consent-open]');
  if (consentButton instanceof HTMLElement) consentButton.hidden = true;
  setStatus(panel, 'loading');

  const script = document.createElement('script');
  script.async = true;
  script.src = WIDGET_SOURCE;
  script.textContent = JSON.stringify(widgetConfiguration(model, locale));
  script.addEventListener('load', () => setStatus(panel, 'loaded'), {once: true});
  script.addEventListener('error', () => {
    script.remove();
    delete panel.dataset.loaded;
    setStatus(panel, 'error');
  }, {once: true});
  mount.appendChild(script);
}

/**
 * Synchronizes every chart with visibility and external-media consent.
 * @param {{externalMedia?: boolean}|null} consent
 * @returns {void}
 */
function syncCharts(consent) {
  document.querySelectorAll('[data-chart-group]').forEach((group) => {
    if (!(group instanceof HTMLElement)) return;
    const locale = group.dataset.chartLocale || 'en';
    group.querySelectorAll('[data-chart-panel]').forEach((panel) => {
      if (!(panel instanceof HTMLElement)) return;
      if (!consent?.externalMedia) unloadChart(panel);
      else loadChart(panel, locale);
    });
  });
}

/**
 * Initializes lazy, consent-gated TradingView widgets.
 * @returns {void}
 */
export function initCharts() {
  let consent = getConsent();
  syncCharts(consent);
  document.querySelectorAll('[data-chart-group]').forEach((group) => {
    group.addEventListener('tabchange', () => syncCharts(consent));
    group.addEventListener('chartchange', () => syncCharts(consent));
  });
  window.addEventListener('consentchange', (event) => {
    const consentEvent = /** @type {CustomEvent} */ (event);
    consent = consentEvent.detail;
    syncCharts(consent);
  });
}
