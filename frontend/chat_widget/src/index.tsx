import React from 'react';
import ReactDOM from 'react-dom';
import ChatWidget from './components/ChatWidget';
import './styles.css';

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

const initChatWidget = (config: ChatWidgetConfig) => {
    // Validate required fields
    if (!config.backend || !config.backend.endpoint || !config.backend.api_key || !config.backend.widget_id) {
        console.error('ChatWidget Error: Backend configuration is incomplete. Please provide endpoint, api_key, and widget_id.');
        return;
    }

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

(window as any).initChatWidget = initChatWidget;