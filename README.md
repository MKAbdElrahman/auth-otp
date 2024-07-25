# Authentication and OTP Microservices

This project consists of two microservices: `auth` and `otp`. These microservices handle user authentication and OTP  generation and verification using Twilio's API. The services are built with Go and the Connect framework.

## Microservices Overview

### Auth Microservice

The `auth` microservice handles user authentication, including signup, verification, login, and profile management.

#### Responsibilities
- **User Signup**: Registering a new user with their phone number.
- **User Verification**: Verifying the user's phone number using an OTP.
- **User Login**: Logging in the user using their phone number and OTP.
- **Profile Management**: Retrieving user profile data.

#### Key Components
- **API Handlers**: Define the gRPC and HTTP handlers for the authentication endpoints.
- **Application Layer**: Contains the business logic for user signup, verification, login, and profile retrieval.
- **Domain Layer**: Defines the domain models and interfaces for the repositories and other services.
- **Infrastructure Layer**: Implements the repositories for user, OTP, and activity data storage, typically using a PostgreSQL database.

### OTP Microservice

The `otp` microservice handles sending of OTPs via Twilio's API. 

#### Responsibilities

- **OTP Sending**: Sending OTPs to users via Twilio's API.

#### Key Components
- **API Handlers**: Define the gRPC and HTTP handlers for the OTP endpoints.
- **Application Layer**: Contains the business logic for generating and sending OTPs.
- **Domain Layer**: Defines the domain models and interfaces for the OTP service.
- **Infrastructure Layer**: Implements the integration with Twilio's API for sending OTPs.

## Getting Started

### Prerequisites

- Docker/Docker Compose
- Task (A modern replacement for make) [Install Task](https://taskfile.dev/installation/)
- REST Client VSCode extension

### Setup Instructions

1. **Install Task:**
   Follow the instructions at [Taskfile.dev](https://taskfile.dev/installation/) to install Task on your machine.

2. **Install Dependencies:**
    ```sh
    task deps-install
    ```
    This will install the migrate tool and GoRPCConnect framework tools.

3. **Tidy Go Modules:**
    ```sh
    go mod tidy
    ```

4. **Start Containers:**
    ```sh
    task up
    ```
    This will run the PostgreSQL and RabbitMQ containers.

5. **Run the Auth Microservice:**
    ```sh
    task run-auth
    ```

6. **Run the OTP Microservice:**
    ```sh
    task run-otp
    ```

7. **Interact with Endpoints:**
    Under the `requests` folder, you will find HTTP files for each endpoint you want to interact with using the REST Client extension in VSCode.

