import React from 'react';

export interface HeaderConfig {
    title: string;
}

interface HeaderProps {
    config: HeaderConfig;
}

const Header: React.FC<HeaderProps> = ({ config }) => {
    return (
        <header className="bg-white shadow">
            <div className="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
                <h1 className="text-3xl font-bold tracking-tight text-gray-900">{config.title}</h1>
            </div>
        </header>
    );
};

export default Header;