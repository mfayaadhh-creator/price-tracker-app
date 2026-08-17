<script>
  import { onMount } from "svelte";
  import Header from "./lib/Header.svelte";
  import BlockCreator from "./lib/BlockCreator.svelte";
  import ProductBlock from "./lib/ProductBlock.svelte";
  import Toast from "./lib/Toast.svelte";
  import AuthModal from "./lib/AuthModal.svelte";
  import ImagePreviewModal from "./lib/ImagePreviewModal.svelte";

  let products = [];
  let isLoading = true;
  let isSubmitting = false;
  let isSyncing = false;
  let searchQuery = "";
  let toasts = [];
  let isAuthModalOpen = false;
  let currentUser = null;
  let previewImageData = null;

  // Drag & drop state
  let draggedIndex = null;

  function showToast(message, type = "info") {
    const id = Date.now() + Math.random();
    toasts = [...toasts, { id, message, type }];
    setTimeout(() => {
      removeToast(id);
    }, 4000);
  }

  function removeToast(id) {
    toasts = toasts.filter((t) => t.id !== id);
  }

  async function fetchProducts() {
    isLoading = true;
    try {
      let endpoint = "/api/v1/tracks";
      if (currentUser && currentUser.user_phone) {
        endpoint += `?chat_id=${encodeURIComponent(currentUser.user_phone)}`;
      }

      const res = await fetch(endpoint);
      if (!res.ok) throw new Error("Gagal mengambil data produk");
      const data = await res.json();
      products = data.data || [];
    } catch (err) {
      console.error(err);
      showToast("Gagal memuat produk dari server", "error");
    } finally {
      isLoading = false;
    }
  }

  function handleLoginSuccess(user) {
    currentUser = user;
    localStorage.setItem("pt_auth_user", JSON.stringify(user));
    showToast(`🎉 Selamat datang kembali, ${user.first_name}!`, "success");
    fetchProducts();
  }

  function handleLogout() {
    currentUser = null;
    localStorage.removeItem("pt_auth_user");
    showToast("Anda telah keluar dari akun.", "info");
    fetchProducts();
  }

  async function handleAddProduct(payload) {
    isSubmitting = true;
    try {
      const res = await fetch("/api/v1/track", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });

      const data = await res.json();

      if (!res.ok) {
        throw new Error(data.error || "Gagal menambahkan produk");
      }

      showToast(`✅ "${data.data.name}" berhasil ditambahkan ke pemantauan!`, "success");
      // Add to front of products list
      products = [data.data, ...products];
      return true;
    } catch (err) {
      console.error(err);
      showToast(`❌ ${err.message}`, "error");
      return false;
    } finally {
      isSubmitting = false;
    }
  }

  async function handleDeleteProduct(id) {
    try {
      const res = await fetch(`/api/v1/tracks/${id}`, {
        method: "DELETE"
      });

      if (!res.ok) throw new Error("Gagal menghapus produk");

      products = products.filter((p) => p.id !== id);
      showToast("🗑️ Block produk berhasil dihapus dari tracking list!", "info");
    } catch (err) {
      console.error(err);
      showToast("Gagal menghapus produk", "error");
    }
  }

  async function handleSyncCron() {
    isSyncing = true;
    try {
      const res = await fetch("/api/cron");
      if (!res.ok) throw new Error("Gagal menjalankan sinkronisasi cron");
      const data = await res.json();
      
      const result = data.result || {};
      const drops = result.price_drops || 0;
      const checked = result.total_checked || 0;

      if (drops > 0) {
        showToast(`🎉 ${drops} penurunan harga terdeteksi & alert terkirim ke Telegram!`, "success");
      } else {
        showToast(`⚡ Evaluasi selesai: ${checked} produk dicek (harga masih stabil).`, "info");
      }

      // Refresh products from database
      await fetchProducts();
    } catch (err) {
      console.error(err);
      showToast("Gagal menjalankan evaluasi harga", "error");
    } finally {
      isSyncing = false;
    }
  }

  // --- Re-order / Drag & Drop Handlers ---
  function handleDragStart(e, index) {
    draggedIndex = index;
    e.dataTransfer.effectAllowed = "move";
  }

  function handleDragOver(e, index) {
    e.preventDefault();
    e.dataTransfer.dropEffect = "move";
  }

  function handleDrop(e, targetIndex) {
    e.preventDefault();
    if (draggedIndex === null || draggedIndex === targetIndex) return;

    const updated = [...products];
    const [movedItem] = updated.splice(draggedIndex, 1);
    updated.splice(targetIndex, 0, movedItem);
    products = updated;
    draggedIndex = null;
  }

  function handleMoveLeft(index) {
    if (index <= 0) return;
    const updated = [...products];
    const temp = updated[index];
    updated[index] = updated[index - 1];
    updated[index - 1] = temp;
    products = updated;
  }

  function handleMoveRight(index) {
    if (index >= products.length - 1) return;
    const updated = [...products];
    const temp = updated[index];
    updated[index] = updated[index + 1];
    updated[index + 1] = temp;
    products = updated;
  }

  $: filteredProducts = products.filter((p) => {
    if (!searchQuery.trim()) return true;
    const query = searchQuery.toLowerCase();
    return (
      (p.name && p.name.toLowerCase().includes(query)) ||
      (p.product_id && p.product_id.toLowerCase().includes(query)) ||
      (p.user_phone && p.user_phone.toLowerCase().includes(query))
    );
  });

  onMount(() => {
    try {
      const savedUser = localStorage.getItem("pt_auth_user");
      if (savedUser) {
        currentUser = JSON.parse(savedUser);
      }
    } catch (e) {
      console.error("Gagal load user dari localStorage:", e);
    }
    fetchProducts();
  });
