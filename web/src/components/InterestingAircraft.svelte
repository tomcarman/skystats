<script>
    import { onMount, onDestroy } from 'svelte'
    import { settings, refreshInterestingData } from '../stores/settings';

    export let endpoint;
    export let title;
    export let icon;
    export let aircraftType;

    let refreshRate = 10000
    let data = [];
    let loading = true;
    let error = null;
    let interval = null;
    let selectedAircraft = null;
    let imageLoadingStates = {
        image1: true,
        image2: true,
        image3: true
    };

    async function fetchData() {
        
        try {
            const response = await fetch(endpoint);
            if(!response.ok) {
                throw new Error(`{response.status}`);
            }
            const result = await response.json();
            data = result;
            error = null
        } catch (err) {
            error = err.message;
        } finally {
            loading = false;
        }
    }

    function showAircraftModal(aircraft) {
        selectedAircraft = aircraft;
        imageLoadingStates = {
            image1: true,
            image2: true,
            image3: true
        };
        // @ts-ignore
        document.getElementById(aircraftType).showModal();
    }

    function closeModal() {
        selectedAircraft = null;
    }

    onMount(() => {
        fetchData();
        interval = setInterval(fetchData, refreshRate)
    })

    onDestroy(() => {
        if (interval) {
            clearInterval(interval)
        }
    });

    // Refresh when settings change
    $: if ($refreshInterestingData) {
        fetchData();
    }

    $: disableTags = $settings['disable_planealertdb_tags']?.setting_value === 'true';

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
                        <thead class="uppercase tracking-wider">
                            <tr>
                                <th>Reg</th>
                                <th>Operator</th>
                                <th>Type</th>
                                <th>Last Seen</th>
                            </tr>
                        </thead>
                        <tbody>
                            {#each data as aircraft}
                            <tr class="hover:bg-base-300 cursor-pointer" on:click={() => showAircraftModal(aircraft)}>
                                <td class="font-mono whitespace-nowrap">{aircraft.registration}</td>
                                <td>{aircraft.operator}</td>
                                <td>{aircraft.type}</td>
                                <td class="whitespace-nowrap">{aircraft.seen ? new Date(aircraft.seen).toLocaleString() : '-'}</td>
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
<dialog id={aircraftType} class="modal" on:close={closeModal}>
    <div class="modal-box max-w-4xl">
        {#if selectedAircraft}
            <div class="flex items-center justify-between mb-2">
                <h3 class="text-lg font-bold">{selectedAircraft.registration} - {selectedAircraft.type}</h3>
                {#if disableTags === false}
                    <div class="flex gap-2">
                        {#if selectedAircraft.tag1}
                            <div class="badge badge-accent text-white">{selectedAircraft.tag1}</div>
                        {/if}
                        {#if selectedAircraft.tag2}
                            <div class="badge badge-accent text-white">{selectedAircraft.tag2}</div>
                        {/if}
                        {#if selectedAircraft.tag3}
                            <div class="badge badge-accent text-white">{selectedAircraft.tag3}</div>
                        {/if}
                    </div>
                {/if}
            </div>
            <p class="text-sm text-gray-600 mb-4">{selectedAircraft.operator} {#if selectedAircraft.flight} - {selectedAircraft.flight} {/if}</p>
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {#if selectedAircraft.image_link_1}
                    <div class="relative">
                        {#if imageLoadingStates.image1}
                            <div class="skeleton h-48 w-full rounded-lg"></div>
                        {/if}
                        <img 
                            src={selectedAircraft.image_link_1} 
                            alt="{selectedAircraft.registration} photo 1" 
                            class="w-full h-auto rounded-lg {imageLoadingStates.image1 ? 'absolute inset-0 opacity-0' : ''}"
                            on:load={() => imageLoadingStates.image1 = false}
                            on:error={() => imageLoadingStates.image1 = false}
                        />
                    </div>
                {/if}
                {#if selectedAircraft.image_link_2}
                    <div class="relative">
                        {#if imageLoadingStates.image2}
                            <div class="skeleton h-48 w-full rounded-lg"></div>
                        {/if}
                        <img 
                            src={selectedAircraft.image_link_2} 
                            alt="{selectedAircraft.registration} photo 2" 
                            class="w-full h-auto rounded-lg {imageLoadingStates.image2 ? 'absolute inset-0 opacity-0' : ''}"
                            on:load={() => imageLoadingStates.image2 = false}
                            on:error={() => imageLoadingStates.image2 = false}
                        />
                    </div>
                {/if}
                {#if selectedAircraft.image_link_3}
                    <div class="relative">
                        {#if imageLoadingStates.image3}
                            <div class="skeleton h-48 w-full rounded-lg"></div>
                        {/if}
                        <img 
                            src={selectedAircraft.image_link_3} 
                            alt="{selectedAircraft.registration} photo 3" 
                            class="w-full h-auto rounded-lg {imageLoadingStates.image3 ? 'absolute inset-0 opacity-0' : ''}"
                            on:load={() => imageLoadingStates.image3 = false}
                            on:error={() => imageLoadingStates.image3 = false}
                        />
                    </div>
                {/if}
            </div>
            {#if !selectedAircraft.image_link_1 && !selectedAircraft.image_link_2 && !selectedAircraft.image_link_3}
                <p class="text-center text-gray-500 py-8">No photos available for this aircraft</p>
            {/if}
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
