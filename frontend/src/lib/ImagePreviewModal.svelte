<script>
  import { onMount, onDestroy } from "svelte";
  import Icon from "./Icon.svelte";

  export let data = null; // { name, image_url, price, base_price, url, platform }
  export let onClose;

  function handleKeydown(e) {
    if (e.key === "Escape") {
      onClose();
    }
  }

  function formatIDR(amount) {
    if (!amount && amount !== 0) return "Rp 0";
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      maximumFractionDigits: 0
    }).format(amount);
  }

  onMount(() => {
    window.addEventListener("keydown", handleKeydown);
  });

  onDestroy(() => {
    window.removeEventListener("keydown", handleKeydown);
  });
</script>

{#if data}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
  <div 
    class="lightbox-backdrop" 
    role="presentation" 
    on:click={onClose}
  >
    <div 
      class="lightbox-card pixel-box" 
      role="dialog" 
      aria-modal="true" 
      tabindex="-1"
      on:click|stopPropagation
    >
      <!-- Header -->
      <div class="lightbox-header">
        <div class="header-left font-mono">
          <span class="platform-tag font-pixel">{data.platform || "E-COMMERCE"}</span>
          <span class="header-id font-mono">PHOTO_PREVIEW // HD</span>
        </div>
        <button type="button" class="close-btn font-mono" on:click={onClose} title="Tutup (Esc)">
          <Icon name="close" size={12} />
          <span>ESC</span>
        </button>
      </div>

      <!-- Main Image Display -->
      <div class="image-container">
        <img 
          src={data.image_url} 
          alt={data.name} 
          class="lightbox-img"
        />
      </div>

      <!-- Footer Info -->
      <div class="lightbox-footer">
        <div class="footer-details">
          <h3 class="product-title font-display">{data.name}</h3>
          <div class="pricing-badge font-mono font-pixel">
            <span>HARGA: {formatIDR(data.last_price || data.base_price)}</span>
            {#if data.base_price > data.last_price && data.last_price > 0}
              <span class="savings-tag">
                <Icon name="trending-down" size={13} color="#00E676" /> HEMAT {formatIDR(data.base_price - data.last_price)}
              </span>
            {/if}
          </div>
        </div>

        {#if data.url}
          <a 
            href={data.url} 
            target="_blank" 
            rel="noopener noreferrer" 
            class="pixel-btn pixel-btn-orange font-mono visit-btn"
          >
            <span>BUKA DI WEB TOKO</span>
            <Icon name="external-link" size={13} />
          </a>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .lightbox-backdrop {
    position: fixed;
    inset: 0;
    background-color: rgba(15, 15, 18, 0.85);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 300;
    padding: 24px;
    animation: fadeIn 0.2s ease-out;
  }

  .lightbox-card {
    background-color: var(--bg-card);
    max-width: 580px;
    width: 100%;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    border-radius: 0;
    box-shadow: 10px 10px 0px #18181B;
    overflow: hidden;
    animation: scaleUp 0.22s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .lightbox-header {
    background-color: var(--bg-card-dark);
    color: #FFFFFF;
    padding: 12px 18px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 2px solid var(--border-color);
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .platform-tag {
    background-color: var(--accent-orange);
    color: #FFFFFF;
    font-size: 0.7rem;
    padding: 2px 6px;
  }

  .header-id {
    font-size: 0.75rem;
    color: #A1A1AA;
  }

  .close-btn {
    background: #27272A;
    border: 1px solid #3F3F46;
    color: #FFFFFF;
    padding: 4px 10px;
    font-size: 0.75rem;
    font-weight: 700;
    cursor: pointer;
    transition: background-color 0.1s ease;
  }

  .close-btn:hover {
    background-color: #DC2626;
    border-color: #DC2626;
  }

  .image-container {
    background-color: #F4F4F5;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    max-height: 58vh;
    overflow: hidden;
    position: relative;
  }

  .lightbox-img {
    max-width: 100%;
    max-height: 52vh;
    object-fit: contain;
    border: 2px solid var(--border-color);
    box-shadow: 4px 4px 0px var(--border-color);
    background-color: #FFFFFF;
  }

  .lightbox-footer {
    padding: 18px 20px;
    background-color: var(--bg-card);
    border-top: 2px solid var(--border-color);
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 16px;
    flex-wrap: wrap;
  }

  .footer-details {
    display: flex;
    flex-direction: column;
    gap: 6px;
    flex: 1;
    min-width: 240px;
  }

  .product-title {
    font-size: 1.05rem;
    font-weight: 700;
    color: var(--text-main);
    line-height: 1.3;
  }

  .pricing-badge {
    font-size: 0.95rem;
    font-weight: 800;
    color: var(--accent-orange);
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .savings-tag {
    background-color: var(--accent-yellow);
    color: #854D0E;
    font-size: 0.75rem;
    padding: 2px 6px;
    border: 1px solid var(--border-color);
  }

  .visit-btn {
    padding: 10px 16px;
    font-size: 0.85rem;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @keyframes scaleUp {
    from {
      opacity: 0;
      transform: scale(0.92) translateY(10px);
    }
    to {
      opacity: 1;
      transform: scale(1) translateY(0);
    }
  }

  @media (max-width: 640px) {
    .lightbox-backdrop {
      padding: 12px;
    }
    .lightbox-footer {
      flex-direction: column;
      align-items: stretch;
    }
    .visit-btn {
      width: 100%;
    }
  }
</style>
