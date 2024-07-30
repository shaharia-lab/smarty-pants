import React from 'react';
import Navbar from '../../components/Navbar';
import DocumentClient from '../../components/DocumentClient';

export default function DocumentHome() {
    return (
        <div className="min-h-full">
            <Navbar/>
            <header className="bg-white shadow">
                <div className="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
                    <h1 className="text-3xl font-bold tracking-tight text-gray-900">Document Dashboard</h1>
                </div>
            </header>
            <main>
                <div className="mx-auto max-w-7xl py-6 sm:px-6 lg:px-8">
                    <div className="px-4 py-6 sm:px-0">
                        <div className="overflow-hidden rounded-lg border-4 border-dashed border-gray-200">
                            <DocumentClient/>
                        </div>
                    </div>
                </div>
            </main>
        </div>
    );
}