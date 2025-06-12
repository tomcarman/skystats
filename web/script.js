// API base URL - adjust if needed
const API_BASE = '/api';

// Load all data when page loads
document.addEventListener('DOMContentLoaded', function() {
    loadAllData();
    
    // Auto-refresh every 30 seconds
    setInterval(loadAllData, 30000);
});

async function loadAllData() {
    updateLastUpdated();
    
    // Load all data concurrently
    await Promise.all([
        loadGeneralStats(),
        loadFastestAircraft(),
        loadSlowestAircraft(),
        loadHighestAircraft(),
        loadLowestAircraft(),
        loadInterestingAircraft()
    ]);
}

async function loadGeneralStats() {
    try {
        const response = await fetch(`${API_BASE}/stats/general`);
        const data = await response.json();
        
        document.getElementById('total-aircraft').textContent = data.total_aircraft?.toLocaleString() || '-';
        document.getElementById('today-aircraft').textContent = data.today_aircraft?.toLocaleString() || '-';
        document.getElementById('unique-types').textContent = data.unique_aircraft_types?.toLocaleString() || '-';
        document.getElementById('interesting-count').textContent = data.interesting_aircraft_count?.toLocaleString() || '-';
        document.getElementById('fastest-speed').textContent = data.fastest_speed ? `${data.fastest_speed.toFixed(1)} kt` : '-';
        document.getElementById('highest-altitude').textContent = data.highest_altitude ? `${data.highest_altitude.toLocaleString()} ft` : '-';
    } catch (error) {
        console.error('Error loading general stats:', error);
    }
}

async function loadFastestAircraft() {
    const container = document.getElementById('fastest-aircraft');
    try {
        const response = await fetch(`${API_BASE}/stats/fastest?limit=5`);
        const data = await response.json();
        
        container.innerHTML = data.map(aircraft => createAircraftItem(aircraft, 'speed')).join('');
    } catch (error) {
        console.error('Error loading fastest aircraft:', error);
        container.innerHTML = '<div class="error">Error loading data</div>';
    }
}

async function loadSlowestAircraft() {
    const container = document.getElementById('slowest-aircraft');
    try {
        const response = await fetch(`${API_BASE}/stats/slowest?limit=5`);
        const data = await response.json();
        
        container.innerHTML = data.map(aircraft => createAircraftItem(aircraft, 'speed')).join('');
    } catch (error) {
        console.error('Error loading slowest aircraft:', error);
        container.innerHTML = '<div class="error">Error loading data</div>';
    }
}

async function loadHighestAircraft() {
    const container = document.getElementById('highest-aircraft');
    try {
        const response = await fetch(`${API_BASE}/stats/highest?limit=5`);
        const data = await response.json();
        
        container.innerHTML = data.map(aircraft => createAircraftItem(aircraft, 'altitude')).join('');
    } catch (error) {
        console.error('Error loading highest aircraft:', error);
        container.innerHTML = '<div class="error">Error loading data</div>';
    }
}

async function loadLowestAircraft() {
    const container = document.getElementById('lowest-aircraft');
    try {
        const response = await fetch(`${API_BASE}/stats/lowest?limit=5`);
        const data = await response.json();
        
        container.innerHTML = data.map(aircraft => createAircraftItem(aircraft, 'altitude')).join('');
    } catch (error) {
        console.error('Error loading lowest aircraft:', error);
        container.innerHTML = '<div class="error">Error loading data</div>';
    }
}

async function loadInterestingAircraft() {
    const container = document.getElementById('interesting-aircraft');
    try {
        const response = await fetch(`${API_BASE}/stats/interesting?limit=10`);
        const data = await response.json();
        
        container.innerHTML = data.map(aircraft => createInterestingAircraftItem(aircraft)).join('');
    } catch (error) {
        console.error('Error loading interesting aircraft:', error);
        container.innerHTML = '<div class="error">Error loading data</div>';
    }
}

