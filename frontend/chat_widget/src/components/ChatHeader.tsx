import React from 'react';
import SVGLogo from './SVGLogo';

interface ChatHeaderProps {
    title: string;
    branding: {
        name: string;
        logo?: string;
    };
    onClose: () => void;
}

const ChatHeader: React.FC<ChatHeaderProps> = ({ title, branding, onClose }) => (
    <div className="flex justify-between items-center p-4 bg-gray-800 text-white">
        <div className="flex items-center">
            {branding.logo ? (
                <img src={branding.logo} alt={branding.name} className="w-8 h-8 mr-2" />
            ) : (
                <div role="img" aria-label="Brand logo">
                    <SVGLogo
                        width={32}
                        height={32}
                        leftBrainColor="#FFF"
                        rightBrainColor="#FFF"
                        centerSquareColor="#8CA6C9"
                        centerSquareBlinkColor="#FFFFFF"
                        onHoverRotationDegree={15}
                    />
                </div>
            )}
            <h3 className="font-semibold ml-2">{title}</h3>
        </div>
        <button onClick={onClose} className="text-white hover:text-gray-300">
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12"></path>
            </svg>
        </button>
    </div>
);

export default ChatHeader;