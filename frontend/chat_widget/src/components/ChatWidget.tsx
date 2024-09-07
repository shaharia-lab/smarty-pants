import React, {useEffect, useRef, useState} from 'react';

interface ChatWidgetProps {
    position: 'bottom-right' | 'bottom-left' | 'top-right' | 'top-left';
    title: string;
    primaryColor: string;
    width?: string;
    height?: string;
}

interface Message {
    text: string;
    isUser: boolean;
}

const ChatWidget: React.FC<ChatWidgetProps> = (
    {
        position,
        title,
        primaryColor,
        width = '320px',  // default width
        height = '480px'  // default height
    }) => {
    const [isOpen, setIsOpen] = useState(false);
    const [messages, setMessages] = useState<Message[]>([]);
    const [inputValue, setInputValue] = useState('');
    const messagesEndRef = useRef<HTMLDivElement>(null);

    const handleToggle = () => setIsOpen(!isOpen);

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (inputValue.trim()) {
            const newMessage: Message = {text: inputValue, isUser: true};
            setMessages([...messages, newMessage]);
            setInputValue('');
            // Simulate bot response
            setTimeout(() => {
                const botMessage: Message = {
                    text: "Thank you for your message. How can I assist you further?",
                    isUser: false
                };
                setMessages(prev => [...prev, botMessage]);
            }, 1000);
        }
    };

    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({behavior: "smooth"});
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
                        className="flex justify-between items-center p-4 bg-gradient-to-r from-blue-500 to-blue-600 text-white"
                        style={{ backgroundColor: primaryColor }}
                    >
                        <h3 className="font-semibold">{title}</h3>
                        <button onClick={handleToggle} className="text-white hover:text-gray-200">
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
                    className="bg-blue-500 text-white p-4 rounded-full shadow-lg hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-opacity-50"
                    style={{ backgroundColor: primaryColor }}
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