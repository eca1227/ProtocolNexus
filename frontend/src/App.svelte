<script>
  import Commander from "./Commander/Commander.svelte";
  import CommanderBP from "./Commander/CommanderBP.svelte";
  import EFEMTest from "./EFEMTest/EFEMTest.svelte";
  import EfemTestBP from "./EFEMTest/EFEMTestBP.svelte";
  import LPMaint from "./LPMaint/LPMaint.svelte";
  import {SetPage} from "../wailsjs/go/main/App.js";

  let activeView = 'Commander';

  function activeMenu(menuName) {
    if (activeView !== menuName) {
      activeView = menuName;
      SetPage(menuName);
    }
  }
</script>

<div class="nexus-container">
  <nav class="main-menu">
    <ul>
      <li>
        <button class="menu-btn" class:active={activeView === 'Commander'} on:click={() => activeMenu('Commander')}>
          Commander
        </button>
      </li>
      <li>
        <button class="menu-btn" class:active={activeView === 'EFEM Test'} on:click={() => activeMenu('EFEM Test')}>
          EFEM Test
        </button>
      </li>
      <li>
        <button class="menu-btn" class:active={activeView === 'LP Maint'} on:click={() => activeMenu('LP Maint')}>
          LP Maint
        </button>
      </li>
    </ul>
  </nav>

  <main class="main-grid">
    <div style="display: {activeView === 'Commander' ? 'contents' : 'none'}">
      <Commander/>
    </div>
    <div style="display: {activeView === 'EFEM Test' ? 'contents' : 'none'}">
      <EFEMTest/>
    </div>
    <div style="display: {activeView === 'LP Maint' ? 'contents' : 'none'}">
      <LPMaint/>
    </div>
  </main>

  <section class="bottom-panel">
    {#if activeView === 'Commander'}
      <CommanderBP />
    {:else if activeView === 'EFEM Test'}
      <EfemTestBP />
    {/if}
  </section>
</div>
