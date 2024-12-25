# Groupchat-Service
Groupchat microservice for PTSS Support, built with GO and FCM SDK and FCM Admin SDK.
This service will use a Azure table storage database.

## Dependencies
### Required Dependency:
- **Viper** (`github.com/spf13/viper v1.19.0`): A powerful library for loading configuration from files, environment variables, and more.

## Getting started
1. Clone the repository:
   ```bash
   git clone <https://github.com/PTSS-Support/Groupchat-Service.git>
   cd Groupchat-Service
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Configure environment variables:
   Create a `.env` file in the root directory of the project using the provided `.env.example` file as a reference. Example:
   ```plain text
   # Firebase Configuration
   FIREBASE_CREDENTIAL_FILE=./config/firebase-credentials-dev.json

   # Azure Storage Configuration
   AZURE_STORAGE_ACCOUNT=your_storage_account
   AZURE_STORAGE_KEY=your_storage_key
   AZURE_CONNECTION_STRING=your_connection_string

   # Application Configuration
   APP_ENV=development
   APP_PORT=8080
   ```

4. Run the application:
   ```bash
   go run main.go
   ```