<script>
    import { onMount } from 'svelte';
    import { settings } from '../stores/settings';
    import { IconBrandGithub } from '@tabler/icons-svelte';


    let activeMenuItem = 'display';
    let isSaving = false;

    let routeTableLimit;
    let interestingTableLimit;
    let recordHolderTableLimit;
    let disablePlaneAlertDbTags;
    let settingsChanged = false;
    let version = { version: '...', commit: '...', date: '...' };

    const menuItems = [
        { id: 'display', label: 'Display' },
        { id: 'about', label: 'About' }
    ];

    $: if (!settingsChanged) {
        if ($settings.route_table_limit) {
            routeTableLimit = parseInt($settings.route_table_limit.setting_value);
        }
        if ($settings.interesting_table_limit) {
            interestingTableLimit = parseInt($settings.interesting_table_limit.setting_value);
        }
        if ($settings.record_holder_table_limit) {
            recordHolderTableLimit = parseInt($settings.record_holder_table_limit.setting_value);
        }
        if ($settings.disable_planealertdb_tags) {
            disablePlaneAlertDbTags = $settings.disable_planealertdb_tags.setting_value === 'true';
        }
    }

    function handleSettingChange() {
        settingsChanged = true;
    }

    async function saveSettings() {
        const form = document.getElementById('display-settings-form');
        if (form && !form.checkValidity()) {
            form.reportValidity();
            return;
        }

        isSaving = true;
        const updates = {
            route_table_limit: routeTableLimit.toString(),
            interesting_table_limit: interestingTableLimit.toString(),
            record_holder_table_limit: recordHolderTableLimit.toString(),
            disable_planealertdb_tags: disablePlaneAlertDbTags.toString()
        };

        const success = await settings.save(updates);
        if (success) {
            settingsChanged = false;
            const modal = document.getElementById('settings-modal');
            if (modal) modal.close();
        }
        isSaving = false;
    }

    async function fetchVersion() {
        try {
            const response = await fetch('/api/version');
            if (response.ok) {
                version = await response.json();
            }
        } catch (error) {
            console.error('Failed to fetch version:', error);
        }
    }

    onMount(() => {
        settings.load();
        fetchVersion();
    });
</script>

