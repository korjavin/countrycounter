# Project Improvements

This document outlines the key improvements, bug fixes, and security enhancements made to the Country Counter application.

## Summary of Changes

The project underwent a thorough review to identify and address discrepancies, security vulnerabilities, and areas for improvement. The following is a summary of the changes implemented:

1.  **Enhanced Security**: The API endpoints were secured by implementing validation of Telegram's `initData`. This ensures that all requests to the backend are authenticated and originate from the legitimate Telegram Web App, preventing unauthorized data access.

2.  **Improved Data Integrity**:
    -   A bug was fixed that allowed duplicate countries to be added to a user's list. The backend now checks for duplicates before adding a new country.
    -   The data saving logic was refactored to be atomic, preventing race conditions during concurrent requests.

3.  **Increased Robustness**:
    -   The backend error handling was improved. The server no longer crashes if the `data.json` file is missing or corrupted. Instead, it logs a warning and starts with a clean dataset.
    -   A bug in the backend's map generation feature was fixed where an incorrect file path to the GeoJSON data was used.

4.  **Standardized Data Sources**:
    -   The frontend was updated to use a local GeoJSON file for the world map, consistent with the backend. This removes the dependency on an external data source and ensures the web app and the bot-generated maps are identical.

5.  **Improved Documentation**:
    -   The `readme.md` was updated with a more robust `docker run` command that includes data persistence by default.
    -   A new section on API security was added to the documentation to inform users and developers about the authentication mechanism.

## Detailed List of Issues and Fixes

### 1. Security Vulnerability: Insecure API Endpoints

-   **Issue**: The API endpoints accepted a `userId` directly from the request without validation, allowing any party to potentially read or write data for any user.
-   **Fix**: Implemented a new authentication middleware that validates the `initData` string from Telegram on every API request. The user's identity is now securely verified using a hash-based message authentication code (HMAC) with the bot token as the secret.

### 2. Bug: Duplicate Country Entries

-   **Issue**: The backend did not check if a country was already in a user's list before adding it, leading to potential data duplication.
-   **Fix**: The `addCountry` handler in the backend now checks for existing entries before appending a new country.

### 3. Discrepancy: Inconsistent GeoJSON Data

-   **Issue**: The frontend fetched GeoJSON data from a remote URL, while the backend used a local file. This could lead to inconsistencies.
-   **Fix**: The frontend was modified to use the same local `countries.geo.json` file as the backend. The file was moved to the `frontend` directory to be served statically.

### 4. Bug: Incorrect File Path in Backend

-   **Issue**: The backend function for generating map images (`generateMapImage`) was looking for the GeoJSON file in a hardcoded path that was only correct inside the Docker container, making local development difficult.
-   **Fix**: The file path was corrected, and the `Dockerfile` was updated to match, ensuring consistent behavior in all environments.

### 5. Bug: Fatal Error on Data Loading

-   **Issue**: The server would crash (`log.Fatalf`) if the `data.json` file was missing, unreadable, or corrupted.
-   **Fix**: The `loadData` function was updated to handle these errors gracefully. It now logs a warning and initializes an empty dataset, allowing the application to start.

### 6. Documentation: Incomplete `docker run` Command

-   **Issue**: The `docker run` command in the `readme.md` was missing the `-v` flag for volume mounting, which could lead to data loss for users.
-   **Fix**: The command was updated to include a named volume for data persistence by default, matching the best practice described in the `docker-compose` section.
