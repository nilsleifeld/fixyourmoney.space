import { initSectionTracking } from './components/analytics.js';
import { initCharts } from './components/charts.js';
import { initConsent } from './components/consent.js';
import { initDialogs } from './components/dialogs.js';
import { initHero } from './components/hero.js';
import { initScrollMotion } from './components/scroll-motion.js';
import { initTabs } from './components/tabs.js';

initConsent();
initDialogs();
initTabs();
initHero();
initCharts();
initScrollMotion();
initSectionTracking();
