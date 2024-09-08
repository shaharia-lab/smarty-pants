import { useState } from 'react';
import { Message } from '../types';

export const useChatMessages = () => {
    const [messages, setMessages] = useState<Message[]>([]);

    const addMessage = (newMessage: Message) => {
        setMessages(prevMessages => [...prevMessages, newMessage]);
    };

    return { messages, addMessage };
};