# Groupchat-Service
Groupchat service for PTSS Support, built with Go and FCM

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
3. Create a configuration file:
   In the root directory, create a `config.yaml` file with the following content as an example:
   ```yaml
   port: "8080"
   database_url: "postgres://user:password@localhost:5432/mydb"
   allowed_origins:
     - "http://localhost:3000"
     - "https://myapp.com"
   jwt_secret: "my-secret-key"
   fcm_server_key: "your-fcm-server-key"
   ```

4. Run the application:
   ```bash
   go run main.go
   ```