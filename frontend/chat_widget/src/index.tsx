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

(window as any).initChatWidget = initChatWidget;