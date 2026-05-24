<script>
    import { onMount, onDestroy } from "svelte";
    import NumberFlow from "@number-flow/svelte";
    import { IconPlane, IconPlaneDeparture, IconPlaneArrival } from "@tabler/icons-svelte";
    import Status from "./Status.svelte";
    import { metricsPollMs } from "../stores/pollConfig";

    let endpoint = "api/stats/above";
    let data = [];
    let loading = true;
    let error = null;
    let interval = null;
    let selectedAircraft = null;
    let selectedAircraftHex = null;
    let selectedAircraftImage = null;

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

    function startPolling(ms) {
        if (interval) {
            clearInterval(interval);
        }
        fetchData();
        interval = setInterval(fetchData, ms);
    }

    function stopPolling() {
        if (interval) {
            clearInterval(interval);
            interval = null;
        }
    }

    function onVisibilityChange() {
        if (document.hidden) {
            stopPolling();
        } else {
            startPolling($metricsPollMs);
        }
    }

    onMount(() => {
        startPolling($metricsPollMs);
        const unsubscribe = metricsPollMs.subscribe((ms) => {
            if (!document.hidden) {
                startPolling(ms);
            }
        });
        document.addEventListener("visibilitychange", onVisibilityChange);

        return () => {
            unsubscribe();
            document.removeEventListener("visibilitychange", onVisibilityChange);
            stopPolling();
        };
    });

    onDestroy(() => {
        stopPolling();
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

            if (distance < 4) idealSlot = 0;
            else if (distance < 8) idealSlot = 1;
            else if (distance < 12) idealSlot = 2;
            else if (distance < 16) idealSlot = 3;
            else if (distance < 20) idealSlot = 4;
            else return;

            // if (distance < 20) idealSlot = 0;
            // else if (distance < 80) idealSlot = 1;
            // else if (distance < 120) idealSlot = 2;
            // else if (distance < 160) idealSlot = 3;
            // else if (distance < 200) idealSlot = 4;
            // else return;

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

    function calculateProgress(aircraft) {
        if (!aircraft || !aircraft.route_distance || !aircraft.destination_distance) {
            return aircraft;
        }
        
        const totalDistance = parseFloat(aircraft.route_distance);
        const remainingDistance = parseFloat(aircraft.destination_distance);
        const traveledDistance = totalDistance - remainingDistance;
        const progressPercent = Math.max(0, Math.min(100, (traveledDistance / totalDistance) * 100));
        
        return {
            ...aircraft,
            totalDistance: totalDistance.toFixed(1),
            remainingDistance: remainingDistance.toFixed(1),
            traveledDistance: traveledDistance.toFixed(1),
            progressPercent: progressPercent.toFixed(1)
        };
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

            return {
                url_photo: photo?.thumbnail_large.src,
                url_photo_photographer: photo?.photographer,
                url_photo_link: photo?.link
            }

        } catch (err) {
            console.error("Error fetching image:", err);
            return null;
        }
    }

    $: if (selectedAircraftHex && data.length > 0) {
        const updatedAircraft = data.find(a => a.hex === selectedAircraftHex);
        if (updatedAircraft) {
            selectedAircraft = calculateProgress(updatedAircraft);
        }   
    }

    async function showAircraftModal(aircraft) {
        selectedAircraftHex = aircraft.hex;
        selectedAircraft = calculateProgress(aircraft);

        const imageData = await getImage(aircraft);
        selectedAircraftImage = imageData;

        // @ts-ignore
        document.getElementById("aircraft-modal").showModal();
    }



    function closeModal() {
        selectedAircraft = null;
        selectedAircraftHex = null;
        selectedAircraftImage = null;
    }
</script>

<div class="w-full">
    {#if loading}
        <div class="flex justify-center py-8">
            <span class="loading loading-ring loading-lg"></span>
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
                <div class="timeline-end">
                    <Status />
                </div>
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
                                        badge badge-accent rounded-sm
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
<div class="modal-box max-w-5xl relative">
        <form method="dialog" class="absolute right-2 top-2">
            <button class="btn btn-md btn-circle btn-ghost text-2xl">✕</button>
        </form>
        {#if selectedAircraft}
            <div id="header">
                <div class="flex items-start gap-3">
                    {#if selectedAircraft.airline_icao}
                        <div class="bg-base-200 p-2 rounded-lg">
                            <img src="https://doj0yisjozhv1.cloudfront.net/square-logos/{selectedAircraft.airline_icao}.png" width="40" height="40" alt="{selectedAircraft.airline_icao}">
                        </div>
                    {/if}
                    <div>
                        <h3 class="text-lg font-bold">{selectedAircraft.registration || 'Unknown'} - {selectedAircraft.flight}</h3>
                        <p class="text-sm uppercase tracking-wider font-mono">{selectedAircraft.hex || ''}</p>
                    </div>
                </div>
            </div>
            <div id=progress>
                <div class="flex items-center gap-4 mt-6">
                    <div class="text-xl text-info font-thin text-accent font-mono">
                        {selectedAircraft.origin_iata_code}
                    </div>
                    <div class="progress-container flex-1">
                        <hr class="progress-hr">
                        <div class="progress-marker text-secondary" style="left: {selectedAircraft.progressPercent}%">
                        <IconPlane size={24} stroke="2" class="fill-base-100"/>
                        </div>
                    </div>
                    <div class="text-xl text-info font-thin text-accent font-mono">
                        {selectedAircraft.destination_iata_code}
                    </div>
                </div>
            </div>
            <div class="text-right text-xs italic">{selectedAircraft.remainingDistance}km remaining<br>
                (estimated)</div>
            <div class="grid grid-cols-1 md:grid-cols-2 mt-6 gap-6">
                <!--image-->
                <div class="mr-8">
                    {#if selectedAircraftImage?.url_photo}
                        <div class="relative w-full max-w-sm">
                            <a href={selectedAircraftImage.url_photo_link} target="_blank" rel="noopener noreferrer">
                                <img
                                    src={selectedAircraftImage.url_photo}
                                    alt="{selectedAircraft.registration}"
                                    class="w-full h-auto rounded-lg cursor-pointer hover:opacity-90 transition-opacity"
                                />
                            </a>
                            {#if selectedAircraftImage.url_photo_photographer}
                                <div class="absolute bottom-1 right-1 bg-black/50 text-white text-[10px] px-1.5 py-0.5 rounded">
                                    © {selectedAircraftImage.url_photo_photographer}
                                </div>
                            {/if}
                        </div>
                    {:else}
                        <div class="bg-base-200 w-full max-w-sm aspect-[3/2] flex items-center justify-center rounded-lg">
                            <p class="text-center text-sm text-gray-500">No photo available</p>
                        </div>
                    {/if}

                    {#if selectedAircraft.manufacturer && selectedAircraft.icao_type}
                        <p class="text-sm mt-2 uppercase font-semibold">{selectedAircraft.manufacturer} {selectedAircraft.icao_type}</p>
                    {:else if selectedAircraft.icao_type}
                        <p class="text-sm mt-2 uppercase font-semibold">{selectedAircraft.icao_type}</p>
                    {/if}
                    {#if selectedAircraft.reg_type}
                        <p class="text-sm text-gray-500">{selectedAircraft.reg_type}</p>
                    {/if}
                </div>
                {#if selectedAircraft.origin_country_iso_name}
                    <div>
                        <p class="font-bold uppercase tracking-wider">Route</p>
                            <div class="flex items-center gap-3 mt-3">
                                <IconPlaneDeparture size={24} />
                                <p><span class="text-sm fi fi-{selectedAircraft.origin_country_iso_name.toLowerCase()}"></span> {selectedAircraft.origin_name}, {selectedAircraft.origin_country_name}</p>
                            </div>
                            <div class="flex items-center mt-4 gap-3">
                                <IconPlaneArrival size={24} />
                                <p><span class="text-sm fi fi-{selectedAircraft.destination_country_iso_name.toLowerCase()}"></span> {selectedAircraft.destination_name}, {selectedAircraft.destination_country_name}</p>
                            </div>
                            <br/>
                            <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mt-3">
                                <div>
                                <p class="font-bold uppercase tracking-wider">Progress</p>
                                    <p class="mt-2"><span class="text-sm font-semibold uppercase">Origin:</span> {selectedAircraft.traveledDistance}km</p>
                                    <p><span class="text-sm font-semibold uppercase">Destination:</span> {selectedAircraft.remainingDistance}km</p>
                                </div>
                                <div class="radial-progress bg-base-300" style="--value:{selectedAircraft.progressPercent}" aria-valuenow={selectedAircraft.progressPercent} role="progressbar">{selectedAircraft.progressPercent}%</div>
                            </div>
                    </div>
                {:else}
                    <div class="bg-base-200 mb-10 w-full max-w-sm flex items-center justify-center rounded-lg">
                        <p class="text-center text-sm text-gray-500">No route data available</p>
                    </div>
                {/if}
            </div>
        {/if}
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

    .progress-container {
        position: relative;
        width: 100%;
        margin: 0;
    }
    
    .progress-hr {
        margin: 0;
        border: none;
        height: 4px;
        background: #ddd;
        border-radius: 2px;
    }
    
    .progress-marker {
        position: absolute;
        top: 50%;
        transform: translate(-50%, -50%);
        transition: left 0.3s ease-out;
        z-index: 10;
    }
</style>
