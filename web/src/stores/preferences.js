import { writable } from 'svelte/store';

/**
 * Indicates if the code is running in a browser environment.
 * Used to avoid errors when accessing localStorage on the server (SSR).
 * @type {boolean}
 */
const browser = typeof window !== 'undefined';

/**
 * Creates a Svelte store that persists its value to localStorage.
 * 
 * If running in the browser, it attempts to read the saved value.
 * It also subscribes to store changes to automatically update localStorage.
 * 
 * @param {string} key - The unique key to save the value in localStorage.
 * @param {*} initialValue - The default initial value if nothing is saved.
 * @returns {import('svelte/store').Writable<*>} A writable Svelte store.
 */
const createPersistedStore = (key, initialValue) => {
    // Check if we're in the browser and have a saved value
    const savedValue = browser ? localStorage.getItem(key) : null;
    const store = writable(savedValue || initialValue);

    if (browser) {
        store.subscribe((value) => {
            localStorage.setItem(key, value);
        });
    }

    return store;
};

/**
 * Definition of available speed units.
 * Each unit contains its display label, internal value, and conversion factor from Knots.
 * 
 * @constant
 * @type {Object.<string, {label: string, value: string, factor: number}>}
 */
export const SPEED_UNITS = {
    /** Knots (Base unit for aviation) */
    KTS: { label: 'Knots', value: 'kts', factor: 1 },
    /** Miles per hour */
    MPH: { label: 'MPH', value: 'mph', factor: 1.15078 },
    /** Kilometers per hour */
    KMH: { label: 'km/h', value: 'km/h', factor: 1.852 },
};

/**
 * Definition of available altitude units.
 * Each unit contains its label, value, and conversion factor from Feet.
 * 
 * @constant
 * @type {Object.<string, {label: string, value: string, factor: number}>}
 */
export const ALTITUDE_UNITS = {
    /** Feet (Base unit for aviation) */
    FT: { label: 'Feet', value: 'ft', factor: 1 },
    /** Meters */
    M: { label: 'Meters', value: 'm', factor: 0.3048 },
};

// The stores

/**
 * Store for the user selected speed unit.
 * Persists in localStorage under the key 'settings_speed_unit'.
 * @type {import('svelte/store').Writable<string>}
 */
export const speedUnit = createPersistedStore('settings_speed_unit', 'kts');

/**
 * Store for the user selected altitude unit.
 * Persists in localStorage under the key 'settings_altitude_unit'.
 * @type {import('svelte/store').Writable<string>}
 */
export const altitudeUnit = createPersistedStore('settings_altitude_unit', 'ft');


// Helpers for formatting

/**
 * Formats a speed value (in knots) to the specified unit.
 * 
 * @param {number|null} knots - The speed in knots (raw API value).
 * @param {string} unitValue - The code of the desired unit (e.g. 'kts', 'mph', 'km/h').
 * @returns {string} The formatted speed with its unit (e.g. "150 km/h") or "-" if value is null.
 */
export const formatSpeed = (knots, unitValue) => {
    if (knots === null || knots === undefined) return '-';
    
    // Find unit object, or default to Knots if not found
    const unit = Object.values(SPEED_UNITS).find(u => u.value === unitValue) || SPEED_UNITS.KTS;
    const val = knots * unit.factor;
    
    // Format without decimals
    return `${val.toLocaleString(undefined, { maximumFractionDigits: 0 })} ${unit.value}`;
};

/**
 * Formats an altitude value (in feet) to the specified unit.
 * 
 * @param {number|null} feet - The altitude in feet (raw API value).
 * @param {string} unitValue - The code of the desired unit (e.g. 'ft', 'm').
 * @returns {string} The formatted altitude with its unit (e.g. "1000 m") or "-" if value is null.
 */
export const formatAltitude = (feet, unitValue) => {
    if (feet === null || feet === undefined) return '-';
    
    // Find unit object, or default to Feet if not found
    const unit = Object.values(ALTITUDE_UNITS).find(u => u.value === unitValue) || ALTITUDE_UNITS.FT;
    const val = feet * unit.factor;
    
    // Format without decimals
    return `${val.toLocaleString(undefined, { maximumFractionDigits: 0 })} ${unit.value}`;
};
