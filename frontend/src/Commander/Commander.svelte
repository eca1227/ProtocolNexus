<script>
    import VirtualList from './VirtualList.svelte';
    import {onMount, onDestroy, tick} from 'svelte';
    import {EventsOn} from "../../wailsjs/runtime/runtime.js";

    // type == {'info', 'received', 'sent', 'error'}
    let logLines = [];
    $: processedLogLines = logLines.flatMap(log => {
        const lines = log.text.split('\n');

        if (lines.length > 1 && lines[lines.length - 1] === '') {
            lines.pop();
        }
        return lines.map((lineText, index) => ({
            ...log, // type과 같은 원래 속성을 복사
            type: index === 0 ? log.type : `${log.type}Multi`,
            text: lineText,
            isFirstLine: index === 0, // 첫 번째 줄(index가 0)인지 여부를 플래그로 저장
        }));
    });

    // --- 스크롤 계산을 위한 변수들 ---
    let logDisplayEl; // 스크롤 컨테이너 div
    let scrollTop = 0;
    let containerHeight = 0;
    let itemHeight = 18; // 각 로그 한 줄의 높이(px)
    let frame;

    function poll() {
        if (logDisplayEl && logDisplayEl.scrollTop !== scrollTop) {
            scrollTop = logDisplayEl.scrollTop;
        }
        frame = requestAnimationFrame(poll);
    }

    onMount(() => {
        // 스크롤 위치 추적 시작
        frame = requestAnimationFrame(poll);

        const cleanup = EventsOn('CommanderLog', (cl) => {
            let isScrolledToBottom = false;
            if (logDisplayEl) {
                isScrolledToBottom = logDisplayEl.scrollTop + logDisplayEl.clientHeight >= logDisplayEl.scrollHeight - 5;
            }

            logLines.push({ type: cl.dataType, text: cl.dataText });
            logLines = logLines;

            if (isScrolledToBottom) {
                // Svelte가 DOM을 업데이트(새 로그를 그리는 것)한 후에 스크롤을 변경해야
                // 정확한 위치로 이동할 수 있습니다. `tick()`이 이 타이밍을 보장해 줍니다.
                tick().then(() => {
                    if (logDisplayEl) {
                        logDisplayEl.scrollTop = logDisplayEl.scrollHeight;
                    }
                });
            }
        });

        onDestroy(() => {
            // 컴포넌트 소멸 시 이벤트 리스너와 프레임 요청 정리
            cleanup();
            cancelAnimationFrame(frame);
        });
    });
</script>

        <div class="commander-container">
            <div class="panel data-display-panel">
                <label for="comm-data">Communication Data</label>

                <div class="log-display"
                        bind:this={logDisplayEl}
                        bind:clientHeight={containerHeight}>
                    <VirtualList items={processedLogLines} {itemHeight} {containerHeight} {scrollTop} let:item let:dummy let:y>

                        <div class="log-line {item.type}" class:dummy style="top: {y}px; height: {itemHeight}px;">
                            {#if !dummy}
                                <span class="log-prefix">
                                    {#if item.isFirstLine}
                                        {#if item.type === 'sent'}
                                            &gt;&gt;
                                        {:else if item.type === 'received'}
                                            &lt;&lt;
                                        {:else if item.type === 'error'}
                                            !!
                                        {/if}
                                    {/if}
                                </span>
                                <span class="log-text">{item.text}</span>
                            {/if}
                        </div>
                    </VirtualList>
                </div>

            </div>
        </div>

<style>
    @import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono&display=swap');

    .commander-container {
        display: grid;
        grid-column: 1 / -1;
        height: 100%;
        width: 100%;
        overflow: hidden;
    }

    /* 1. 상단 데이터 출력 패널 위치 */
    .data-display-panel {
        grid-column: 1 / -1;
        grid-row: 1 / 2;
        display: flex;
        flex-direction: column;
        overflow: hidden;
    }
    .log-display {
        flex-grow: 1;
        min-height: 0;
        position: relative;
        background-color: var(--input-bg);
        border: 1px solid var(--border-color);
        color: var(--text-color);
        border-radius: 0.375rem;
        font-size: 13px;
        overflow-y: auto;
        padding-top: 0.2rem;
        padding-bottom: 0.2rem;
        box-sizing: border-box; /* 패딩이 요소 크기에 포함되도록 설정 */
    }

    /* 각 로그 라인의 기본 스타일 */
    .log-line {
        position: absolute; /* y 값으로 위치를 잡기 위해 필수 */
        width: 100%;        /* 부모 (.log-display) 너비에 맞춤 */
        white-space: pre-wrap; /* 공백과 줄바꿈 유지 */
        box-sizing: border-box; /* 패딩을 포함한 높이 계산 */
        padding: 0 0.5rem;      /* 좌우 여백, 높이에 포함되지 않음 */
    }
    .log-line::after {
        content: "\A";
    }
    .dummy {
        /* 보이지 않는 더미 아이템 */
        visibility: hidden;
    }

    /* === 타입별 색상 지정 === */
    .log-line[class*="info"] {
        color: var(--text-muted-color);
    }
    .log-line[class*="sent"] {
        color: #60a5fa;
    }
    .log-line[class*="received"] {
        color: #4ade80;
    }
    .log-line[class*="error"] {
        color: var(--danger-color);
        font-weight: bold;
    }

</style>