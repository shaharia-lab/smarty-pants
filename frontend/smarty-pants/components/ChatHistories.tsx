import React, {useEffect, useState} from 'react';

interface InteractionSummary {
    uuid: string;
    title: string;
}

interface InteractionsResponse {
    interactions: InteractionSummary[];
    limit: number;
    per_page: number;
}

interface ChatHistoriesProps {
    onSelectInteraction: (uuid: string) => void;
}

const ChatHistories: React.FC<ChatHistoriesProps> = ({onSelectInteraction}) => {
    const [histories, setHistories] = useState<InteractionSummary[]>([]);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        fetchHistories();
    }, []);

    const fetchHistories = async () => {
        setIsLoading(true);
        try {
            const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/v1/interactions`);
            const data: InteractionsResponse = await response.json();
            setHistories(data.interactions);
        } catch (error) {
            console.error('Error fetching chat histories:', error);
        } finally {
            setIsLoading(false);
        }
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
                            <h3 className="text-lg font-medium text-gray-900">{history.title}</h3>
                        </li>
                    ))}
                </ul>
            )}
        </div>
    );
};

export default ChatHistories;