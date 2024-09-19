// app/index.tsx
import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet, TouchableOpacity, TextInput, Animated } from 'react-native';
import { useRouter } from 'expo-router';
import AsyncStorage from '@react-native-async-storage/async-storage';
import SVGLogo from '../components/SVGLogo';

export default function Welcome() {
    const [endpoint, setEndpoint] = useState('');
    const [isValid, setIsValid] = useState(false);
    const router = useRouter();
    const animatedValue = new Animated.Value(0);

    useEffect(() => {
        Animated.loop(
            Animated.sequence([
                Animated.timing(animatedValue, {
                    toValue: 1,
                    duration: 15000,
                    useNativeDriver: true,
                }),
                Animated.timing(animatedValue, {
                    toValue: 0,
                    duration: 15000,
                    useNativeDriver: true,
                }),
            ])
        ).start();
    }, []);

    const backgroundStyle = {
        transform: [
            {
                translateX: animatedValue.interpolate({
                    inputRange: [0, 1],
                    outputRange: [0, 50],
                }),
            },
            {
                translateY: animatedValue.interpolate({
                    inputRange: [0, 1],
                    outputRange: [0, 50],
                }),
            },
        ],
    };

    const handleEndpointChange = (text: string) => {
        setEndpoint(text);
        setIsValid(isValidUrl(text));
    };

    const handleGetStarted = async () => {
        if (isValid) {
            const existingEndpoint = await AsyncStorage.getItem('backendEndpoint');
            if (!existingEndpoint) {
                await AsyncStorage.setItem('backendEndpoint', endpoint);
                router.replace('/home');
            } else {
                // Endpoint already exists, just navigate
                router.replace('/home');
            }
        }
    };

    const isValidUrl = (url: string) => {
        const pattern = new RegExp(
            '^(https?:\\/\\/)?' +
            '((([a-z\\d]([a-z\\d-]*[a-z\\d])*)\\.)+[a-z]{2,}|' +
            '((\\d{1,3}\\.){3}\\d{1,3}))' +
            '(\\:\\d+)?(\\/[-a-z\\d%_.~+]*)*' +
            '(\\?[;&a-z\\d%_.~+=-]*)?' +
            '(\\#[-a-z\\d_]*)?$',
            'i'
        );
        return pattern.test(url);
    };

    return (
        <View style={styles.container}>
            <Animated.View style={[styles.backgroundAnimation, backgroundStyle]} />
            <View style={styles.content}>
                <SVGLogo
                    width={150}
                    height={150}
                    leftBrainColor="#ffffff"
                    rightBrainColor="#ffffff"
                    centerSquareColor="#4b5563"
                    centerSquareBlinkColor="#6b7280"
                />
                <Text style={styles.title}>SmartyPants</Text>
                <TextInput
                    style={styles.input}
                    placeholder="Enter backend endpoint"
                    placeholderTextColor="#a0aec0"
                    value={endpoint}
                    onChangeText={handleEndpointChange}
                />
                <TouchableOpacity
                    style={[styles.button, !isValid && styles.buttonDisabled]}
                    onPress={handleGetStarted}
                    disabled={!isValid}
                >
                    <Text style={styles.buttonText}>Get Started</Text>
                </TouchableOpacity>
            </View>
        </View>
    );
}

const styles = StyleSheet.create({
    container: {
        flex: 1,
        backgroundColor: '#1f2937',
        alignItems: 'center',
        justifyContent: 'center',
    },
    backgroundAnimation: {
        position: 'absolute',
        top: -100,
        left: -100,
        right: -100,
        bottom: -100,
        backgroundColor: '#2e4c77',
        opacity: 0.1,
    },
    content: {
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 1,
    },
    title: {
        fontSize: 32,
        fontWeight: 'bold',
        color: '#ffffff',
        marginBottom: 40,
    },
    input: {
        width: 300,
        height: 50,
        backgroundColor: '#374151',
        borderRadius: 10,
        paddingHorizontal: 15,
        fontSize: 16,
        color: '#ffffff',
        marginBottom: 20,
    },
    button: {
        backgroundColor: '#4b5563',
        paddingHorizontal: 40,
        paddingVertical: 15,
        borderRadius: 10,
    },
    buttonDisabled: {
        backgroundColor: '#374151',
        opacity: 0.6,
    },
    buttonText: {
        color: '#ffffff',
        fontSize: 18,
        fontWeight: '600',
    },
});