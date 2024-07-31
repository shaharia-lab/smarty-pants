<h1 align="center">SmartyPants</h1>
<p align="center">Build and manage AI solutions fast without complicated learning curve about AI. An open source 
initiative by Shaharia Lab OÜ</p>
<p align="center"><a href="https://github.com/shaharia-lab/smarty-pants">shaharia-lab/smarty-pants</a> </p>

<p align="center">
  <a href="https://github.com/shaharia-lab/smarty-pants/actions/workflows/base_branch.yml"><img src="https://github.
com/shaharia-lab/smarty-pants/actions/workflows/base_branch.yml/badge.svg" height="20"/></a>
  <a href="https://sonarcloud.io/summary/new_code?id=shaharia-lab_smarty-pants"><img src="https://sonarcloud.io/api/project_badges/measure?project=shaharia-lab_smarty-pants&metric=reliability_rating" height="20"/></a>
  <a href="https://sonarcloud.io/summary/new_code?id=shaharia-lab_smarty-pants"><img src="https://sonarcloud.io/api/project_badges/measure?project=shaharia-lab_smarty-pants&metric=vulnerabilities" height="20"/></a>
  <a href="https://sonarcloud.io/summary/new_code?id=shaharia-lab_smarty-pants"><img src="https://sonarcloud.io/api/project_badges/measure?project=shaharia-lab_smarty-pants&metric=security_rating" height="20"/></a>
  <a href="https://sonarcloud.io/summary/new_code?id=shaharia-lab_smarty-pants"><img src="https://sonarcloud.io/api/project_badges/measure?project=shaharia-lab_smarty-pants&metric=sqale_rating" height="20"/></a>
  <a href="https://sonarcloud.io/summary/new_code?id=shaharia-lab_smarty-pants"><img src="https://sonarcloud.io/api/project_badges/measure?project=shaharia-lab_smarty-pants&metric=code_smells" height="20"/></a>
  <a href="https://sonarcloud.io/summary/new_code?id=shaharia-lab_smarty-pants"><img src="https://sonarcloud.io/api/project_badges/measure?project=shaharia-lab_smarty-pants&metric=ncloc" height="20"/></a>
  <a href="https://sonarcloud.io/summary/new_code?id=shaharia-lab_smarty-pants"><img src="https://sonarcloud.io/api/project_badges/measure?project=shaharia-lab_smarty-pants&metric=alert_status" height="20"/></a>
  <a href="https://sonarcloud.io/summary/new_code?id=shaharia-lab_smarty-pants"><img src="https://sonarcloud.io/api/project_badges/measure?project=shaharia-lab_smarty-pants&metric=duplicated_lines_density" height="20"/></a>
  <a href="https://sonarcloud.io/summary/new_code?id=shaharia-lab_smarty-pants"><img src="https://sonarcloud.io/api/project_badges/measure?project=shaharia-lab_smarty-pants&metric=bugs" height="20"/></a>
  <a href="https://sonarcloud.io/summary/new_code?id=shaharia-lab_smarty-pants"><img src="https://sonarcloud.io/api/project_badges/measure?project=shaharia-lab_smarty-pants&metric=sqale_index" height="20"/></a>
</p><br/><br/>

<p align="center">
  <a href="https://github.com/shaharia-lab/smarty-pants"><img src="https://github.com/user-attachments/assets/999b1cc2-dbcc-448d-9cfc-d2a77bfdcf6b" height="225"/></a>
</p><br/>

## What is SmartyPants?
**SmartyPants AI** is an intelligent, AI-driven platform that seamlessly integrates multiple data sources, embedding models, and LLM providers. It offers powerful semantic search capabilities and an intuitive chat interface, allowing users to easily configure and interact with various AI models. Whether you're building a smart chatbot or need advanced data processing and querying, SmartyPants provides a flexible, user-friendly solution for your AI-powered applications.

## Why SmartyPants?
We named this project **SmartyPants**. A lighthearted name implying the system is incredibly intelligent, able to handle complex queries with ease.

## Key Features:
- Multi-source data integration and embedding generation
- Configurable LLM and embedding models
- Semantic search functionality
- Built-in chat interface and chatbot creation tools
- Easy-to-use API for seamless integration

Empower your projects with SmartyPants – where AI meets simplicity!

## Installation

