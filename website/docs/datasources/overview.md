---
sidebar_position: 1
title: Overview
---

# Data Sources

Data sources are the core of the SmartyPants platform. They are responsible for fetching the documents from the configured data sources.

## How it Works?

We need data sources to power our semantic search features. When a data source fetches new documents, it sends those documents to the configured embedding provider to generate embeddings for those documents.
These embeddings are used to make the documents searchable. Later we can use those embeddings to power our semantic search and generative AI features.

## Configuring Data Sources

To configure data sources, you need to enable and configure at least one of the supported data sources.

You can configure data sources by going to `http://your-frontend-app.dev/datasources` page.