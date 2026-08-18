<script>
  import { onMount } from "svelte";
  import Header from "./lib/Header.svelte";
  import BlockCreator from "./lib/BlockCreator.svelte";
  import ProductBlock from "./lib/ProductBlock.svelte";
  import Toast from "./lib/Toast.svelte";
  import AuthModal from "./lib/AuthModal.svelte";
  import AboutModal from "./lib/AboutModal.svelte";
  import ImagePreviewModal from "./lib/ImagePreviewModal.svelte";
  import Icon from "./lib/Icon.svelte";

  let products = [];
  let isLoading = true;
  let isSubmitting = false;
  let isSyncing = false;
  let searchQuery = "";
  let toasts = [];
  let isAuthModalOpen = false;
  let isAboutModalOpen = false;
  let currentUser = null;
  let previewImageData = null;
  let pendingProduct = null;

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
    if (!currentUser || !currentUser.user_phone) {
      products = [];
      isLoading = false;
      return;
    }

    isLoading = true;
    try {
      const endpoint = `/api/v1/tracks?chat_id=${encodeURIComponent(currentUser.user_phone)}`;
      const res = await fetch(endpoint);
      if (!res.ok) throw new Error("Gagal mengambil data produk dari server");
      const data = await res.json();
      products = data.data || [];
    } catch (err) {
      console.error(err);
      showToast(err.message, "error");
    } finally {
      isLoading = false;
    }
  }

  async function handleLoginSuccess(user) {
    currentUser = user;
    try {
      localStorage.setItem("pt_auth_user", JSON.stringify(user));
    } catch (e) {
      console.error("Storage save error:", e);
    }
    showToast(`Selamat datang, ${user.first_name || "User"}! Akun Telegram terhubung.`, "success");

    // Jika ada produk yang sedang dipantau saat tamu mengklik tombol Telegram
    if (pendingProduct) {
      const toAdd = {
        url: pendingProduct.url,
        user_phone: user.user_phone,
        target_price: pendingProduct.target_price || 0
      };
      pendingProduct = null;
      await handleAddProduct(toAdd);
    } else {
      await fetchProducts();
    }
  }

  function handleRequestGuestAuth(productData) {
    pendingProduct = productData;
    isAuthModalOpen = true;
    showToast("Silakan tekan START di Telegram untuk menghubungkan akun & menyimpan produk.", "info");
  }

  function handleLogout() {
    currentUser = null;
    try {
      localStorage.removeItem("pt_auth_user");
    } catch (e) {
      console.error("Storage remove error:", e);
    }
    showToast("Anda telah keluar dari sesi.", "info");
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

      showToast(`Produk "${data.data?.name || "Baru"}" berhasil dipantau!`, "success");
      await fetchProducts();
      return true;
    } catch (err) {
      console.error(err);
      showToast(err.message, "error");
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

      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || "Gagal menghapus produk");
      }

      showToast("Block produk berhasil dihapus!", "info");
      products = products.filter((p) => p.id !== id);
    } catch (err) {
      console.error(err);
      showToast(err.message, "error");
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
        showToast(`${drops} penurunan harga terdeteksi & alert terkirim ke Telegram!`, "success");
      } else {
        showToast(`Evaluasi selesai: ${checked} produk dicek (harga masih stabil).`, "info");
      }

      await fetchProducts();
    } catch (err) {
      console.error(err);
      showToast("Gagal menjalankan evaluasi harga", "error");
    } finally {
      isSyncing = false;
    }
  }

  // --- Re-order / Drag & Drop Handlers ---
  function handleDragStart(index) {
    draggedIndex = index;
  }

  function handleDragOver(index) {
    // Handled visually by ProductBlock component
  }

  function handleDrop(sourceIndex, targetIndex) {
    if (sourceIndex === targetIndex || sourceIndex === null) return;

    const reordered = [...products];
    const [movedItem] = reordered.splice(sourceIndex, 1);
    reordered.splice(targetIndex, 0, movedItem);
    products = reordered;
    draggedIndex = null;
    saveReorderOrder();
  }

  function handleMoveLeft(index) {
    if (index <= 0) return;
    const reordered = [...products];
    const temp = reordered[index];
    reordered[index] = reordered[index - 1];
    reordered[index - 1] = temp;
    products = reordered;
    saveReorderOrder();
  }

  function handleMoveRight(index) {
    if (index >= products.length - 1) return;
    const reordered = [...products];
    const temp = reordered[index];
    reordered[index] = reordered[index + 1];
    reordered[index + 1] = temp;
    products = reordered;
    saveReorderOrder();
  }

  function saveReorderOrder() {
    try {
      const orderIds = products.map(p => p.id);
      localStorage.setItem("price_tracker_block_order", JSON.stringify(orderIds));
    } catch (err) {
      console.error("Failed to save reorder state:", err);
    }
  }

  $: filteredProducts = products.filter((p) => {
    if (!searchQuery.trim()) return true;
    const q = searchQuery.toLowerCase();
    const nameMatch = p.name && p.name.toLowerCase().includes(q);
    const phoneMatch = p.user_phone && p.user_phone.includes(q);
    const platformMatch = p.platform && p.platform.toLowerCase().includes(q);
    return nameMatch || phoneMatch || platformMatch;
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

    fetchProducts().then(() => {
      try {
        const savedOrder = localStorage.getItem("price_tracker_block_order");
        if (savedOrder && products.length > 0) {
          const parsed = JSON.parse(savedOrder);
          const orderMap = new Map(parsed.map((id, idx) => [id, idx]));
          products = [...products].sort((a, b) => {
            const idxA = orderMap.has(a.id) ? orderMap.get(a.id) : 9999;
            const idxB = orderMap.has(b.id) ? orderMap.get(b.id) : 9999;
            return idxA - idxB;
          });
        }
      } catch (e) {
        console.error("Order restore err:", e);
      }
    });
  });
