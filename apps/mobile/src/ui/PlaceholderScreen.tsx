import { StyleSheet, Text, View } from "react-native";

export function PlaceholderScreen({ title }: { title: string }) {
  return (
    <View style={styles.container}>
      <Text accessibilityRole="header" style={styles.title}>
        {title}
      </Text>
      <Text>Reserved for a later, backend-backed vertical-slice feature.</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, gap: 8, padding: 20 },
  title: { fontSize: 22, fontWeight: "600" },
});
