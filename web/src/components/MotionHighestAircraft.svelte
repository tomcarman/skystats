<script>
    import MotionStats from './MotionStats.svelte';
    import { IconArrowUpDashed } from '@tabler/icons-svelte';
    import { altitudeUnit, formatAltitude } from '../stores/preferences';

    $: columns = [
        { header: 'Reg', field: 'registration', class: 'font-mono' },
        { header: 'Model', field: 'type' },
        // { header: 'Flight', field: 'flight' },
        { 
            header: `Altitude`, 
            field: 'barometric_altitude',
            formatter: (value) => formatAltitude(value, $altitudeUnit)
        },
        { 
            header: 'First Seen', 
            field: 'first_seen',
            formatter: (value) => value ? new Date(value).toLocaleString() : '-'
        }
    ];
</script>

<MotionStats 
    endpoint="api/stats/motion/highest"
    title="Highest Aircraft"
    {columns}
    icon={IconArrowUpDashed}
/>
