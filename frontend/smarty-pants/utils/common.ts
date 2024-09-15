const truncateMessage = (message: string, maxLength: number = 100): string => {
    return message.length > maxLength ? `${message.substring(0, maxLength)}...` : message;
};
export { truncateMessage };