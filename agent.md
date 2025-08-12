# TODO List for Country Counter Development

This document outlines the tasks for a coding agent to develop the Country Counter application.

## High-Level Tasks

1. **Frontend Development:**
    * Create a responsive, mobile-optimized frontend using pure JavaScript.
    * Implement the world map visualization using a suitable library (e.g., Leaflet, Mapbox).
    * Display basic statistics (number of visited countries).
    * Implement the searchable/sortable list of visited countries, hidden by default, with a toggle to show/hide.
    * Implement the "Add country" mode with a user-friendly country selection interface.
2. **Backend Development (Golang):**
    * Implement the backend using the standard HTTP router.
    * Handle Telegram user authentication.
    * Store and retrieve user data (user ID, visited countries) in a JSON file.
    * Serve the frontend files.
3. **Dockerization:**
    * Create a Dockerfile to build a single container image containing both frontend and backend.
    * Configure GitHub Actions to build and push the image to `ghcr.io/korjavin/countrycounter`.

## Detailed Tasks

### Frontend

* **Map Visualization:**
    * Use a JavaScript mapping library to display a world map.
    * Color countries based on visited status.
* **Statistics:**
    * Display the number of visited countries.
* **Visited Countries List:**
    * Create a hidden table/list of visited countries.
    * Implement search and sort functionality for the list.
    * Add a button/toggle to show/hide the list.
* **Add Country Mode:**
    * Implement a user interface for adding countries.
    * Provide a list of countries for selection (can be hardcoded initially).
    * Upon selection, update the map and statistics.

### Backend

* **Telegram Integration:**
    * Use the Telegram Bot API to authenticate users.
* **Data Handling:**
    * Store user data (user ID, visited countries) in a JSON file.
    * Implement API endpoints for retrieving and updating user data.
* **Frontend Serving:**
    * Serve the frontend files (HTML, CSS, JS) using the Golang HTTP router.

### Docker

* **Dockerfile:**
    * Create a `Dockerfile` to build the image.
    * Include necessary dependencies for both frontend and backend.
* **GitHub Actions:**
    * Create a workflow file (`.github/workflows/docker-image.yml`) to automate the build and push process to `ghcr.io/korjavin/countrycounter`.

## Additional Information

* Target platform: Telegram mini web app.
* Focus on mobile optimization for the frontend.
* Use pure JavaScript for the frontend.
* Use Golang with standard HTTP router for the backend.
* Store user data in a JSON file.
