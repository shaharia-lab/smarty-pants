import React, { useRef, useEffect } from 'react';
import { Message } from '../types';

interface MessageListProps {
    messages: Message[];
}

const MessageList: React.FC<MessageListProps> = ({ messages }) => {
    const messagesEndRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [messages]);

    return (
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
    );
};

export default MessageList;