<dialog id="settings-modal" class="modal">
    <div class="modal-box w-11/12 max-w-5xl h-[600px] p-0 relative">
        <form method="dialog" class="absolute right-2 top-2 z-10">
            <button class="btn btn-md btn-circle btn-ghost text-2xl">✕</button>
        </form>
        <div class="flex h-full">
            <!-- Settings Menu -->
            <div class="w-56 bg-base-200 p-4">
                <h3 class="text-xl font-bold mb-6 px-3">Settings</h3>
                <ul class="menu">
                    {#each menuItems as item}
                        <li>
                            <button
                                type="button"
                                class="{activeMenuItem === item.id ? 'active' : ''}"
                                on:click={() => activeMenuItem = item.id}
                            >
                                {item.label}
                            </button>
                        </li>
                    {/each}
                </ul>
            </div>

            <!-- Settings -->
            <div class="flex-1 p-6 flex flex-col">
                <div class="flex-1 {activeMenuItem === 'about' ? 'flex items-center justify-center' : ''}">
                    {#if activeMenuItem === 'display'}
                        <h4 class="text-lg font-semibold mb-6">Display Settings</h4>

                        <form id="display-settings-form" class="space-y-6">
                            
                            <!-- Route Table Display Settings -->
                            <div>
                                <p class="text-xl font-extralight tracking-wider mb-4">Route Information</p>
                                <p class="text-m text-base-content/70 mb-2">
                                    Number of rows to display in "Route Information" tables
                                </p>
                                <input
                                    id="route-table-limit"
                                    type="number"
                                    bind:value={routeTableLimit}
                                    on:input={handleSettingChange}
                                    min="1"
                                    max="100"
                                    step="1"
                                    required
                                    class="input w-20"
                                />
                                <span class="ml-2 text-sm text-base-content/70">(1-100)</span>
                            </div>

                            <!-- Interesting Table Display Settings -->

                            <!-- Interesting Table Display Settings: Table Limit -->
                            <div>
                                <p class="text-xl font-extralight tracking-wider mb-4">Interesting Aircraft</p>
                                <p class="text-m text-base-content/70 mb-2">
                                    Number of rows to display in "Interesting Aircraft" tables
                                </p>
                                <input
                                    id="interesting-table-limit"
                                    type="number"
                                    bind:value={interestingTableLimit}
                                    on:input={handleSettingChange}
                                    min="1"
                                    max="100"
                                    step="1"
                                    required
                                    class="input w-20"
                                />
                                <span class="ml-2 text-sm text-base-content/70">(1-100)</span>
                            </div>

                            <!-- Interesting Table Display Settings: Disable Tags -->
                            <div>
                                <!-- <p class="text-xl font-extralight tracking-wider mb-4">PlaneAlertDB Tags</p> -->
                                <label class="flex items-center gap-3">
                                    <input
                                        type="checkbox"
                                        bind:checked={disablePlaneAlertDbTags}
                                        on:change={handleSettingChange}
                                        class="checkbox"
                                    />
                                    <span class="text-m text-base-content/70">
                                        Disable <a href="https://github.com/sdr-enthusiasts/plane-alert-db?tab=readme-ov-file#description-of-categories" target="_blank" rel="noopener noreferrer" class="text-accent hover:text-primary transition-colors">plane-alert-db</a> tags in modal
                                    </span>
                                </label>
                            </div>


                            <!-- Record Holder Display Settings -->
                            <div>
                                <p class="text-xl font-extralight tracking-wider mb-4">Record Holders</p>
                                <p class="text-m text-base-content/70 mb-2">
                                    Number of rows to display in "Record Holders" tables
                                </p>
                                <input
                                    id="record-holder-table-limit"
                                    type="number"
                                    bind:value={recordHolderTableLimit}
                                    on:input={handleSettingChange}
                                    min="1"
                                    max="100"
                                    step="1"
                                    required
                                    class="input w-20"
                                />
                                <span class="ml-2 text-sm text-base-content/70">(1-100)</span>
                            </div>
                        </form>

                    {:else if activeMenuItem === 'about'}
                        <div class="text-center mx-auto">
                            <div class="flex items-center justify-center gap-6 mb-2">
                                <img src="/logo_icon.png" alt="Skystats Logo" class="w-32 h-32" />
                                <h1 class="text-7xl font-normal text-primary drop-shadow-[0_0_15px_rgba(59,130,246,0.5)]">
                                    Skystats
                                </h1>
                            </div>
                            <div class="mb-6 text-base-content/50">
                                {#if version.version === "dev"}
                                    {version.version} • {version.commit} • {version.date.toLocaleString()}
                                {:else}
                                    {version.version}
                                {/if}
                            </div>
                            <p class="mt-6 mb-6 text-sm text-base-content/50">
                                Created by <a href="https://github.com/tomcarman" target="_blank" rel="noopener noreferrer" class="text-accent hover:text-primary transition-colors">@tomcarman</a> with support from the SDR Enthusiasts community. Join us on <a href="https://discord.gg/znkBr2eyev" target="_blank" rel="noopener noreferrer" class="text-accent hover:text-primary transition-colors">Discord</a>.
                            </p>
                            <a href="https://github.com/tomcarman/skystats" target="_blank" rel="noopener noreferrer" class="inline-flex items-center gap-2 opacity-50 hover:opacity-100 transition-opacity">
                                <IconBrandGithub stroke={2} size={32} />
                                <span>GitHub</span>
                            </a>
                        </div>
                    {/if}
                </div>

                {#if activeMenuItem !== 'about'}
                    <div class="modal-action justify-end">
                        <button
                            class="btn btn-primary"
                            on:click={saveSettings}
                            disabled={!settingsChanged || isSaving}
                        >
                            {isSaving ? 'Saving...' : 'Save'}
                        </button>
                    </div>
                {/if}
            </div>
        </div>
    </div>
</dialog>
