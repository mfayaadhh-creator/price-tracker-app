<script>
  import { onMount, onDestroy } from "svelte";
  import Icon from "./Icon.svelte";

  export let isOpen = false;
  export let onClose;
  export let onLoginSuccess;

  let activeTab = "deeplink"; // 'deeplink' | 'manual'
  let authCode = "";
  let deepLink = "";
  let isInitializing = false;
  let isPolling = false;
  let pollInterval = null;
  let manualChatId = "";
  let manualName = "";
  let errorMsg = "";

  async function initAuthSession() {
    isInitializing = true;
    errorMsg = "";
    try {
      const res = await fetch("/api/v1/auth/telegram/init", {
        method: "POST"
      });
      if (!res.ok) throw new Error("Gagal membuat sesi login Telegram");
      const data = await res.json();
      authCode = data.code;
      deepLink = data.deep_link;
      startPolling(authCode);
    } catch (err) {
      console.error(err);
      errorMsg = "Gagal terhubung ke server auth. Silakan gunakan opsi Chat ID manual.";
    } finally {
      isInitializing = false;
    }
  }

  let pollAttempts = 0;
  const maxPollAttempts = 30; // 30 x 2s = 60 detik maksimal polling
  let isTimedOut = false;

  function startPolling(code) {
    if (pollInterval) clearInterval(pollInterval);
    isPolling = true;
    pollAttempts = 0;
    isTimedOut = false;

    pollInterval = setInterval(async () => {
      // Pause jika tab browser sedang tidak aktif/fokus
      if (document.hidden) return;

      pollAttempts++;
      if (pollAttempts > maxPollAttempts) {
        clearInterval(pollInterval);
        isPolling = false;
        isTimedOut = true;
        return;
      }

      try {
        const res = await fetch(`/api/v1/auth/telegram/poll?code=${code}`);
        if (res.ok) {
          const session = await res.json();
          if (session.verified) {
            clearInterval(pollInterval);
            isPolling = false;
            onLoginSuccess({
              user_phone: session.user_phone,
              first_name: session.first_name || "User Telegram",
              username: session.username || ""
            });
            onClose();
          }
        }
      } catch (err) {
        console.error("Polling auth error:", err);
      }
    }, 2000);
  }

  async function handleManualLogin() {
    if (!manualChatId.trim()) {
      errorMsg = "Chat ID wajib diisi!";
      return;
    }

    try {
      const res = await fetch("/api/v1/auth/instant", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          chat_id: manualChatId.trim(),
          name: manualName.trim()
        })
      });

      if (!res.ok) throw new Error("Gagal login manual");
      const data = await res.json();
      onLoginSuccess({
        user_phone: data.user_phone,
        first_name: data.first_name,
        username: data.username || ""
      });
      onClose();
    } catch (err) {
      errorMsg = err.message;
    }
  }

  let hasInitialized = false;

  $: if (isOpen && !hasInitialized) {
    hasInitialized = true;
    initAuthSession();
  } else if (!isOpen) {
    hasInitialized = false;
    if (pollInterval) {
      clearInterval(pollInterval);
      pollInterval = null;
    }
  }

  onDestroy(() => {
    if (pollInterval) clearInterval(pollInterval);
  });
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
          <Icon name="lock" size={16} color="#00E676" />
          <span>LOGIN_TELEGRAM // AUTH</span>
        </div>
        <button type="button" class="close-btn" on:click={onClose} title="Tutup">
          <Icon name="close" size={14} />
        </button>
      </div>

      <!-- Tab Buttons -->
      <div class="tabs-header font-mono">
        <button 
          type="button" 
          class="tab-btn {activeTab === 'deeplink' ? 'active' : ''}"
          on:click={() => activeTab = 'deeplink'}
        >
          <Icon name="bolt" size={14} color="#FF5722" />
          <span>DEEP-LINK BOT (AUTO)</span>
        </button>
        <button 
          type="button" 
          class="tab-btn {activeTab === 'manual' ? 'active' : ''}"
          on:click={() => activeTab = 'manual'}
        >
          <Icon name="user" size={14} />
          <span>INPUT CHAT ID</span>
        </button>
      </div>

      <!-- Tab Content: Deep-link -->
      {#if activeTab === 'deeplink'}
        <div class="tab-body">
          <p class="tab-instruction font-mono">
            Klik tombol di bawah untuk membuka Bot Telegram. Cukup tekan <strong>START</strong> di Telegram, dan browser ini akan otomatis login!
          </p>

          {#if isInitializing}
            <div class="loading-state font-mono">
              <div class="spinner-pixel"></div>
              <span>Menyiapkan sesi login...</span>
            </div>
          {:else}
            <!-- Big Deep Link CTA -->
            <a 
              href={deepLink} 
              target="_blank" 
              rel="noopener noreferrer" 
              class="pixel-btn pixel-btn-blue telegram-cta font-mono"
            >
              <Icon name="send" size={16} color="#FFFFFF" />
              <span>BUKA TELEGRAM & VERIFIKASI</span>
              <Icon name="external-link" size={14} color="#FFFFFF" />
            </a>

            <!-- Polling Live Status -->
            {#if isTimedOut}
              <div class="timeout-box font-mono">
                <Icon name="warning" size={14} color="#DC2626" />
                <span>Sesi login berakhir (1 menit).</span>
                <button type="button" on:click={initAuthSession} class="pixel-btn retry-btn font-mono">
                  <Icon name="refresh" size={13} />
                  <span>REFRESH SESI</span>
                </button>
              </div>
            {:else}
              <div class="polling-box font-mono">
                <span class="pulse-indicator"></span>
                <span>Menunggu konfirmasi tombol Start di Telegram...</span>
              </div>
            {/if}

            <div class="code-badge font-mono">
              SESSION CODE: <strong>{authCode}</strong>
            </div>
          {/if}
        </div>
      {:else}
        <!-- Tab Content: Manual Chat ID -->
        <div class="tab-body">
          <p class="tab-instruction font-mono">
            Masukkan Chat ID Telegram Anda untuk login instan:
          </p>

          <form on:submit|preventDefault={handleManualLogin} class="manual-form">
            <div class="input-group">
              <label for="manual-chat-id" class="input-label font-mono">CHAT ID TELEGRAM:</label>
              <input 
                id="manual-chat-id"
                type="text" 
                bind:value={manualChatId} 
                placeholder="Contoh: 123456789"
                class="pixel-input"
                required
              />
            </div>

            <div class="input-group">
              <label for="manual-name" class="input-label font-mono">NAMA LENGKAP:</label>
              <input 
                id="manual-name"
                type="text" 
                bind:value={manualName} 
                placeholder="Nama Anda"
                class="pixel-input"
              />
            </div>

            <button type="submit" class="pixel-btn pixel-btn-orange submit-btn font-mono">
              <span>MASUK SEKARANG</span>
              <span>➔</span>
            </button>
          </form>
        </div>
      {/if}

      {#if errorMsg}
        <div class="error-box font-mono">
          ⚠️ {errorMsg}
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background-color: rgba(24, 24, 27, 0.7);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 200;
    padding: 20px;
  }

  .modal-card {
    background-color: var(--bg-card);
    width: 100%;
    max-width: 520px;
    border-radius: 0;
    box-shadow: 8px 8px 0px #18181B;
    overflow: hidden;
  }

  .modal-header {
    background-color: var(--bg-card-dark);
    color: #FFFFFF;
    padding: 14px 18px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 2px solid var(--border-color);
  }

  .header-title {
    font-size: 1rem;
    font-weight: 700;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .close-btn {
    background: transparent;
    border: none;
    color: #FFFFFF;
    font-size: 1.1rem;
    font-weight: 800;
    cursor: pointer;
    padding: 2px 6px;
  }

  .tabs-header {
    display: flex;
    border-bottom: 2px solid var(--border-color);
    background-color: var(--bg-canvas-subtle);
  }

  .tab-btn {
    flex: 1;
    padding: 10px;
    border: none;
    border-right: 2px solid var(--border-color);
    background: transparent;
    font-family: inherit;
    font-size: 0.78rem;
    font-weight: 700;
    cursor: pointer;
  }

  .tab-btn:last-child {
    border-right: none;
  }

  .tab-btn.active {
    background-color: var(--bg-card);
    color: var(--accent-orange);
    box-shadow: inset 0 -3px 0 var(--accent-orange);
  }

  .tab-body {
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .tab-instruction {
    font-size: 0.85rem;
    color: var(--text-muted);
    line-height: 1.4;
  }

  .telegram-cta {
    padding: 16px;
    font-size: 1rem;
    width: 100%;
  }

  .polling-box {
    display: flex;
    align-items: center;
    gap: 10px;
    background-color: #FEF3C7;
    border: 1.5px solid #D97706;
    padding: 10px 14px;
    font-size: 0.78rem;
    color: #92400E;
    font-weight: 600;
  }

  .pulse-indicator {
    width: 10px;
    height: 10px;
    background-color: #D97706;
    border-radius: 50%;
    animation: pulse 1.5s infinite;
  }

  .code-badge {
    text-align: center;
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  .manual-form {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .input-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .input-label {
    font-size: 0.78rem;
    font-weight: 700;
  }

  .submit-btn {
    width: 100%;
    padding: 12px;
  }

  .error-box {
    margin: 0 24px 20px 24px;
    background-color: #FEE2E2;
    border: 1.5px solid #DC2626;
    color: #991B1B;
    padding: 8px 12px;
    font-size: 0.8rem;
    font-weight: 600;
  }

  .loading-state {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    padding: 20px;
  }

  .spinner-pixel {
    width: 16px;
    height: 16px;
    border: 3px solid var(--border-color);
    border-top-color: var(--accent-orange);
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; transform: scale(1); }
    50% { opacity: 0.3; transform: scale(0.85); }
  }
</style>
