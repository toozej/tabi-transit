import { Component, type ReactNode } from "react";

type Props = { children: ReactNode };
type State = { failed: boolean };

export class ErrorBoundary extends Component<Props, State> {
  state: State = { failed: false };

  static getDerivedStateFromError(): State {
    return { failed: true };
  }

  componentDidCatch() {
    // Telemetry is intentionally not emitted until web-consent policy is approved.
  }

  render() {
    if (this.state.failed) {
      return (
        <main className="page">
          <h1>Something went wrong</h1>
          <p>Tabi could not show this page. Refresh to try again.</p>
        </main>
      );
    }
    return this.props.children;
  }
}
