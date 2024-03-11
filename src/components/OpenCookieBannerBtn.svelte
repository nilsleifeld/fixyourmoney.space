<script lang="ts">
  import { isBannerOpen } from './cookie';

  function openBanner() {
    isBannerOpen.set(true);
  }
</script>

<button on:click={openBanner}>Cookies</button>
