import React, { useEffect, useState } from 'react';
import { Stack, Redirect } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { View, ActivityIndicator } from 'react-native';

export default function Layout() {
    const [isLoading, setIsLoading] = useState(true);
    const [hasEndpoint, setHasEndpoint] = useState(false);

    useEffect(() => {
        checkEndpoint();
    }, []);

    const checkEndpoint = async () => {
        try {
            const savedEndpoint = await AsyncStorage.getItem('backendEndpoint');
            setHasEndpoint(!!savedEndpoint);
        } catch (error) {
            console.error('Error checking endpoint:', error);
        } finally {
            setIsLoading(false);
        }
    };

    if (isLoading) {
        return (
            <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center', backgroundColor: '#1f2937' }}>
                <ActivityIndicator size="large" color="#ffffff" />
            </View>
        );
    }

    return (
        <>
            <StatusBar style="light" />
            {hasEndpoint ? <Redirect href="/home" /> : <Redirect href="/" />}
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