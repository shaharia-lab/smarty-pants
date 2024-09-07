export interface ChatWidgetConfig {
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

export interface Message {
    text: string;
    isUser: boolean;
}