<script>
    // item     현재 위치에 표시될 실제 데이터 객체. (예: { id: 1, text: '...' }) dummy가 true이면 undefined
    // dummy	현재 항목이 실제 데이터가 아닌, 공간을 채우기 위한 가짜(dummy) 항목인지 여부를 나타내는 boolean 값
    // y	    해당 항목이 위치해야 할 **수직(top) 위치값(px)** style="top: {y}px;" 형태로 반드시 적용
    export let footerHeight = 0; // 컨테이너 내부에 position: sticky로 고정된 푸터가 있을 경우 그 높이를 지정
    export let headerHeight = 0; // 컨테이너 내부에 position: sticky로 고정된 헤더가 있을 경우 그 높이를 지정
    export let items; // 전체 데이터 배열
    export let itemHeight; // 각 항목의 고정된 높이(px) CSS에 설정된 높이와 반드시 일치해야 함. 가변 높이는 지원하지 않음
    export let containerHeight; // 스크롤이 일어나는 컨테이너의 높이(px) 보통 bind:clientHeight를 통해 자동으로 값을 얻음
    export let scrollTop; // 스크롤 컨테이너의 현재 스크롤 위치(px) requestAnimationFrame을 사용해 부모 컴포넌트에서 지속적으로 추적하고 전달
    export let overscanCount = 5; // 미리 렌더링할 항목 수

    const dummySymbol = Symbol('dummy item');

    $: lostHeight = headerHeight + footerHeight;
    $: visibleHeight = containerHeight - lostHeight;
    $: spacerHeight = items.length * itemHeight;
    $: numItems = Math.ceil(visibleHeight / itemHeight) + 1;
    $: startIndex = Math.max(0, Math.floor(scrollTop / itemHeight) - overscanCount);
    $: endIndex = Math.min(items.length, startIndex + numItems + overscanCount * 2);
    $: slice = items.slice(startIndex, endIndex);
</script>

<div class="spacer" style="height: {spacerHeight}px;" tabindex="-1" on:keydown on:wheel>
    {#each slice as item, index}
        <slot
                item={item === dummySymbol ? undefined : item}
                dummy={item === dummySymbol}
                y={(startIndex + index) * itemHeight}
        />
    {/each}
</div>

<style>
    .spacer {
        width: 100%;

        /* Prevent the translated items from bleeding through, causing more scrolling */
        overflow: hidden;

        /* 2021 inline-block happy fun time  */
        /*font-size: 0;*/
        /*line-height: 0;*/
    }
</style>
