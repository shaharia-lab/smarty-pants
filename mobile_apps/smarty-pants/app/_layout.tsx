import { Drawer } from 'expo-router/drawer';
import { GestureHandlerRootView } from 'react-native-gesture-handler';
import { useState } from 'react';
import UnderConstructionModal from '../components/UnderConstructionModal';

export default function Layout() {
  const [modalVisible, setModalVisible] = useState(false);

  const handleDrawerItemPress = () => {
    setModalVisible(true);
  };

  return (
      <GestureHandlerRootView style={{ flex: 1 }}>
        <Drawer
            screenOptions={{
              headerShown: true,
            }}
        >
          <Drawer.Screen
              name="index"
              options={{
                drawerLabel: 'New Conversation',
                title: 'New Conversation',
              }}
              listeners={{
                focus: handleDrawerItemPress,
              }}
          />
          <Drawer.Screen
              name="about"
              options={{
                drawerLabel: 'About',
                title: 'About',
              }}
              listeners={{
                focus: handleDrawerItemPress,
              }}
          />
        </Drawer>
        <UnderConstructionModal
            visible={modalVisible}
            onClose={() => setModalVisible(false)}
        />
      </GestureHandlerRootView>
  );
}