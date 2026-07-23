import { createElement, type PropsWithChildren } from "react";

type HostProps = PropsWithChildren<Record<string, unknown>>;

function host(name: string) {
  return function HostComponent({ children, ...props }: HostProps) {
    return createElement(name, props, children);
  };
}

export const View = host("View");
export const Text = host("Text");
export const FlatList = host("FlatList");
export const ScrollView = host("ScrollView");
export const Pressable = host("Pressable");
export const TextInput = host("TextInput");
export const StyleSheet = { create: <T>(styles: T): T => styles };
export const AppState = {
  currentState: "active",
  addEventListener: () => ({ remove: () => undefined }),
};