</script>

<div class="app-wrapper">
  <!-- Sticky Pixel Header -->
  <Header 
    totalTracks={products.length} 
    {isSyncing} 
    onSync={handleSyncCron}
    {currentUser}
    onOpenLogin={() => isAuthModalOpen = true}
    onLogout={handleLogout}
  />

  <main class="main-content">
    <!-- Hero / Info Banner -->
    <section class="hero-banner pixel-box">
      <div class="hero-content">
        <div class="hero-badge font-pixel">
          <span>⚡</span> MISTRAL NEO-PIXEL EDITION
        </div>
        <h1 class="hero-title font-display">
          PANTAU DISKON FASHION DENGAN <span class="highlight-text">MODULAR BLOCKS</span>
        </h1>
        <p class="hero-desc font-mono">
          Tambahkan link produk dari toko favorit Anda (Uniqlo, Zara, Zalora, H&M, dll). Setiap kali sistem mendeteksi penurunan harga atau diskon baru, bot Telegram kami akan langsung mengirimkan pesan ke HP Anda!
        </p>

        <!-- Quick 3-Step Guide -->
        <div class="steps-grid font-mono">
          <div class="step-card">
            <span class="step-num font-pixel">01</span>
            <span class="step-text">Login via Telegram untuk Dashboard Privat</span>
          </div>
          <div class="step-card">
            <span class="step-num font-pixel">02</span>
            <span class="step-text">Input Link Produk Toko Favorit Anda</span>
          </div>
          <div class="step-card">
            <span class="step-num font-pixel">03</span>
            <span class="step-text">Sistem Memantau 24/7 & Kirim Alert Diskon</span>
          </div>
        </div>
      </div>
    </section>

    <!-- Block Creator Form -->
    <section class="creator-section">
      <BlockCreator 
        onAddProduct={handleAddProduct} 
        {isSubmitting}
        {currentUser}
      />
    </section>

    <!-- Tracked Products Grid Section -->
    <section class="blocks-section">
      <div class="section-header">
        <div class="section-title-group">
          <h2 class="section-title font-pixel">
            <span>📦</span> ACTIVE_BLOCKS ({filteredProducts.length})
          </h2>
          <span class="drag-hint font-mono">
            // Tahan & geser (drag) atau gunakan tombol ◀ ▶ untuk mengatur urutan block
          </span>
        </div>

        <!-- Search / Filter Input -->
        {#if products.length > 0}
          <div class="search-wrapper">
            <input 
              type="text" 
              bind:value={searchQuery}
              placeholder="Cari nama produk / Chat ID..."
              class="pixel-input search-input font-mono"
            />
          </div>
        {/if}
      </div>

      <!-- Loading State -->
      {#if isLoading}
        <div class="loading-box pixel-box font-mono">
          <div class="spinner-large"></div>
          <p class="loading-text font-pixel">MEMUAT DAFTAR BLOCKS DARI DATABASE...</p>
        </div>

      <!-- Empty State -->
      {:else if products.length === 0}
        <div class="empty-box pixel-box">
          <div class="empty-pixel-icon font-pixel">[ ? ]</div>
          <h3 class="empty-title font-display">Belum Ada Block Produk yang Dipantau</h3>
          <p class="empty-desc font-mono">
            Salin link produk dari website toko favorit Anda (Uniqlo, Zara, Zalora, H&M, dll) dan masukkan pada form di atas untuk memulai pemantauan harga otomatis!
          </p>
        </div>

      <!-- Product Blocks Grid -->
      {:else}
        {#if filteredProducts.length === 0}
          <div class="empty-search pixel-box font-mono">
            <p>Tidak ada block produk yang cocok dengan kata kunci "{searchQuery}".</p>
          </div>
        {:else}
          <div class="blocks-grid">
            {#each filteredProducts as product, index (product.id || index)}
              <ProductBlock 
                {product}
                {index}
                total={filteredProducts.length}
                onDelete={handleDeleteProduct}
                onMoveLeft={handleMoveLeft}
                onMoveRight={handleMoveRight}
                onDragStart={handleDragStart}
                onDragOver={handleDragOver}
                onDrop={handleDrop}
                onPreviewImage={(prod) => previewImageData = prod}
              />
            {/each}
          </div>
        {/if}
      {/if}
    </section>
  </main>

  <!-- Image Preview Lightbox Modal -->
  <ImagePreviewModal 
    data={previewImageData} 
    onClose={() => previewImageData = null} 
  />

  <!-- Auth Modal -->
  <AuthModal 
    isOpen={isAuthModalOpen} 
    onClose={() => isAuthModalOpen = false} 
    onLoginSuccess={handleLoginSuccess} 
  />

  <!-- Floating Toast Feedback System -->
  <Toast {toasts} onRemove={removeToast} />
</div>

<style>
  .app-wrapper {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
  }

  .main-content {
    max-width: 1280px;
    margin: 0 auto;
    padding: 32px 20px 80px 20px;
    width: 100%;
  }

  .hero-banner {
    background-color: var(--bg-card);
    padding: 32px;
    margin-bottom: 28px;
    border-radius: 0;
  }

  .hero-badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    background-color: var(--accent-yellow);
    border: 1.5px solid var(--border-color);
    padding: 4px 10px;
    font-size: 0.75rem;
    font-weight: 700;
    color: #854D0E;
    box-shadow: 2px 2px 0px var(--border-color);
    margin-bottom: 14px;
  }

  .hero-title {
    font-size: clamp(1.6rem, 3.5vw, 2.4rem);
    font-weight: 800;
    line-height: 1.2;
    margin-bottom: 12px;
    letter-spacing: -0.5px;
  }

  .highlight-text {
    color: var(--accent-orange);
    text-decoration: underline;
    text-decoration-thickness: 4px;
    text-underline-offset: 4px;
  }

  .hero-desc {
    font-size: 0.95rem;
    color: var(--text-muted);
    max-width: 760px;
    line-height: 1.5;
    margin-bottom: 24px;
  }

  .steps-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: 12px;
    margin-top: 20px;
  }

  .step-card {
    background-color: var(--bg-canvas);
    border: 2px solid var(--border-color);
    padding: 12px 14px;
    display: flex;
    align-items: center;
    gap: 12px;
    box-shadow: 2px 2px 0px var(--border-color);
  }

  .step-num {
    background-color: var(--border-color);
    color: #FFFFFF;
    font-size: 0.85rem;
    padding: 4px 8px;
    font-weight: 700;
  }

  .step-text {
    font-size: 0.8rem;
    color: var(--text-main);
    line-height: 1.35;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    flex-wrap: wrap;
    gap: 16px;
    margin-bottom: 20px;
    padding-bottom: 12px;
    border-bottom: var(--border-width) solid var(--border-color);
  }

  .section-title {
    font-size: 1.25rem;
    color: var(--text-main);
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 4px;
  }

  .drag-hint {
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  .search-input {
    min-width: 280px;
    padding: 8px 12px;
    font-size: 0.85rem;
  }

  .blocks-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 20px;
  }

  .loading-box, .empty-box, .empty-search {
    background-color: var(--bg-card);
    padding: 48px 24px;
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 14px;
  }

  .spinner-large {
    width: 32px;
    height: 32px;
    border: 4px solid var(--border-color);
    border-top-color: var(--accent-orange);
    border-radius: 0;
    animation: spin 0.8s linear infinite;
  }

  .loading-text {
    font-size: 0.9rem;
    color: var(--text-main);
  }

  .empty-pixel-icon {
    font-size: 2rem;
    font-weight: 800;
    color: var(--accent-orange);
    background-color: var(--accent-yellow);
    border: 2px solid var(--border-color);
    padding: 8px 18px;
    box-shadow: 3px 3px 0px var(--border-color);
  }

  .empty-title {
    font-size: 1.3rem;
    font-weight: 700;
  }

  .empty-desc {
    font-size: 0.9rem;
    color: var(--text-muted);
    max-width: 520px;
    line-height: 1.4;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  @media (max-width: 768px) {
    .main-content {
      padding: 20px 14px 60px 14px;
    }

    .hero-banner {
      padding: 20px 16px;
    }

    .section-header {
      flex-direction: column;
      align-items: flex-start;
    }

    .search-input {
      width: 100%;
      min-width: 100%;
    }
  }
</style>
