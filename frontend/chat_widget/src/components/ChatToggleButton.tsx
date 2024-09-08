import React from 'react';
import SVGLogo from './SVGLogo';

interface ChatToggleButtonProps {
    onClick: () => void;
}

const ChatToggleButton: React.FC<ChatToggleButtonProps> = ({ onClick }) => (
    <button
        onClick={onClick}
        className="bg-gray-800 text-white p-2 rounded-full shadow-lg hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-opacity-50"
        style={{
            width: '48px',
            height: '48px',
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center'
        }}
        aria-label="Toggle chat"
    >
        <span role="img" aria-label="Chat logo">
            <SVGLogo
                width={32}
                height={32}
                leftBrainColor="#FFF"
                rightBrainColor="#FFF"
                centerSquareColor="#8CA6C9"
                centerSquareBlinkColor="#FFFFFF"
                onHoverRotationDegree={15}
            />
        </span>
    </button>
);

export default ChatToggleButton;