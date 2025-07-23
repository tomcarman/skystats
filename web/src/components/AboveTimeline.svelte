<script>
    import { onMount, onDestroy } from "svelte";
    import NumberFlow from "@number-flow/svelte";
    import { IconPlane } from "@tabler/icons-svelte";

    let endpoint = "api/stats/above";

    let refreshRate = 2000;
    let data = [];
    let loading = true;
    let error = null;
    let interval = null;
    let selectedAircraft = null;
    let imageLoading = false;

    async function fetchData() {
        try {
            const response = await fetch(endpoint);
            if (!response.ok) {
                throw new Error(`{response.status}`);
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

    onMount(() => {
        fetchData();
        interval = setInterval(fetchData, refreshRate);
    });

    onDestroy(() => {
        if (interval) {
            clearInterval(interval);
        }
    });

    function getSlottedAircraft(aircraftList) {
        const slots = [null, null, null, null, null];

        // Sort by distance ascending
        const sortedAircraft = [...aircraftList].sort(
            (a, b) =>
                parseFloat(a.last_seen_distance) -
                parseFloat(b.last_seen_distance),
        );

        sortedAircraft.forEach((aircraft) => {
            const distance = parseFloat(aircraft.last_seen_distance);
            let idealSlot;

            // if (distance < 4) idealSlot = 0;
            // else if (distance < 8) idealSlot = 1;
            // else if (distance < 12) idealSlot = 2;
            // else if (distance < 16) idealSlot = 3;
            // else if (distance < 20) idealSlot = 4;
            // else return;

            if (distance < 20) idealSlot = 0;
            else if (distance < 40) idealSlot = 1;
            else if (distance < 60) idealSlot = 2;
            else if (distance < 80) idealSlot = 3;
            else if (distance < 100) idealSlot = 4;
            else return;

            let placed = false;

            // try idea slot
            if (slots[idealSlot] === null) {
                slots[idealSlot] = aircraft;
                placed = true;
            } else {
                // if no ideal, check later slots
                for (let i = idealSlot + 1; i < 5 && !placed; i++) {
                    if (slots[i] === null) {
                        slots[i] = aircraft;
                        placed = true;
                    }
                }

                // if no ideal or later, check earlier slots
                for (let i = idealSlot - 1; i >= 0 && !placed; i--) {
                    if (slots[i] === null) {
                        slots[i] = aircraft;
                        placed = true;
                    }
                }
            }
        });

        // final reorder to ensure always in logical order
        for (let i = 0; i < 4; i++) {
            for (let j = i + 1; j < 5; j++) {
                if (slots[i] && slots[j]) {
                    const dist1 = parseFloat(slots[i].last_seen_distance);
                    const dist2 = parseFloat(slots[j].last_seen_distance);
                    if (dist1 > dist2) {
                        // swap
                        const temp = slots[i];
                        slots[i] = slots[j];
                        slots[j] = temp;
                    }
                }
            }
        }

        return slots;
    }

    $: slottedData = getSlottedAircraft(data);

    function showAircraftModal(aircraft) {
        selectedAircraft = aircraft;
        imageLoading = true;
        // @ts-ignore
        document.getElementById("aircraft-modal").showModal()
    }

    function closeModal() {
        selectedAircraft = null;
    }
</script>

<!-- <div class="card bg-base-100 mb-4 w96 shadow-sm rounded-xl hover:shadow-md transition-all duration-200"> -->
<!-- <div class="card-body"> -->
<div class="w-full">
    {#if loading}
        <div class="flex justify-center py-8">
            <span class="loading loading-spinner loading-lg"></span>
        </div>
    {:else if error}
        <div class="flex alert alert-error">
            <svg
                xmlns="http://www.w3.org/2000/svg"
                class="stroke-current shrink-0 h-6 w-6"
                fill="none"
                viewBox="0 0 24 24"
                ><path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
                /></svg
            >
            <span>Something went wrong: {error}</span>
        </div>
    {:else}
        <ul class="timeline timeline-horizontal w-full">
            <!--Home-->
            <li>
                <!-- <hr /> -->
                <div class="timeline-start"></div>
                <div class="timeline-middle">
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="24"
                        height="24"
                        viewBox="0 0 24 24"
                        fill="currentColor"
                        class="icon icon-tabler icons-tabler-filled icon-tabler-home"
                    >
                        <path stroke="none" d="M0 0h24v24H0z" fill="none" />
                        <path
                            d="M12.707 2.293l9 9c.63 .63 .184 1.707 -.707 1.707h-1v6a3 3 0 0 1 -3 3h-1v-7a3 3 0 0 0 -2.824 -2.995l-.176 -.005h-2a3 3 0 0 0 -3 3v7h-1a3 3 0 0 1 -3 -3v-6h-1c-.89 0 -1.337 -1.077 -.707 -1.707l9 -9a1 1 0 0 1 1.414 0m.293 11.707a1 1 0 0 1 1 1v7h-4v-7a1 1 0 0 1 .883 -.993l.117 -.007z"
                        />
                    </svg>
                </div>
                <div class="timeline-end"></div>
                <hr />
            </li>
            <!-- End Home-->
            {#each slottedData as aircraft, index (index)}
                {#if aircraft}
                    <li>
                        <hr />
                        <div class="timeline-start mb-5">
                            <button type="button"
                                    class="cursor-pointer
                                        badge badge-accent 
                                        uppercase font-bold tracking-wider text-white text-[8px] sm:text-xs"
                                    on:click={() => showAircraftModal(aircraft)}>
                                {aircraft.registration || aircraft.hex}
                            </button>
                        </div>
                        <div class="timeline-middle">
                            <IconPlane
                                size={24}
                                style="transform: rotate({aircraft.track -
                                    90}deg)"
                            />
                        </div>
                        <div class="timeline-end text-xs sm:text-sm">
                            <NumberFlow
                                value={Number.parseFloat(
                                    aircraft.last_seen_distance,
                                ).toFixed(0)}
                                suffix=" km"
                                willChange={true}
                                respectMotionPreference={false}
                            />
                        </div>
                        <hr />
                    </li>
                {:else}
                    <li>
                        <hr />
                        <div class="timeline-start mb-5">
                            <div class="invisible text-xs sm:text-xs">
                                PLACEHOLDER
                            </div>
                        </div>
                        <div class="timeline-middle opacity-20">
                            <IconPlane size={24} />
                        </div>
                        <div class="timeline-end invisible">0 km</div>
                        <hr />
                    </li>
                {/if}
            {/each}
            <!--Far-->
            <li>
                <hr />
                <div class="timeline-start"></div>
                <div class="timeline-middle">
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="24"
                        height="24"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        class="icon icon-tabler icons-tabler-outline icon-tabler-world"
                    >
                        <path stroke="none" d="M0 0h24v24H0z" fill="none" />
                        <path d="M3 12a9 9 0 1 0 18 0a9 9 0 0 0 -18 0" />
                        <path d="M3.6 9h16.8" />
                        <path d="M3.6 15h16.8" />
                        <path d="M11.5 3a17 17 0 0 0 0 18" />
                        <path d="M12.5 3a17 17 0 0 1 0 18" />
                    </svg>
                </div>
                <div class="timeline-end text-xs sm:text-sm">20km+</div>
                <!-- <hr /> -->
            </li>
            <!-- End Far-->
        </ul>
    {/if}
</div>

<dialog id="aircraft-modal" class="modal" on:close={closeModal}>
<div class="modal-box max-w-5xl">
        {#if selectedAircraft}
            <!--header-->
            <div class="flex items-start gap-3">
                    {#if selectedAircraft.airline_icao}
                    <div class="bg-base-200 p-2 rounded-lg">
                        <img src="https://doj0yisjozhv1.cloudfront.net/square-logos/{selectedAircraft.airline_icao}.png" width="40" height="40" alt="{selectedAircraft.airline_icao}">
                    </div>
                {/if}
                <div>
                    <h3 class="text-lg font-bold">{selectedAircraft.registration || 'Unknown'}</h3>
                    <p class="text-sm uppercase tracking-wider font-mono">{selectedAircraft.hex || ''}</p>
                </div>
            </div>
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 mt-6 gap-6">
                
                <!--image-->
                <!-- <div class="md:mr-4 lg:mr-8"> -->
                <div class="mt-6 mr-8">
                    {#if selectedAircraft.url_photo}
                        {#if imageLoading}
                            <div class="skeleton w-full max-w-sm rounded-lg aspect-[3/2]"></div>
                        {/if}
                        <img 
                            src={selectedAircraft.url_photo} 
                            alt="{selectedAircraft.registration}" 
                            class="w-full max-w-sm h-auto rounded-lg {imageLoading ? 'hidden' : ''}"
                            on:load={() => imageLoading = false}
                            on:error={() => imageLoading = false}
                        />
                    {/if}
                    {#if !selectedAircraft.url_photo}
                        <div class="bg-base-200 w-full max-w-sm aspect-[3/2] flex items-center justify-center rounded-lg">
                            <p class="text-center text-sm text-gray-500">No photo available</p>
                        </div>
                    {/if}
                    {#if selectedAircraft.icao_type}
                        <p class="text-sm mt-4 font-bold">{selectedAircraft.icao_type}</p>
                        <p class="text-sm text-gray-500 italic">{selectedAircraft.reg_type}</p>
                    {/if}  
                </div>
                <!-- reg info -->
                <div>
                    <p class="font-semibold uppercase tracking-wider">Registration</p>
                    {#if selectedAircraft.reg_type}
                        <p>icao_type: {selectedAircraft.icao_type}</p>
                        <p>manufacturer: {selectedAircraft.manufacturer}</p>
                        <p>registered_owner_country_name: {selectedAircraft.registered_owner_country_name}</p>
                        <p>registered_owner_country_iso: {selectedAircraft.registered_owner_country_iso}</p>
                        <p>registered_owner_operator_flag: {selectedAircraft.registered_owner_operator_flag}</p>
                        <p>registered_owner: {selectedAircraft.registered_owner}</p>
                    {:else}
                        <div class="bg-base-200 w-full max-w-sm flex items-center justify-center rounded-lg">
                            <p class="text-center text-sm text-gray-500">No registration data available</p>
                        </div>
                    {/if}
                </div>
                <!-- route info -->
                <div>
                    <p class="font-semibold uppercase tracking-wider">Route</p>
                    {#if selectedAircraft.origin_country_iso_name}
                        <p>airline_name: {selectedAircraft.airline_name}</p>
                        <p>airline_icao: {selectedAircraft.airline_icao}</p>
                        <p>origin_country_name: {selectedAircraft.origin_country_name}</p>
                        <p>origin_country_iso_name: {selectedAircraft.origin_country_iso_name}</p>
                        <p>origin_iata_code: {selectedAircraft.origin_iata_code}</p>
                        <p>origin_icao_code: {selectedAircraft.origin_icao_code}</p>
                        <p>origin_name: {selectedAircraft.origin_name}</p>
                        <p>destination_country_name: {selectedAircraft.destination_country_name}</p>
                        <p>destination_country_iso_name: {selectedAircraft.destination_country_iso_name}</p>
                        <p>destination_iata_code: {selectedAircraft.destination_iata_code}</p>
                        <p>destination_icao_code: {selectedAircraft.destination_icao_code}</p>
                        <p>destination_name: {selectedAircraft.destination_name}    </p>
                    
                    {:else}
                        <div class="bg-base-200 w-full max-w-sm flex items-center justify-center rounded-lg">
                            <p class="text-center text-sm text-gray-500">No route data available</p>
                        </div>
                    {/if}
                </div>
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

<style>
    .timeline-horizontal {
        justify-content: space-between;
    }

    .timeline-horizontal li {
        flex: 1;
    }

    .timeline-horizontal hr {
        width: 100%;
    }
</style>
