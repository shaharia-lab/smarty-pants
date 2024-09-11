
export const availableEmbeddingProviders = [
    {
        id: 'openai',
        name: 'OpenAI',
        imageUrl: 'https://static-00.iconduck.com/assets.00/openai-icon-2021x2048-4rpe5x7n.png',
        description: 'Use OpenAI\'s powerful language models for embeddings.',
        configurationUrl: '/embedding-providers/openai/add',
    },
    {
        id: 'huggingface',
        name: 'Hugging Face',
        imageUrl: 'https://huggingface.co/datasets/huggingface/brand-assets/resolve/main/hf-logo.svg',
        description: 'Leverage Hugging Face\'s extensive model hub for embeddings.',
        configurationUrl: '/embedding-providers/add?type=huggingface',
    },
    // Add more embedding providers as needed
];
