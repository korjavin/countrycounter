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

    // --- Add Country Functionality ---
    const allCountries = [
        'Afghanistan', 'Albania', 'Algeria', 'Andorra', 'Angola', 'Antigua and Barbuda', 'Argentina', 'Armenia', 'Australia', 'Austria',
        'Azerbaijan', 'Bahamas', 'Bahrain', 'Bangladesh', 'Barbados', 'Belarus', 'Belgium', 'Belize', 'Benin', 'Bhutan', 'Bolivia',
        'Bosnia and Herzegovina', 'Botswana', 'Brazil', 'Brunei', 'Bulgaria', 'Burkina Faso', 'Burundi', 'Cabo Verde', 'Cambodia',
        'Cameroon', 'Canada', 'Central African Republic', 'Chad', 'Chile', 'China', 'Colombia', 'Comoros', 'Congo, Democratic Republic of the',
        'Congo, Republic of the', 'Costa Rica', 'Croatia', 'Cuba', 'Cyprus', 'Czech Republic', 'Denmark', 'Djibouti', 'Dominica',
        'Dominican Republic', 'Ecuador', 'Egypt', 'El Salvador', 'Equatorial Guinea', 'Eritrea', 'Estonia', 'Eswatini', 'Ethiopia',
        'Fiji', 'Finland', 'France', 'Gabon', 'Gambia', 'Georgia', 'Germany', 'Ghana', 'Greece', 'Grenada', 'Guatemala', 'Guinea',
        'Guinea-Bissau', 'Guyana', 'Haiti', 'Honduras', 'Hungary', 'Iceland', 'India', 'Indonesia', 'Iran', 'Iraq', 'Ireland', 'Israel',
        'Italy', 'Jamaica', 'Japan', 'Jordan', 'Kazakhstan', 'Kenya', 'Kiribati', 'Kuwait', 'Kyrgyzstan', 'Laos', 'Latvia', 'Lebanon',
        'Lesotho', 'Liberia', 'Libya', 'Liechtenstein', 'Lithuania', 'Luxembourg', 'Madagascar', 'Malawi', 'Malaysia', 'Maldives',
        'Mali', 'Malta', 'Marshall Islands', 'Mauritania', 'Mauritius', 'Mexico', 'Micronesia', 'Moldova', 'Monaco', 'Mongolia',
        'Montenegro', 'Morocco', 'Mozambique', 'Myanmar', 'Namibia', 'Nauru', 'Nepal', 'Netherlands', 'New Zealand', 'Nicaragua',
        'Niger', 'Nigeria', 'North Korea', 'North Macedonia', 'Norway', 'Oman', 'Pakistan', 'Palau', 'Palestine State', 'Panama',
        'Papua New Guinea', 'Paraguay', 'Peru', 'Philippines', 'Poland', 'Portugal', 'Qatar', 'Romania', 'Russia', 'Rwanda',
        'Saint Kitts and Nevis', 'Saint Lucia', 'Saint Vincent and the Grenadines', 'Samoa', 'San Marino', 'Sao Tome and Principe',
        'Saudi Arabia', 'Senegal', 'Serbia', 'Seychelles', 'Sierra Leone', 'Singapore', 'Slovakia', 'Slovenia', 'Solomon Islands',
        'Somalia', 'South Africa', 'South Korea', 'South Sudan', 'Spain', 'Sri Lanka', 'Sudan', 'Suriname', 'Sweden', 'Switzerland',
        'Syria', 'Taiwan', 'Tajikistan', 'Tanzania', 'Thailand', 'Timor-Leste', 'Togo', 'Tonga', 'Trinidad and Tobago', 'Tunisia',
        'Turkey', 'Turkmenistan', 'Tuvalu', 'Uganda', 'Ukraine', 'United Arab Emirates', 'United Kingdom', 'United States of America',
        'Uruguay', 'Uzbekistan', 'Vanuatu', 'Vatican City', 'Venezuela', 'Vietnam', 'Yemen', 'Zambia', 'Zimbabwe'
    ];

    const countrySelect = document.getElementById('country-select');
    allCountries.forEach(country => {
        const option = document.createElement('option');
        option.value = country;
        option.textContent = country;
        countrySelect.appendChild(option);
    });

    const addCountryBtn = document.getElementById('add-country-btn');
    const visitedCountSpan = document.getElementById('visited-count');
    const countriesUl = document.getElementById('countries-ul');
    const toggleListBtn = document.getElementById('toggle-list-btn');
    const visitedListDiv = document.getElementById('visited-list');

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
    updateStatsAndList();
});
