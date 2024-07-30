import OpenAIEmbeddingProviderForm from '../../../../components/OpenAIEmbeddingProviderForm';

export default function EditOpenAIEmbeddingProviderPage({params}: { params: Readonly<{ providerId: string }> }) {
    return <OpenAIEmbeddingProviderForm providerId={params.providerId}/>;
}