function createAircraftItem(aircraft, type) {
    const flight = aircraft.flight?.trim() || aircraft.hex || 'Unknown';
    const registration = aircraft.registration || '-';
    const aircraftType = aircraft.type || '-';
    
    let primaryMetric = '';
    let secondaryMetrics = [];
    
    if (type === 'speed') {
        primaryMetric = `${aircraft.ground_speed?.toFixed(1) || '-'} kt`;
        secondaryMetrics = [
            { label: 'IAS', value: aircraft.indicated_air_speed ? `${aircraft.indicated_air_speed} kt` : '-' },
            { label: 'TAS', value: aircraft.true_air_speed ? `${aircraft.true_air_speed} kt` : '-' }
        ];
    } else if (type === 'altitude') {
        primaryMetric = `${aircraft.barometric_altitude?.toLocaleString() || '-'} ft`;
        secondaryMetrics = [
            { label: 'Geometric', value: aircraft.geometric_altitude ? `${aircraft.geometric_altitude.toLocaleString()} ft` : '-' }
        ];
    }
    
    const seenDate = aircraft.first_seen ? formatDate(aircraft.first_seen) : '-';
    
    return `
        <div class="aircraft-item">
            <div class="aircraft-header">
                <span class="aircraft-flight">${flight}</span>
                <span class="aircraft-type">${aircraftType}</span>
            </div>
            <div class="aircraft-details">
                <div class="aircraft-detail">
                    <span class="detail-label">${type === 'speed' ? 'Ground Speed' : 'Altitude'}</span>
                    <span class="detail-value">${primaryMetric}</span>
                </div>
                <div class="aircraft-detail">
                    <span class="detail-label">Registration</span>
                    <span class="detail-value">${registration}</span>
                </div>
                ${secondaryMetrics.map(metric => `
                    <div class="aircraft-detail">
                        <span class="detail-label">${metric.label}</span>
                        <span class="detail-value">${metric.value}</span>
                    </div>
                `).join('')}
                <div class="aircraft-detail">
                    <span class="detail-label">First Seen</span>
                    <span class="detail-value">${seenDate}</span>
                </div>
            </div>
        </div>
    `;
}

function createInterestingAircraftItem(aircraft) {
    const flight = aircraft.flight?.trim() || aircraft.hex || 'Unknown';
    const registration = aircraft.registration || '-';
    const operator = aircraft.operator || '-';
    const aircraftType = aircraft.type || '-';
    const category = aircraft.category || '-';
    const group = aircraft.group || '-';
    const seenDate = aircraft.seen ? formatDate(aircraft.seen) : '-';
    
    return `
        <div class="aircraft-item">
            <div class="aircraft-header">
                <span class="aircraft-flight">${flight}</span>
                <span class="aircraft-type">${category}</span>
            </div>
            <div class="aircraft-details">
                <div class="aircraft-detail">
                    <span class="detail-label">Registration</span>
                    <span class="detail-value">${registration}</span>
                </div>
                <div class="aircraft-detail">
                    <span class="detail-label">Operator</span>
                    <span class="detail-value">${operator}</span>
                </div>
                <div class="aircraft-detail">
                    <span class="detail-label">Type</span>
                    <span class="detail-value">${aircraftType}</span>
                </div>
                <div class="aircraft-detail">
                    <span class="detail-label">Group</span>
                    <span class="detail-value">${group}</span>
                </div>
                <div class="aircraft-detail">
                    <span class="detail-label">Last Seen</span>
                    <span class="detail-value">${seenDate}</span>
                </div>
            </div>
        </div>
    `;
}

function formatDate(dateString) {
    try {
        const date = new Date(dateString);
        return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
    } catch (error) {
        return dateString;
    }
}

function updateLastUpdated() {
    const now = new Date();
    document.getElementById('last-updated').textContent = now.toLocaleTimeString();
}