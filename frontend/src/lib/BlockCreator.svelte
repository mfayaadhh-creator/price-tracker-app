<script>
  export let onAddProduct;
  export let isSubmitting = false;
  export let currentUser = null;

  let url = "";
  let userPhone = "7514771766";
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
      <span>[+]</span> TAMBAH BLOCK BARU
    </div>
    <span class="header-desc font-mono">
      Tempel link toko apa saja (Uniqlo, Zara, Zalora, H&M, dll) → Auto scraping & pantau 24/7
    </span>
  </div>

  <form on:submit|preventDefault={handleSubmit} class="creator-form">
    <!-- URL Input -->
    <div class="input-group">
      <label for="product-url" class="input-label font-mono">
        <span class="label-badge font-pixel">URL</span>
        LINK PRODUK E-COMMERCE (UNIQLO, ZARA, ZALORA, H&M, DLL):
      </label>
      <div class="input-wrapper">
        <input 
          id="product-url"
          type="url" 
          bind:value={url} 
          placeholder="https://www.uniqlo.com/... atau https://www.zara.com/... atau https://www.zalora.co.id/..."
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

    <!-- Row: Chat ID & Target Budget -->
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
          placeholder="Contoh: 7514771766"
          class="pixel-input"
          disabled={isSubmitting}
          required
        />
        <small class="helper-text font-mono">
          Dapatkan Chat ID via bot <code>@userinfobot</code>
        </small>
      </div>

      <!-- Target Price (Optional) -->
      <div class="input-group flex-1">
        <label for="target-price" class="input-label font-mono">
          <span class="label-badge font-pixel">IDR</span>
          TARGET BUDGET (OPSIONAL):
        </label>
        <input 
          id="target-price"
          type="number" 
          bind:value={targetPrice} 
          placeholder="Contoh: 180000 (Opsional)"
          class="pixel-input"
          disabled={isSubmitting}
        />
        <small class="helper-text font-mono">
          Kosongkan jika ingin notifikasi untuk setiap penurunan harga
        </small>
      </div>
    </div>

    <!-- Error Message -->
    {#if formError}
      <div class="error-box font-mono">
        <span>⚠️</span>
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
          <span>⚡</span>
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

  .helper-text code {
    background-color: #E4E4E7;
    padding: 1px 4px;
    border: 1px solid #D4D4D8;
    color: #18181B;
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
