<script>
  export let product;
  export let index;
  export let total;
  export let onDelete;
  export let onMoveLeft;
  export let onMoveRight;
  export let onDragStart;
  export let onDragOver;
  export let onDrop;
  export let onPreviewImage = () => {};

  let isDeleting = false;
  let isDragging = false;
  let isDragOver = false;
  let imageError = false;

  function formatIDR(amount) {
    if (!amount && amount !== 0) return "Rp 0";
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      maximumFractionDigits: 0
    }).format(amount);
  }

  function calculateSavings(base, current) {
    if (base > current) {
      return base - current;
    }
    return 0;
  }

  async function handleDelete() {
    if (confirm(`Apakah Anda yakin ingin menghapus pemantauan untuk "${product.name}"?`)) {
      isDeleting = true;
      await onDelete(product.id);
      isDeleting = false;
    }
  }

  function handleInternalDragStart(e) {
    isDragging = true;
    onDragStart(e, index);
  }

  function handleInternalDragEnd() {
    isDragging = false;
    isDragOver = false;
  }

  function handleInternalDragOver(e) {
    onDragOver(e, index);
    isDragOver = true;
  }

  function handleInternalDragLeave() {
    isDragOver = false;
  }

  function handleInternalDrop(e) {
    isDragOver = false;
    isDragging = false;
    onDrop(e, index);
  }

  $: savings = calculateSavings(product.base_price, product.last_price);
  $: hasDiscount = product.is_discount || (product.base_price > product.last_price);
  $: isUnavailable = product.status === "unavailable";
</script>

<article 
  class="product-block pixel-box {isDragging ? 'is-dragging' : ''} {isDragOver ? 'is-drag-over' : ''} {isUnavailable ? 'is-unavailable' : ''}"
  aria-label={product.name}
  draggable="true"
  on:dragstart={handleInternalDragStart}
  on:dragend={handleInternalDragEnd}
  on:dragover={handleInternalDragOver}
  on:dragleave={handleInternalDragLeave}
  on:drop={handleInternalDrop}
