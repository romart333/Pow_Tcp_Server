# Word of Wisdom TCP Server

## Task Description

Design and implement a **“Word of Wisdom” TCP server**.

### Requirements:
- **TCP server** should be protected from **DDOS attacks** using **Proof of Work** (PoW).
  - The **PoW** algorithm will be used to prevent excessive requests.
  - The **challenge-response protocol** should be employed to enforce PoW.
- The **PoW algorithm** should be chosen and explained in detail. It should be a secure and efficient method to prevent abuse of the server.
- After **Proof of Work** verification, the server should send a **quote** from the “Word of Wisdom” book or any other collection of quotes. The quote should be selected randomly from the collection.
- **Docker files** must be provided for both the **server** and the **client** that solves the **PoW challenge**. This will allow easy deployment and execution in a containerized environment.

## Choosing a PoW Algorithm

We use **Hashcash** as the PoW algorithm. It is a widely used and simple approach where the server challenges the client to find a nonce that satisfies a computational puzzle. This is computationally expensive, preventing excessive requests and mitigating DDoS attacks. The difficulty level determines how challenging the PoW task is.

## Setup

### Prerequisites

- **Go 1.24** or later.
- **Docker** for building and running the server and client containers.
- **Docker Compose** for managing multi-container applications.

### Environment Variables

The project relies on environment variables to configure certain parameters. However, the configuration is primarily loaded from YAML files (`client.yml` and `server.yml`). To use these configurations in local development, ensure your `.env` file or environment variables are properly mapped to the keys in the YAML files.

#### Example of `client.yml`:
```yaml
client_configuration:
  server_address: "localhost:8080"  # Target server address
  pow_timeout: "30s"                # Timeout for POW calculation
  dial_timeout: "3s"                # Connection establishment timeout
  max_retries: 3                    # Maximum number of retry attempts
  base_retry_delay: "1s"            # Initial retry delay
  max_retry_delay: "10s"            # Maximum retry delay
```
#### Example of `server.yml`:
```yaml
 server_configuration:
  address: ":8080"                  # TCP listen address
  pow_difficulty: 3                 # PoW difficulty level (higher = more difficult)
  pow_calc_timeout: "25s"           # Timeout for PoW calculation
  read_timeout: "5s"                # Socket read timeout
  write_timeout: "5s"               # Socket write timeout
  accept_timeout: "500ms"           # New connection accept timeout
  shutdown_timeout: "15s"           # Graceful shutdown timeout
  max_connections: 1000             # Max concurrent connections
```

For local development, you can create an .env file or set environment variables based on these values. If you use Docker Compose, ensure that the environment variables are passed correctly to the containers.

#### Example .env file:
```bash
# Client config
CLIENT_SERVER_ADDRESS="localhost:8080"
CLIENT_POW_TIMEOUT="30s"
CLIENT_DIAL_TIMEOUT="3s"
CLIENT_MAX_RETRIES=3
CLIENT_BASE_RETRY_DELAY="1s"
CLIENT_MAX_RETRY_DELAY="10s"

# Server config
SERVER_ADDRESS=":8080"
POW_DIFFICULTY=3
POW_CALC_TIMEOUT="25s"
SERVER_READ_TIMEOUT="5s"
SERVER_WRITE_TIMEOUT="5s"
SERVER_ACCEPT_TIMEOUT="500ms"
SERVER_SHUTDOWN_TIMEOUT="15s"
SERVER_MAX_CONNECTIONS=1000
```

### Available Make Commands

The Makefile supports various commands to manage your development workflow.

#### Linting and Testing

1. **Install Lint Tool**:
   Installs `golangci-lint` for running lint checks.

   ```bash
   make install-lint
   ```

2. **Run Lint Checks**:
   Runs `golangci-lint` to check the Go code for style, bugs, and best practices.

   ```bash
   make lint
   ```

3. **Go Vet**:
   Runs `go vet` to report any suspicious constructs in the code.

   ```bash
   make vet
   ```

4. **Run Tests**:
   Runs all the tests in the project using `go test`.

   ```bash
   make test
   ```

#### Building the Project

1. **Build the Server**:
   Builds the server binary and places it in the `bin/` directory.

   ```bash
   make build-server
   ```

2. **Build the Client**:
   Builds the client binary and places it in the `bin/` directory.

   ```bash
   make build-client
   ```

#### Running the Project

1. **Run the Server**:
   Runs the server locally using environment variables specified in the `.env` file.

   ```bash
   make run-server
   ```

2. **Run the Client**:
   Runs the client locally using environment variables specified in the `.env` file.

   ```bash
   make run-client
   ```

#### Docker Commands

1. **Build Docker Image for Server**:
   Builds a Docker image for the server using the Dockerfile in `Dockerfile.server`.

   ```bash
   make docker-build-server
   ```

2. **Build Docker Image for Client**:
   Builds a Docker image for the client using the Dockerfile in `Dockerfile.client`.

   ```bash
   make docker-build-client
   ```

3. **Build Docker Images for Both Server and Client**:
   Builds Docker images for both the server and the client.

   ```bash
   make docker-build-all
   ```