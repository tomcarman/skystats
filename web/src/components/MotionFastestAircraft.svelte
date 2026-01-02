<script>
    import MotionStats from './MotionStats.svelte';
    import { IconRocket } from '@tabler/icons-svelte';
    import { speedUnit, formatSpeed } from '../stores/preferences';

    $: columns = [
        { header: 'Reg', field: 'registration', class: 'font-mono whitespace-nowrap' },
        { header: 'Type', field: 'type' },
        // { header: 'Flight', field: 'flight' },
        { 
            header: `Speed`, 
            field: 'ground_speed',
            formatter: (value) => formatSpeed(value, $speedUnit)
        },
        { 
            header: 'First Seen', 
            field: 'first_seen',
            class: 'whitespace-nowrap',
            formatter: (value) => value ? new Date(value).toLocaleString() : '-'
        }
    ];
</script>

<MotionStats 
    endpoint="api/stats/motion/fastest"
    title="Fastest Aircraft"
    {columns}
    icon={IconRocket}
/>
