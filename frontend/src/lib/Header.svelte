<script>
  import Icon from "./Icon.svelte";

  export let totalTracks = 0;
  export let isSyncing = false;
  export let onSync;
  export let currentUser = null;
  export let onOpenLogin;
  export let onLogout;
  export let onOpenAbout;
  export let deferredPrompt = null;

  let isMobileMenuOpen = false;

  $: isInstallable = !!deferredPrompt;

  async function handleInstallPWA() {
    if (!deferredPrompt) return;
    deferredPrompt.prompt();
    const { outcome } = await deferredPrompt.userChoice;
    if (outcome === "accepted") {
      closeMobileMenu();
    }
  }

  const githubUrl = "https://github.com/mfayaadhh-creator/price-tracker-app";
  const botUrl = "https://t.me/mf_pricetracker_bot";

  function closeMobileMenu() {
    isMobileMenuOpen = false;
  }

  function handleAboutClick() {
    closeMobileMenu();
    onOpenAbout();
  }

  function handleLoginClick() {
    closeMobileMenu();
    onOpenLogin();
  }

  function handleLogoutClick() {
    closeMobileMenu();
    onLogout();
  }

  function handleSyncClick() {
    closeMobileMenu();
    onSync();
  }
</script>

<svelte:window on:keydown={(e) => { if (e.key === "Escape") closeMobileMenu(); }} />

