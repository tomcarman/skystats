import { writable, get } from 'svelte/store';
import { metricsPollMs } from './pollConfig';

const TOTALS_POLL_MS = 60000;

const state = writable({
    data: {},
    loading: true,
    error: null,
});

let recentTimer = null;
let totalsTimer = null;
let subscriberCount = 0;
let pollMs = 2000;

export const seenMetrics = {
    subscribe: state.subscribe,

    start() {
        subscriberCount += 1;
        if (subscriberCount > 1) {
            return;
        }

        pollMs = get(metricsPollMs);
        const unsubscribePoll = metricsPollMs.subscribe((ms) => {
            pollMs = ms;
            if (subscriberCount > 0) {
                restartTimers();
            }
        });

        seenMetrics._unsubscribePoll = unsubscribePoll;
        restartTimers();

        if (typeof document !== 'undefined') {
            document.addEventListener('visibilitychange', onVisibilityChange);
        }
    },

    stop() {
        subscriberCount = Math.max(0, subscriberCount - 1);
        if (subscriberCount > 0) {
            return;
        }

        clearTimers();
        if (seenMetrics._unsubscribePoll) {
            seenMetrics._unsubscribePoll();
            seenMetrics._unsubscribePoll = null;
        }
        if (typeof document !== 'undefined') {
            document.removeEventListener('visibilitychange', onVisibilityChange);
        }
    },
};

function onVisibilityChange() {
    if (document.hidden) {
        clearTimers();
    } else if (subscriberCount > 0) {
        fetchRecent();
        restartTimers();
    }
}

function clearTimers() {
    if (recentTimer) {
        clearInterval(recentTimer);
        recentTimer = null;
    }
    if (totalsTimer) {
        clearInterval(totalsTimer);
        totalsTimer = null;
    }
}

function restartTimers() {
    clearTimers();
    if (subscriberCount === 0 || (typeof document !== 'undefined' && document.hidden)) {
        return;
    }

    fetchRecent();
    recentTimer = setInterval(fetchRecent, pollMs);

    fetchTotals();
    totalsTimer = setInterval(fetchTotals, TOTALS_POLL_MS);
}

async function fetchRecent() {
    try {
        const tz = Intl.DateTimeFormat().resolvedOptions().timeZone;
        const response = await fetch(`api/stats/seen/recent?tz=${encodeURIComponent(tz)}`);
        if (!response.ok) {
            throw new Error(`${response.status}`);
        }
        const result = await response.json();
        state.update((s) => ({
            ...s,
            data: { ...s.data, ...result },
            error: null,
            loading: false,
        }));
    } catch (err) {
        state.update((s) => ({
            ...s,
            error: err.message,
            loading: false,
        }));
    }
}

async function fetchTotals() {
    try {
        const response = await fetch('api/stats/seen/totals');
        if (!response.ok) {
            throw new Error(`${response.status}`);
        }
        const result = await response.json();
        state.update((s) => ({
            ...s,
            data: { ...s.data, ...result },
            error: null,
        }));
    } catch (err) {
        state.update((s) => ({
            ...s,
            error: err.message,
        }));
    }
}
