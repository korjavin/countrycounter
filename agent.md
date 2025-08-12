# Task Description for Coding Agent

## Overall Goal

Your primary objective is to implement the "Country Counter" Telegram Mini Web App. This involves creating a Go backend to manage user data, a pure JavaScript frontend for visualization and user interaction, and setting up the necessary Docker and GitHub Actions configurations for deployment.

## Part 1: Backend Implementation (Go)

Your first set of tasks is to build the backend server.

**Key Files:** `backend/main.go`, `backend/data.json`

---

### **Task 1.1: Basic HTTP Server Setup**

-   **Description**: Create a basic HTTP server using Go's standard `net/http` package. It should listen on port `8080`. For routing, you can use the standard `http.ServeMux` or a lightweight router like `julienschmidt/httprouter`. [9, 32]
-   **File to Modify**: `backend/main.go`
-   **Details**:
    -   Initialize a new router.
    -   Create a main function that starts the HTTP server on port `8080`.
    -   Add a simple health check endpoint at `/health` that returns a `200 OK` status.

---

### **Task 1.2: Data Structures**

-   **Description**: Define the Go structs needed to handle the application's data. The primary data model is a map where the key is the user's Telegram ID and the value is a list of visited country names.
-   **File to Modify**: `backend/main.go`
-   **Details**:
    -   Create a `UserData` struct or a `map[int64][]string` to hold the data. `int64` will be used for the Telegram User ID.

---

### **Task 1.3: JSON Data Persistence**

-   **Description**: Implement functions to read from and write to a `data.json` file. This file will act as our simple database.
-   **File to Modify**: `backend/main.go`
-   **Details**:
    -   Create a `loadData()` function that reads `backend/data.json` into your data structure when the server starts. If the file doesn't exist, initialize an empty data structure.
    -   Create a `saveData()` function that writes the current data structure back to `backend/data.json` in a pretty-printed JSON format. This should be called after any data modification.
    -   Use the standard `encoding/json` package. There are also simple file-based DB libraries like `simdb` if you prefer. [14, 22]

---

### **Task 1.4: API Endpoints**

-   **Description**: Create the API endpoints that the frontend will use to get and update user data.
-   **File to Modify**: `backend/main.go`
-   **Details**:
    -   **`GET /api/countries`**: This endpoint should retrieve the Telegram user ID from the request (e.g., from a header or query parameter), look up the user's visited countries in your data structure, and return them as a JSON array.
    -   **`POST /api/countries`**: This endpoint should read the user ID and a new country from the request body, add the country to the user's list, save the data to `data.json`, and return a success status.

---

### **Task 1.5: Serve Frontend Files**

-   **Description**: Configure the Go server to serve the static files for the frontend (`index.html`, CSS, and JS).
-   **File to Modify**: `backend/main.go`
-   **Details**:
    -   Use `http.FileServer` to serve the `frontend` directory. [23]
    -   Create a handler that serves `frontend/index.html` for the root path (`/`).
    -   Set up routes for `/css` and `/js` to serve files from the corresponding frontend subdirectories.

## Part 2: Frontend Implementation (JavaScript)

Now, build the user-facing part of the application.

**Key Files:** `frontend/index.html`, `frontend/js/app.js`, `frontend/css/style.css`

---

### **Task 2.1: HTML and CSS Structure**

-   **Description**: Create the basic HTML structure and CSS for a clean, mobile-friendly layout.
-   **Files to Modify**: `frontend/index.html`, `frontend/css/style.css`
-   **Details**:
    -   In `index.html`, include a `div` for the map (e.g., `<div id="map"></div>`), a section for stats, and a hidden `div` for the list of visited countries.
    -   In `style.css`, ensure the map container has a defined height and the layout is responsive.

---

### **Task 2.2: Telegram Web App Integration**

-   **Description**: Initialize the Telegram Web App script and retrieve the user's data.
-   **File to Modify**: `frontend/index.html`, `frontend/js/app.js`
-   **Details**:
    -   Include the Telegram Web App script in `index.html`: `<script src="https://telegram.org/js/telegram-web-app.js"></script>`.
    -   In `app.js`, initialize the app and get the user's ID from `window.Telegram.WebApp.initDataUnsafe.user.id`. Store this ID for API requests.

---

### **Task 2.3: World Map Visualization**

-   **Description**: Integrate Leaflet.js to display an interactive world map. Fetch the user's visited countries from the backend and color them on the map.
-   **File to Modify**: `frontend/js/app.js`
-   **Details**:
    -   Initialize a Leaflet map. [1, 46]
    -   You will need a GeoJSON file with country boundaries. You can find one online or use a service that provides map tiles.
    -   On page load, make a `fetch` request to your `GET /api/countries` endpoint.
    -   When the data is received, iterate through the GeoJSON features and apply a different style (e.g., a fill color) to the countries that are in the user's visited list.

---

### **Task 2.4: Add Country Functionality**

-   **Description**: Implement the "add country" mode. This should include a user-friendly way to select a country from a list.
-   **File to Modify**: `frontend/js/app.js`, `frontend/index.html`
-   **Details**:
    -   Create a button to enter "add mode".
    -   In this mode, display a searchable dropdown (`<select>` or a custom component) of all countries. You can hardcode this list in your JavaScript.
    -   When a user selects a country and confirms, send a `POST` request to `/api/countries` with the user ID and the selected country name.
    -   On success, refresh the map to show the newly added country.

---

### **Task 2.5: Stats and Visited List**

-   **Description**: Display the count of visited countries and populate the hidden list.
-   **File to Modify**: `frontend/js/app.js`, `frontend/index.html`
-   **Details**:
    -   After fetching the country data, update a `<span>` or `<div>` to show the total count.
    -   Populate an `<ul>` with the names of the visited countries.
    -   Add a button to toggle the visibility of this list.

## Part 3: Docker and CI/CD

Finally, containerize the application and set up the automated build pipeline.

---

### **Task 3.1: Create the Dockerfile**

-   **Description**: Write a multi-stage `Dockerfile` to create a lean production image.
-   **File to Create**: `Dockerfile`
-   **Details**:
    -   **Build Stage**: Use a `golang` base image to build the backend executable. [26]
    -   **Final Stage**: Use a minimal base image (like `alpine` or `scratch`). Copy the compiled Go binary from the build stage and the `frontend` directory.
    -   Expose port `8080`.
    -   Set the `CMD` to run the backend executable.

---

### **Task 3.2: GitHub Actions Workflow**

-   **Description**: Create a GitHub Actions workflow to build and push the Docker image to `ghcr.io`.
-   **File to Create**: `.github/workflows/docker-publish.yml`
-   **Details**:
    -   Trigger the workflow on `push` to the `main` branch.
    -   Use the `docker/login-action` to log in to `ghcr.io` using the `GITHUB_TOKEN`. [12, 34]
    -   Use the `docker/build-push-action` to build the image and push it to `ghcr.io/korjavin/countrycounter`. [29, 39]
    -   Tag the image with `latest`.
