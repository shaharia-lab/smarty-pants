---
sidebar_position: 1
title: Overview
---

# LLM Provider

LLM (Language Model) Providers are the core of the SmartyPants platform. They are responsible for generating responses to user queries based on the input data and the model they are trained on.

## How it Works?

We need LLM provider to power our generative AI features such as question answering, text generation, chatbot and more.

## RAG (Retrieval Augmented Generation) Pipeline

RAG is a technique that enhances the capabilities of Large Language Models (LLMs) by providing them with relevant context from an external knowledge base. Here's how our RAG pipeline works:

- User Query:
    - The process begins when a user submits a query or question.

- Semantic Search:
    - The user's query is processed by SmartyPant's semantic search engine.
    - This engine searches through all knowledge base to find the most relevant information.

- Retrieval of Top Results:
    - The semantic search engine retrieves the top N most relevant results.
    - These results serve as context for the LLM.

- Context Preparation:
    - The retrieved results are formatted and prepared to be used as context for the LLM.

- LLM Processing:
    - The original user query and the retrieved context are sent to the LLM.
    - The LLM uses this information to generate a response.

- Answer Generation:
    - The LLM produces an answer based on both its pre-trained knowledge and the provided context.

- Response Delivery:
    - The generated answer is returned to the user.

This RAG pipeline allows SmartyPants to provide more accurate, up-to-date, and contextually relevant responses by 
combining the power of LLMs with specific information from our knowledge base.

## Configuring LLM Providers
To configure LLM providers, you need to enable and configure at least one of the supported LLM providers.

Without LLM provider, you can't use the generative AI features of SmartyPants.
