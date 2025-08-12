
document.addEventListener('DOMContentLoaded', () => {
    const tg = window.Telegram.WebApp;
    tg.ready();

    let userId = null;
    if (tg.initDataUnsafe && tg.initDataUnsafe.user) {
        userId = tg.initDataUnsafe.user.id;
    } else {
        // For testing in a normal browser
        // You can manually set a user ID here, e.g., userId = 12345;
        console.log("Telegram user data not found. Running in standalone mode.");
    }

    console.log('User ID:', userId);

    const map = L.map('map', {
        center: [20, 0],
        zoom: 2,
        zoomControl: false,
        attributionControl: false
    });

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        maxZoom: 18,
    }).addTo(map);

    let visitedCountries = [];
    let geojsonLayer;
    let markerLayers = [];

    const customIcon = L.icon({
        iconUrl: 'images/marker.png',
        iconSize: [25, 41],
        iconAnchor: [12, 41],
        popupAnchor: [1, -34],
        shadowSize: [41, 41]
    });

    const countrySelect = document.getElementById('country-select');
    const addCountryBtn = document.getElementById('add-country-btn');
    const visitedCountSpan = document.getElementById('visited-count');
    const countriesUl = document.getElementById('countries-ul');
    const toggleListBtn = document.getElementById('toggle-list-btn');
    const visitedListDiv = document.getElementById('visited-list');
    const statusBar = document.getElementById('status-bar');

    function updateStatus(message, isError = false) {
        statusBar.textContent = message;
        statusBar.style.backgroundColor = isError ? '#d9534f' : '#5cb85c';
        setTimeout(() => {
            statusBar.textContent = '';
        }, 3000); // Hide after 3 seconds
    }

    async function getVisitedCountries() {
        if (!userId) {
            visitedCountries = [];
            updateStatus("Running in standalone mode.", true);
            updateMap();
            return;
        }
        try {
            updateStatus("Loading visited countries...");
            const response = await fetch(`/api/countries?userId=${userId}`);
            if (!response.ok) {
                if (response.status === 404) { // Handles new user case
                    visitedCountries = [];
                    updateStatus("New user, ready to add countries.");
                } else {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
            } else {
                visitedCountries = await response.json();
                updateStatus("Visited countries loaded.");
            }
            updateMap();
        } catch (error) {
            console.error('Error fetching visited countries:', error);
            updateStatus(`Error: ${error.message}`, true);
        }
    }

    async function addCountry(country) {
        if (!userId) {
            console.error("Cannot add country without a user ID.");
            updateStatus("Cannot save country without a user.", true);
            return;
        }
        try {
            updateStatus(`Adding ${country}...`);
            const response = await fetch('/api/countries', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ userId: userId, country: country }),
            });

            if (response.status === 201) {
                updateStatus(`${country} added successfully!`);
                await getVisitedCountries(); // Refresh the list from the backend
            } else {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
        } catch (error) {
            console.error('Error adding country:', error);
            updateStatus(`Error adding ${country}: ${error.message}`, true);
        }
    }

    async function loadAllCountries() {
        try {
            // Assuming `all_countries.json` is in the `frontend` root
            const response = await fetch('all_countries.json');
            if (!response.ok) throw new Error('Failed to load all_countries.json');
            const countries = await response.json();

            countrySelect.innerHTML = '<option value="">Select a country</option>';
            countries.forEach(country => {
                const option = document.createElement('option');
                option.value = country;
                option.textContent = country;
                countrySelect.appendChild(option);
            });
        } catch (error) {
            console.error('Error loading country list:', error);
        }
    }

    async function loadMapData() {
        try {
            const response = await fetch('countries.geo.json');
            const geojsonData = await response.json();
            geojsonLayer = L.geoJSON(geojsonData, {
                style: countryStyle,
            }).addTo(map);
            await getVisitedCountries(); // Fetch user data after map is ready
        } catch (error) {
            console.error('Error loading GeoJSON data:', error);
        }
    }

    function countryStyle(feature) {
        return {
            fillColor: visitedCountries.includes(feature.properties.name) ? 'green' : '#3388ff',
            weight: 2,
            opacity: 1,
            color: 'white',
            dashArray: '3',
            fillOpacity: 0.7
        };
    }

    function updateMap() {
        // Clear existing markers
        markerLayers.forEach(marker => {
            map.removeLayer(marker);
        });
        markerLayers = [];

        if (geojsonLayer) {
            geojsonLayer.eachLayer(layer => {
                const countryName = layer.feature.properties.name;
                if (visitedCountries.includes(countryName)) {
                    const center = layer.getBounds().getCenter();
                    const marker = L.marker(center, { icon: customIcon }).addTo(map);
                    markerLayers.push(marker);
                }
            });
            geojsonLayer.setStyle(countryStyle);
        }
        visitedCountSpan.textContent = visitedCountries.length;
        countriesUl.innerHTML = '';
        visitedCountries.sort().forEach(country => {
            const li = document.createElement('li');
            li.textContent = country;
            countriesUl.appendChild(li);
        });
    }

    addCountryBtn.addEventListener('click', () => {
        const selectedCountry = countrySelect.value;
        if (selectedCountry && !visitedCountries.includes(selectedCountry)) {
            addCountry(selectedCountry);
        }
    });

    toggleListBtn.addEventListener('click', () => {
        const isHidden = visitedListDiv.classList.toggle('hidden');
        toggleListBtn.textContent = isHidden ? 'Show List' : 'Hide List';
    });

    // Initial Load
    loadAllCountries();
    loadMapData();
});

