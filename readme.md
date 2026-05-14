# Country Counter

[![Logo](logo.png)](https://github.com/korjavin/countrycounter)

**TL;DR: A fun, map-based Telegram Mini Web App to visualize your travel history. See the world you've explored, one country at a time!**

Country Counter is a Telegram Mini Web App designed for travel enthusiasts to track the countries they have visited. It provides a simple and visually appealing way to see your travel progress on a world map, optimized for a seamless mobile experience.

The application is built with a lightweight Go backend and a pure JavaScript frontend, all packaged within a single Docker container for easy deployment.

## Features

-   **Telegram Integration**: Securely authenticates users via their Telegram account.
-   **Monochrome World Map**: Display-only Equal-Earth SVG map (d3-geo + topojson) that fills in your visited countries against a two-color paper / oilcloth backdrop.
-   **Travel Statistics**: Displays a running count of visited countries.
-   **Visited Countries List**: A collapsible list to review your visited countries.
-   **Add Countries Mode**: An intuitive interface to search and add new countries to your list.
-   **Location-Based Country Detection**: Simply send your location in Telegram, and the bot automatically detects and adds the country to your visited list! Perfect for adding countries as you travel in real-time.
-   **Telegram Bot Commands**: 
    -   `/map` - Get a PNG image of your map right in your chat
    -   `/list` - Get a text list of your visited countries
    -   `/suggest` - Get personalized country recommendations based on your travel history
    -   **Send Location** - Share any location to automatically detect and add the country
-   **Mobile-First Design**: A clean and responsive UI optimized for mobile devices.
-   **Simple & Efficient Backend**: A lightweight Go server handles data persistence.
-   **Containerized**: Packaged as a single Docker container for easy deployment and scaling.
-   **Automated CI/CD**: GitHub Actions automatically build and publish the Docker image to GitHub Container Registry.

## How to Use

### Adding Countries via the Web App

1. Open the Telegram Mini App from your bot
2. Click the "+" button to enter add mode
3. Search for countries and click to add them to your list
4. Your map will update automatically to fill visited countries against the two-color (paper / oilcloth) world map

### Adding Countries via Location (NEW!)

The easiest way to track your travels is to send your location directly in Telegram:

1. Open your chat with the bot
2. Click the attachment icon (📎) and select "Location"
3. Choose "Send My Current Location" or select any location on the map
4. The bot will automatically:
   - Detect which country the location is in
   - Check if it's already in your list
   - Add it if it's new (or let you know if you've already added it)
   - Show you how many countries you've visited so far

**Pro tip**: You can send locations from photos, saved places, or manually select any point on the map. This makes it easy to add countries from past trips by sharing locations from your photo gallery!

### Using Bot Commands

- `/map` - Generate and receive a beautiful PNG map of all countries you've visited
- `/list` - Get a formatted text list of all your visited countries
- `/suggest` - Get personalized recommendations for your next destination based on your travel history

## Data Collection and Privacy

We respect your privacy. Here's a transparent look at the data we collect and how we use it:

-   **What we store**: We only store your Telegram User ID and the list of countries you have visited. We do not store your name, username, or any other personal information.
-   **Location data**: When you send a location, we use the coordinates only to determine the country name. **We do not store your GPS coordinates, specific locations, addresses, or any location history.** Only the country name is saved.
-   **How we use it**: Your User ID is used as a key to retrieve your list of visited countries. The list of countries is used to generate your personalized map and statistics.
-   **Data Storage**: All data is stored in a SQLite database file (`data.db`) on the server where the application is hosted. You have full control over this data. (Legacy: instances upgrading from a previous version may still have a `data.json` file — on first start the backend auto-imports it once and then ignores it.)
-   **No Third-Party Tracking**: The application does not use any third-party analytics or tracking services.
-   **Offline Processing**: Location-to-country conversion happens entirely on your server without sending data to external geocoding services.

## Tech Stack

