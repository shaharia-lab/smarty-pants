import React from 'react';

interface AlertProps {
    children: React.ReactNode;
    variant?: 'default' | 'destructive';
}

export const Alert: React.FC<AlertProps> = ({ children, variant = 'default' }) => {
    const baseClasses = 'px-4 py-3 rounded-md mb-4 text-sm';
    const variantClasses = {
        default: 'bg-green-100 text-green-800 border border-green-300',
        destructive: 'bg-red-100 text-red-800 border border-red-300'
    };

    return (
        <div className={`${baseClasses} ${variantClasses[variant]}`} role="alert">
            {children}
        </div>
    );
};

export const AlertDescription: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    return <p className="text-sm">{children}</p>;
};