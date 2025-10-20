<script>
    import { writable } from 'svelte/store';
    import { fly } from 'svelte/transition';
    import { quintOut } from 'svelte/easing';

    // 1. 스토어 로직을 컴포넌트 내부에 포함시킵니다.
    const notifications = writable([]);

    // 2. 알림을 추가하는 함수를 `export`하여 외부에서 호출할 수 있도록 합니다.
    export function add(message, type = 'INFO', duration = 3000) {
        const id = Date.now() + Math.random();
        const notification = { id, message, type, duration };

        notifications.update(n => [...n, notification]);

        // 일정 시간 후 자동으로 알림 제거
        if (duration > 0) {
            setTimeout(() => {
                remove(notification.id);
            }, duration);
        }
    }

    // 내부적으로만 사용하는 알림 제거 함수
    function remove(id) {
        notifications.update(n => n.filter(n => n.id !== id));
    }

    function handleKeydown(event, id) {
        if (event.key === 'Enter' || event.key === ' ') {
            remove(id);
        }
    }
</script>

<div class="notifier-container">
    {#each $notifications as notification (notification.id)}
        <div
                in:fly={{ y: -30, duration: 400, easing: quintOut }}
                out:fly={{ x: '100%', duration: 300, easing: quintOut }}
                class="notification type-{notification.type}"
                on:click={() => remove(notification.id)}

                role="button"
                tabindex="0"
                on:keydown={(event) => handleKeydown(event, notification.id)}
        >
            <div class="message">{notification.message}</div>
        </div>
    {/each}
</div>

<style>
    .notifier-container {
        position: fixed;
        bottom: 1.5rem;
        right: 1.5rem;
        display: flex;
        flex-direction: column;
        align-items: flex-end;
        gap: 0.75rem;
        z-index: 9999;
    }
    .notification {
        min-width: 250px;
        max-width: 350px;
        padding: 0.8rem 1.2rem;
        border-radius: 6px;
        color: white;
        cursor: pointer;
        box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        font-weight: 500;
        font-family: var(--font-family, sans-serif);
    }
    .type-info { background: linear-gradient(to right, #3b82f6, #60a5fa); }
    .type-success { background: linear-gradient(to right, #16a34a, #22c55e); }
    .type-error { background: linear-gradient(to right, #dc2626, #ef4444); }
</style>