<script>
  export let toasts = [];
  export let onRemove;
</script>

<div class="toast-container">
  {#each toasts as t (t.id)}
    <div class="toast-item pixel-box toast-{t.type} font-mono">
      <div class="toast-content">
        <span class="toast-icon">
          {#if t.type === 'success'}✅
          {:else if t.type === 'error'}❌
          {:else}⚡{/if}
        </span>
        <span class="toast-text">{t.message}</span>
      </div>
      <button 
        type="button" 
        on:click={() => onRemove(t.id)} 
        class="toast-close"
      >
        ✕
      </button>
    </div>
  {/each}
</div>

<style>
  .toast-container {
    position: fixed;
    bottom: 24px;
    right: 24px;
    z-index: 100;
    display: flex;
    flex-direction: column;
    gap: 10px;
    max-width: 420px;
    width: calc(100% - 48px);
  }

  .toast-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    font-size: 0.88rem;
    font-weight: 700;
    border-radius: 0;
    animation: slideUp 0.2s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  }

  .toast-success {
    background-color: var(--accent-green-light);
    border-color: #065F46;
    color: #065F46;
  }

  .toast-error {
    background-color: #FEE2E2;
    border-color: #DC2626;
    color: #991B1B;
  }

  .toast-info {
    background-color: var(--accent-yellow);
    border-color: #854D0E;
    color: #854D0E;
  }

  .toast-content {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .toast-close {
    background: transparent;
    border: none;
    font-size: 0.9rem;
    font-weight: 800;
    cursor: pointer;
    color: inherit;
    padding: 2px 6px;
  }

  @keyframes slideUp {
    from {
      transform: translateY(20px);
      opacity: 0;
    }
    to {
      transform: translateY(0);
      opacity: 1;
    }
  }
</style>
