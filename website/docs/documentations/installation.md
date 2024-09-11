---
sidebar_position: 1
title: Installation
---

# Getting Started

SmartyPants has both backend and frontend components. We have made it easy for everyone to get started with SmartyPants.

## Installation

### Prerequisites

Before installing SmartyPants, ensure you have a PostgreSQL database with the pgvector extension enabled. You can use the following Docker command to set up a compatible database:

```bash
docker run -d \
  --name smarty-pants-postgresql \
  -e POSTGRES_DB=<your_db_name> \
  -e POSTGRES_USER=<your_db_user> \
  -e POSTGRES_PASSWORD=<your_db_password> \
  -p <your_db_port>:5432 \
  ankane/pgvector
```

### Method 1: Using Docker

Before proceeding, check the [Releases page](https://github.com/shaharia-lab/smarty-pants/releases) for the latest version number. Replace `<version>` in the following commands with the latest release version.

#### Backend Installation

1. Pull the backend Docker image:

```bash
docker pull ghcr.io/shaharia-lab/smarty-pants-backend:<version>
```

2. Run the backend container:

```bash
docker run -d \
  --name smarty-pants-backend \
  -p 8080:8080 \
  -p 2223:2223 \
  -e DB_HOST=<your_db_host> \
  -e DB_PORT=<your_db_port> \
  -e DB_USER=<your_db_user> \
  -e DB_PASS=<your_db_password> \
  -e DB_NAME=<your_db_name> \
  ghcr.io/shaharia-lab/smarty-pants-backend:<version>
```

Replace `<your_db_host>`, `<your_db_port>`, `<your_db_user>`, `<your_db_password>`, and `<your_db_name>` with your actual database connection details.

#### Frontend Installation

1. Pull the frontend Docker image:

```bash
docker pull ghcr.io/shaharia-lab/smarty-pants-frontend:<version>
```

2. Run the frontend container:

```bash
docker run -d \
  --name smarty-pants-frontend \
  -p 3000:3000 \
  ghcr.io/shaharia-lab/smarty-pants-frontend:<version>
```

### Method 2: Using Docker Compose

1. Create a file named `docker-compose.yml` in your project directory with the following content:

```yaml
version: '3.8'

services:
  database:
    image: ankane/pgvector
    environment:
      POSTGRES_DB: app
      POSTGRES_USER: app
      POSTGRES_PASSWORD: pass
    ports:
      - "5432:5432"

  backend:
    image: ghcr.io/shaharia-lab/smarty-pants-backend:<version>
    ports:
      - "8080:8080"
      - "2223:2223"
    environment:
      DB_HOST: database
      DB_PORT: 5432
      DB_USER: app
      DB_PASS: pass
      DB_NAME: app
    depends_on:
      - database

  frontend:
    image: ghcr.io/shaharia-lab/smarty-pants-frontend:<version>
    ports:
      - "3000:3000"
    depends_on:
      - backend
```

Replace `<version>` with the latest release version from the [Releases page](https://github.com/shaharia-lab/smarty-pants/releases).

2. Run the following command in the same directory as your `docker-compose.yml` file:

```bash
docker-compose up -d
```

This command will start all the services defined in the `docker-compose.yml` file: the PostgreSQL database, the backend, and the frontend.

## Accessing the Application

After installation, you can access:

- The frontend application at `http://localhost:3000`
- The backend API at `http://localhost:8080`

## Verify Installation

To ensure that both the frontend and backend components are working correctly, you can perform the following checks:

### Frontend Verification

1. Open a web browser and navigate to:
   ```
   http://localhost:3000/auth
   ```
2. You should see the login page of the SmartyPants application.

### Backend Verification

1. To verify the backend, you can use a web browser or a tool like curl to send a request to the following endpoint:
   ```
   http://localhost:8080/system/ping
   ```
2. The response should be:
   ```
   pong
   ```

If you can see the login page and receive the "pong" response, it means both your frontend and backend are installed and running correctly.

## Troubleshooting

If you encounter any issues during installation or while running the application, please check the following:

1. Ensure all required ports (5432, 8080, 2223, 3000) are available and not being used by other applications.
2. Verify that your Docker installation is up-to-date and running correctly.
3. Check the logs of each container for any error messages:
   ```bash
   docker logs smarty-pants-backend
   docker logs smarty-pants-frontend
   ```

For further assistance:

- Check if your issue has already been reported by visiting our GitHub Issues page from [here](https://github.com/shaharia-lab/smarty-pants/issues?q=is%3Aissue+is%3Aopen+label%3Ainstallation)

- If you can't find a solution to your problem, please submit a new issue on our GitHub repository from [here](https://github.com/shaharia-lab/smarty-pants/issues)
  When submitting a new issue, please provide detailed information about your setup, the steps to reproduce the problem, and any error messages you've encountered.

We appreciate your contributions to improving SmartyPants!