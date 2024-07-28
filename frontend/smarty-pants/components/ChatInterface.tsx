import React, { useState, useEffect, useRef } from 'react';
import ReactMarkdown from 'react-markdown';
import rehypeRaw from 'rehype-raw';
import rehypeSanitize from 'rehype-sanitize';
import rehypeHighlight from 'rehype-highlight';
import 'highlight.js/styles/github.css';
import { ComponentProps } from 'react';

interface Message {
    role: 'system' | 'user';
    text: string;
}

interface Interaction {
    uuid: string;
    query: string;
    conversations: Message[];
}

interface ChatInterfaceProps {
    interactionId: string | null;
}

const ChatInterface: React.FC<ChatInterfaceProps> = ({ interactionId }) => {
    const [interaction, setInteraction] = useState<Interaction | null>(null);
    const [inputMessage, setInputMessage] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const messagesEndRef = useRef<HTMLDivElement>(null);
    const textareaRef = useRef<HTMLTextAreaElement>(null);

    useEffect(() => {
        if (interactionId) {
            fetchInteraction(interactionId);
        } else {
            startNewSession();
        }
    }, [interactionId]);

    useEffect(() => {
        scrollToBottom();
    }, [interaction]);

    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    };

    const fetchInteraction = async (id: string) => {
        setIsLoading(true);
        try {
            const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/interactions/${id}`);
            const data: Interaction = await response.json();
            setInteraction(data);
        } catch (error) {
            console.error('Error fetching interaction:', error);
        } finally {
            setIsLoading(false);
        }
    };

    const startNewSession = async () => {
        setIsLoading(true);
        try {
            const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/interactions`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ query: 'Start new session' }),
            });
            const data: Interaction = await response.json();

            setInteraction({
                ...data,
                conversations: [
                    { role: 'system', text: "I am your smart brain! How may I help you today?" }
                ]
            });
        } catch (error) {
            console.error('Error starting new session:', error);
        } finally {
            setIsLoading(false);
        }
    };

    const handleInputChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        setInputMessage(e.target.value);
        adjustTextareaHeight();
    };

    const adjustTextareaHeight = () => {
        if (textareaRef.current) {
            textareaRef.current.style.height = 'auto';
            textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`;
        }
    };

    const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSubmit(e);
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!inputMessage.trim() || !interaction) return;

        setIsLoading(true);
        try {
            const response = await fetch(`${process.env.NEXT_PUBLIC_API_BASE_URL}/api/interactions/${interaction.uuid}/message`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ query: inputMessage }),
            });
            const data = await response.json();

            setInteraction(prev => prev ? {
                ...prev,
                conversations: [...prev.conversations, { role: 'user', text: inputMessage }, { role: 'system', text: data.response }]
            } : null);

            setInputMessage('');
            adjustTextareaHeight();
        } catch (error) {
            console.error('Error sending message:', error);
        } finally {
            setIsLoading(false);
        }
    };

    const renderMessage = (message: Message) => (
        <div className={`mb-4 flex ${message.role === 'user' ? 'justify-end' : 'justify-start'}`}>
            <div className={`flex items-start max-w-[70%] ${message.role === 'user' ? 'flex-row-reverse' : 'flex-row'}`}>
                {message.role === 'system' ? (
                    <SystemIcon className="w-6 h-6 mt-1 mx-2 text-gray-600" />
                ) : (
                    <UserIcon className="w-6 h-6 mt-1 mx-2 text-blue-600" />
                )}
                <div className={`p-3 rounded-lg ${message.role === 'user' ? 'bg-blue-100' : 'bg-gray-100'}`}>
                    <ReactMarkdown
                        rehypePlugins={[rehypeRaw, rehypeSanitize, rehypeHighlight]}
                        components={{
                            code: ({ inline, className, children, ...props }: ComponentProps<'code'> & { inline?: boolean }) => {
                                const match = /language-(\w+)/.exec(className || '')
                                return !inline && match ? (
                                    <pre className={className}>
                <code {...props} className={className}>
                    {children}
                </code>
            </pre>
                                ) : (
                                    <code {...props} className={className}>
                                        {children}
                                    </code>
                                )
                            }
                        }}
                    >
                        {message.text}
                    </ReactMarkdown>
                </div>
            </div>
        </div>
    );


    return (
        <div className="bg-white shadow-md rounded-lg overflow-hidden flex flex-col h-[calc(100vh-200px)]">
            <div className="border-b border-gray-200 p-4 flex justify-between items-center">
                <h2 className="text-xl font-semibold">Chat Session</h2>
                <button
                    onClick={startNewSession}
                    className="bg-blue-500 text-white px-4 py-2 rounded-lg hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                    Start New Session
                </button>
            </div>
            <div className="flex-grow overflow-y-auto p-4">
                {interaction?.conversations.map((message, index) => (
                    <React.Fragment key={index}>
                        {renderMessage(message)}
                    </React.Fragment>
                ))}
                {isLoading && (
                    <div className="flex items-center justify-center">
                        <LoadingSpinner />
                        <p className="ml-2">Preparing an answer for you. Please wait...</p>
                    </div>
                )}
                <div ref={messagesEndRef} />
            </div>
            <form onSubmit={handleSubmit} className="border-t border-gray-200 p-4">
                <div className="flex items-end">
                    <textarea
                        ref={textareaRef}
                        value={inputMessage}
                        onChange={handleInputChange}
                        onKeyDown={handleKeyDown}
                        placeholder="Type your message here... (Shift+Enter for new line)"
                        className="flex-grow mr-2 p-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 min-h-[40px] max-h-[200px] resize-none"
                        rows={1}
                    />
                    <button
                        type="submit"
                        disabled={isLoading}
                        className="bg-blue-500 text-white px-4 py-2 rounded-lg hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
                    >
                        Send
                    </button>
                </div>
            </form>
        </div>
    );
};

const SystemIcon: React.FC<{ className?: string }> = ({ className }) => (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
    </svg>
);

const UserIcon: React.FC<{ className?: string }> = ({ className }) => (
    <svg className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
    </svg>
);

const LoadingSpinner: React.FC = () => (
    <svg className="animate-spin h-5 w-5 text-blue-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
    </svg>
);

export default ChatInterface;