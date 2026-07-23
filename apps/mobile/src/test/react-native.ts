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
export const StyleSheet = { create: <T>(styles: T): T => styles };
