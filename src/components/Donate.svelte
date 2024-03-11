<script lang="ts">
  const lnurl = 'LNURL1DP68GURN8GHJ7MRW9E6XJURN9UH8WETVDSKKKMN0WAHZ7MRWW4EXCUP0XPURGVPKVSCXGDECVF3KVD3SX3SNG3XWVEM';

  let isCopied = false;

  function copyToClipboard() {
    navigator.clipboard.writeText(lnurl);
    isCopied = true;

    setTimeout(() => {
      isCopied = false;
    }, 3000);
  }
</script>

<div class="flex justify-center items-center">
  <img alt="LNURL QRcode" src="/img/donate.jpg" />
</div>

<div class=" flex px-5 py-4 justify-center items-center border-black border-4 mb-5 rounded-xl">
  <pre class="overflow-auto break-words whitespace-pre-wrap">{lnurl}</pre>
  <button class="flex justify-center items-center ps-5" on:click={copyToClipboard}>
    {#if isCopied}
      <i class="bi text-2xl bi-clipboard-check-fill" />
    {:else}
      <i class="bi text-2xl bi-clipboard" />
    {/if}
  </button>
</div>
