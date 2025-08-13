# Country Counter

[![Logo](logo.png)](https://github.com/korjavin/countrycounter)

**TL;DR: A fun, map-based Telegram Mini Web App to visualize your travel history. See the world you've explored, one country at a time!**

Country Counter is a Telegram Mini Web App designed for travel enthusiasts to track the countries they have visited. It provides a simple and visually appealing way to see your travel progress on a world map, optimized for a seamless mobile experience.

The application is built with a lightweight Go backend and a pure JavaScript frontend, all packaged within a single Docker container for easy deployment.

## Features

-   **Telegram Integration**: Securely authenticates users via their Telegram account.
-   **Interactive World Map**: Visualizes visited countries with custom colors. The map is zoomable and pannable.
-   **Travel Statistics**: Displays a running count of visited countries.
-   **Visited Countries List**: A collapsible list to review your visited countries.
-   **Add Countries Mode**: An intuitive interface to search and add new countries to your list.
-   **Telegram Bot Commands**: Use `/map` to get a PNG image of your map and `/list` to get a text list of your visited countries, right in your chat.
-   **Mobile-First Design**: A clean and responsive UI optimized for mobile devices.
-   **Simple & Efficient Backend**: A lightweight Go server handles data persistence.
-   **Containerized**: Packaged as a single Docker container for easy deployment and scaling.
-   **Automated CI/CD**: GitHub Actions automatically build and publish the Docker image to GitHub Container Registry.

## Data Collection and Privacy

We respect your privacy. Here's a transparent look at the data we collect and how we use it:

-   **What we store**: We only store your Telegram User ID and the list of countries you have visited. We do not store your name, username, or any other personal information.
-   **How we use it**: Your User ID is used as a key to retrieve your list of visited countries. The list of countries is used to generate your personalized map and statistics.
-   **Data Storage**: All data is stored in a `data.json` file on the server where the application is hosted. You have full control over this data.
-   **No Third-Party Tracking**: The application does not use any third-party analytics or tracking services.

## Tech Stack

-   **Frontend**:
    -   Pure JavaScript (ES6+)
    -   [Leaflet.js](https://leafletjs.com/): An open-source JavaScript library for mobile-friendly interactive maps.
    -   HTML5 & CSS3
-   **Backend**:
    -   **Go (Golang)**: For building a simple, efficient, and reliable web server.
    -   **Standard `net/http` package**: For routing and serving files.
    -   **JSON**: For simple, file-based data storage.
-   **Deployment**:
    -   **Docker**: To containerize the application.
    -   **GitHub Actions**: For continuous integration and deployment.
    -   **GitHub Container Registry (ghcr.io)**: For hosting the Docker image.

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

The application stores user data in a JSON file located at `/app/backend/data.json` inside the container. To ensure that this data is not lost when the container is stopped or removed, you should mount a Docker volume to this path.

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

This setup will create a named volume `countrycounter-data` on your host machine and mount it to `/app/backend` in the container. All data saved by the application will be stored in this volume, ensuring it persists across container restarts.

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
