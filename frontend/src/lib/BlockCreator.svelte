<script>
  import Icon from "./Icon.svelte";

  export let onAddProduct;
  export let isSubmitting = false;
  export let currentUser = null;

  let url = "";
  let userPhone = "";
  let targetPrice = "";
  let formError = "";

  $: if (currentUser && currentUser.user_phone) {
    userPhone = currentUser.user_phone;
  }

  const sampleUrl = "https://www.uniqlo.com/id/id/products/E493232-000/00?colorDisplayCode=00&sizeDisplayCode=004";

  function setSample() {
    url = sampleUrl;
    formError = "";
  }

  async function handleSubmit() {
    formError = "";
    
    if (!url.trim()) {
      formError = "URL produk wajib diisi!";
      return;
    }

    if (!url.startsWith("http://") && !url.startsWith("https://")) {
      formError = "Link harus berupa URL web yang valid (dimulai dengan https:// atau http://)";
      return;
    }

    if (!userPhone.trim()) {
      formError = "Chat ID Telegram wajib diisi agar notifikasi bisa dikirim!";
      return;
    }

    const payload = {
      url: url.trim(),
      user_phone: userPhone.trim(),
      target_price: targetPrice ? parseFloat(targetPrice) : 0
    };

    const success = await onAddProduct(payload);
    if (success) {
      url = "";
      targetPrice = "";
    }
  }
</script>

