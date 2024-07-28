import OpenAILLMProviderForm from "@/components/OpenAILLMProviderForm";

export default function EditOpenAIEmbeddingProviderPage({params}: { params: { providerId: string } }) {
    return <OpenAILLMProviderForm providerId={params.providerId}/>;
}