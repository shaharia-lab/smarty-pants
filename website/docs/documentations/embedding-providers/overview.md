---
sidebar_position: 1
title: Overview
---

# Embedding Provider

Embedding Providers are the core of the SmartyPants platform. They are responsible for generating embeddings for the 
documents fetched from configured datasources.

## How it Works?

We need Embedding provider to power our semantic search features. When a datasource fetches new documents, it sends 
those documents to the configured embedding provider to generate embeddings for those documents. These embeddings are
used to make the documents searchable.

## Configuring Embedding Providers

To configure Embedding providers, you need to enable and configure at least one of the supported Embedding providers.
Without Embedding provider, you can't use the semantic search features of SmartyPants.

You can configure Embedding providers by going to `http://your-frontend-app.dev/embedding-providers` page.