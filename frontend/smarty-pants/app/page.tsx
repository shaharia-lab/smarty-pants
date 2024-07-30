import React from 'react';
import dynamic from 'next/dynamic';
import Navbar from '../components/Navbar';
import Header, {HeaderConfig} from "@/components/Header";

const DynamicDashboard = dynamic(() => import('../components/Dashboard'), {ssr: false});

export default function Home() {
    const headerConfig: HeaderConfig = {
        title: "Dashboard"
    };

    return (
        <div className="min-h-screen bg-gray-50">
            <Navbar/>
            <Header config={headerConfig}/>
            <main className="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
                <DynamicDashboard/>
            </main>
        </div>
    );
}