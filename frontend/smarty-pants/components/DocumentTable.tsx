'use client';

import React from 'react';
import Link from 'next/link';
import { Document } from '@/types';

interface DocumentTableProps {
    documents: Document[];
}

const DocumentTable: React.FC<DocumentTableProps> = ({ documents }) => {
    return (
        <div className="mt-8 flow-root">
            <div className="-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
                <div className="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
                    <table className="min-w-full divide-y divide-gray-300">
                        <thead>
                        <tr>
                            <th scope="col" className="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0">Title</th>
                            <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Status</th>
                            <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Created At</th>
                        </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-200">
                        {documents.map((doc) => (
                            <tr key={doc.uuid}>
                                <td className="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-0">
                                    <Link href={`/documents/${doc.uuid}`} className="text-indigo-600 hover:text-indigo-900">
                                        {doc.title}
                                    </Link>
                                </td>
                                <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{doc.status}</td>
                                <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{new Date(doc.created_at).toLocaleString()}</td>
                            </tr>
                        ))}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
};

export default DocumentTable;