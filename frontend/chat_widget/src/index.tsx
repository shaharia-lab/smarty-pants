import React from 'react';
import ReactDOM from 'react-dom';
import ChatWidget from './components/ChatWidget';

interface ChatWidgetConfig {
    position: 'bottom-right' | 'bottom-left' | 'top-right' | 'top-left';
    title: string;
    primaryColor: string;
}

const initChatWidget = (config: ChatWidgetConfig) => {
    const widgetContainer = document.createElement('div');
    widgetContainer.id = 'chat-widget-container';
    document.body.appendChild(widgetContainer);

    ReactDOM.render(
        <React.StrictMode>
            <ChatWidget {...config} />
        </React.StrictMode>,
        widgetContainer
    );
};

// Expose the init function globally
(window as any).initChatWidget = initChatWidget;