<div class="creator-card pixel-box">
  <div class="creator-header">
    <div class="header-tag font-pixel">
      <Icon name="plus" size={14} color="#FF5722" strokeWidth={2.5} /> TAMBAH BLOCK BARU
    </div>
    <span class="header-desc font-mono">
      Tempel link produk online apa saja (Laptop, Elektronik, Fashion, dll) → Auto scraping & pantau 24/7
    </span>
  </div>

  <form on:submit|preventDefault={handleSubmit} class="creator-form">
    <!-- URL Input -->
    <div class="input-group">
      <label for="product-url" class="input-label font-mono">
        <span class="label-badge font-pixel">URL</span>
        LINK PRODUK E-COMMERCE (AGRES.ID, UNIQLO, ZARA, ZALORA, H&M, DLL):
      </label>
      <div class="input-wrapper">
        <input 
          id="product-url"
          type="url" 
          bind:value={url} 
          placeholder="https://www.agres.id/... atau https://www.uniqlo.com/... atau https://www.zara.com/..."
          class="pixel-input"
          disabled={isSubmitting}
          required
        />
        <button 
          type="button" 
          on:click={setSample} 
          class="sample-btn font-mono"
          title="Isi dengan contoh link produk"
        >
          [SAMPLE]
        </button>
      </div>
    </div>

    <!-- Row: Chat ID & Target Budget (Conditional on Login Status) -->
    {#if currentUser && currentUser.user_phone}
      <div class="form-row">
        <!-- Target Price -->
        <div class="input-group flex-1">
          <label for="target-price" class="input-label font-mono">
            <span class="label-badge font-pixel">IDR</span>
            TARGET BUDGET / HARGA MAKSIMAL (OPSIONAL):
          </label>
          <input 
            id="target-price"
            type="number" 
            bind:value={targetPrice} 
            placeholder="Contoh: 15000000 (Kosongkan jika ingin alert di setiap penurunan)"
            class="pixel-input"
            disabled={isSubmitting}
          />
          <small class="helper-text font-mono">
            Kosongkan jika ingin selalu menerima alert setiap kali ada penurunan harga.
          </small>
        </div>

        <!-- Connected Telegram Account Card -->
        <div class="account-card flex-1 font-mono">
          <div class="account-header">
            <span class="label-badge font-pixel">TG</span>
            <span class="account-label">PENERIMA ALERT NOTIFIKASI:</span>
          </div>
          <div class="account-details">
            <div class="account-icon-wrap">
              <Icon name="check" size={14} color="#065F46" strokeWidth={2.5} />
            </div>
            <div class="account-info">
              <strong class="account-name">{currentUser.first_name || "User Telegram"}</strong>
              <span class="account-chatid">Chat ID: <code>{currentUser.user_phone}</code> (Otomatis Terhubung)</span>
            </div>
          </div>
        </div>
      </div>
    {:else}
      <div class="form-row">
        <!-- Telegram Chat ID -->
        <div class="input-group flex-1">
          <label for="telegram-id" class="input-label font-mono">
            <span class="label-badge font-pixel">TG</span>
            CHAT ID TELEGRAM PENERIMA:
          </label>
          <input 
            id="telegram-id"
            type="text" 
            bind:value={userPhone} 
            placeholder="Contoh: 123456789"
            class="pixel-input"
            disabled={isSubmitting}
            required
          />
          <small class="helper-text font-mono">
            💡 Tips: Gunakan tombol <strong>LOGIN TG</strong> di header agar Chat ID terisi otomatis.
          </small>
        </div>

        <!-- Target Price (Optional) -->
        <div class="input-group flex-1">
          <label for="target-price-guest" class="input-label font-mono">
            <span class="label-badge font-pixel">IDR</span>
            TARGET BUDGET (OPSIONAL):
          </label>
          <input 
            id="target-price-guest"
            type="number" 
            bind:value={targetPrice} 
            placeholder="Contoh: 180000 (Opsional)"
            class="pixel-input"
            disabled={isSubmitting}
          />
          <small class="helper-text font-mono">
            Kosongkan jika ingin notifikasi untuk setiap penurunan harga.
          </small>
        </div>
      </div>
    {/if}

    <!-- Error Message -->
    {#if formError}
      <div class="error-box font-mono">
        <Icon name="warning" size={15} color="#DC2626" />
        <span>{formError}</span>
      </div>
    {/if}

    <!-- Submit Button -->
    <div class="form-footer">
      <button 
        type="submit" 
        disabled={isSubmitting}
        class="pixel-btn pixel-btn-orange submit-btn font-mono"
      >
        {#if isSubmitting}
          <span class="spinner-pixel"></span>
          <span>SEDANG MENGAMBIL DATA HARGA LIVE...</span>
        {:else}
          <Icon name="tag" size={15} color="#FFFFFF" />
          <span>MULAI PANTAU HARGA PRODUK INI</span>
        {/if}
      </button>
    </div>
  </form>
</div>

<style>
  .creator-card {
    background-color: var(--bg-card-highlight);
    padding: 24px;
    margin-bottom: 32px;
    border-radius: 0;
  }

  .creator-header {
    margin-bottom: 20px;
    padding-bottom: 14px;
    border-bottom: 2px dashed var(--border-color);
  }

  .header-tag {
    font-size: 1.15rem;
    font-weight: 700;
    color: var(--text-main);
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 4px;
  }

  .header-desc {
    font-size: 0.8rem;
    color: var(--text-muted);
  }

  .creator-form {
    display: flex;
    flex-direction: column;
    gap: 18px;
  }

  .form-row {
    display: flex;
    gap: 16px;
    flex-wrap: wrap;
  }

  .flex-1 {
    flex: 1;
    min-width: 260px;
  }

  .input-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .input-label {
    font-size: 0.82rem;
    font-weight: 700;
    color: var(--text-main);
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .label-badge {
    background-color: var(--border-color);
    color: #FFFFFF;
    font-size: 0.65rem;
    padding: 2px 6px;
  }

  .input-wrapper {
    position: relative;
    display: flex;
    align-items: center;
  }

  .sample-btn {
    position: absolute;
    right: 8px;
    background-color: var(--accent-yellow);
    border: 1.5px solid var(--border-color);
    padding: 4px 8px;
    font-size: 0.7rem;
    font-weight: 700;
    cursor: pointer;
    box-shadow: 1px 1px 0px var(--border-color);
  }

  .sample-btn:hover {
    background-color: #FDE047;
  }

  .helper-text {
    font-size: 0.72rem;
    color: var(--text-muted);
  }

  .account-card {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .account-header {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 0.82rem;
    font-weight: 700;
    color: var(--text-main);
  }

  .account-details {
    background-color: var(--accent-green-light);
    border: 2px solid #065F46;
    padding: 10px 14px;
    display: flex;
    align-items: center;
    gap: 10px;
    box-shadow: 2px 2px 0px #065F46;
    min-height: 48px;
  }

  .account-icon-wrap {
    background-color: #FFFFFF;
    border: 1.5px solid #065F46;
    padding: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 1px 1px 0px #065F46;
  }

  .account-info {
    display: flex;
    flex-direction: column;
    line-height: 1.2;
  }

  .account-name {
    font-size: 0.85rem;
    color: #065F46;
  }

  .account-chatid {
    font-size: 0.72rem;
    color: #047857;
  }

  .account-chatid code {
    background-color: rgba(255, 255, 255, 0.6);
    padding: 1px 4px;
    border: 1px solid #065F46;
    font-weight: 700;
  }

  .error-box {
    background-color: #FEE2E2;
    border: 2px solid #DC2626;
    color: #991B1B;
    padding: 10px 14px;
    font-size: 0.85rem;
    font-weight: 600;
    display: flex;
    align-items: center;
    gap: 8px;
    box-shadow: 3px 3px 0px #DC2626;
  }

  .form-footer {
    display: flex;
    justify-content: flex-end;
    margin-top: 4px;
  }

  .submit-btn {
    width: 100%;
    padding: 14px 20px;
    font-size: 0.95rem;
  }

  .spinner-pixel {
    width: 14px;
    height: 14px;
    border: 3px solid #FFFFFF;
    border-top-color: transparent;
    border-radius: 0;
    display: inline-block;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
</style>