</script>

<div class="app-shell">
  <!-- Sticky Pixel Header -->
  <Header 
    totalTracks={products.length} 
    {isSyncing} 
    onSync={handleSyncCron}
    {currentUser}
    onOpenLogin={() => isAuthModalOpen = true}
    onLogout={handleLogout}
    onOpenAbout={() => isAboutModalOpen = true}
  />

  <main class="main-content">
    <!-- Hero / Info Banner -->
    <section class="hero-banner pixel-box">
      <div class="hero-content">
        <div class="hero-badge font-pixel">
          <Icon name="tag" size={13} color="#FF5722" strokeWidth={2.2} />
          <span>UNIVERSAL E-COMMERCE PRICE TRACKER v2.0</span>
        </div>
        <h1 class="hero-title font-display">
          PANTAU HARGA & DISKON E-COMMERCE DENGAN <span class="highlight-text">NOTIFIKASI INSTAN TELEGRAM</span>
        </h1>
        <p class="hero-desc font-mono">
          Pantau penurunan harga laptop, gadget, pakaian, dan barang impian Anda dari toko online mana saja (Agres.id, Uniqlo, Zara, Zalora, H&M, dll) secara otomatis 24/7. Notifikasi instan langsung dikirim ke Telegram Anda saat harga turun!
        </p>

        <!-- Quick 3-Step Guide -->
        <div class="steps-grid font-mono">
          <div class="step-card">
            <span class="step-num font-pixel">01</span>
            <span class="step-text">Login via Telegram untuk Dashboard Privat</span>
          </div>
          <div class="step-card">
            <span class="step-num font-pixel">02</span>
            <span class="step-text">Tempel Link Produk (Elektronik, Fashion, dll)</span>
          </div>
          <div class="step-card">
            <span class="step-num font-pixel">03</span>
            <span class="step-text">Sistem Pantau 24/7 & Kirim Alert Diskon</span>
          </div>
        </div>
      </div>
    </section>

    <!-- Block Creator Form -->
    <section class="creator-section">
      <BlockCreator 
        onAddProduct={handleAddProduct} 
        onRequestAuth={handleRequestGuestAuth}
        {isSubmitting}
        {currentUser}
      />

      <!-- Supported Store Categories Ticker -->
      <div class="category-pills font-mono">
        <span class="category-pill">
          <Icon name="laptop" size={13} color="#FF5722" />
          <span>LAPTOP & ELEKTRONIK (AGRES.ID, DLL)</span>
        </span>
        <span class="category-pill">
          <Icon name="tag" size={13} color="#00E676" />
          <span>FASHION (UNIQLO, ZARA, H&M)</span>
        </span>
        <span class="category-pill">
          <Icon name="shopping-bag" size={13} color="#3B82F6" />
          <span>SNEAKERS & SEPATU (ZALORA)</span>
        </span>
        <span class="category-pill">
          <Icon name="package" size={13} color="#A855F7" />
          <span>SHOPIFY & ONLINE STORES</span>
        </span>
      </div>
    </section>

    <!-- Tracked Products Grid Section -->
    <section class="blocks-section">
      <div class="section-header">
        <div class="section-title-group">
          <h2 class="section-title font-pixel">
            <Icon name="grid" size={16} color="#FF5722" />
            <span>PRODUK_TERPANTAU ({filteredProducts.length})</span>
          </h2>
          <span class="drag-hint font-mono">
            // Tahan & geser (drag) atau gunakan tombol ◀ ▶ untuk mengatur urutan produk
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

      <!-- Guest State (Not Logged In) -->
      {:else if !currentUser}
        <div class="empty-box pixel-box">
          <div class="empty-pixel-icon">
            <Icon name="lock" size={48} color="#FF5722" />
          </div>
          <h3 class="empty-title font-display">Dashboard Privat & Terisolasi</h3>
          <p class="empty-desc font-mono">
            Setiap pengguna memiliki ruang pemantauan harga pribadi yang terenkripsi dan terpisah. Silakan hubungkan akun Telegram Anda untuk melihat dan memantau produk secara privat.
          </p>
          <button class="pixel-btn btn-primary font-pixel mt-4" on:click={() => isAuthModalOpen = true}>
            <Icon name="send" size={15} color="#000" />
            <span>LOGIN DENGAN TELEGRAM</span>
          </button>
        </div>

      <!-- Empty State (Logged In with 0 tracks) -->
      {:else if products.length === 0}
        <div class="empty-box pixel-box">
          <div class="empty-pixel-icon">
            <Icon name="package" size={48} color="#A1A1AA" />
          </div>
          <h3 class="empty-title font-display">Belum Ada Block Produk yang Dipantau</h3>
          <p class="empty-desc font-mono">
            Halo <strong>{currentUser.first_name || "User"}</strong>! Salin link produk dari website toko favorit Anda (Agres.id, Uniqlo, Zara, Zalora, H&M, dll) dan masukkan pada form di atas untuk memulai pemantauan harga otomatis!
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

  <!-- About & How It Works Modal -->
  <AboutModal 
    isOpen={isAboutModalOpen} 
    onClose={() => isAboutModalOpen = false} 
  />

  <!-- Floating Toast Feedback System -->
  <Toast {toasts} onRemove={removeToast} />

  <!-- Neo-Brutalist Technical Footer -->
  <footer class="app-footer font-mono">
    <div class="footer-inner">
      <div class="footer-brand">
        <span class="font-pixel footer-title">PRICE_TRACKER v2.0</span>
        <span class="footer-tagline">// High-Performance Multi-Store Price Monitor & Telegram Dispatcher</span>
      </div>
      <div class="footer-meta">
        <span>Go + Svelte + Supabase</span>
        <span class="footer-divider">•</span>
        <a href="https://github.com/mfayaadhh-creator/price-tracker-app" target="_blank" rel="noopener noreferrer" class="footer-link">
          <Icon name="github" size={13} />
          <span>GitHub</span>
        </a>
        <span class="footer-divider">•</span>
        <a href="https://t.me/mf_pricetracker_bot" target="_blank" rel="noopener noreferrer" class="footer-link">
          <Icon name="send" size={13} color="var(--accent-blue)" />
          <span>@mf_pricetracker_bot</span>
        </a>
      </div>
    </div>
  </footer>