<header class="header-container">
  <div class="header-inner">
    <!-- Brand Logo -->
    <div class="brand-group">
      <div class="brand-badge">
        <span class="pixel-dot"></span>
        <span class="font-pixel brand-title">PRICE_TRACKER</span>
        <span class="version-tag">v2.0</span>
      </div>
      <p class="brand-subtitle font-mono">
        // UNIVERSAL E-COMMERCE PRICE MONITOR & AUTOMATED TELEGRAM ALERTS
      </p>
    </div>

    <!-- Desktop Actions & Navigation (Hidden on mobile) -->
    <div class="actions-group desktop-only">
      <!-- Nav Links -->
      <div class="nav-links font-mono">
        <button 
          type="button" 
          class="nav-btn" 
          on:click={onOpenAbout} 
          title="Pelajari cara kerja & website yang didukung"
        >
          <Icon name="info" size={13} />
          <span>INFO</span>
        </button>

        <a 
          href={githubUrl} 
          target="_blank" 
          rel="noopener noreferrer" 
          class="nav-btn" 
          title="Lihat source code di GitHub"
        >
          <Icon name="github" size={13} />
          <span>GITHUB</span>
        </a>

        <a 
          href={botUrl} 
          target="_blank" 
          rel="noopener noreferrer" 
          class="nav-btn nav-btn-tg" 
          title="Buka Bot Telegram Resmi"
        >
          <Icon name="send" size={13} color="var(--accent-blue)" />
          <span>BOT TG</span>
        </a>
      </div>

      <!-- Status Ticker -->
      <div class="status-ticker font-mono">
        <span class="ticker-item">
          <span class="dot-green"></span>
          TRACKED: <strong>{totalTracks}</strong>
        </span>
      </div>

      <!-- Auth Action Button (Login or User Profile) -->
      {#if currentUser}
        <div class="user-profile-badge font-mono">
          <Icon name="user" size={15} color="var(--accent-orange)" />
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
          <Icon name="lock" size={14} color="#00E676" />
          <span>LOGIN TG</span>
        </button>
      {/if}

      <!-- Manual Cron Trigger Button -->
      <button 
        on:click={onSync} 
        disabled={isSyncing} 
        class="pixel-btn pixel-btn-orange font-mono"
        title="Jalankan evaluasi harga ke seluruh produk"
      >
        <Icon name="refresh" size={14} className={isSyncing ? "spin-icon" : ""} />
        <span>{isSyncing ? "SYNCING..." : "SYNC"}</span>
      </button>
    </div>

    <!-- Mobile Top Actions (Login Direct + Hamburger Menu) -->
    <div class="mobile-top-actions mobile-only">
      <!-- Direct Login or Compact Profile Badge on Mobile Top Bar -->
      {#if currentUser}
        <div class="mobile-user-badge font-mono" title="Terhubung sebagai {currentUser.first_name || 'User'}">
          <Icon name="user" size={13} color="var(--accent-orange)" />
          <span class="mobile-user-name">{currentUser.first_name || "User"}</span>
        </div>
      {:else}
        <button 
          type="button" 
          class="pixel-btn pixel-btn-dark mobile-direct-login font-mono"
          on:click={onOpenLogin}
          title="Login langsung via Telegram"
        >
          <Icon name="lock" size={12} color="#00E676" />
          <span>LOGIN TG</span>
        </button>
      {/if}

      <!-- Hamburger Menu Button -->
      <button 
        type="button" 
        class="pixel-btn menu-toggle-btn font-mono" 
        on:click={() => isMobileMenuOpen = !isMobileMenuOpen}
        aria-label="Buka Menu Navigasi"
      >
        <Icon name="menu" size={16} color="var(--text-main)" strokeWidth={2.5} />
      </button>
    </div>
  </div>
</header>

<!-- Mobile Drawer Sidebar Sheet -->
{#if isMobileMenuOpen}
  <!-- Backdrop Blur -->
  <div 
    class="drawer-backdrop" 
    on:click={closeMobileMenu}
    role="presentation"
  ></div>

  <!-- Sidebar Panel -->
  <aside class="mobile-drawer pixel-box">
    <div class="drawer-header">
      <div class="drawer-brand">
        <span class="pixel-dot"></span>
        <span class="font-pixel drawer-title">MENU NAVIGASI</span>
      </div>
      <button 
        type="button" 
        class="drawer-close-btn font-pixel" 
        on:click={closeMobileMenu}
        aria-label="Tutup Menu"
      >
        <Icon name="close" size={16} strokeWidth={2.5} />
      </button>
    </div>

    <div class="drawer-body font-mono">
      <!-- Active Tracks Stats -->
      <div class="drawer-stat-box">
        <span class="dot-green"></span>
        <span>STATUS: <strong>{totalTracks} PRODUK AKTIF DIPANTAU</strong></span>
      </div>

      <!-- User Profile Card (if logged in) -->
      {#if currentUser}
        <div class="drawer-user-card">
          <div class="drawer-user-info">
            <Icon name="user" size={18} color="var(--accent-orange)" />
            <div>
              <strong class="drawer-user-name">{currentUser.first_name || "User Telegram"}</strong>
              <small class="drawer-user-chatid">Chat ID: <code>{currentUser.user_phone}</code></small>
            </div>
          </div>
          <button type="button" class="drawer-logout-btn" on:click={handleLogoutClick}>
            [KELUAR / LOGOUT]
          </button>
        </div>
      {:else}
        <button 
          type="button" 
          class="pixel-btn pixel-btn-dark drawer-action-btn font-mono"
          on:click={handleLoginClick}
        >
          <Icon name="lock" size={15} color="#00E676" />
          <span>LOGIN DENGAN TELEGRAM</span>
        </button>
      {/if}

      <div class="drawer-divider"></div>

      <!-- Navigation Links -->
      <nav class="drawer-nav">
        <button type="button" class="drawer-nav-item" on:click={handleAboutClick}>
          <div class="nav-item-icon">
            <Icon name="info" size={16} color="#FF5722" />
          </div>
          <div class="nav-item-content">
            <strong>CARA KERJA & INFO TOKO</strong>
            <small>Panduan scraper & e-commerce yang didukung</small>
          </div>
        </button>

        <a 
          href={botUrl} 
          target="_blank" 
          rel="noopener noreferrer" 
          class="drawer-nav-item drawer-nav-tg"
          on:click={closeMobileMenu}
        >
          <div class="nav-item-icon">
            <Icon name="send" size={16} color="var(--accent-blue)" />
          </div>
          <div class="nav-item-content">
            <strong>BOT TELEGRAM RESMI</strong>
            <small>@mf_pricetracker_bot (Penerima Alert)</small>
          </div>
          <Icon name="external-link" size={13} color="#666" />
        </a>

        <a 
          href={githubUrl} 
          target="_blank" 
          rel="noopener noreferrer" 
          class="drawer-nav-item"
          on:click={closeMobileMenu}
        >
          <div class="nav-item-icon">
            <Icon name="github" size={16} color="#000" />
          </div>
          <div class="nav-item-content">
            <strong>SOURCE CODE GITHUB</strong>
            <small>Repository proyek & dokumentasi</small>
          </div>
          <Icon name="external-link" size={13} color="#666" />
        </a>
      </nav>

      <div class="drawer-divider"></div>

      <!-- PWA Install Button (If browser supports install prompt) -->
      {#if isInstallable}
        <button 
          type="button" 
          class="pixel-btn pixel-btn-green drawer-install-btn font-mono"
          on:click={handleInstallPWA}
        >
          <Icon name="download" size={15} color="#000000" />
          <span>INSTALL APP KE HOME SCREEN</span>
        </button>
      {/if}

      <!-- Sync Button inside Drawer -->
      <button 
        on:click={handleSyncClick} 
        disabled={isSyncing} 
        class="pixel-btn pixel-btn-orange drawer-sync-btn font-mono"
      >
        <Icon name="refresh" size={15} className={isSyncing ? "spin-icon" : ""} />
        <span>{isSyncing ? "SEDANG EVALUASI HARGA..." : "EVALUASI SEMUA HARGA (SYNC)"}</span>
      </button>
    </div>
  </aside>
{/if}

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
    padding: 14px 20px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 16px;
  }

  .brand-group {
    display: flex;
    flex-direction: column;
    gap: 3px;
  }

  .brand-badge {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .pixel-dot {
    width: 11px;
    height: 11px;
    background-color: var(--accent-orange);
    border: 2px solid var(--border-color);
    box-shadow: 2px 2px 0px var(--border-color);
  }

  .brand-title {
    font-size: 1.25rem;
    font-weight: 700;
    color: var(--text-main);
    letter-spacing: -0.5px;
  }

  .version-tag {
    font-family: var(--font-mono);
    font-size: 0.68rem;
    font-weight: 700;
    background-color: var(--accent-yellow);
    color: #854D0E;
    padding: 2px 6px;
    border: 1.5px solid var(--border-color);
    box-shadow: 1px 1px 0px var(--border-color);
  }

  .brand-subtitle {
    font-size: 0.72rem;
    color: var(--text-muted);
    font-weight: 600;
  }

  /* Desktop Actions */
  .actions-group {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .nav-links {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .nav-btn {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    background-color: var(--bg-canvas);
    border: 2px solid var(--border-color);
    padding: 6px 10px;
    font-size: 0.75rem;
    font-weight: 700;
    color: var(--text-main);
    text-decoration: none;
    cursor: pointer;
    box-shadow: 2px 2px 0px var(--border-color);
    transition: transform 0.1s ease, box-shadow 0.1s ease;
  }

  .nav-btn:hover {
    transform: translate(-1px, -1px);
    box-shadow: 3px 3px 0px var(--border-color);
    background-color: #FFFFFF;
  }

  .nav-btn-tg {
    background-color: var(--accent-blue-light);
  }

  .status-ticker {
    display: flex;
    align-items: center;
    gap: 8px;
    background-color: var(--bg-canvas);
    padding: 7px 12px;
    border: 2px solid var(--border-color);
    font-size: 0.78rem;
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
    border: 1px solid var(--border-color);
  }

  .user-profile-badge {
    display: flex;
    align-items: center;
    gap: 8px;
    background-color: var(--accent-blue-light);
    border: 2px solid var(--border-color);
    padding: 6px 12px;
    box-shadow: 2px 2px 0px var(--border-color);
    font-size: 0.78rem;
  }

  .user-name {
    display: flex;
    flex-direction: column;
    line-height: 1.1;
  }

  .user-name small {
    font-size: 0.65rem;
    color: var(--text-muted);
  }

  .logout-btn {
    background: transparent;
    border: none;
    color: #DC2626;
    font-family: inherit;
    font-weight: 800;
    font-size: 0.72rem;
    cursor: pointer;
    margin-left: 4px;
  }

  .logout-btn:hover {
    text-decoration: underline;
  }

  /* Mobile Display Switchers */
  .mobile-only {
    display: none;
  }

  .desktop-only {
    display: flex;
  }

  /* Mobile Styles (<= 768px) */
  @media (max-width: 768px) {
    .desktop-only {
      display: none;
    }

    .mobile-only {
      display: flex;
    }

    .header-inner {
      padding: 12px 18px;
    }

    .brand-title {
      font-size: 1.05rem;
    }

    .brand-subtitle {
      display: none;
    }

    .version-tag {
      display: none;
    }

    .mobile-top-actions {
      align-items: center;
      gap: 10px;
    }

    .mobile-direct-login {
      padding: 6px 11px;
      font-size: 0.72rem;
      gap: 5px;
      box-shadow: 2px 2px 0px var(--border-color);
    }

    .mobile-user-badge {
      display: flex;
      align-items: center;
      gap: 5px;
      background-color: var(--accent-blue-light);
      border: 2px solid var(--border-color);
      padding: 5px 9px;
      font-size: 0.75rem;
      font-weight: 700;
      box-shadow: 2px 2px 0px var(--border-color);
    }

    .mobile-user-name {
      max-width: 90px;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .menu-toggle-btn {
      padding: 7px 10px;
      background-color: var(--bg-canvas);
      box-shadow: 2px 2px 0px var(--border-color);
    }
  }

  @media (max-width: 480px) {
    .header-inner {
      padding: 10px 16px;
    }

    .brand-title {
      font-size: 0.95rem;
    }

    .pixel-dot {
      width: 9px;
      height: 9px;
    }

    .mobile-top-actions {
      gap: 8px;
    }

    .mobile-direct-login {
      padding: 5px 9px;
      font-size: 0.68rem;
    }
  }

  /* Drawer / Sidebar Styles */
  .drawer-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(2px);
    z-index: 99;
  }

  .mobile-drawer {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    width: 320px;
    max-width: 88vw;
    background-color: var(--bg-card);
    border-left: var(--border-width) solid var(--border-color);
    box-shadow: -8px 0px 0px rgba(0, 0, 0, 0.15);
    z-index: 100;
    display: flex;
    flex-direction: column;
    padding: 0;
    border-radius: 0;
    animation: slideInRight 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes slideInRight {
    from {
      transform: translateX(100%);
    }
    to {
      transform: translateX(0);
    }
  }

  .drawer-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    border-bottom: 2px solid var(--border-color);
    background-color: var(--bg-canvas);
  }

  .drawer-brand {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .drawer-title {
    font-size: 0.95rem;
    font-weight: 700;
    color: var(--text-main);
  }

  .drawer-close-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    background-color: #FEE2E2;
    border: 2px solid var(--border-color);
    color: #DC2626;
    cursor: pointer;
    box-shadow: 2px 2px 0px var(--border-color);
  }

  .drawer-close-btn:active {
    transform: translate(1px, 1px);
    box-shadow: 1px 1px 0px var(--border-color);
  }

  .drawer-body {
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 14px;
    overflow-y: auto;
    flex: 1;
  }

  .drawer-stat-box {
    display: flex;
    align-items: center;
    gap: 8px;
    background-color: var(--bg-canvas);
    border: 2px solid var(--border-color);
    padding: 8px 12px;
    font-size: 0.75rem;
    box-shadow: 2px 2px 0px var(--border-color);
  }

  .drawer-user-card {
    background-color: var(--accent-blue-light);
    border: 2px solid var(--border-color);
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 8px;
    box-shadow: 2px 2px 0px var(--border-color);
  }

  .drawer-user-info {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .drawer-user-name {
    font-size: 0.85rem;
    display: block;
  }

  .drawer-user-chatid {
    font-size: 0.7rem;
    color: var(--text-muted);
  }

  .drawer-logout-btn {
    align-self: flex-start;
    background: transparent;
    border: none;
    color: #DC2626;
    font-family: inherit;
    font-size: 0.75rem;
    font-weight: 800;
    cursor: pointer;
    padding: 2px 0;
  }

  .drawer-action-btn {
    width: 100%;
    padding: 10px;
    font-size: 0.82rem;
  }

  .drawer-divider {
    height: 2px;
    background-color: var(--border-color);
    margin: 4px 0;
  }

  .drawer-nav {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .drawer-nav-item {
    display: flex;
    align-items: center;
    gap: 12px;
    background-color: var(--bg-card);
    border: 2px solid var(--border-color);
    padding: 10px 12px;
    text-decoration: none;
    color: var(--text-main);
    text-align: left;
    box-shadow: 2px 2px 0px var(--border-color);
    cursor: pointer;
    transition: transform 0.1s ease, box-shadow 0.1s ease;
  }

  .drawer-nav-item:active {
    transform: translate(1px, 1px);
    box-shadow: 1px 1px 0px var(--border-color);
  }

  .drawer-nav-tg {
    background-color: var(--accent-blue-light);
  }

  .nav-item-icon {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .nav-item-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .nav-item-content strong {
    font-size: 0.78rem;
  }

  .nav-item-content small {
    font-size: 0.68rem;
    color: var(--text-muted);
  }

  .drawer-install-btn {
    width: 100%;
    padding: 12px;
    font-size: 0.8rem;
    font-weight: 800;
    letter-spacing: -0.2px;
  }

  .drawer-sync-btn {
    width: 100%;
    padding: 12px;
    font-size: 0.82rem;
    margin-top: auto;
  }
</style>