- Pre-requisites:
  - [PostgreSQL](https://www.postgresql.org/download/) 13 or higher with [pgvector](https://github.
    com/pgvector/pgvector) extension enabled for [vector search](https://www.elastic.co/what-is/vector-search) capabilities.

### Use as a Docker Image

To use this application with Docker, follow these steps:

#### Pulling the Docker Images

1. For the backend:
   ```
   docker pull ghcr.io/shaharia-lab/smarty-pants-backend:$VERSION
   ```

2. For the frontend:
   ```
   docker pull ghcr.io/shaharia-lab/smarty-pants-frontend:$VERSION
   ```

Replace `$VERSION` with the desired version tag.

#### Running the Backend

To run the backend, you need to set the following environment variables:

- `DB_HOST`: Database host
- `DB_PORT`: Database port
- `DB_USER`: Database user
- `DB_PASS`: Database password
- `DB_NAME`: Database name
- `DB_MIGRATION_PATH`: Database migration path

Run the backend with this command:

```
docker run \
  --name smarty-pants-backend \
  -p 8080:8080 \
  -e DB_HOST=<value> \
  -e DB_PORT=<value> \
  -e DB_USER=<value> \
  -e DB_PASS=<value> \
  -e DB_NAME=<value> \
  -e DB_MIGRATION_PATH=<value> \
  ghcr.io/shaharia-lab/smarty-pants-backend:$VERSION start
```

Replace `<value>` with the appropriate values for your environment.

#### Running the Frontend

To run the frontend, use this command:

```
docker run \
  --name smarty-pants-frontend \
  -e NEXT_PUBLIC_API_BASE_URL="http://localhost:8080" \
  -p 3000:3000 \
  ghcr.io/shaharia-lab/smarty-pants-frontend:$VERSION
```

Make sure to configure any necessary network settings to allow the frontend to communicate with the backend.

## Environment Variables

| Variable                             | Required | Default value             | Description                                                   |
|--------------------------------------|----------|---------------------------|---------------------------------------------------------------|
| `APP_NAME`                           | No       | `"smarty-pants-ai"`       | Name of the application                                       |
| `DB_ENGINE`                          | No       | `"postgres"`              | Database engine to use. Currently it only supports `postgres` |
| `DB_HOST`                            | No       | `"localhost"`             | Database host address                                         |
| `DB_PORT`                            | No       | `5432`                    | Database port number                                          |
| `DB_USER`                            | No       | `"app"`                   | Database user name                                            |
| `DB_PASS`                            | No       | `"pass"`                  | Database password                                             |
| `DB_NAME`                            | No       | `"app"`                   | Database name                                                 |
| `DB_MIGRATION_PATH`                  | No       | `"migrations"`            | Path to database migration files                              |
| `API_PORT`                           | No       | `8080`                    | Port number for the API server                                |
| `API_SERVER_READ_TIMEOUT_IN_SECS`    | No       | `10`                      | API server read timeout in seconds                            |
| `API_SERVER_WRITE_TIMEOUT_IN_SECS`   | No       | `30`                      | API server write timeout in seconds                           |
| `API_SERVER_IDLE_TIMEOUT_IN_SECS`    | No       | `120`                     | API server idle timeout in seconds                            |
| `TRACING_ENABLED`                    | No       | `false`                   | Enable or disable tracing                                     |
| `OTLP_TRACER_HOST`                   | No       | `"localhost"`             | OpenTelemetry Protocol (OTLP) tracer host                     |
| `OTLP_TRACER_PORT`                   | No       | `4317`                    | OTLP tracer port                                              |
| `OTEL_METRICS_ENABLED`               | No       | `false`                   | Enable or disable OpenTelemetry metrics                       |
| `OTEL_METRICS_EXPOSED_PORT`          | No       | `2223`                    | Port to expose OpenTelemetry metrics                          |
| `COLLECTOR_WORKER_COUNT`             | No       | `1`                       | Number of collector workers                                   |
| `GRACEFUL_SHUTDOWN_TIMEOUT_IN_SECS`  | No       | `30`                      | Graceful shutdown timeout in seconds                          |
| `PROCESSOR_WORKER_COUNT`             | No       | `1`                       | Number of processor workers                                   |
| `PROCESSOR_BATCH_SIZE`               | No       | `2`                       | Batch size for the processor                                  |
| `PROCESSOR_INTERVAL_IN_SECS`         | No       | `10`                      | Processor interval in seconds                                 |
| `PROCESSOR_RETRY_ATTEMPTS`           | No       | `3`                       | Number of processor retry attempts                            |
| `PROCESSOR_RETRY_DELAY_IN_SECS`      | No       | `5`                       | Delay between processor retry attempts in seconds             |
| `PROCESSOR_SHUTDOWN_TIMEOUT_IN_SECS` | No       | `10`                      | Processor shutdown timeout in seconds                         |
| `PROCESSOR_REFRESH_INTERVAL_IN_SECS` | No       | `60`                      | Processor refresh interval in seconds                         |



### 🤝 Contributing

Contributions are welcome! Please follow the guidelines outlined in the [CONTRIBUTING](https://github.com/shaharia-lab/smarty-pants/blob/master/CONTRIBUTING.md) file.

### 📝 License

This project is licensed under the MIT License - see the [LICENSE](https://github.com/shaharia-lab/smarty-pants/blob/master/LICENSE) file for details.