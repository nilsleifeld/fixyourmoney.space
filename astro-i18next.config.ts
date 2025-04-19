import type { AstroI18nextConfig } from "astro-i18next";

const config: AstroI18nextConfig = {
  defaultLocale: "en",
  locales: ['en', 'de', 'ar', 'cs', 'es', 'fr', 'hi', 'it', 'ja', 'ko', 'pl', 'pt', 'ru', 'tr'],
  i18nextServer: {
    debug: true,
  },
};

export default config;
