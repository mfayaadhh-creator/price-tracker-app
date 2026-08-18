<script>
  import Icon from "./Icon.svelte";
  import { onMount } from "svelte";

  export let deferredPrompt = null;
  export let onInstalled;

  let isBannerVisible = false;
  let isDismissed = false;

  $: if (deferredPrompt && !isDismissed) {
    isBannerVisible = true;
  } else {
    isBannerVisible = false;
  }

  async function handleInstall() {
    if (!deferredPrompt) return;
    deferredPrompt.prompt();
    const { outcome } = await deferredPrompt.userChoice;
    if (outcome === "accepted") {
      isBannerVisible = false;
      if (onInstalled) onInstalled();
    }
  }

  function handleDismiss() {
    isDismissed = true;
    isBannerVisible = false;
    try {
      sessionStorage.setItem("pt_pwa_dismissed", "true");
    } catch (e) {}
  }

  onMount(() => {
    try {
      if (sessionStorage.getItem("pt_pwa_dismissed") === "true") {
        isDismissed = true;
      }
    } catch (e) {}
  });
</script>

{#if isBannerVisible}
  <div class="pwa-floating-banner pixel-box font-mono" role="dialog" aria-label="Install Aplikasi">
    <div class="banner-icon-wrap">
      <img src="/favicon.svg" alt="Price Tracker Logo" class="banner-logo-img" />
    </div>
    <div class="banner-text">
      <strong class="banner-title font-pixel">PASANG APLIKASI (PWA)</strong>
      <span class="banner-desc">Buka instan 1-klik di Home Screen tanpa browser bar!</span>
    </div>
    <div class="banner-actions">
      <button type="button" class="pixel-btn pixel-btn-green install-action-btn font-pixel" on:click={handleInstall}>
        INSTALL
      </button>
      <button type="button" class="dismiss-btn" on:click={handleDismiss} title="Tutup">
        <Icon name="close" size={14} />
      </button>
    </div>
  </div>
{/if}

<style>
  .pwa-floating-banner {
    position: fixed;
    bottom: 20px;
    left: 20px;
    right: 20px;
    max-width: 480px;
    margin: 0 auto;
    background-color: var(--bg-card);
    border: var(--border-width) solid var(--border-color);
    box-shadow: 0 8px 0px rgba(0,0,0,0.18);
    padding: 12px 16px;
    display: flex;
    align-items: center;
    gap: 12px;
    z-index: 90;
    animation: slideUp 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes slideUp {
    from {
      transform: translateY(100px);
      opacity: 0;
    }
    to {
      transform: translateY(0);
      opacity: 1;
    }
  }

  .banner-icon-wrap {
    width: 36px;
    height: 36px;
    border-radius: 8px;
    border: 2px solid var(--border-color);
    box-shadow: 2px 2px 0px var(--border-color);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    overflow: hidden;
  }

  .banner-logo-img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .banner-text {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .banner-title {
    font-size: 0.85rem;
    color: var(--text-main);
  }

  .banner-desc {
    font-size: 0.72rem;
    color: var(--text-muted);
    line-height: 1.2;
  }

  .banner-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }

  .install-action-btn {
    padding: 6px 12px;
    font-size: 0.75rem;
    box-shadow: 2px 2px 0px var(--border-color);
  }

  .dismiss-btn {
    background: transparent;
    border: none;
    cursor: pointer;
    color: var(--text-muted);
    padding: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .dismiss-btn:hover {
    color: #DC2626;
  }
</style>
