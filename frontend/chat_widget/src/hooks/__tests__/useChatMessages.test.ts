import { renderHook, act } from '@testing-library/react-hooks';
import { useChatMessages } from '../useChatMessages';

describe('useChatMessages', () => {
    it('initializes with an empty array of messages', () => {
        const { result } = renderHook(() => useChatMessages());
        expect(result.current.messages).toEqual([]);
    });

    it('adds a new message correctly', () => {
        const { result } = renderHook(() => useChatMessages());
        act(() => {
            result.current.addMessage({ text: 'Hello', isUser: true });
        });
        expect(result.current.messages).toEqual([{ text: 'Hello', isUser: true }]);
    });

    it('preserves existing messages when adding a new one', () => {
        const { result } = renderHook(() => useChatMessages());
        act(() => {
            result.current.addMessage({ text: 'Hello', isUser: true });
            result.current.addMessage({ text: 'Hi there!', isUser: false });
        });
        expect(result.current.messages).toEqual([
            { text: 'Hello', isUser: true },
            { text: 'Hi there!', isUser: false },
        ]);
    });
});