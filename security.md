# Security

This document outlines the security measures taken to protect the "Country Counter" Telegram Mini Web App.

## Vulnerability Description

A security vulnerability was identified in the initial version of the application. The backend API endpoints (`/api/countries`) were accepting a `userId` directly from the frontend without proper validation. This meant that a malicious actor could potentially access or modify the country list of any user if they knew the user's Telegram ID.

The vulnerability was present in the following files:
- `frontend/js/app.js`: The frontend was sending the `userId` as a query parameter in GET requests and in the request body of POST requests. The `userId` was obtained from `window.Telegram.WebApp.initDataUnsafe.user.id`, which, as the name suggests, is not secure.
- `backend/main.go`: The backend was accepting the `userId` from the request without any validation, trusting the client to provide the correct information.

## Solution

To address this vulnerability, we have implemented a robust authentication mechanism based on the official Telegram Mini App documentation. The solution involves the following changes:

### Frontend Changes

- The frontend (`frontend/js/app.js`) has been modified to send the `window.Telegram.WebApp.initData` string in the `Authorization` header of all API requests. This string contains the user's data, along with a hash that can be used to verify its authenticity.

### Backend Changes

- The backend (`backend/main.go`) has been updated to include a new authentication middleware for all `/api` routes.
- This middleware extracts the `initData` string from the `Authorization` header.
- It then validates the `initData` by performing an HMAC-SHA-256 validation, as described in the [Telegram documentation](https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app).
- The validation process uses a secret key derived from the bot's token, ensuring that the data has not been tampered with and is from a legitimate Telegram user.
- If the validation is successful, the user's ID is extracted from the validated `initData` and added to the request context. The API handlers then use this validated user ID to process the request.
- If the validation fails, the request is rejected with a `401 Unauthorized` error.

## Security Measures Taken

- **Authentication:** All API requests are now authenticated to ensure that they are coming from a legitimate Telegram user.
- **Authorization:** By using the validated user ID, we ensure that users can only access and modify their own data.
- **Data Integrity:** The HMAC-SHA-256 validation ensures that the data sent from the Mini App has not been tampered with.

These changes effectively close the security vulnerability and ensure that the application is secure.
