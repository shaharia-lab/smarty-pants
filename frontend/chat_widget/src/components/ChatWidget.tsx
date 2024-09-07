import React, { useState, useEffect } from 'react';
import ChatHeader from './ChatHeader';
import MessageList from './MessageList';
import MessageInput from './MessageInput';
import ChatToggleButton from './ChatToggleButton';
import { ChatWidgetConfig, Message } from '../types';
import { useChatMessages } from '../hooks/useChatMessages';

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
    const { messages, addMessage } = useChatMessages();

    useEffect(() => {
        if (!backend.endpoint || !backend.api_key || !backend.widget_id) {
            console.error('ChatWidget Error: Backend configuration is incomplete. Please provide endpoint, api_key, and widget_id.');
        }
    }, [backend]);

    const handleToggle = () => setIsOpen(!isOpen);

    const handleSendMessage = (text: string) => {
        addMessage({ text, isUser: true });
        // Simulate API call with dummy response
        setTimeout(() => {
            addMessage({ text: `This is a dummy response. Your message was: ${text}`, isUser: false });
        }, 1000);
    };

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
                    <ChatHeader title={title} branding={branding} onClose={handleToggle} />
                    <MessageList messages={messages} />
                    <MessageInput onSendMessage={handleSendMessage} primaryColor={primaryColor} />
                </div>
            ) : (
                <ChatToggleButton onClick={handleToggle} />
            )}
        </div>
    );
};

export default ChatWidget;