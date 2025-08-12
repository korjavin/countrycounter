# Country Counter

A Telegram mini web app for tracking visited countries.  Visualize your travel history on a world map and keep a simple list of visited countries.

## Features

* Telegram integration for user authorization.
* World map visualization of visited countries.
* Simple statistics (number of visited countries).
* Searchable/sortable list of visited countries (hidden by default).
* "Add country" mode for updating the list.
* Data persistence via a JSON file.
* Dockerized deployment.

## Technical Details

* **Frontend:** Pure JavaScript optimized for mobile.
* **Backend:** Golang with standard HTTP router, serving frontend files and user data.
* **Data Storage:** JSON file for storing user ID and visited countries.
* **Deployment:** Single Docker container, built and pushed to `ghcr.io/korjavin/countrycounter` via GitHub Actions.

## Development

1. Clone the repository: `git clone https://github.com/korjavin/countrycounter.git`
2. Build the Docker image: `docker build -t countrycounter .`
3. Run the Docker container: `docker run -p 8080:8080 countrycounter`
4. Access the app: `http://localhost:8080`

## Telegram Bot Integration

1. Create a Telegram bot and obtain its API token.
2. Set the `TELEGRAM_BOT_TOKEN` environment variable in the Docker container.

## Contributing

Contributions are welcome!
