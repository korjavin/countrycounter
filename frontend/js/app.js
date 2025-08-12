
document.addEventListener('DOMContentLoaded', () => {
    const tg = window.Telegram.WebApp;
    tg.ready();

    let userId = null;
    if (tg.initDataUnsafe && tg.initDataUnsafe.user) {
        userId = tg.initDataUnsafe.user.id;
    }

    console.log('Telegram User ID:', userId);

    // Initialize Leaflet map
    const map = L.map('map', {
        center: [20, 0],
        zoom: 2,
        zoomControl: false,
        attributionControl: false
    });

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        maxZoom: 18,
    }).addTo(map);

    L.control.attribution({
        prefix: '<a href="https://leafletjs.com" title="A JS library for interactive maps">Leaflet</a>',
        position: 'bottomright'
    }).addTo(map);

    // Mock data for visited countries (as backend is not ready)
    const visitedCountries = ['Germany', 'France', 'Spain']; // Example data

    // Later, we will fetch this from the backend and use the GeoJSON to color the map.
    // For now, this is just a placeholder.

    // --- UI Elements and Data ---
    const countrySelect = document.getElementById('country-select');
    const addCountryBtn = document.getElementById('add-country-btn');
    const visitedCountSpan = document.getElementById('visited-count');
    const countriesUl = document.getElementById('countries-ul');
    const toggleListBtn = document.getElementById('toggle-list-btn');
    const visitedListDiv = document.getElementById('visited-list');

    // --- Core Functions ---
    async function loadCountries() {
        try {
            const response = await fetch('all_countries');
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const text = await response.text();
            const countries = text.trim().split('\n');

            countrySelect.innerHTML = '<option value="">Select a country</option>'; // Clear existing options
            countries.forEach(country => {
                const option = document.createElement('option');
                option.value = country;
                option.textContent = country;
                countrySelect.appendChild(option);
            });
        } catch (error) {
            console.error('Error loading countries:', error);
            // Optionally, display an error to the user in the UI
        }
    }

    function updateStatsAndList() {
        // Update count
        visitedCountSpan.textContent = visitedCountries.length;

        // Update list
        countriesUl.innerHTML = ''; // Clear existing list
        visitedCountries.sort().forEach(country => {
            const li = document.createElement('li');
            li.textContent = country;
            countriesUl.appendChild(li);
        });
    }

    addCountryBtn.addEventListener('click', () => {
        const selectedCountry = countrySelect.value;
        if (selectedCountry && !visitedCountries.includes(selectedCountry)) {
            visitedCountries.push(selectedCountry);
            console.log(`Simulating POST request: Add ${selectedCountry}`);
            updateStatsAndList(); // Refresh UI
        }
    });

    toggleListBtn.addEventListener('click', () => {
        const isHidden = visitedListDiv.classList.toggle('hidden');
        toggleListBtn.textContent = isHidden ? 'Show List' : 'Hide List';
    });

    // Initial UI update on page load
    loadCountries();
    updateStatsAndList();
});