>
  <!-- Block Header Bar -->
  <div class="block-header">
    <div class="header-left">
      <span class="drag-handle font-mono" title="Tahan & geser untuk mengubah urutan">
        <span class="drag-icon">⠿</span>
        <span class="block-num font-pixel">BLOCK_{index + 1}</span>
      </span>
      <span class="pixel-tag pixel-tag-dark">
        {product.platform || "UNIQLO"}
      </span>
    </div>

    <!-- Quick Move & Delete Actions -->
    <div class="header-right">
      <div class="move-controls">
        <button 
          type="button" 
          class="move-btn"
          disabled={index === 0}
          on:click={() => onMoveLeft(index)}
          title="Geser block ke kiri"
        >
          ◀
        </button>
        <button 
          type="button" 
          class="move-btn"
          disabled={index === total - 1}
          on:click={() => onMoveRight(index)}
          title="Geser block ke kanan"
        >
          ▶
        </button>
      </div>

      <button 
        type="button" 
        class="delete-btn"
        disabled={isDeleting}
        on:click={handleDelete}
        title="Hapus pemantauan produk ini"
      >
        ✕
      </button>
    </div>
  </div>

  <!-- Block Body -->
  <div class="block-body">
    <!-- Product Showcase: Image & Meta Info -->
    <div class="product-showcase">
      {#if product.image_url && !imageError}
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
        <div 
          class="image-wrapper" 
          role="button" 
          tabindex="0"
          on:click={() => onPreviewImage(product)} 
          title="Klik untuk melihat foto ukuran penuh (HD)"
        >
          <img 
            src={product.image_url} 
            alt={product.name} 
            loading="lazy" 
            class="product-image"
            on:error={() => imageError = true}
          />
          <span class="zoom-badge font-mono">🔍 HD</span>
        </div>
      {:else}
        <div class="image-placeholder font-pixel" title="Foto produk default">
          <span>👕</span>
        </div>
      {/if}

      <div class="info-group">
        <!-- Product Name -->
        <h3 class="product-title font-display" title={product.name}>
          {product.name}
        </h3>

        <!-- Product Metadata Badges -->
        <div class="meta-badges">
          {#if isUnavailable}
            <span class="pixel-tag pixel-tag-danger font-mono">
              ⚠️ TIDAK TERSEDIA (404)
            </span>
          {/if}
          <span class="pixel-tag pixel-tag-orange font-mono">
            ID: {product.product_id || "N/A"}
          </span>
          <span class="pixel-tag pixel-tag-blue font-mono" title="Notifikasi dikirim ke Chat ID ini">
            📱 TG: {product.user_phone}
          </span>
          {#if product.target_price > 0}
            <span class="pixel-tag pixel-tag-green font-mono">
              🎯 TARGET: {formatIDR(product.target_price)}
            </span>
          {/if}
        </div>
      </div>
    </div>

    <!-- Pricing Showcase Box -->
    <div class="pricing-card">
      <div class="price-row">
        <div class="price-group">
          <span class="price-label font-mono">HARGA SAAT INI</span>
          <span class="price-current font-mono font-pixel">
            {formatIDR(product.last_price)}
          </span>
        </div>

        {#if hasDiscount}
          <div class="price-original-group">
            <span class="price-label font-mono">HARGA AWAL</span>
            <span class="price-original font-mono">
              {formatIDR(product.base_price)}
            </span>
          </div>
        {/if}
      </div>

      <!-- Status / Discount Banner -->
      {#if isUnavailable}
        <div class="unavailable-banner font-mono font-pixel">
          <span>⚠️ PRODUK HABIS / HALAMAN TIDAK TERSEDIA</span>
        </div>
      {:else if hasDiscount}
        <div class="savings-banner font-mono font-pixel">
          <span>🔥 HEMAT: {formatIDR(savings)}</span>
        </div>
      {:else}
        <div class="normal-banner font-mono">
          <span>● HARGA NORMAL (STANDAR)</span>
        </div>
      {/if}
    </div>
  </div>

  <!-- Block Footer Actions -->
  <div class="block-footer">
    <a 
      href={product.url} 
      target="_blank" 
      rel="noopener noreferrer"
      class="pixel-btn visit-btn font-mono"
    >
      <span>BUKA WEB {product.platform || "TOKO"}</span>
      <span>↗</span>
    </a>
  </div>
</article>

<style>
  .product-block {
    background-color: var(--bg-card);
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    cursor: grab;
    user-select: none;
    border-radius: 0;
    transition: transform 0.18s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.18s ease, box-shadow 0.18s ease;
  }

  .product-block:active {
    cursor: grabbing;
  }

  /* Award-Winning Translucent Ghost Effect saat Di-Hold / Di-Drag */
  .product-block.is-dragging {
    opacity: 0.28 !important;
    transform: scale(0.96) rotate(-1.2deg) !important;
    filter: grayscale(35%) !important;
    border-style: dashed !important;
    border-color: var(--accent-orange) !important;
    box-shadow: none !important;
  }

  /* Magnetic Drop Slot Indicator */
  .product-block.is-drag-over {
    transform: scale(1.025) translateY(-3px) !important;
    border-color: var(--accent-orange) !important;
    background-color: #FFFDF5 !important;
    box-shadow: 0 0 0 3px var(--accent-orange), var(--shadow-lg) !important;
  }

  .block-header {
    background-color: var(--bg-canvas-subtle);
    padding: 10px 14px;
    border-bottom: var(--border-width) solid var(--border-color);
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 8px;
  }

  .header-left, .header-right {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .drag-handle {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 0.75rem;
    font-weight: 700;
    color: var(--text-main);
  }

  .drag-icon {
    font-size: 1.1rem;
    color: var(--accent-orange);
    letter-spacing: -2px;
  }

  .block-num {
    font-size: 0.75rem;
    background-color: #E4E4E7;
    padding: 2px 6px;
    border: 1px solid var(--border-color);
  }

  .move-controls {
    display: flex;
    gap: 2px;
  }

  .move-btn {
    background-color: #FFFFFF;
    border: 1.5px solid var(--border-color);
    padding: 2px 6px;
    font-size: 0.7rem;
    cursor: pointer;
    box-shadow: 1px 1px 0px var(--border-color);
  }

  .move-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  .delete-btn {
    background-color: #FEE2E2;
    border: 1.5px solid #DC2626;
    color: #DC2626;
    padding: 2px 8px;
    font-weight: 700;
    font-size: 0.8rem;
    cursor: pointer;
    box-shadow: 1px 1px 0px #DC2626;
    transition: background-color 0.1s ease;
  }

  .delete-btn:hover {
    background-color: #DC2626;
    color: #FFFFFF;
  }

  .block-body {
    padding: 18px 16px;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .product-showcase {
    display: flex;
    gap: 12px;
    align-items: flex-start;
  }

  .image-wrapper {
    width: 80px;
    height: 100px;
    flex-shrink: 0;
    border: 2px solid var(--border-color);
    background-color: #FAFAFA;
    box-shadow: 2px 2px 0px var(--border-color);
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
    position: relative;
    cursor: zoom-in;
    transition: transform 0.15s ease, box-shadow 0.15s ease;
  }

  .image-wrapper:hover {
    transform: translateY(-2px);
    box-shadow: 3px 3px 0px var(--border-color);
  }

  .zoom-badge {
    position: absolute;
    bottom: 2px;
    right: 2px;
    background-color: rgba(24, 24, 27, 0.85);
    color: #FFFFFF;
    font-size: 0.6rem;
    font-weight: 700;
    padding: 1px 4px;
    border: 1px solid var(--border-color);
    opacity: 0.85;
    transition: opacity 0.15s ease, background-color 0.15s ease;
  }

  .image-wrapper:hover .zoom-badge {
    opacity: 1;
    background-color: var(--accent-orange);
  }

  .product-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
    transition: transform 0.25s ease;
  }

  .image-wrapper:hover .product-image {
    transform: scale(1.08);
  }

  .image-placeholder {
    width: 80px;
    height: 100px;
    flex-shrink: 0;
    border: 2px solid var(--border-color);
    background-color: var(--bg-canvas-subtle);
    box-shadow: 2px 2px 0px var(--border-color);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.8rem;
  }

  .info-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
    flex: 1;
    min-width: 0;
  }

  .product-title {
    font-size: 0.95rem;
    font-weight: 700;
    line-height: 1.35;
    color: var(--text-main);
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .meta-badges {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .pricing-card {
    background-color: #FAFAFA;
    border: 2px solid var(--border-color);
    padding: 12px;
    box-shadow: 2px 2px 0px rgba(0,0,0,0.06);
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .price-row {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    flex-wrap: wrap;
    gap: 8px;
  }

  .price-group, .price-original-group {
    display: flex;
    flex-direction: column;
  }

  .price-label {
    font-size: 0.65rem;
    color: var(--text-muted);
    font-weight: 700;
  }

  .price-current {
    font-size: 1.25rem;
    font-weight: 800;
    color: var(--accent-orange);
    letter-spacing: -0.5px;
  }

  .price-original {
    font-size: 0.85rem;
    text-decoration: line-through;
    color: #71717A;
    font-weight: 600;
  }

  .product-block.is-unavailable {
    opacity: 0.82;
    filter: grayscale(15%);
    border-color: #B91C1C !important;
  }

  .unavailable-banner {
    background-color: #FEE2E2;
    border: 1.5px solid #DC2626;
    color: #991B1B;
    padding: 6px 8px;
    font-size: 0.72rem;
    font-weight: 800;
    text-align: center;
    box-shadow: 1px 1px 0px #DC2626;
  }

  .savings-banner {
    background-color: var(--accent-yellow);
    border: 1.5px solid var(--border-color);
    color: #854D0E;
    padding: 4px 8px;
    font-size: 0.72rem;
    font-weight: 800;
    text-align: center;
  }

  .normal-banner {
    background-color: #F4F4F5;
    color: var(--text-muted);
    padding: 3px 6px;
    font-size: 0.7rem;
    font-weight: 600;
    text-align: center;
  }

  .block-footer {
    padding: 12px 16px;
    background-color: var(--bg-canvas);
    border-top: 2px solid var(--border-color);
  }

  .visit-btn {
    width: 100%;
    font-size: 0.8rem;
    padding: 8px 12px;
  }
</style>