</div>

<style>
  .app-shell {
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

  .creator-section {
    margin-bottom: 32px;
  }

  .category-pills {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 14px;
  }

  .category-pill {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    background-color: var(--bg-card);
    border: 1.5px solid var(--border-color);
    padding: 6px 10px;
    font-size: 0.72rem;
    font-weight: 700;
    color: var(--text-main);
    box-shadow: 2px 2px 0px rgba(0,0,0,0.06);
  }

  .app-footer {
    background-color: var(--bg-card);
    border-top: 2px solid var(--border-color);
    padding: 24px 20px;
    margin-top: auto;
  }

  .footer-inner {
    max-width: 1280px;
    margin: 0 auto;
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 16px;
  }

  .footer-brand {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .footer-title {
    font-size: 0.95rem;
    font-weight: 700;
    color: var(--text-main);
  }

  .footer-tagline {
    font-size: 0.74rem;
    color: var(--text-muted);
  }

  .footer-meta {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 10px;
    font-size: 0.76rem;
    color: var(--text-muted);
  }

  .footer-link {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    color: var(--text-main);
    text-decoration: none;
    font-weight: 600;
  }

  .footer-link:hover {
    color: var(--accent-orange);
    text-decoration: underline;
  }

  .footer-divider {
    color: #CBD5E1;
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

    .footer-inner {
      flex-direction: column;
      align-items: flex-start;
    }
  }
</style>