-   **Frontend**:
    -   Pure JavaScript (ES6+) with a monospace, two-color (paper / oilcloth) treatment that matches the [Claude Design](https://api.anthropic.com/v1/design/h/NLFBTX80T4dYFOn0kpkwlw) handoff bundle vendored at `docs/design/`.
    -   [d3-geo](https://github.com/d3/d3-geo) + [topojson-client](https://github.com/topojson/topojson-client): Equal-Earth SVG world map rendered inline, no tile server.
    -   [JetBrains Mono](https://www.jetbrains.com/lp/mono/) (variable, self-hosted under `frontend/fonts/`) as the single typeface.
    -   HTML5 & CSS3 with design tokens in `frontend/css/monospace.css`.
-   **Backend**:
    -   **Go (Golang)**: For building a simple, efficient, and reliable web server.
    -   **Standard `net/http` package**: For routing and serving files.
    -   **[revgeo](https://github.com/filipkroca/revgeo)**: Fast offline reverse geocoding library for converting GPS coordinates to country codes.
    -   **[Telegram Bot API](https://github.com/go-telegram-bot-api/telegram-bot-api)**: Official Go wrapper for the Telegram Bot API.
    -   **SQLite** (via [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite), a pure-Go driver): Embedded database for durable storage of visited countries.
    -   **[pressly/goose](https://github.com/pressly/goose)**: Embedded SQL migrations applied automatically on startup.
-   **Deployment**:
    -   **Docker**: To containerize the application.
    -   **GitHub Actions**: For continuous integration and deployment.
    -   **GitHub Container Registry (ghcr.io)**: For hosting the Docker image.

## How Location Detection Works

The location-based country detection feature uses offline reverse geocoding, which means:

-   **No external API calls**: The bot determines countries entirely on your server using embedded GeoJSON data
-   **Fast response**: Country detection happens in milliseconds
-   **Privacy-friendly**: Your location coordinates are never sent to third parties
-   **Reliable**: Works even without internet connectivity (after initial setup)

**Technical Details:**
1. When you send a location, Telegram provides latitude and longitude coordinates
2. The `revgeo` library uses pre-loaded GeoJSON polygon data to determine which country contains those coordinates
3. The library returns an ISO 3166-1 alpha-3 country code (e.g., "USA", "FRA", "JPN")
4. The bot maps this code to the full country name used in your map data
5. The country is added to your list if it's not already there

This approach is more privacy-conscious and faster than using external geocoding APIs like Google Maps or OpenStreetMap.

## Getting Started

To run this project, you will need Docker installed on your machine.

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/korjavin/countrycounter.git
    cd countrycounter
    ```

2.  **Set up your Telegram Bot:**
    -   Create a new bot by talking to the [@BotFather](https://t.me/BotFather) on Telegram.
    -   You will receive a token for your bot. This will be needed for the bot commands and for the web app to authenticate users.
    -   Set the `TELEGRAM_BOT_TOKEN` environment variable.

3.  **Build and run the Docker container:**
    ```sh
    docker build -t ghcr.io/korjavin/countrycounter:latest .
    docker run -p 8080:8080 -e TELEGRAM_BOT_TOKEN="YOUR_BOT_TOKEN" ghcr.io/korjavin/countrycounter:latest
    ```

4.  **Access the application:**
    -   The application will be running and accessible through your Telegram bot's Mini App interface.

## Data Persistence

The application stores user data in a SQLite database at `/app/backend/data.db` inside the container (configurable via the `DB_PATH` environment variable). To ensure data is not lost when the container is stopped or removed, mount a Docker volume to `/app/backend`.

Schema is managed by `pressly/goose` migrations embedded in the binary and applied automatically on startup.

**Auto-import from legacy `data.json`**: If on startup the database is empty *and* a `data.json` file exists at `/app/backend/data.json`, the backend reads it once and imports all `(user, country)` pairs into the new schema, logging `Auto-imported N rows from data.json`. On subsequent restarts — or if `data.json` is absent — the import step is skipped silently. This makes the upgrade path from older versions a no-op: leave the old `data.json` mounted on the first start and the new image will migrate it for you.

You can use a `docker-compose.yml` file to manage the container and its volume easily.

1.  **Create a `docker-compose.yml` file:**
    ```yml
    version: '3.8'
    services:
      countrycounter:
        image: ghcr.io/korjavin/countrycounter:latest
        ports:
          - "8080:8080"
        environment:
          - TELEGRAM_BOT_TOKEN="YOUR_BOT_TOKEN"
          - DB_PATH=/app/backend/data.db
        volumes:
          - countrycounter-data:/app/backend
        restart: unless-stopped

    volumes:
      countrycounter-data:
    ```

2.  **Run the container using Docker Compose:**
    ```sh
    docker-compose up -d
    ```

This setup will create a named volume `countrycounter-data` on your host machine and mount it to `/app/backend` in the container. The SQLite database file and any legacy `data.json` will be stored in this volume, ensuring data persists across container restarts.

## Running with Portainer

You can easily deploy this application using Portainer's "Stacks" feature.

1.  In Portainer, go to **Stacks** and click **Add stack**.
2.  Give the stack a name (e.g., `country-counter`).
3.  In the **Web editor**, paste the following `docker-compose` configuration. **Remember to replace `YOUR_BOT_TOKEN` with your actual Telegram bot token.**

```yml
version: '3.8'

services:
  countrycounter:
    image: ghcr.io/korjavin/countrycounter:latest
    ports:
      - "8080:8080"
    environment:
      - TELEGRAM_BOT_TOKEN="YOUR_BOT_TOKEN"
      - DB_PATH=/app/backend/data.db
    volumes:
      - countrycounter-data:/app/backend
    restart: unless-stopped

volumes:
  countrycounter-data:
```

4.  Click **Deploy the stack**.

Portainer will pull the image and deploy the container with the specified configuration. Your application will be running and accessible.

## License

This project is licensed under the **MIT License**.

## Contributing

Contributions are welcome! If you have any ideas, suggestions, or bug reports, please open an issue or submit a pull request. We appreciate any help to make this project better.
