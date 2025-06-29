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
    
    // Load all data concurrently (matching visual order)
    await Promise.all([
        loadGeneralStats(),
        loadRouteStats(),
        loadMilitaryAircraft(),
        loadGovernmentAircraft(),
        loadPoliceAircraft(),
        loadCivilianAircraft(),
        loadFastestAircraft(),
        loadSlowestAircraft(),
        loadHighestAircraft(),
        loadLowestAircraft()
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
    } catch (error) {
        console.error('Error loading general stats:', error);
    }
}

async function loadFastestAircraft() {
    const container = document.getElementById('fastest-aircraft');
    try {
        const response = await fetch(`${API_BASE}/stats/fastest?limit=5`);
        const data = await response.json();
        
        container.innerHTML = data.map(aircraft => createSimpleTableRow(aircraft, 'speed')).join('');
    } catch (error) {
        console.error('Error loading fastest aircraft:', error);
        container.innerHTML = '<tr><td colspan="4" class="px-6 py-4 text-center text-red-600 bg-red-50">Error loading data</td></tr>';
    }
}

async function loadSlowestAircraft() {
    const container = document.getElementById('slowest-aircraft');
    try {
        const response = await fetch(`${API_BASE}/stats/slowest?limit=5`);
        const data = await response.json();
        
        container.innerHTML = data.map(aircraft => createSimpleTableRow(aircraft, 'speed')).join('');
    } catch (error) {
        console.error('Error loading slowest aircraft:', error);
        container.innerHTML = '<tr><td colspan="4" class="px-6 py-4 text-center text-red-600 bg-red-50">Error loading data</td></tr>';
    }
}

async function loadHighestAircraft() {
    const container = document.getElementById('highest-aircraft');
    try {
        const response = await fetch(`${API_BASE}/stats/highest?limit=5`);
        const data = await response.json();
        
        container.innerHTML = data.map(aircraft => createSimpleTableRow(aircraft, 'altitude')).join('');
    } catch (error) {
        console.error('Error loading highest aircraft:', error);
        container.innerHTML = '<tr><td colspan="4" class="px-6 py-4 text-center text-red-600 bg-red-50">Error loading data</td></tr>';
    }
}

async function loadLowestAircraft() {
    const container = document.getElementById('lowest-aircraft');
    try {
        const response = await fetch(`${API_BASE}/stats/lowest?limit=5`);
        const data = await response.json();
        
        container.innerHTML = data.map(aircraft => createSimpleTableRow(aircraft, 'altitude')).join('');
    } catch (error) {
        console.error('Error loading lowest aircraft:', error);
        container.innerHTML = '<tr><td colspan="4" class="px-6 py-4 text-center text-red-600 bg-red-50">Error loading data</td></tr>';
    }
}

async function loadCivilianAircraft() {
    const container = document.getElementById('civilian-aircraft');
    try {
        const response = await fetch(`${API_BASE}/stats/interesting/civilian?limit=5`);
        const data = await response.json();
        
        container.innerHTML = data.map(aircraft => createCategoryAircraftItem(aircraft)).join('');
    } catch (error) {
        console.error('Error loading civilian aircraft:', error);
        container.innerHTML = '<tr><td colspan="7" class="px-4 py-4 text-center text-red-600 bg-red-50">Error loading data</td></tr>';
    }
}

async function loadPoliceAircraft() {
    const container = document.getElementById('police-aircraft');
    try {
        const response = await fetch(`${API_BASE}/stats/interesting/police?limit=5`);
        const data = await response.json();
        
        container.innerHTML = data.map(aircraft => createCategoryAircraftItem(aircraft)).join('');
    } catch (error) {
        console.error('Error loading police aircraft:', error);
        container.innerHTML = '<tr><td colspan="7" class="px-4 py-4 text-center text-red-600 bg-red-50">Error loading data</td></tr>';
    }
}

async function loadMilitaryAircraft() {
    const container = document.getElementById('military-aircraft');
    try {
        const response = await fetch(`${API_BASE}/stats/interesting/military?limit=5`);
        const data = await response.json();
        
        container.innerHTML = data.map(aircraft => createCategoryAircraftItem(aircraft)).join('');
    } catch (error) {
        console.error('Error loading military aircraft:', error);
        container.innerHTML = '<tr><td colspan="7" class="px-4 py-4 text-center text-red-600 bg-red-50">Error loading data</td></tr>';
    }
}

