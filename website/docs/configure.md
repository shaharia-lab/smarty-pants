---
sidebar_position: 2
title: Configure
---

# Configure SmartyPants

To get started with SmartyPants to enjoy all of its features, you need to configure it properly.
There are several components that you need to configure to make it work properly.

To make SmaryPants work, you need to configure the following components:

- First you need to configure at least one Embedding Provider.
- Then you need to configure at least one datasource.
- Finally, you need to configure at least one LLM Provider.

Without those components, you can't use the full potential of SmartyPants.

## Configuration Components
### LLM Providers
To be able to ask question to SmartyPants, and be able to get the answer, you need to configure at least one LLM provider.
Please go to http://your-frontend-app.dev/llm-providers and configure your first LLM provider.

To learn more about supported LLM providers and it's configuration, please go to [LLM Providers](/docs/llm-providers) section.

### Embedding Provider
For semantic search capability, you need to configure at least one embedding provider. So, when your datasource will 
discover new documents, it will be able to generate the embeddings for those documents to make it searchable.

You need to go to http://your-frontend-app.dev/embedding-providers and configure your first embedding provider.

To learn more about supported embedding providers and it's configuration, please go to [Embedding Providers](/docs/documentations/embedding-providers) section.


### Datasource Integration

It's time to integrate your datasource with SmartyPants. You need to go to http://your-frontend-app.dev/datasources and configure your first datasource.

To learn more about supported datasource and it's configuration, please go to [Datasources](/docs/documentations/datasources) section.