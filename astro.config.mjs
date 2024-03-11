import { defineConfig } from 'astro/config';
import svelte from '@astrojs/svelte';
import tailwind from '@astrojs/tailwind';
import astroI18next from 'astro-i18next';

import mdx from '@astrojs/mdx';

export default defineConfig({
  integrations: [svelte(), tailwind(), astroI18next(), mdx()],
  site: 'https://nilsleifeld.github.io/fixyourmoney.space',
});
