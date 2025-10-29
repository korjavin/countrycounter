document.addEventListener("DOMContentLoaded", () => {
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

  console.log("User ID:", userId);

  // Debug panel helper
  function debugLog(message) {
    const debugPanel = document.getElementById("debug-panel");
    const debugContent = document.getElementById("debug-content");
    if (debugPanel && debugContent) {
      debugPanel.style.display = "block";
      const line = document.createElement("div");
      line.textContent = message;
      line.style.borderBottom = "1px solid #333";
      line.style.paddingBottom = "2px";
      line.style.marginBottom = "2px";
      debugContent.appendChild(line);
      while (debugContent.children.length > 20) {
        debugContent.removeChild(debugContent.firstChild);
      }
    }
    console.log(message);
  }

  const map = L.map("map", {
    center: [20, 0],
    zoom: 2,
    zoomControl: false,
    attributionControl: false,
  });

  L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
    maxZoom: 18,
  }).addTo(map);

  let visitedCountries = [];
  let geojsonLayer;
  let markerLayers = [];

  const customIcon = L.icon({
    iconUrl: "images/marker.png",
    iconSize: [25, 41],
    iconAnchor: [12, 41],
    popupAnchor: [1, -34],
  });

  const countryInput = document.getElementById("country-input");
  const countryList = document.getElementById("country-list");
  const addCountryBtn = document.getElementById("add-country-btn");
  const visitedCountSpan = document.getElementById("visited-count");
  const countriesUl = document.getElementById("countries-ul");
  const toggleListBtn = document.getElementById("toggle-list-btn");
  const visitedListDiv = document.getElementById("visited-list");
  const statusBar = document.getElementById("status-bar");

  function updateStatus(message, isError = false) {
    statusBar.textContent = message;
    statusBar.style.backgroundColor = isError ? "#d9534f" : "#5cb85c";
    setTimeout(() => {
      statusBar.textContent = "";
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
        if (response.status === 404) {
          // Handles new user case
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
      console.error("Error fetching visited countries:", error);
      updateStatus(`Error: ${error.message}`, true);
    }
  }

  async function addCountry(country) {
    if (!userId) {
      // Standalone mode: just update the local list and map
      if (!visitedCountries.includes(country)) {
        visitedCountries.push(country);
        updateMap();
        updateStatus(`${country} added locally.`);
      }
      return;
    }
    try {
      updateStatus(`Adding ${country}...`);
      const response = await fetch("/api/countries", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
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
      console.error("Error adding country:", error);
      updateStatus(`Error adding ${country}: ${error.message}`, true);
    }
  }

  async function deleteCountry(country) {
    if (!userId) {
      // Standalone mode: just update the local list and map
      const index = visitedCountries.indexOf(country);
      if (index > -1) {
        visitedCountries.splice(index, 1);
        updateMap();
        updateStatus(`${country} deleted locally.`);
      }
      return;
    }
    try {
      updateStatus(`Deleting ${country}...`);
      const response = await fetch("/api/countries", {
        method: "DELETE",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ userId: userId, country: country }),
      });

      if (response.status === 200) {
        updateStatus(`${country} deleted successfully!`);
        await getVisitedCountries(); // Refresh the list from the backend
      } else {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
    } catch (error) {
      console.error("Error deleting country:", error);
      updateStatus(`Error deleting ${country}: ${error.message}`, true);
    }
  }

  async function loadAllCountries() {
    try {
      const response = await fetch("all_countries.json");
      if (!response.ok) throw new Error("Failed to load all_countries.json");
      const countries = await response.json();

      countryList.innerHTML = "";
      countries.forEach((country) => {
        const option = document.createElement("option");
        option.value = country;
        countryList.appendChild(option);
      });
    } catch (error) {
      console.error("Error loading country list:", error);
    }
  }

  async function loadMapData() {
    try {
      const response = await fetch("/data/countries.geo.json");
      const geojsonData = await response.json();
      geojsonLayer = L.geoJSON(geojsonData, {
        style: countryStyle,
      }).addTo(map);
      await getVisitedCountries(); // Fetch user data after map is ready
    } catch (error) {
      console.error("Error loading GeoJSON data:", error);
    }
  }

  function countryStyle(feature) {
    const countryName = feature.properties.name;
    const isVisited = visitedCountries.includes(countryName);

    // Debug logging for Czech countries
    if (
      countryName &&
      (countryName.includes("Czech") || countryName === "Czechia")
    ) {
      debugLog(
        `Country: '${countryName}' len=${countryName.length} isVisited=${isVisited}`,
      );
      debugLog(`visitedCountries has ${visitedCountries.length} items`);
      const czechInList = visitedCountries.filter((c) => c.includes("Czech"));
      debugLog(`Czech in list: ${JSON.stringify(czechInList)}`);
    }

    return {
      fillColor: isVisited ? "#d4ac0d" : "#f0f0f0",
      weight: 2,
      opacity: 1,
      color: "white",
      dashArray: "3",
      fillOpacity: 0.7,
    };
  }

  function updateMap() {
    if (geojsonLayer) {
      geojsonLayer.setStyle(countryStyle);
    }
    visitedCountSpan.textContent = visitedCountries.length;
    countriesUl.innerHTML = "";
    visitedCountries.sort().forEach((country) => {
      const li = document.createElement("li");

      const countryName = document.createElement("span");
      countryName.textContent = country;
      li.appendChild(countryName);

      const deleteBtn = document.createElement("button");
      deleteBtn.textContent = "Delete";
      deleteBtn.classList.add("delete-btn");
      deleteBtn.onclick = () => deleteCountry(country);
      li.appendChild(deleteBtn);

      countriesUl.appendChild(li);
    });
  }

  addCountryBtn.addEventListener("click", () => {
    const selectedCountry = countryInput.value;
    if (selectedCountry && !visitedCountries.includes(selectedCountry)) {
      addCountry(selectedCountry);
      countryInput.value = ""; // Clear the input after adding
    }
  });

  toggleListBtn.addEventListener("click", () => {
    const isHidden = visitedListDiv.classList.toggle("hidden");
    toggleListBtn.textContent = isHidden ? "Show List" : "Hide List";
  });

  // Initial Load
  loadAllCountries();
  loadMapData();
});
