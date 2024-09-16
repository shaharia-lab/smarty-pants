import React from 'react';
import { Interaction } from '@/types/api';
import { truncateMessage } from "@/utils/common";
import { useChatHistories } from "@/hooks/useChatHistories";

interface ChatHistoriesProps {
    onSelectInteraction: (uuid: string) => void;
}

const ChatHistories: React.FC<ChatHistoriesProps> = ({ onSelectInteraction }) => {
    const { histories, isLoading } = useChatHistories();

    return (
        <div className="bg-white shadow-md rounded-lg overflow-hidden">
            <h2 className="text-xl font-semibold p-4 border-b">Chat Histories</h2>
            {isLoading ? (
                <div className="p-4">Loading...</div>
            ) : (
                <ChatHistoryList
                    histories={histories}
                    onSelectInteraction={onSelectInteraction}
                />
            )}
        </div>
    );
};

interface ChatHistoryListProps {
    histories: Interaction[] | null;
    onSelectInteraction: (uuid: string) => void;
}

const ChatHistoryList: React.FC<ChatHistoryListProps> = ({ histories, onSelectInteraction }) => {
    if (!histories || histories.length === 0) {
        return <p className="p-4 text-gray-500">No chat histories available.</p>;
    }

    return (
        <ul className="divide-y divide-gray-200">
            {histories.map((history) => (
                <ChatHistoryItem
                    key={history.uuid}
                    history={history}
                    onSelect={onSelectInteraction}
                />
            ))}
        </ul>
    );
};

interface ChatHistoryItemProps {
    history: Interaction;
    onSelect: (uuid: string) => void;
}

const ChatHistoryItem: React.FC<ChatHistoryItemProps> = ({ history, onSelect }) => (
    <li
        className="p-4 hover:bg-gray-50 cursor-pointer"
        onClick={() => onSelect(history.uuid)}
    >
        <h3 className="text-lg font-medium text-gray-900">
            {truncateMessage(getFirstUserMessage(history), 100)}
        </h3>
    </li>
);

const getFirstUserMessage = (interaction: Interaction): string => {
    if (!interaction.conversations || interaction.conversations.length === 0) {
        return interaction.query;
    }
    const userMessage = interaction.conversations.find(msg => msg.role === 'user');
    return userMessage ? userMessage.text : interaction.query;
};

export default ChatHistories;