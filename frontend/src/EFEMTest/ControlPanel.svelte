<script>
    // 부모로부터 '제목'과 '데이터 배열'을 전달받습니다.
    export let title = '';
    export let items = [];

    let selectedIndex = 0;

    // 데이터가 변경될 때마다 선택된 항목을 안전하게 업데이트합니다.
    $: selectedItem = items[selectedIndex];

    function navigate(direction) {
        selectedIndex = (selectedIndex + direction + items.length) % items.length;
    }
</script>

<div class="panel">
    <div class="panel-header">
        <h2>{title}</h2>
        {#if items.length > 0}
            <div class="selector">
                <div class="coin-selector">
                    {#each items as item, i}
                        <button
                                class="coin"
                                class:active={selectedIndex === i}
                                on:click={() => selectedIndex = i}
                                style="z-index: {items.length - i};"
                        >
                            <span>{i + 1}</span>
                        </button>
                    {/each}
                </div>
                <div class="selector-nav">
                    <button class="nav-btn" on:click={() => navigate(-1)}>&lt;</button>
                    <span class="nav-index">{selectedIndex + 1}</span>
                    <button class="nav-btn" on:click={() => navigate(1)}>&gt;</button>
                </div>
            </div>
        {/if}
    </div>

    {#if selectedItem}
        <div class="content-grid" style="column-gap: 0;">
            <div>
                <div style="margin-top: 0.3rem;">
                {#each ['Run', 'Stop', 'Origin', 'Overrun', 'Alarm'] as status}
                    <div class="status-indicator">
                        <span class="light-indicator"
                              class:on={selectedItem[status.toLowerCase()]}
                              class:red={status === 'Overrun' || status === 'Alarm'}
                              class:green={status !== 'Overrun' && status !== 'Alarm'}
                        ></span>
                        <span>{status}</span>
                    </div>
                {/each}
                </div>
                <button class="btn" style="margin-top: 0.85rem; padding-left: 10px; padding-right: 10px;">ReConnect</button>
            </div>
            <div class="input-group">
                <div class="input-row">
                    <span style="margin-right: 5px;">IP/Port</span>
                    <input type="text" bind:value={selectedItem.port}/>
                </div>
                <div class="grid-4-buttons">
                    <button class="status-btn">ERR</button>
                    <button class="status-btn">DRT</button>
                    <button class="status-btn">ORG</button>
                    <button class="status-btn">Demo</button>
                </div>
                <div class="write-action">
                    <input type="text" placeholder="Command..." bind:value={selectedItem.writeCmd} />
                    <button class="btn">Write</button>
                </div>
            </div>
        </div>
        <div class="bottom-controls">
            <slot name="actions"></slot>
        </div>
    {:else}
        <div class="no-data-message">
            No data available.
        </div>
    {/if}
</div>

<style>
    .no-data-message {
        flex-grow: 1;
        display: flex;
        align-items: center;
        justify-content: center;
        color: var(--text-muted-color);
        border-top: 1px solid var(--border-color);
        margin-top: 0.4rem;
        padding-top: 0.75rem;
    }
</style>