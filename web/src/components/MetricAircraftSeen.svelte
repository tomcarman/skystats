<script>
    import { onMount, onDestroy } from 'svelte';
    import { IconPlaneTilt, IconCalendar, IconClock } from '@tabler/icons-svelte';
    import NumberFlow from '@number-flow/svelte';
    import SkeletonMetrics from './SkeletonMetrics.svelte';
    import { seenMetrics } from '../stores/seenMetrics';

    $: data = $seenMetrics.data;
    $: loading = $seenMetrics.loading;
    $: error = $seenMetrics.error;

    onMount(() => {
        seenMetrics.start();
    });

    onDestroy(() => {
        seenMetrics.stop();
    });
</script>
{#if loading}
    <SkeletonMetrics />
{:else if error}
    <div class="alert alert-error">
        <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
        <span>Something went wrong: {error}</span>
    </div>
{:else}
    <div class="p-2 stats stats-vertical lg:stats-horizontal bg-base-100 shadow-sm rounded-xl hover:shadow-md transition-all duration-200 w-full">
        <div class="stat">
            <div class="stat-figure">
                <div class="icon icon-tabler icons-tabler-outline">
                    <IconClock stroke={2}/>
                </div>
            </div>
            <div class="stat-title">Aircraft seen</div>
            <div class="stat-value"><NumberFlow willChange={true} respectMotionPreference={false} value={data.hour_aircraft} /></div>
            <div class="stat-desc">past hour</div>
        </div>
        <div class="stat">
            <div class="stat-figure">
                <div class="icon icon-tabler icons-tabler-outline">
                    <IconCalendar stroke={2}/>
                </div>
            </div>
            <div class="stat-title">Aircraft seen</div>
            <div class="stat-value"><NumberFlow willChange={true} respectMotionPreference={false} value={data.today_aircraft} /></div>
            <div class="stat-desc">today</div>
        </div>
        <div class="stat">
            <div class="stat-figure">
                <div class="icon icon-tabler icons-tabler-outline">
                    <IconPlaneTilt stroke={2}/>
                </div>
            </div>
            <div class="stat-title">Aircraft seen</div>
            <div class="stat-value"><NumberFlow willChange={true} respectMotionPreference={false} value={data.total_aircraft} /></div>
            <div class="stat-desc">all time</div>
        </div>
    </div>
{/if}
