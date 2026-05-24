import { writable } from 'svelte/store';

const DEFAULT_METRICS_POLL_MS = 2000;

export const metricsPollMs = writable(DEFAULT_METRICS_POLL_MS);

export async function loadPollConfig() {
    try {
        const response = await fetch('/api/config');
        if (response.ok) {
            const data = await response.json();
            if (typeof data.metrics_poll_ms === 'number' && data.metrics_poll_ms > 0) {
                metricsPollMs.set(data.metrics_poll_ms);
            }
        }
    } catch (error) {
        console.error('Failed to load poll config:', error);
    }
}

loadPollConfig();
