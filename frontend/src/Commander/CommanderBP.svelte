<script>
    import { onMount } from 'svelte';
    import { EventsOn } from '../../wailsjs/runtime';
    import Notifier from '../module/Notifier.svelte';
    import {SendData, CommanderConn, SerialList} from "../../wailsjs/go/main/App.js";
    let activeCommOption = 'tcp';
    let serialPort = 'COM1';
    let baudRate = '57600';
    let targetIp = '192.168.0.1';
    let tcpPort = '4000';

    let notifier;

    let sendOnceBoxes = ''
    let sendBoxes = [
        { id: 1, value: '' },
        { id: 2, value: '' },
        { id: 3, value: '' },
        { id: 4, value: '' },
    ];
    let SendDataHistory = '';
    let memo = '';

    // --- Telnet 모드 변수 ---
    let isTelnetMode = false; // Telnet 모드인지 여부
    let isHoveringButton = false; // 버튼 전체에 마우스가 올라와 있는지
    let isHoveringSeam = false;   // '틈'(seam) 부분에 마우스가 올라와 있는지

    // 0 = 연결 끊김, 1 = 연결됨, 2 = 대기 세 가지 상태를 가짐
    let connectionState = 0;
    async function handleConnect() {
        if (connectionState === 2) return;

        let result = false;
        if (connectionState === 1) {
            if (activeCommOption === 'serial') {
                result = await CommanderConn(-1, serialPort, baudRate);
            } else {
                result = await CommanderConn((isTelnetMode) ? -3:-2, targetIp, tcpPort);
            }
            if (result) {notifier?.add("연결 해제 실패", "error", 3000);}
        } else {
            connectionState = 2;
            // 실제 Go 함수를 호출하고 응답을 기다림 (Promise)
            if (activeCommOption === 'serial') {
                result = await CommanderConn(1, serialPort, baudRate);
            } else {
                result = await CommanderConn((isTelnetMode) ? 3:2, targetIp, tcpPort);
            }
            if (!result) {notifier?.add("연결 실패", "error", 3000);}
        }
        if (result) {
            connectionState = 1;
        } else {
            connectionState = 0;
        }
    }

    onMount(() => {
        EventsOn("disconnCommander", () => {
            if (connectionState !== 0) {
                connectionState = 0;
                isTelnetMode = false;
            }
        });
    });

    function handleIpInput(event) {
        const input = event.target;
        const value = input.value;

        // 현재 커서 위치를 기억
        const cursorPosition = input.selectionStart;
        const dotsBefore = (value.substring(0, cursorPosition).match(/\./g) || []).length;

        let cleaned = value.replace(/[^0-9.]/g, '').replace(/\.+/g, '.');
        let parts = cleaned.split('.');

        for (let i = 0; i < 4 && i < parts.length; i++) {
            if (i === 3) {
                if (parts[i].length > 3) {
                    parts[i] = parts[i].substring(0, 3);
                }
            } else {
                if (parts[i].length > 3) {
                    if (parts[i + 1] !== undefined) {
                        parts[i + 1] = parts[i].substring(3) + parts[i + 1];
                    } else {
                        parts.push(parts[i].substring(3));
                    }
                    parts[i] = parts[i].substring(0, 3);
                }
            }
        }
        const newIp = parts.join('.');

        targetIp = newIp;
        const dotsAfter = (newIp.substring(0, cursorPosition).match(/\./g) || []).length;
        const newCursorPosition = cursorPosition + (dotsAfter - dotsBefore);

        // Svelte가 화면을 업데이트한 후 커서 위치를 재설정
        setTimeout(() => {
            input.setSelectionRange(newCursorPosition, newCursorPosition);
        }, 0);
    }

    let serialPortsList = [];
    async function fetchSerialPorts() {
        try {
            serialPortsList = await SerialList();
            if (serialPortsList.length > 0 && !serialPortsList.includes(serialPort)) {
                serialPort = serialPortsList[0];
            }
        } catch (err) {
            console.error("Failed to fetch serial ports:", err);
            serialPortsList = ['Error'];
        }
    }

    function handleEnterKey(event) {
        if (event.key === 'Enter') {
            event.preventDefault();
            handleConnect();
        }
    }
</script>

