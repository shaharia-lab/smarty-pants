import React, { useState, useRef, useEffect } from 'react';
import SVGLogo from './SVGLogo';

interface ChatWidgetConfig {
    position: 'bottom-right' | 'bottom-left' | 'top-right' | 'top-left';
    title: string;
    primaryColor: string;
    width?: string;
    height?: string;
    branding?: {
        name: string;
        logo?: string;
    };
    backend: {
        endpoint: string;
        api_key: string;
        widget_id: string;
    };
}

interface Message {
    text: string;
    isUser: boolean;
}

const ChatWidget: React.FC<ChatWidgetConfig> = ({
                                                    position,
                                                    title,
                                                    primaryColor,
                                                    width = '320px',
                                                    height = '480px',
                                                    branding = { name: 'Smarty Pants' },
                                                    backend
                                                }) => {
    const [isOpen, setIsOpen] = useState(false);
    const [messages, setMessages] = useState<Message[]>([]);
    const [inputValue, setInputValue] = useState('');
    const messagesEndRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        // Validate backend configuration
        if (!backend.endpoint || !backend.api_key || !backend.widget_id) {
            console.error('ChatWidget Error: Backend configuration is incomplete. Please provide endpoint, api_key, and widget_id.');
        }
    }, [backend]);

    const handleToggle = () => setIsOpen(!isOpen);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (inputValue.trim()) {
            const newMessage: Message = { text: inputValue, isUser: true };
            setMessages(prev => [...prev, newMessage]);
            setInputValue('');

            // Simulate API call with dummy response
            setTimeout(() => {
                const dummyResponse: Message = { text: "This is a dummy response. Your message was: " + inputValue, isUser: false };
                setMessages(prev => [...prev, dummyResponse]);
            }, 1000);
        }
    };

    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [messages]);

    const positionClass = {
        'bottom-right': 'bottom-4 right-4',
        'bottom-left': 'bottom-4 left-4',
        'top-right': 'top-4 right-4',
        'top-left': 'top-4 left-4',
    }[position];

    return (
        <div className={`fixed ${positionClass} z-50`}>
            {isOpen ? (
                <div
                    className="bg-white rounded-lg shadow-xl flex flex-col overflow-hidden"
                    style={{ width, height }}
                >
                    <div
                        className="flex justify-between items-center p-4 bg-gray-800 text-white"
                    >
                        <div className="flex items-center">
                            {branding.logo ? (
                                <img src={branding.logo} alt={branding.name} className="w-8 h-8 mr-2" />
                            ) : (
                                <SVGLogo width={32} height={32} leftBrainColor="#FFF"
                                         rightBrainColor="#FFF"
                                         centerSquareColor="#8CA6C9"
                                         centerSquareBlinkColor="#FFFFFF"
                                         onHoverRotationDegree={15} />
                            )}
                            <h3 className="font-semibold ml-2">{title}</h3>
                        </div>
                        <button onClick={handleToggle} className="text-white hover:text-gray-300">
                            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12"></path>
                            </svg>
                        </button>
                    </div>
                    <div className="flex-1 p-4 overflow-y-auto">
                        {messages.map((msg, index) => (
                            <div key={index} className={`chat-message ${msg.isUser ? 'user-message' : 'bot-message'}`}>
                                <div className="chat-message-content">
                                    {msg.text}
                                </div>
                            </div>
                        ))}
                        <div ref={messagesEndRef} />
                    </div>
                    <form onSubmit={handleSubmit} className="p-4 border-t">
                        <div className="flex items-center">
                            <input
                                type="text"
                                value={inputValue}
                                onChange={(e) => setInputValue(e.target.value)}
                                placeholder="Type a message..."
                                className="flex-1 p-2 border rounded-l-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                            />
                            <button
                                type="submit"
                                className="p-2 bg-blue-500 text-white rounded-r-lg hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500"
                                style={{ backgroundColor: primaryColor }}
                            >
                                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"></path>
                                </svg>
                            </button>
                        </div>
                    </form>
                </div>
            ) : (
                <button
                    onClick={handleToggle}
                    className="bg-gray-800 text-white p-4 rounded-full shadow-lg hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-opacity-50"
                >
                    <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z"></path>
                    </svg>
                </button>
            )}
        </div>
    );
};

export default ChatWidget;