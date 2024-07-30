'use client';

import React, {useState} from 'react';
import Navbar from '../../components/Navbar';
import ChatInterface from '../../components/ChatInterface';
import ChatHistories from '../../components/ChatHistories';
import Header, {HeaderConfig} from "@/components/Header";

export default function AskPage() {
    const [selectedInteractionId, setSelectedInteractionId] = useState<string | null>(null);

    const handleSelectInteraction = (uuid: string) => {
        setSelectedInteractionId(uuid);
    };

    const headerConfig: HeaderConfig = {
        title: "Ask Smart Brain"
    };

    return (
        <div className="min-h-screen bg-gray-50">
            <Navbar/>
            <Header config={headerConfig}/>
            <main className="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
                <div className="flex space-x-6">
                    <div className="w-1/3">
                        <ChatHistories onSelectInteraction={handleSelectInteraction}/>
                    </div>
                    <div className="w-2/3">
                        <ChatInterface interactionId={selectedInteractionId}/>
                    </div>
                </div>
            </main>
        </div>
    );
}