<script>
  export let totalTracks = 0;
  export let isSyncing = false;
  export let onSync;
  export let currentUser = null;
  export let onOpenLogin;
  export let onLogout;

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
        // MULTI-STORE PRICE TRACKER & REAL-TIME TELEGRAM ALERT ENGINE
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

      <!-- Auth Action Button (Login or User Profile) -->
      {#if currentUser}
        <div class="user-profile-badge font-mono">
          <span class="user-avatar">👤</span>
          <span class="user-name">
            <strong>{currentUser.first_name || "User"}</strong>
            <small>({currentUser.user_phone})</small>
          </span>
          <button 
            type="button" 
            class="logout-btn" 
            on:click={onLogout}
            title="Keluar dari akun"
          >
            [LOGOUT]
          </button>
        </div>
      {:else}
        <button 
          type="button" 
          class="pixel-btn pixel-btn-dark font-mono"
          on:click={onOpenLogin}
          title="Masuk menggunakan akun Telegram Anda"
        >
          <span>🔐</span>
          <span>LOGIN VIA TELEGRAM</span>
        </button>
      {/if}

      <!-- Manual Cron Trigger Button -->
      <button 
        on:click={onSync} 
        disabled={isSyncing} 
        class="pixel-btn pixel-btn-orange font-mono"
        title="Jalankan evaluasi harga ke seluruh produk"
      >
        <span>⚡</span>
        <span>{isSyncing ? "SYNCING..." : "SYNC CRON"}</span>
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

  .user-profile-badge {
    display: flex;
    align-items: center;
    gap: 8px;
    background-color: var(--accent-blue-light);
    border: 2px solid var(--border-color);
    padding: 6px 12px;
    box-shadow: 2px 2px 0px var(--border-color);
    font-size: 0.8rem;
  }

  .user-avatar {
    font-size: 1rem;
  }

  .user-name {
    display: flex;
    flex-direction: column;
    line-height: 1.1;
  }

  .user-name small {
    font-size: 0.68rem;
    color: var(--text-muted);
  }

  .logout-btn {
    background: transparent;
    border: none;
    color: #DC2626;
    font-family: inherit;
    font-weight: 800;
    font-size: 0.75rem;
    cursor: pointer;
    margin-left: 4px;
    padding: 2px 4px;
  }

  .logout-btn:hover {
    text-decoration: underline;
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
