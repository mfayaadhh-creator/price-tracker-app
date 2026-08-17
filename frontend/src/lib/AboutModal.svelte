<script>
  import Icon from "./Icon.svelte";

  export let isOpen = false;
  export let onClose;

  const githubUrl = "https://github.com/mfayaadhh-creator/price-tracker-app";
  const botUrl = "https://t.me/mf_pricetracker_bot";
</script>

{#if isOpen}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
  <div 
    class="modal-backdrop" 
    role="presentation" 
    on:click={onClose}
  >
    <div 
      class="modal-card pixel-box" 
      role="dialog" 
      aria-modal="true" 
      tabindex="-1"
      on:click|stopPropagation
    >
      <!-- Modal Header -->
      <div class="modal-header">
        <div class="header-title font-pixel">
          <Icon name="info" size={16} color="#FF5722" />
          <span>TENTANG // PRICE_TRACKER v2.0</span>
        </div>
        <button type="button" class="close-btn" on:click={onClose} title="Tutup">
          <Icon name="close" size={14} />
        </button>
      </div>

      <!-- Modal Body -->
      <div class="modal-body font-mono">
        <section class="info-section">
          <h3 class="info-heading font-display">
            Apa itu PRICE_TRACKER?
          </h3>
          <p class="info-paragraph">
            <strong>PRICE_TRACKER</strong> adalah aplikasi pemantau harga dan pemburu diskon otomatis untuk berbagai toko online (e-commerce). Sistem memantau perubahan harga produk secara berkala dan langsung mengirimkan pesan ke Telegram Anda ketika harga turun atau mencapai target budget!
          </p>
        </section>

        <!-- Feature Grid -->
        <div class="features-grid">
          <div class="feature-card">
            <div class="feature-icon">
              <Icon name="laptop" size={18} color="#FF5722" />
            </div>
            <div class="feature-content">
              <strong>Multi-Store Universal</strong>
              <p>Mendukung toko elektronik (Agres.id, dll), fashion (Uniqlo, Zara, Zalora, H&M), hingga toko berbasis Shopify & WooCommerce.</p>
            </div>
          </div>

          <div class="feature-card">
            <div class="feature-icon">
              <Icon name="send" size={18} color="var(--accent-blue)" />
            </div>
            <div class="feature-content">
              <strong>Alert Telegram Instan</strong>
              <p>Notifikasi real-time langsung ke akun Telegram pribadi tanpa delay, lengkap dengan rincian total uang yang Anda hemat.</p>
            </div>
          </div>

          <div class="feature-card">
            <div class="feature-icon">
              <Icon name="tag" size={18} color="#00E676" />
            </div>
            <div class="feature-content">
              <strong>Target Budget Alerts</strong>
              <p>Atur batas budget maksimal Anda. Bot akan memprioritaskan peringatan khusus saat harga menyentuh target.</p>
            </div>
          </div>

          <div class="feature-card">
            <div class="feature-icon">
              <Icon name="lock" size={18} color="#A855F7" />
            </div>
            <div class="feature-content">
              <strong>Privat & Tanpa Password</strong>
              <p>Login aman menggunakan Deep-Link Bot Telegram atau Chat ID Anda. Tidak ada password atau data sensitif yang disimpan.</p>
            </div>
          </div>
        </div>

        <!-- Links CTA -->
        <div class="links-footer">
          <a 
            href={githubUrl} 
            target="_blank" 
            rel="noopener noreferrer" 
            class="pixel-btn pixel-btn-dark link-btn"
          >
            <Icon name="github" size={16} />
            <span>GITHUB REPOSITORY</span>
            <Icon name="external-link" size={13} />
          </a>

          <a 
            href={botUrl} 
            target="_blank" 
            rel="noopener noreferrer" 
            class="pixel-btn pixel-btn-blue link-btn"
          >
            <Icon name="send" size={16} />
            <span>BOT TELEGRAM RESMI</span>
            <Icon name="external-link" size={13} />
          </a>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background-color: rgba(15, 15, 18, 0.85);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 350;
    padding: 20px;
    animation: fadeIn 0.2s ease-out;
  }

  .modal-card {
    background-color: var(--bg-card);
    max-width: 580px;
    width: 100%;
    border-radius: 0;
    box-shadow: 10px 10px 0px #18181B;
    overflow: hidden;
    animation: scaleUp 0.22s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .modal-header {
    background-color: var(--bg-card-dark);
    color: #FFFFFF;
    padding: 14px 18px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    border-bottom: 2.5px solid var(--border-color);
  }

  .header-title {
    font-size: 0.88rem;
    display: flex;
    align-items: center;
    gap: 8px;
    letter-spacing: 0.5px;
  }

  .close-btn {
    background: transparent;
    border: none;
    color: #FFFFFF;
    cursor: pointer;
    padding: 4px 8px;
    font-weight: bold;
    display: flex;
    align-items: center;
  }

  .close-btn:hover {
    color: var(--accent-red);
  }

  .modal-body {
    padding: 24px;
    max-height: 80vh;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .info-heading {
    font-size: 1.15rem;
    font-weight: 800;
    color: var(--text-main);
    margin-bottom: 8px;
  }

  .info-paragraph {
    font-size: 0.88rem;
    line-height: 1.6;
    color: #3F3F46;
  }

  .features-grid {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .feature-card {
    display: flex;
    gap: 12px;
    background-color: var(--bg-canvas);
    border: 1.5px solid var(--border-color);
    padding: 12px 14px;
    align-items: flex-start;
  }

  .feature-icon {
    background-color: #FFFFFF;
    border: 1.5px solid var(--border-color);
    padding: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 2px 2px 0px var(--border-color);
  }

  .feature-content strong {
    display: block;
    font-size: 0.85rem;
    color: var(--text-main);
    margin-bottom: 2px;
  }

  .feature-content p {
    font-size: 0.78rem;
    line-height: 1.45;
    color: #52525B;
    margin: 0;
  }

  .links-footer {
    display: flex;
    gap: 12px;
    margin-top: 8px;
  }

  .link-btn {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 10px 14px;
    font-size: 0.78rem;
    text-decoration: none;
  }

  @media (max-width: 600px) {
    .links-footer {
      flex-direction: column;
    }
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @keyframes scaleUp {
    from { transform: scale(0.95); opacity: 0; }
    to { transform: scale(1); opacity: 1; }
  }
</style>