<Notifier bind:this={notifier} />
<div class="commander-bp-container">
    <section class="panel comm-option-panel">
        <div class="comm-menu">
            <button
                    style="text-align: center; position: relative; height: 28px;"
                    class="menu-btn tcp-telnet-button"
                    class:active={activeCommOption === 'tcp'}
                    class:hover-button={isHoveringButton || isTelnetMode}
                    class:hover-seam={isHoveringSeam || isTelnetMode}
                    disabled={connectionState !== 0}
                    on:click={() => {
                        if (isTelnetMode) {
                            isTelnetMode = false;
                        } else {
                            if (isHoveringSeam) {
                                isTelnetMode = true;
                            }
                        }
                        if (isTelnetMode) {
                            tcpPort = '23';
                        } else {
                            tcpPort = '4000';
                        }
                        activeCommOption = 'tcp';
                    }}
                    on:mouseenter={() => isHoveringButton = true}
                    on:mouseleave={() => { isHoveringSeam = false; isHoveringButton = false; }}
            >
                <span class="btn-text tcp-text">TCP</span>
                <span class="btn-text telnet-text">Telnet</span>

                <div class="seam" on:mouseenter={() => isHoveringSeam = true}>
                    <div class="fold-shadow"></div>
                    <div class="fold-face"></div>
                </div>
            </button>
            <button style="text-align: center;" class="menu-btn" class:active={activeCommOption === 'serial'}
                    disabled={connectionState !== 0} on:click={() => {
                        activeCommOption = 'serial';
                        isTelnetMode = false;
                        fetchSerialPorts();
                    }}>
                Serial
            </button>
        </div>
            {#if activeCommOption === 'serial'}
                <div class="form-grid">
                        <label for="serial-port">Serial Port</label>
                        <select id="serial-port" style="font-size: 0.875rem; height: 32px !important;" bind:value={serialPort}
                                disabled={connectionState !== 0}>
                            {#each serialPortsList as port}
                                <option value={port}>{port}</option>
                            {/each}
                        </select>
                        <label for="baud-rate">Baud Rate</label>
                        <select id="baud-rate" style="font-size: 0.875rem; height: 32px !important;" bind:value={baudRate}
                                disabled={connectionState !== 0}>
                            <option>4800</option>
                            <option>9600</option>
                            <option>14400</option>
                            <option>19200</option>
                            <option>28800</option>
                            <option>57600</option>
                            <option>115200</option>
                        </select>
                        <button class="btn btn-primary"
                                class:connecting={connectionState === 2}
                                class:connected={connectionState === 1}
                                disabled={connectionState === 2}
                                style="width: 100%;"
                                on:click={() => handleConnect()}>
                            {#if connectionState === 2}
                                Connecting
                            {:else if connectionState === 1}
                                Disconnect
                            {:else}
                                Connect
                            {/if}
                        </button>
                </div>
            {:else if activeCommOption === 'tcp'}
                    <div class="form-grid">
                        <label for="target-ip">Target IP</label>
                        <input type="text" id="target-ip" bind:value={targetIp}
                               disabled={connectionState !== 0}
                               on:input={handleIpInput}
                               on:keydown={handleEnterKey}
                               style="height: 32px !important; font-size: 0.68rem; "/>
                        <label for="tcp-port">Port</label>
                        <input style="height: 32px !important;" type="text" id="tcp-port" bind:value={tcpPort}
                               disabled={connectionState !== 0}
                               on:keydown={handleEnterKey}/>

                        <button class="btn btn-primary"
                                class:connecting={connectionState === 2}
                                class:connected={connectionState === 1}
                                disabled={connectionState === 2}
                                style="width: 100%;"
                                on:click={handleConnect}>
                            {#if connectionState === 2}
                                Connecting
                            {:else if connectionState === 1}
                                Disconnect
                            {:else}
                                Connect
                            {/if}
                        </button>
                </div>
            {/if}
    </section>

    <div class="panel send-data-panel">
        <div class="send-box-list">
            <div class="section-title" style="margin-bottom: 0; font-size: 0.875rem; color: var(--text-muted-color); ">Send</div>
            <div class="send-data-box">
                <input type="text" bind:value={sendOnceBoxes} placeholder="Send Once"
                    on:keydown={(event) => {
                    if (event.key === 'Enter') {
                    if (sendOnceBoxes !== '') {
                        event.preventDefault(); // 기본 Enter 동작 방지
                        SendData((activeCommOption === 'serial') ? serialPort:targetIp, tcpPort, sendOnceBoxes)
                        sendOnceBoxes = '';
                    }
                    }
                }}/>
                <button class="btn" on:click={() => {
                    if (sendOnceBoxes !== '') {
                    SendData((activeCommOption === 'serial') ? serialPort:targetIp, tcpPort, sendOnceBoxes)
                    sendOnceBoxes = '';
                    }
                }}>Send Once</button>
            </div>
            {#each sendBoxes as box (box.id)}
                <div class="send-data-box">
                    <input type="text" bind:value={box.value} placeholder="Send to Data"
                        on:keydown={(event) => {
                        if (event.key === 'Enter') {
                        if (box.value !== '') {
                            event.preventDefault();
                            SendData((activeCommOption === 'serial') ? serialPort:targetIp, tcpPort, box.value)
                        }
                        }
                    }}/>
                    <button class="btn" on:click={() => {
                        if (box.value !== '') {
                        SendData((activeCommOption === 'serial') ? serialPort:targetIp, tcpPort, box.value)
                        }
                    }}>Send</button>
                </div>
            {/each}
        </div>

        <div class="bott-box">
            <label for="data-send-history">Sent Data History</label>
            <textarea id="data-send-history" readonly bind:value={SendDataHistory}></textarea>
        </div>
        <div class="bott-box">
            <label for="memo">Memo</label>
            <textarea id="memo" bind:value={memo}></textarea>
        </div>
    </div>
</div>


<style>
    .commander-bp-container {
        display: grid;
        gap: 1rem;
        grid-template-columns: 0.1fr auto; /* 왼쪽(콘텐츠크기), 오른쪽(남은공간) */
        height: 100%;
    }

    /* 통신 설정 패널 */
    .comm-option-panel {
        display: flex;
        gap: 0.2rem;
        width: 100px;
        border-radius: 0.5rem;
        padding: 0.5rem;
    }

    /* 통신 설정 내부의 미니 메뉴 */
    .comm-menu {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }

    /* 전역 .menu-btn과 충돌을 피하기 위해 이름을 유지 */
    .comm-menu .menu-btn {
        padding: 0.3rem;
        font-size: 0.9rem;
        background-color: var(--btn-bg);
        border: 1px solid var(--border-color);
        color: var(--text-muted-color);
        cursor: pointer;
        transition: all 0.2s;
    }
    .comm-menu .menu-btn:hover { background-color: var(--btn-hover-bg); }
    .comm-menu .menu-btn.active { background-color: var(--accent-color); color: white; }

    /* 통신 설정 콘텐츠 영역 */
    .form-grid {
        display: flex;
        flex-direction: column;
        align-items: center;
        height: 100%;
    }

    /* 데이터 전송 패널 */
    .send-data-panel {
        display: grid;
        grid-template-columns: 1fr 1fr 1fr;
        gap: 0.5rem;
    }
    .send-box-list {
        display: flex;
        flex-direction: column;
        gap: 0.4rem;
    }
    .send-data-box {
        display: flex;
        gap: 0.3rem;
        align-items: center;
    }
    .send-data-box button {
        flex-shrink: 0;
    }
    .bott-box {
        display: flex;
        flex-direction: column;
        height: 100%;
    }
    .bott-box textarea {
        flex-grow: 1; /* 남은 공간을 모두 차지 */
        resize: none;
    }

    .btn.connected {
        /* 평소에는 일반 버튼과 비슷하게 보이되, 살짝 다른 색상 사용 */
        background-color: var(--btn-bg);
        border: 1px solid var(--border-color);
    }
    /* '연결됨' 상태에서 호버 시, 위험/해제 액션임을 알리는 색상으로 변경 */
    .btn.connected:hover {
        background-color: var(--danger-color);
        border-color: var(--danger-hover-color);
        color: white;
        font-weight: 600;
    }
    /* '연결 중' 상태 */
    .btn.connecting {
        cursor: wait;

        background-size: 40px 40px;
        background-image: linear-gradient(
                -45deg,
                rgba(255, 255, 255, 0.1) 25%,
                transparent 25%,
                transparent 50%,
                rgba(255, 255, 255, 0.1) 50%,
                rgba(255, 255, 255, 0.1) 75%,
                transparent 75%,
                transparent
        );
        animation: connecting-stripes 2s linear infinite;
    }
    .tcp-telnet-button {
        position: relative;
        overflow: hidden;
        transition: background-color 0.4s ease-out;
    }

    /* '틈' 역할을 하는 컨테이너 */
    .seam {
        position: absolute;
        right: 0;
        bottom: 0;
        width: 20px;
        height: 20px;
        z-index: 10;
        cursor: pointer;

        transform-origin: bottom right; /* 오른쪽 아래를 기준으로 커짐 */
        transition: transform 0.4s cubic-bezier(0.2, 0.8, 0.2, 1);
    }

    /* 접힌 모양을 만드는 삼각형들 */
    .fold-shadow, .fold-face {
        content: '';
        position: absolute;
        right: 0;
        bottom: 0;
        border-style: solid;
        border-width: 0; /* 평상시엔 안 보임 */
        transition: all 0.3s ease-out;
    }
    .fold-shadow {
        border-color: transparent rgba(0, 0, 0, 0.4) transparent transparent;
    }
    .fold-face {
        border-color: transparent #cbd5e1 transparent transparent; /* var(--text-color) */
    }

    /* 1단계: 버튼에 마우스 올리면 fold 효과 나타남 */
    .tcp-telnet-button.hover-button .fold-shadow,
    .tcp-telnet-button.hover-button .fold-face {
        border-width: 20px 20px 0 0;
    }

    /* 2단계: seam에 마우스 올리면 seam 컨테이너 자체가 커짐 */
    .tcp-telnet-button.hover-seam .seam {
        /* 버튼을 완전히 덮을 만큼 충분히 크게 확대 */
        transform: scale(15);
    }

    /* 2단계: 버튼 배경색과 글자 전환 */
    .tcp-telnet-button.hover-seam {
        /* Telnet 모드일 때의 배경색 */
        background-color: var(--btn-bg);
    }

    .btn-text {
        position: absolute;
        transition: opacity 0.3s ease-in-out;
        white-space: nowrap;
        z-index: 11;
        left: 50%;
        top: 48%;
        transform: translate(-50%, -50%);
    }
    .tcp-text { opacity: 1; }
    .telnet-text { opacity: 0;
        color: var(--accent-color); }

    .tcp-telnet-button.hover-seam .tcp-text { opacity: 0; }
    .tcp-telnet-button.hover-seam .telnet-text { opacity: 1; }
    @keyframes connecting-stripes {
        from { background-position: 40px 0; }
        to { background-position: 0 0; }
    }
</style>