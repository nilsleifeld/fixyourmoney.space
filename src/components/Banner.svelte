<script lang="ts">
  import { slide } from 'svelte/transition';
  import { isBannerOpen, acceptAll, acceptOnlyRequired, rejectAll } from './cookie';
</script>

{#if $isBannerOpen}
  <div class="z-50 fixed w-full h-full bg-black bg-opacity-50 flex justify-center items-center">
    <div transition:slide class="bg-white p-8 md:p-10 overflow-y-scroll max-h-screen md:max-w-xl rounded-xl">
      <h1 class="text-black mb-5 text-lg md:text-xl font-bold text-center">
        <i class="bi bi-lock-fill" /><span class="ms-4">Privatsphäre-Einstellungen</span>
      </h1>
      <div class="mb-5">
        <p class="text-justify mb-3">
          Wir verwenden Cookies und ähnliche Technologien auf unserer Website und verarbeiten personenbezogene Daten von
          dir (z.B. IP-Adresse), um z.B. Inhalte und Anzeigen zu personalisieren, Medien von Drittanbietern einzubinden
          oder Zugriffe auf unsere Website zu analysieren. Die Datenverarbeitung kann auch erst in Folge gesetzter
          Cookies stattfinden. Wir teilen diese Daten mit Dritten, die wir in den Privatsphäre-Einstellungen benennen.
        </p>
        <p class="text-justify">
          Die Datenverarbeitung kann mit deiner Einwilligung oder auf Basis eines berechtigten Interesses erfolgen, dem
          du in den Privatsphäre-Einstellungen widersprechen kannst. Du hast das Recht, nicht einzuwilligen und deine
          Einwilligung zu einem späteren Zeitpunkt zu ändern oder zu widerrufen. Weitere Informationen zur Verwendung
          deiner Daten findest du in unserer Datenschutzerklärung.
        </p>
      </div>
      <div class="grid grid-cols-1 gap-2">
        <button class="btn btn-primary" on:click={acceptAll}>Alle akzeptieren</button>
        <button class="btn btn-secondary" on:click={acceptOnlyRequired}>Nur zur darstellung Benötiegte</button>
        <button class="btn btn-secondary" on:click={rejectAll}>Weiter ohne Einwilligung</button>
      </div>
    </div>
  </div>
{/if}

<style lang="scss">
  .btn {
    @apply w-full py-3 border-4 rounded-xl transition-all;
  }

  .btn-secondary {
    @apply border-4 border-black hover:scale-105;
  }

  .btn-primary {
    @apply bg-black border-black text-white font-bold hover:scale-105;
  }
</style>
