<script>
  export let totalTracks = 0;
  export let isSyncing = false;
  export let onSync;

  const botUrl = "https://t.me/mf_pricetracker_bot";
</script>

<header class="header-container">
  <div class="header-inner">
    <!-- Brand Logo -->
    <div class="brand-group">
      <div class="brand-badge">
        <span class="pixel-dot"></span>
        <span class="font-pixel brand-title">PRICE_TRACKER</span>
        <span class="version-tag">v1.0</span>
      </div>
      <p class="brand-subtitle font-mono">
        // MONITOR DISKON UNIQLO & AUTOMATED TELEGRAM ALERT ENGINE
      </p>
    </div>

    <!-- Actions & Stats -->
    <div class="actions-group">
      <!-- Status Ticker -->
      <div class="status-ticker font-mono">
        <span class="ticker-item">
          <span class="dot-green"></span>
          TRACKED: <strong>{totalTracks}</strong>
        </span>
        <span class="ticker-divider">/</span>
        <span class="ticker-item">
          BOT: <strong>ACTIVE</strong>
        </span>
      </div>

      <!-- Telegram Direct Link Button -->
      <a 
        href={botUrl} 
        target="_blank" 
        rel="noopener noreferrer" 
        class="pixel-btn pixel-btn-blue font-mono"
        title="Buka bot di aplikasi Telegram"
      >
        <span>🤖</span>
        <span>BUKA BOT TELEGRAM</span>
        <span class="arrow-icon">↗</span>
      </a>

      <!-- Manual Cron Trigger Button -->
      <button 
        on:click={onSync} 
        disabled={isSyncing} 
        class="pixel-btn pixel-btn-orange font-mono"
        title="Jalankan evaluasi harga ke seluruh produk"
      >
        <span>⚡</span>
        <span>{isSyncing ? "MENGECEK HARGA..." : "SYNC HARGA SEKARANG"}</span>
      </button>
    </div>
  </div>
</header>

<style>
  .header-container {
    background-color: var(--bg-card);
    border-bottom: var(--border-width) solid var(--border-color);
    box-shadow: 0 4px 0px rgba(0,0,0,0.06);
    position: sticky;
    top: 0;
    z-index: 50;
  }

  .header-inner {
    max-width: 1280px;
    margin: 0 auto;
    padding: 16px 20px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 16px;
  }

  .brand-group {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .brand-badge {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .pixel-dot {
    width: 12px;
    height: 12px;
    background-color: var(--accent-orange);
    border: 2px solid var(--border-color);
    box-shadow: 2px 2px 0px var(--border-color);
  }

  .brand-title {
    font-size: 1.35rem;
    font-weight: 700;
    color: var(--text-main);
    letter-spacing: -0.5px;
  }

  .version-tag {
    font-family: var(--font-mono);
    font-size: 0.7rem;
    font-weight: 700;
    background-color: var(--accent-yellow);
    color: #854D0E;
    padding: 2px 6px;
    border: 1.5px solid var(--border-color);
    box-shadow: 1px 1px 0px var(--border-color);
  }

  .brand-subtitle {
    font-size: 0.75rem;
    color: var(--text-muted);
    font-weight: 600;
  }

  .actions-group {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  .status-ticker {
    display: flex;
    align-items: center;
    gap: 8px;
    background-color: var(--bg-canvas);
    padding: 8px 12px;
    border: 2px solid var(--border-color);
    font-size: 0.8rem;
    box-shadow: 2px 2px 0px var(--border-color);
  }

  .ticker-item {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .dot-green {
    width: 8px;
    height: 8px;
    background-color: var(--accent-green);
    border-radius: 0;
    border: 1px solid var(--border-color);
  }

  .ticker-divider {
    color: #A1A1AA;
  }

  .arrow-icon {
    font-size: 0.95rem;
  }

  @media (max-width: 768px) {
    .header-inner {
      flex-direction: column;
      align-items: flex-start;
    }
    
    .actions-group {
      width: 100%;
      justify-content: flex-start;
    }

    .status-ticker {
      width: 100%;
      justify-content: space-around;
    }
  }
</style>
