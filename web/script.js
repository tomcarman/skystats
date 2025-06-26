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
        container.innerHTML = '<div class="text-center py-8 text-red-600 bg-red-50 border border-red-200 rounded-lg">Error loading data</div>';
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
        container.innerHTML = '<div class="text-center py-8 text-red-600 bg-red-50 border border-red-200 rounded-lg">Error loading data</div>';
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
        container.innerHTML = '<div class="text-center py-8 text-red-600 bg-red-50 border border-red-200 rounded-lg">Error loading data</div>';
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
        container.innerHTML = '<div class="text-center py-8 text-red-600 bg-red-50 border border-red-200 rounded-lg">Error loading data</div>';
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
        container.innerHTML = '<tr><td colspan="7" class="px-6 py-4 text-center text-red-600 bg-red-50">Error loading data</td></tr>';
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
        <div class="bg-white border border-gray-200 rounded-lg p-4 hover:shadow-md transition-all duration-200 hover:border-gray-300">
            <div class="flex justify-between items-start mb-3">
                <div class="flex items-center space-x-2">
                    <div class="w-2 h-2 bg-blue-500 rounded-full"></div>
                    <span class="font-bold text-gray-900 text-lg">${flight}</span>
                </div>
                <span class="bg-blue-100 text-blue-800 px-2 py-1 rounded-md text-sm font-semibold">${aircraftType}</span>
            </div>
            <div class="grid grid-cols-2 lg:grid-cols-4 gap-4 text-sm">
                <div class="bg-gray-50 rounded-lg p-2">
                    <span class="block text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">${type === 'speed' ? 'Ground Speed' : 'Altitude'}</span>
                    <span class="block font-bold text-gray-900">${primaryMetric}</span>
                </div>
                <div class="bg-gray-50 rounded-lg p-2">
                    <span class="block text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">Registration</span>
                    <span class="block font-bold text-gray-900">${registration}</span>
                </div>
                ${secondaryMetrics.map(metric => `
                    <div class="bg-gray-50 rounded-lg p-2">
                        <span class="block text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">${metric.label}</span>
                        <span class="block font-bold text-gray-900">${metric.value}</span>
                    </div>
                `).join('')}
                <div class="bg-gray-50 rounded-lg p-2">
                    <span class="block text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">First Seen</span>
                    <span class="block font-bold text-gray-900">${seenDate}</span>
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
        <tr class="hover:bg-gray-50 transition-colors duration-200">
            <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center">
                    <div class="w-2 h-2 bg-yellow-500 rounded-full mr-3"></div>
                    <span class="text-sm font-medium text-gray-900">${flight}</span>
                </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${registration}</td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${operator}</td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${aircraftType}</td>
            <td class="px-6 py-4 whitespace-nowrap">
                <span class="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-yellow-100 text-yellow-800">${category}</span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${group}</td>
            <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">${seenDate}</td>
        </tr>
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
            airlinesContainer.innerHTML = '<div class="text-center py-8 text-gray-500 bg-gray-50 border border-gray-200 rounded-lg">No airline data available</div>';
        }
        
        // Update top routes
        const routesContainer = document.getElementById('top-routes');
        if (data.top_routes && data.top_routes.length > 0) {
            routesContainer.innerHTML = data.top_routes.map(route => createRouteItem(route)).join('');
        } else {
            routesContainer.innerHTML = '<div class="text-center py-8 text-gray-500 bg-gray-50 border border-gray-200 rounded-lg">No route data available</div>';
        }
        
        // Update top airports (combining origin and destination)
        const airportsContainer = document.getElementById('top-airports');
        if (data.top_origin_airports && data.top_origin_airports.length > 0) {
            airportsContainer.innerHTML = data.top_origin_airports.slice(0, 5).map(airport => createAirportItem(airport)).join('');
        } else {
            airportsContainer.innerHTML = '<div class="text-center py-8 text-gray-500 bg-gray-50 border border-gray-200 rounded-lg">No airport data available</div>';
        }
        
        // Update top countries
        const countriesContainer = document.getElementById('top-countries');
        if (data.top_countries && data.top_countries.length > 0) {
            countriesContainer.innerHTML = data.top_countries.map(country => createCountryItem(country)).join('');
        } else {
            countriesContainer.innerHTML = '<div class="text-center py-8 text-gray-500 bg-gray-50 border border-gray-200 rounded-lg">No country data available</div>';
        }
        
    } catch (error) {
        console.error('Error loading route stats:', error);
        // Set error messages for all route stat containers
        const containers = ['top-airlines', 'top-routes', 'top-airports', 'top-countries'];
        containers.forEach(containerId => {
            const container = document.getElementById(containerId);
            if (container) {
                container.innerHTML = '<div class="text-center py-8 text-red-600 bg-red-50 border border-red-200 rounded-lg">Error loading data</div>';
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
        <div class="bg-gray-50 rounded-lg p-3 hover:bg-gray-100 transition-colors duration-200 border border-gray-200">
            <div class="flex justify-between items-center mb-2">
                <span class="font-semibold text-gray-900 text-sm">${name}</span>
                <span class="bg-blue-100 text-blue-800 px-2 py-1 rounded-md text-xs font-bold min-w-[35px] text-center">${count.toLocaleString()}</span>
            </div>
            <div class="flex items-center">
                <span class="text-xs text-gray-600 font-medium">${codes || 'No codes'}</span>
            </div>
        </div>
    `;
}

function createRouteItem(route) {
    const routeName = route.route || 'Unknown Route';
    const count = route.count || 0;
    
    return `
        <div class="bg-gray-50 rounded-lg p-3 hover:bg-gray-100 transition-colors duration-200 border border-gray-200">
            <div class="flex justify-between items-center">
                <span class="font-semibold text-gray-900 text-sm">${routeName}</span>
                <span class="bg-green-100 text-green-800 px-2 py-1 rounded-md text-xs font-bold min-w-[35px] text-center">${count.toLocaleString()}</span>
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
        <div class="bg-gray-50 rounded-lg p-3 hover:bg-gray-100 transition-colors duration-200 border border-gray-200">
            <div class="flex justify-between items-center mb-2">
                <span class="font-semibold text-gray-900 text-sm">${code} - ${name}</span>
                <span class="bg-purple-100 text-purple-800 px-2 py-1 rounded-md text-xs font-bold min-w-[35px] text-center">${count.toLocaleString()}</span>
            </div>
            <div class="flex items-center">
                <span class="text-xs text-gray-600 font-medium">${country}</span>
            </div>
        </div>
    `;
}

function createCountryItem(country) {
    const name = country.country || 'Unknown Country';
    const iso = country.country_iso || '';
    const count = country.count || 0;
    
    return `
        <div class="bg-gray-50 rounded-lg p-3 hover:bg-gray-100 transition-colors duration-200 border border-gray-200">
            <div class="flex justify-between items-center mb-2">
                <span class="font-semibold text-gray-900 text-sm">${name}</span>
                <span class="bg-orange-100 text-orange-800 px-2 py-1 rounded-md text-xs font-bold min-w-[35px] text-center">${count.toLocaleString()}</span>
            </div>
            <div class="flex items-center">
                <span class="text-xs text-gray-600 font-medium">${iso}</span>
            </div>
        </div>
    `;
}

function updateLastUpdated() {
    const now = new Date();
    document.getElementById('last-updated').textContent = now.toLocaleTimeString();
}