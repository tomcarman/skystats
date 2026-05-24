<script>
  import { IconSettings } from '@tabler/icons-svelte';
  import ThemeSelector from './components/ThemeSelector.svelte';
  import AboveTimeline from './components/AboveTimeline.svelte';
  import TabRouteStats from './components/TabRouteStats.svelte';
  import TabMotionStats from './components/TabMotionStats.svelte';
  import TabInterestingStats from './components/TabInterestingStats.svelte';
  import TabActivity from './components/TabActivity.svelte';
  import Footer from './components/Footer.svelte';
  import Settings from './components/Settings.svelte';

  let activeTab = 'activity';
  let tabsElement;

  const tabs = [
    { name: 'activity', label: 'Activity', component: TabActivity },
    { name: 'route-stat', label: 'Route Information', component: TabRouteStats },
    { name: 'interesting-stat', label: 'Interesting Aircraft', component: TabInterestingStats },
    { name: 'motion-stat', label: 'Record Holders', component: TabMotionStats }
  ];

  function setActiveTab(tabName) {
    activeTab = tabName;
    if (tabsElement) {
      const yOffset = -60;
      const y = tabsElement.getBoundingClientRect().top + window.pageYOffset + yOffset;
      window.scrollTo({ top: y, behavior: 'smooth' });
    }
  }

  function openSettingsModal() {
      document.getElementById("settings-modal").showModal();
  }

</script>


<div class="navbar bg-base-100 shadow-sm">
  <div class="navbar-start">
    <!-- <div class="dropdown">
      <div tabindex="0" role="button" class="btn btn-ghost btn-circle">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"> <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7" /> </svg>
      </div>
      <ul
        tabindex="0"
        class="menu menu-sm dropdown-content bg-base-100 rounded-box z-1 mt-3 w-52 p-2 shadow">
        <li><a>Menu 1</a></li>
        <li><a>Menu 2</a></li>
      </ul>
    </div> -->
  </div>
  <div class="navbar-center">
    <h1 class="text-4xl font-normal text-primary drop-shadow-[0_0_15px_rgba(59,130,246,0.5)]">
      Skystats
    </h1>
  </div>
  <div class="navbar-end">
    <div class="mr-4">
      <ThemeSelector/>
    </div>
    <button
        class="btn btn-ghost btn-circle"
        on:click={() => openSettingsModal()}
    >
        <IconSettings class="h-5 w-5" />
    </button>
    <!-- <button class="btn btn-ghost btn-circle">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"> <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /> </svg>
    </button>
    <button class="btn btn-ghost btn-circle">
      <div class="indicator">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"> <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" /> </svg>
        <span class="badge badge-xs badge-primary indicator-item"></span>
      </div>
    </button> -->
  </div>
</div>

<div class="container max-w-8xl mx-auto p-8">
  <div class="grid grid-cols-1 mt-10 mb-15">
    <AboveTimeline />
  </div>

  <!-- tabs -->
  <div bind:this={tabsElement} class="tabs mb-6 flex justify-center">
    {#each tabs as tab}
      <button class="mr-4
                    { activeTab === tab.name ?
                      'badge badge-lg badge-primary tab-active text-white' :
                      'badge badge-lg badge-primary badge-outline'
                    }"
      on:click={() => setActiveTab(tab.name)}>
      {tab.label}
      </button>
    {/each}
  </div>

  <!-- tab content: only mount the active tab to avoid loading all API endpoints at once -->
  <div style="min-height: 1000px;">
    {#each tabs as tab}
      {#if activeTab === tab.name}
        <div class="fade-in">
          <svelte:component this={tab.component} />
        </div>
      {/if}
    {/each}
  </div>

  <Footer />
</div>

<Settings />

<style>

  .fade-in {
    animation: fadeIn 0.5s ease-in;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  /* Override the divider for DaisyUI list component, as its stopped working in recent versions */
  :global(.soft-divider > :not(:last-child).list-row)::after,
  :global(.soft-divider > :not(:last-child) .list-row)::after {
    opacity: 0.05 !important;
  }
</style>
