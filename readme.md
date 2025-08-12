# Country Counter

Country Counter is a Telegram Mini Web App designed for travel enthusiasts to track the countries they have visited. It provides a simple and visually appealing way to see your travel progress on a world map, optimized for a seamless mobile experience.

The application is built with a lightweight Go backend and a pure JavaScript frontend, all packaged within a single Docker container for easy deployment.

## Features

-   **Telegram Integration**: Securely authenticates users via their Telegram account.
-   **Interactive World Map**: Visualizes visited countries with custom colors.
-   **Travel Statistics**: Displays a running count of visited countries.
-   **Visited Countries List**: A collapsible list to review your visited countries.
-   **Add Countries Mode**: An intuitive interface to search and add new countries to your list.
-   **Mobile-First Design**: A clean and responsive UI optimized for mobile devices.
-   **Simple & Efficient Backend**: A lightweight Go server handles data persistence.
-   **Containerized**: Packaged as a single Docker container for easy deployment and scaling.
-   **Automated CI/CD**: GitHub Actions automatically build and publish the Docker image to GitHub Container Registry.

## Tech Stack

-   **Frontend**:
    -   Pure JavaScript (ES6+)
    -   [Leaflet.js](https://leafletjs.com/): An open-source JavaScript library for mobile-friendly interactive maps. [1]
    -   HTML5 & CSS3
-   **Backend**:
    -   **Go (Golang)**: For building a simple, efficient, and reliable web server.
    -   **Standard `net/http` package**: For routing and serving files.
    -   **JSON**: For simple, file-based data storage.
-   **Deployment**:
    -   **Docker**: To containerize the application. [5, 6]
    -   **GitHub Actions**: For continuous integration and deployment. [11, 12]
    -   **GitHub Container Registry (ghcr.io)**: For hosting the Docker image.

## Project Structure
/
├── .github/
│ └── workflows/
│ └── docker-publish.yml
├── backend/
│ ├── main.go
│ └── data.json
├── frontend/
│ ├── index.html
│ ├── css/
│ │ └── style.css
│ └── js/
│ └── app.js
├── Dockerfile
└── .dockerignore

## Getting Started

To run this project locally, you will need Docker installed on your machine.

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/korjavin/countrycounter.git
    cd countrycounter
    ```

2.  **Set up your Telegram Bot:**
    -   Create a new bot by talking to the [@BotFather](https://t.me/BotFather) on Telegram. [7, 19]
    -   You will receive a token for your bot. This will be needed to authenticate requests.

3.  **Build and run the Docker container:**
    ```sh
    docker build -t ghcr.io/korjavin/countrycounter:latest .
    docker run -p 8080:8080 ghcr.io/korjavin/countrycounter:latest
    ```

4.  **Access the application:**
    -   The application will be running and accessible through your Telegram bot's Mini App interface.

## GitHub Actions

The project is configured with a GitHub Actions workflow in `.github/workflows/docker-publish.yml`. This workflow triggers on every push to the `main` branch. It performs the following steps:
1.  Logs in to the GitHub Container Registry (ghcr.io). [12]
2.  Builds the Docker image.
3.  Pushes the image to `ghcr.io/korjavin/countrycounter`. [11, 13]