async function loadGovernmentAircraft() {
    const container = document.getElementById('government-aircraft');
    try {
        const response = await fetch(`${API_BASE}/stats/interesting/government?limit=5`);
        const data = await response.json();
        
        container.innerHTML = data.map(aircraft => createCategoryAircraftItem(aircraft)).join('');
    } catch (error) {
        console.error('Error loading government aircraft:', error);
        container.innerHTML = '<tr><td colspan="7" class="px-4 py-4 text-center text-red-600 bg-red-50">Error loading data</td></tr>';
    }
}

function createSimpleTableRow(aircraft, type) {
    const registration = aircraft.registration || '-';
    const aircraftType = aircraft.type || '-';
    const seenDate = aircraft.first_seen ? formatDate(aircraft.first_seen) : '-';
    
    let primaryMetric = '';
    if (type === 'speed') {
        primaryMetric = aircraft.ground_speed ? `${aircraft.ground_speed.toFixed(1)} kt` : '-';
    } else {
        primaryMetric = aircraft.barometric_altitude ? `${aircraft.barometric_altitude.toLocaleString()} ft` : '-';
    }
    
    return `
        <tr class="hover:bg-gray-50 transition-colors duration-200">
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${registration}</td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${aircraftType}</td>
            <td class="px-6 py-4 whitespace-nowrap text-sm font-semibold text-gray-900">${primaryMetric}</td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${seenDate}</td>
        </tr>
    `;
}

