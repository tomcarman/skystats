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
        loadInterestingAircraft(),
        loadRouteStats()
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

async function loadRouteStats() {
    try {
        const response = await fetch(`${API_BASE}/stats/routes`);
        const data = await response.json();
        
        // Update overview stats
        document.getElementById('total-routes').textContent = data.total_routes?.toLocaleString() || '-';
        document.getElementById('domestic-flights').textContent = data.international_vs_domestic?.domestic?.toLocaleString() || '-';
        document.getElementById('international-flights').textContent = data.international_vs_domestic?.international?.toLocaleString() || '-';
        document.getElementById('average-distance').textContent = data.average_route_distance ? `${data.average_route_distance.toFixed(0)} km` : '-';
        
        // Update top airlines
        const airlinesContainer = document.getElementById('top-airlines');
        if (data.top_airlines && data.top_airlines.length > 0) {
            airlinesContainer.innerHTML = data.top_airlines.map(airline => createAirlineItem(airline)).join('');
        } else {
            airlinesContainer.innerHTML = '<div class="no-data">No airline data available</div>';
        }
        
        // Update top routes
        const routesContainer = document.getElementById('top-routes');
        if (data.top_routes && data.top_routes.length > 0) {
            routesContainer.innerHTML = data.top_routes.map(route => createRouteItem(route)).join('');
        } else {
            routesContainer.innerHTML = '<div class="no-data">No route data available</div>';
        }
        
        // Update top airports (combining origin and destination)
        const airportsContainer = document.getElementById('top-airports');
        if (data.top_origin_airports && data.top_origin_airports.length > 0) {
            airportsContainer.innerHTML = data.top_origin_airports.slice(0, 5).map(airport => createAirportItem(airport)).join('');
        } else {
            airportsContainer.innerHTML = '<div class="no-data">No airport data available</div>';
        }
        
        // Update top countries
        const countriesContainer = document.getElementById('top-countries');
        if (data.top_countries && data.top_countries.length > 0) {
            countriesContainer.innerHTML = data.top_countries.map(country => createCountryItem(country)).join('');
        } else {
            countriesContainer.innerHTML = '<div class="no-data">No country data available</div>';
        }
        
    } catch (error) {
        console.error('Error loading route stats:', error);
        // Set error messages for all route stat containers
        const containers = ['top-airlines', 'top-routes', 'top-airports', 'top-countries'];
        containers.forEach(containerId => {
            const container = document.getElementById(containerId);
            if (container) {
                container.innerHTML = '<div class="error">Error loading data</div>';
            }
        });
    }
}

function createAirlineItem(airline) {
    const name = airline.airline_name || 'Unknown Airline';
    const icao = airline.airline_icao || '';
    const iata = airline.airline_iata || '';
    const count = airline.count || 0;
    
    const codes = [icao, iata].filter(code => code).join(' / ');
    
    return `
        <div class="route-item">
            <div class="route-header">
                <span class="route-name">${name}</span>
                <span class="route-count">${count.toLocaleString()}</span>
            </div>
            <div class="route-detail">
                <span class="route-code">${codes || 'No codes'}</span>
            </div>
        </div>
    `;
}

function createRouteItem(route) {
    const routeName = route.route || 'Unknown Route';
    const count = route.count || 0;
    
    return `
        <div class="route-item">
            <div class="route-header">
                <span class="route-name">${routeName}</span>
                <span class="route-count">${count.toLocaleString()}</span>
            </div>
        </div>
    `;
}

function createAirportItem(airport) {
    const code = airport.airport_code || 'Unknown';
    const name = airport.airport_name || 'Unknown Airport';
    const country = airport.country || '';
    const count = airport.count || 0;
    
    return `
        <div class="route-item">
            <div class="route-header">
                <span class="route-name">${code} - ${name}</span>
                <span class="route-count">${count.toLocaleString()}</span>
            </div>
            <div class="route-detail">
                <span class="route-code">${country}</span>
            </div>
        </div>
    `;
}

function createCountryItem(country) {
    const name = country.country || 'Unknown Country';
    const iso = country.country_iso || '';
    const count = country.count || 0;
    
    return `
        <div class="route-item">
            <div class="route-header">
                <span class="route-name">${name}</span>
                <span class="route-count">${count.toLocaleString()}</span>
            </div>
            <div class="route-detail">
                <span class="route-code">${iso}</span>
            </div>
        </div>
    `;
}

function updateLastUpdated() {
    const now = new Date();
    document.getElementById('last-updated').textContent = now.toLocaleTimeString();
}