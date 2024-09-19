import { Stack } from 'expo-router/stack';
import { StatusBar } from 'expo-status-bar';

export default function Layout() {
  return (
      <>
        <StatusBar style="auto" />
        <Stack screenOptions={{
          headerShown: false,
        }} />
      </>
  );
}