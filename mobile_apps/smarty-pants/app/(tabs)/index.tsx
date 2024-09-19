import React, { useState } from 'react';
import { StyleSheet, View, TouchableOpacity, Text, Modal } from 'react-native';
import SVGLogo from '@/components/SVGLogo';

export default function HomeScreen() {
    const [modalVisible, setModalVisible] = useState(false);

    return (
        <View style={styles.container}>
            <SVGLogo
                width={150}
                height={150}
                leftBrainColor="#FFF"
                rightBrainColor="#FFF"
                centerSquareColor="#8CA6C9"
                centerSquareBlinkColor="#FFFFFF"
            />
            <Text style={styles.welcomeText}>Welcome to SmartyPants</Text>
            <TouchableOpacity style={styles.button} onPress={() => setModalVisible(true)}>
                <Text style={styles.buttonText}>Ask Me</Text>
            </TouchableOpacity>

            <Modal
                animationType="fade"
                transparent={true}
                visible={modalVisible}
                onRequestClose={() => setModalVisible(false)}
            >
                <View style={styles.modalContainer}>
                    <View style={styles.modalContent}>
                        <Text style={styles.modalText}>Coming soon!</Text>
                        <TouchableOpacity
                            style={styles.closeButton}
                            onPress={() => setModalVisible(false)}
                        >
                            <Text style={styles.closeButtonText}>Close</Text>
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
        backgroundColor: '#1F2937',
        alignItems: 'center',
        justifyContent: 'center',
    },
    welcomeText: {
        color: '#FFFFFF',
        fontSize: 24,
        fontWeight: 'bold',
        marginTop: 20,
        marginBottom: 30,
    },
    button: {
        backgroundColor: '#FFFFFF',
        paddingHorizontal: 20,
        paddingVertical: 10,
        borderRadius: 5,
    },
    buttonText: {
        color: '#1F2937',
        fontSize: 18,
        fontWeight: 'bold',
    },
    modalContainer: {
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
        backgroundColor: 'rgba(0, 0, 0, 0.5)',
    },
    modalContent: {
        backgroundColor: '#FFFFFF',
        padding: 20,
        borderRadius: 10,
        alignItems: 'center',
    },
    modalText: {
        fontSize: 18,
        marginBottom: 15,
        color: '#1F2937',
    },
    closeButton: {
        backgroundColor: '#1F2937',
        padding: 10,
        borderRadius: 5,
    },
    closeButtonText: {
        color: '#FFFFFF',
        fontWeight: 'bold',
    },
});