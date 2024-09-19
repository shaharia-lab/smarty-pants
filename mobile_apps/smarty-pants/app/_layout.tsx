import { Stack } from 'expo-router/stack';
import { StatusBar } from 'expo-status-bar';

export default function Layout() {
    return (
        <>
            <StatusBar style="light" />
            <Stack
                screenOptions={{
                    headerShown: false,
                    contentStyle: { backgroundColor: '#1f2937' },
                }}
            >
                <Stack.Screen name="index" />
                <Stack.Screen name="home" />
            </Stack>
        </>
    );
}