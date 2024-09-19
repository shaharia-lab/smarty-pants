// app/home.tsx
import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet, TouchableOpacity, Modal } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import SVGLogo from '../components/SVGLogo';

export default function Home() {
    const [modalVisible, setModalVisible] = useState(false);
    const [endpoint, setEndpoint] = useState('');

    useEffect(() => {
        loadEndpoint();
    }, []);

    const loadEndpoint = async () => {
        const savedEndpoint = await AsyncStorage.getItem('backendEndpoint');
        if (savedEndpoint) {
            setEndpoint(savedEndpoint);
        }
    };

    return (
        <View style={styles.container}>
            <SVGLogo
                width={150}
                height={150}
                leftBrainColor="#ffffff"
                rightBrainColor="#ffffff"
                centerSquareColor="#4b5563"
                centerSquareBlinkColor="#6b7280"
            />
            <Text style={styles.title}>SmartyPants</Text>
            <Text style={styles.subtitle}>Connected to: {endpoint}</Text>
            <TouchableOpacity
                style={styles.button}
                onPress={() => setModalVisible(true)}
            >
                <Text style={styles.buttonText}>Ask me</Text>
            </TouchableOpacity>

            <Modal
                animationType="fade"
                transparent={true}
                visible={modalVisible}
                onRequestClose={() => setModalVisible(false)}
            >
                <View style={styles.modalContainer}>
                    <View style={styles.modalContent}>
                        <Text style={styles.modalText}>Under Construction</Text>
                        <TouchableOpacity
                            style={styles.modalButton}
                            onPress={() => setModalVisible(false)}
                        >
                            <Text style={styles.modalButtonText}>Close</Text>
                        </TouchableOpacity>
                    </View>
                </View>
            </Modal>
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
    title: {
        fontSize: 32,
        fontWeight: 'bold',
        color: '#ffffff',
        marginBottom: 10,
    },
    subtitle: {
        fontSize: 16,
        color: '#d1d5db', // Lighter color for better readability
        marginBottom: 40,
    },
    button: {
        backgroundColor: '#2e4c77',
        paddingHorizontal: 40,
        paddingVertical: 15,
        borderRadius: 10,
    },
    buttonText: {
        color: '#ffffff',
        fontSize: 18,
        fontWeight: '600',
    },
    modalContainer: {
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
        backgroundColor: 'rgba(0, 0, 0, 0.5)',
    },
    modalContent: {
        backgroundColor: '#374151',
        borderRadius: 20,
        padding: 30,
        alignItems: 'center',
    },
    modalText: {
        fontSize: 20,
        fontWeight: '600',
        marginBottom: 20,
        color: '#ffffff',
    },
    modalButton: {
        backgroundColor: '#4b5563',
        paddingHorizontal: 30,
        paddingVertical: 10,
        borderRadius: 10,
    },
    modalButtonText: {
        color: '#ffffff',
        fontSize: 16,
        fontWeight: '600',
    },
});