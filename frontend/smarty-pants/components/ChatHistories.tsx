import React, {useEffect, useState} from 'react';
import {createApiService} from "@/services/apiService";
import AuthService from "@/services/authService";
import axios from "axios";
import {Interaction} from "@/types/api";
import {truncateMessage} from "@/utils/common";

interface ChatHistoriesProps {
    onSelectInteraction: (uuid: string) => void;
}

const ChatHistories: React.FC<ChatHistoriesProps> = ({onSelectInteraction}) => {
    const [histories, setHistories] = useState<Interaction[]>([]);
    const [isLoading, setIsLoading] = useState(true);

    const apiService = createApiService(AuthService);

    useEffect(() => {
        const source = axios.CancelToken.source();
        const fetchHistories = async () => {
            setIsLoading(true);
            try {
                const data = await apiService.chatHisories.getChatHistories(source.token);
                setHistories(data.interactions);
            } catch (error) {
                console.error('Error fetching chat histories:', error);
            } finally {
                setIsLoading(false);
            }
        };

        fetchHistories();
    }, []);

    const getFirstUserMessage = (interaction: Interaction): string => {
        if (!interaction.conversations || interaction.conversations.length === 0) {
            return interaction.query;
        }
        const userMessage = interaction.conversations.find(msg => msg.role === 'user');
        return userMessage ? userMessage.text : interaction.query;
    };

    return (
        <div className="bg-white shadow-md rounded-lg overflow-hidden">
            <h2 className="text-xl font-semibold p-4 border-b">Chat Histories</h2>
            {isLoading ? (
                <div className="p-4">Loading...</div>
            ) : (
                <ul className="divide-y divide-gray-200">
                    {histories.map((history) => (
                        <li
                            key={history.uuid}
                            className="p-4 hover:bg-gray-50 cursor-pointer"
                            onClick={() => onSelectInteraction(history.uuid)}
                        >
                            <h3 className="text-lg font-medium text-gray-900">{truncateMessage(getFirstUserMessage(history), 100)}</h3>
                        </li>
                    ))}
                </ul>
            )}
        </div>
    );
};

export default ChatHistories;