import React, { useEffect, useState } from 'react';
import { Stack } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { View, ActivityIndicator } from 'react-native';

export default function Layout() {
    const [isLoading, setIsLoading] = useState(true);
    const [initialRoute, setInitialRoute] = useState<string | null>(null);

    useEffect(() => {
        checkEndpoint();
    }, []);

    const checkEndpoint = async () => {
        try {
            const savedEndpoint = await AsyncStorage.getItem('backendEndpoint');
            setInitialRoute(savedEndpoint ? 'home' : 'index');
        } catch (error) {
            console.error('Error checking endpoint:', error);
            setInitialRoute('index');
        } finally {
            setIsLoading(false);
        }
    };

    if (isLoading || initialRoute === null) {
        return (
            <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center', backgroundColor: '#1f2937' }}>
                <StatusBar style="light" />
                <ActivityIndicator size="large" color="#ffffff" />
            </View>
        );
    }

    return (
        <>
            <StatusBar style="light" />
            <Stack
                initialRouteName={initialRoute}
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