function createCategoryAircraftItem(aircraft) {
    const flight = aircraft.flight?.trim() || aircraft.hex || 'Unknown';
    const registration = aircraft.registration || '-';
    const operator = aircraft.operator || '-';
    const aircraftType = aircraft.type || '-';
    const category = aircraft.category || '-';
    const seenDate = aircraft.seen ? formatDate(aircraft.seen) : '-';
    
    // Combine tags into a compact display
    const tags = [aircraft.tag1, aircraft.tag2, aircraft.tag3].filter(tag => tag && tag.trim() !== '');
    const tagsDisplay = tags.length > 0 ? 
        tags.map(tag => `<span class="inline-block px-1.5 py-0.5 text-xs bg-gray-100 text-gray-700 rounded">${tag}</span>`).join(' ') : 
        '-';
    
    // Set color based on group
    let dotColor = 'bg-blue-500';
    if (aircraft.group === 'Pol') dotColor = 'bg-indigo-500';
    else if (aircraft.group === 'Mil') dotColor = 'bg-red-500';
    else if (aircraft.group === 'Gov') dotColor = 'bg-green-500';
    
    // Collect image links
    const imageLinks = [aircraft.image_link_1, aircraft.image_link_2, aircraft.image_link_3]
        .filter(link => link && link.trim() !== '')
        .join('|');
    
    // Add hoverable-row class and data attributes if images exist
    const rowClass = imageLinks ? 'hoverable-row' : '';
    const dataAttributes = imageLinks ? `data-images="${imageLinks}" data-registration="${registration}" data-type="${aircraftType}"` : '';
    
    return `
        <tr class="hover:bg-gray-50 transition-colors duration-200 ${rowClass}" ${dataAttributes}>
            <td class="px-4 py-3 whitespace-nowrap">
                <div class="flex items-center">
                    <div class="w-2 h-2 ${dotColor} rounded-full mr-2"></div>
                    <span class="text-sm font-medium text-gray-900">${flight}</span>
                </div>
            </td>
            <td class="px-4 py-3 whitespace-nowrap text-sm text-gray-900">${registration}</td>
            <td class="px-4 py-3 whitespace-nowrap text-sm text-gray-900 max-w-32 truncate">${operator}</td>
            <td class="px-4 py-3 whitespace-nowrap text-sm text-gray-900">${aircraftType}</td>
            <td class="px-4 py-3 whitespace-nowrap text-sm text-gray-900">${category}</td>
            <td class="px-4 py-3 whitespace-nowrap text-sm text-gray-900">${tagsDisplay}</td>
            <td class="px-4 py-3 whitespace-nowrap text-sm font-medium text-gray-900">${seenDate}</td>
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

// Aircraft image hover functionality
document.addEventListener('DOMContentLoaded', function() {
    // Add event delegation for hover events on rows with images
    document.addEventListener('mouseenter', function(e) {
        if (e.target.closest('.hoverable-row')) {
            const row = e.target.closest('.hoverable-row');
            showAircraftImageOverlay(e, row);
        }
    }, true);

    document.addEventListener('mouseleave', function(e) {
        if (e.target.closest('.hoverable-row')) {
            hideAircraftImageOverlay();
        }
    }, true);
});

function showAircraftImageOverlay(event, row) {
    const images = row.dataset.images;
    const registration = row.dataset.registration;
    const type = row.dataset.type;
    
    if (!images) return;
    
    const overlay = document.getElementById('aircraftImageOverlay');
    const infoContainer = overlay.querySelector('.aircraft-info-overlay');
    const imageContainer = overlay.querySelector('.aircraft-image-container');
    
    // Set aircraft info
    infoContainer.textContent = `${registration} - ${type}`;
    
    // Clear previous images and add new ones
    imageContainer.innerHTML = '';
    
    const imageLinks = images.split('|');
    imageLinks.forEach(link => {
        if (link && link.trim()) {
            const img = document.createElement('img');
            img.src = link.trim();
            img.alt = `Aircraft ${registration}`;
            img.onerror = function() {
                this.style.display = 'none';
            };
            imageContainer.appendChild(img);
        }
    });
    
    // Show overlay to get its actual dimensions
    overlay.style.visibility = 'hidden';
    overlay.classList.add('show');
    
    // Calculate position to keep overlay within viewport
    const rect = row.getBoundingClientRect();
    const overlayRect = overlay.getBoundingClientRect();
    const overlayWidth = overlayRect.width || 650; // Use actual width or max-width
    const overlayHeight = overlayRect.height || 400; // Use actual height or estimate
    const padding = 15;
    const viewportWidth = window.innerWidth;
    const viewportHeight = window.innerHeight;
    const scrollY = window.scrollY;
    const scrollX = window.scrollX;
    
    // Try to position the overlay in this order of preference:
    // 1. To the right of the row
    // 2. To the left of the row
    // 3. Below the row
    // 4. Above the row
    // 5. Centered on screen if all else fails
    
    let left, top;
    
    // Option 1: Try right side first
    if (rect.right + padding + overlayWidth <= viewportWidth) {
        left = rect.right + padding + scrollX;
        top = rect.top + scrollY - (overlayHeight - rect.height) / 2; // Center vertically relative to row
    }
    // Option 2: Try left side
    else if (rect.left - padding - overlayWidth >= 0) {
        left = rect.left - padding - overlayWidth + scrollX;
        top = rect.top + scrollY - (overlayHeight - rect.height) / 2; // Center vertically relative to row
    }
    // Option 3: Try below
    else if (rect.bottom + padding + overlayHeight <= viewportHeight) {
        left = Math.max(padding, Math.min(rect.left + scrollX, viewportWidth - overlayWidth - padding));
        top = rect.bottom + padding + scrollY;
    }
    // Option 4: Try above
    else if (rect.top - padding - overlayHeight >= 0) {
        left = Math.max(padding, Math.min(rect.left + scrollX, viewportWidth - overlayWidth - padding));
        top = rect.top - padding - overlayHeight + scrollY;
    }
    // Option 5: Center on screen
    else {
        left = (viewportWidth - overlayWidth) / 2 + scrollX;
        top = (viewportHeight - overlayHeight) / 2 + scrollY;
    }
    
    // Ensure overlay stays within viewport bounds
    left = Math.max(padding, Math.min(left, viewportWidth - overlayWidth - padding + scrollX));
    top = Math.max(padding + scrollY, Math.min(top, viewportHeight - overlayHeight - padding + scrollY));
    
    overlay.style.left = `${left}px`;
    overlay.style.top = `${top}px`;
    overlay.style.visibility = 'visible';
    
    // Show the overlay
    overlay.classList.add('show');
}

function hideAircraftImageOverlay() {
    const overlay = document.getElementById('aircraftImageOverlay');
    overlay.classList.remove('show');
}