import { render } from "@testing-library/react-native";
import { Text, View } from "react-native";
import { expect, it } from "vitest";

it("proves the Vitest RNTL harness can query accessible text", () => {
  const screen = render(
    <View>
      <Text accessibilityRole="header">Tabi test harness</Text>
    </View>,
  );

  expect(screen.getByText("Tabi test harness")).toBeTruthy();
});
