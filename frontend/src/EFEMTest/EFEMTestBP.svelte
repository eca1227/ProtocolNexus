<script>
    const systemList = ['12CH (4port)', 'CTS 2P', 'highspeedWTR', 'highspeedWTR_4', 'K5000_4port', 'KCTECH CMP EFEM (4port)', 'KCTECH CMP STA EFEM (3port)', 'KCTECH CMP STBa EFEM (3port)'];
    let repeatCount = 1;
</script>

<div class="panel">
    <div class="cycle-test-grid">
        <div class="system-list-container">
            <label for="system-list"><h2>Cycle Test</h2></label>
            <select id="system-list" size="8">
                {#each systemList as system}
                    <option>{system}</option>
                {/each}
            </select>
        </div>

        <div class="cycle-controls">
            <div>
                <label for="test-condition">Test Condition</label>
                <textarea id="test-condition" style="min-height: 162px; resize: none;"></textarea>
            </div>
            <div style="flex-direction: column">
                <div class="repeat-current-row">
                    <div class="repeat-container">
                        <label for="repeat">Repeat</label>
                        <input id="repeat" type="number" bind:value={repeatCount} />
                    </div>
                    <div class="current-container">
                        <label for="current">Current</label>
                        <input id="current" type="number" bind:value={repeatCount} />
                    </div>
                </div>
                <div class="action-buttons">
                    <button class="btn btn-full btn-primary">Start</button>
                    <button class="btn btn-full">Stop</button>
                    <button class="btn btn-full btn-danger">EMO</button>
                </div>
            </div>
        </div>
    </div>
</div>
