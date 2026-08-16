<script>
// @ts-nocheck
    import { onMount } from 'svelte';
    import { refreshRecordHolderData } from '../stores/settings';

    export let endpoint;
    export let title;
    export let columns = [];
    export let icon = null;

    let data = [];
    let loading = true;
    let error = null;
    let selectedAircraft = null;
    let selectedAircraftImage = null;
    let imageLoading = true;

    // Unique modal id per table (fastest/slowest/highest/lowest)
    const modalId = 'motion-modal-' + endpoint.replace(/[^a-zA-Z0-9]/g, '-');

    async function fetchData() {

        try {
            const response = await fetch(endpoint);
            if (!response.ok) {
                throw new Error(`${response.status}`);
            }
            const result = await response.json();
            data = result;
            error = null;
        } catch (err) {
            error = err.message;
        } finally {
            loading = false;
        }
    }

    async function getImage(aircraft) {
        if (!aircraft?.hex) {
            return null;
        }

        try {
            const response = await fetch(`https://api.planespotters.net/pub/photos/hex/${aircraft.hex}`);
            if (!response.ok) {
                return null;
            }
            const result = await response.json();
            const photo = result.photos?.[0];
            if (!photo) {
                return null;
            }
            return {
                url_photo: photo.thumbnail_large?.src,
                url_photo_photographer: photo.photographer,
                url_photo_link: photo.link
            };
        } catch (err) {
            console.error("Error fetching image:", err);
            return null;
        }
    }

    async function showAircraftModal(aircraft) {
        selectedAircraft = aircraft;
        selectedAircraftImage = null;
        imageLoading = true;
        // @ts-ignore
        document.getElementById(modalId).showModal();

        selectedAircraftImage = await getImage(aircraft);
        imageLoading = false;
    }

    function closeModal() {
        selectedAircraft = null;
        selectedAircraftImage = null;
    }

    onMount(() => {
        fetchData();
    })

    // Refresh when settings change
    $: if ($refreshRecordHolderData) {
        fetchData();
    }
</script>

<div>
    <div class="card bg-base-100 mb-4 w96 shadow-sm rounded hover:shadow-md transition-all duration-200">
        <div class="card-body">
            <div class="overflow-x-auto">
                {#if loading}
                    <div class="flex justify-center py-8">
                        <span class="loading loading-ring loading-lg"></span>
                    </div>
                {:else if error}
                    <div class="flex alert alert-error">
                        <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                        <span>Something went wrong: {error}</span>
                    </div>
                {:else if data.length === 0}
                    <div class="alert alert-info">
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
                        <span>No data available</span>
                    </div>
                {:else}
                    <!-- table header-->
                    <div class="flex items-center gap-2 mb-5">
                    {#if icon}
                        <div class="w-8 h-8 rounded-lg flex items-center justify-center">
                            <svelte:component this={icon} class="w-6 h-6 text-primary" />
                        </div>
                    {/if}
                    <h2 class="text-2xl font-extralight tracking-wider">{title}</h2>
                    </div>
                    <!-- table-->
                    <table class="table">
                        <thead>
                            <tr class="uppercase tracking-wider">
                                {#each columns as column}
                                    <th>{column.header}</th>
                                {/each}
                            </tr>
                        </thead>
                        <tbody>
                            {#each data as aircraft}
                            <tr class="hover:bg-base-300 cursor-pointer" on:click={() => showAircraftModal(aircraft)}>
                                {#each columns as column}
                                    <td class={column.class || ''}>
                                        {#if column.formatter}
                                            {@html column.formatter(aircraft[column.field])}
                                        {:else}
                                            {aircraft[column.field] || '-'}
                                        {/if}
                                    </td>
                                {/each}
                            </tr>
                            {/each}
                        </tbody>
                    </table>
                {/if}
            </div>
        </div>
    </div>
</div>

<!--modal-->
<dialog id={modalId} class="modal" on:close={closeModal}>
    <div class="modal-box max-w-2xl">
        {#if selectedAircraft}
            <div class="flex items-center justify-between mb-1">
                <h3 class="text-lg font-bold">{selectedAircraft.registration || 'Unknown'} - {selectedAircraft.type || ''}</h3>
                {#if selectedAircraft.hex}
                    <p class="text-sm uppercase tracking-wider font-mono">{selectedAircraft.hex}</p>
                {/if}
            </div>
            {#if selectedAircraft.flight}
                <p class="text-sm text-gray-600 mb-4">{selectedAircraft.flight}</p>
            {/if}

            <!-- photo -->
            {#if imageLoading}
                <div class="skeleton h-64 w-full rounded-lg mb-4"></div>
            {:else if selectedAircraftImage?.url_photo}
                <div class="relative mb-4">
                    <a href={selectedAircraftImage.url_photo_link} target="_blank" rel="noopener noreferrer">
                        <img
                            src={selectedAircraftImage.url_photo}
                            alt="{selectedAircraft.registration}"
                            class="w-full h-auto rounded-lg"
                        />
                    </a>
                    {#if selectedAircraftImage.url_photo_photographer}
                        <span class="absolute bottom-1 right-2 text-xs text-white opacity-80">© {selectedAircraftImage.url_photo_photographer}</span>
                    {/if}
                </div>
            {:else}
                <p class="text-center text-gray-500 py-8">No photo available for this aircraft</p>
            {/if}

            <!-- details -->
            <div class="grid grid-cols-2 gap-x-6 gap-y-1">
                {#each columns as column}
                    <div class="flex justify-between border-b border-base-200 py-1">
                        <span class="text-xs uppercase tracking-wider text-gray-500">{column.header}</span>
                        <span class="text-sm text-right">
                            {#if column.formatter}
                                {@html column.formatter(selectedAircraft[column.field])}
                            {:else}
                                {selectedAircraft[column.field] || '-'}
                            {/if}
                        </span>
                    </div>
                {/each}
            </div>
        {/if}
        <div class="modal-action">
            <form method="dialog">
                <button class="btn">Close</button>
            </form>
        </div>
    </div>
    <form method="dialog" class="modal-backdrop">
        <button>close</button>
    </form>
</dialog>
