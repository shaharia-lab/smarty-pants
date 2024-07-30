import React, {useState} from 'react';
import {Document} from '@/types';

interface DocumentDetailsClientProps {
    document: Document;
}

const DocumentDetailsClient: React.FC<DocumentDetailsClientProps> = ({document}) => {
    const [expandedEmbeddings, setExpandedEmbeddings] = useState<{ [key: number]: boolean }>({});

    const toggleEmbedding = (index: number) => {
        setExpandedEmbeddings(prev => ({...prev, [index]: !prev[index]}));
    };

    const formatEmbedding = (embedding: number[]) => {
        return embedding.map(value => value.toFixed(6)).join(', ');
    };

    const renderEmbedding = (embedding: number[], index: number) => {
        const formattedEmbedding = formatEmbedding(embedding);
        const lines = formattedEmbedding.split(', ');
        const previewLines = lines.slice(0, 3);
        const remainingLines = lines.slice(3);
        const isExpanded = expandedEmbeddings[index];

        return (
            <div>
        <pre className="mt-2 text-sm text-gray-900 bg-gray-100 p-2 rounded overflow-x-auto">
          <code>
            {previewLines.join(', ')}
              {remainingLines.length > 0 && (
                  <>
                      {isExpanded ? (
                          <>
                              ,<br/>
                              {remainingLines.join(',\n')}
                          </>
                      ) : ', ...'}
                  </>
              )}
          </code>
        </pre>
                {remainingLines.length > 0 && (
                    <button
                        onClick={() => toggleEmbedding(index)}
                        className="mt-2 text-sm text-blue-600 hover:text-blue-800"
                    >
                        {isExpanded ? 'Show less' : `Show ${remainingLines.length} more`}
                    </button>
                )}
            </div>
        );
    };

    return (
        <div className="bg-white shadow overflow-hidden sm:rounded-lg">
            <div className="px-4 py-5 sm:px-6">
                <h3 className="text-lg leading-6 font-medium text-gray-900">Document Overview</h3>
            </div>
            <div className="border-t border-gray-200 px-4 py-5 sm:p-0">
                <dl className="sm:divide-y sm:divide-gray-200">
                    <div className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                        <dt className="text-sm font-medium text-gray-500">Title</dt>
                        <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{document.title}</dd>
                    </div>
                    <div className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                        <dt className="text-sm font-medium text-gray-500">Body</dt>
                        <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{document.body}</dd>
                    </div>
                    <div className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                        <dt className="text-sm font-medium text-gray-500">Status</dt>
                        <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{document.status}</dd>
                    </div>
                    <div className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                        <dt className="text-sm font-medium text-gray-500">Created At</dt>
                        <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{new Date(document.created_at).toLocaleString()}</dd>
                    </div>
                    <div className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                        <dt className="text-sm font-medium text-gray-500">Updated At</dt>
                        <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{new Date(document.updated_at).toLocaleString()}</dd>
                    </div>
                </dl>
            </div>

            <div className="px-4 py-5 sm:px-6">
                <h3 className="text-lg leading-6 font-medium text-gray-900">Embedding</h3>
            </div>
            <div className="border-t border-gray-200 px-4 py-5 sm:p-0">
                {document.embedding.embedding ? (
                    <div className="py-4 sm:py-5 sm:px-6">
                        {document.embedding.embedding.map((item, index) => (
                            <div key={index} className="mb-4">
                                <p className="text-sm font-medium text-gray-500">Content: {item.content}</p>
                                {renderEmbedding(item.embedding, index)}
                            </div>
                        ))}
                    </div>
                ) : (
                    <p className="py-4 sm:py-5 sm:px-6 text-sm text-gray-500">No embedding available</p>
                )}
            </div>

            <div className="px-4 py-5 sm:px-6">
                <h3 className="text-lg leading-6 font-medium text-gray-900">Metadata</h3>
            </div>
            <div className="border-t border-gray-200 px-4 py-5 sm:p-0">
                <dl className="sm:divide-y sm:divide-gray-200">
                    {document.metadata.map((item, index) => (
                        <div key={index} className="py-4 sm:py-5 sm:grid sm:grid-cols-3 sm:gap-4 sm:px-6">
                            <dt className="text-sm font-medium text-gray-500">{item.key}</dt>
                            <dd className="mt-1 text-sm text-gray-900 sm:mt-0 sm:col-span-2">{item.value}</dd>
                        </div>
                    ))}
                </dl>
            </div>
        </div>
    );
};

export default DocumentDetailsClient;