import React, { useState } from 'react';

interface ChatWidgetProps {
    position: 'bottom-right' | 'bottom-left' | 'top-right' | 'top-left';
    title: string;
    primaryColor: string;
}

const ChatWidget: React.FC<ChatWidgetProps> = ({ position, title, primaryColor }) => {
    const [isOpen, setIsOpen] = useState(false);
    const [messages, setMessages] = useState<string[]>([]);
    const [inputValue, setInputValue] = useState('');

    const handleToggle = () => setIsOpen(!isOpen);

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (inputValue.trim()) {
            setMessages([...messages, inputValue]);
            setInputValue('');
            // Here you would typically send the message to your backend
            // and handle the response
        }
    };

    const positionStyle = {
        position: 'fixed',
        ...(position.includes('bottom') ? { bottom: '20px' } : { top: '20px' }),
        ...(position.includes('right') ? { right: '20px' } : { left: '20px' }),
    } as React.CSSProperties;

    return (
        <div style={positionStyle}>
            <button onClick={handleToggle} style={{ backgroundColor: primaryColor, color: 'white' }}>
                {isOpen ? 'Close Chat' : title}
            </button>
            {isOpen && (
                <div style={{ width: '300px', height: '400px', border: `1px solid ${primaryColor}` }}>
                    <div style={{ height: '350px', overflowY: 'auto' }}>
                        {messages.map((msg, index) => (
                            <div key={index}>{msg}</div>
                        ))}
                    </div>
                    <form onSubmit={handleSubmit}>
                        <input
                            type="text"
                            value={inputValue}
                            onChange={(e) => setInputValue(e.target.value)}
                            style={{ width: '100%' }}
                        />
                    </form>
                </div>
            )}
        </div>
    );
};

export default ChatWidget;