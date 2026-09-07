import { Component, type ErrorInfo, type ReactNode } from "react";
import { CLINIC } from "../lib/clinic";
import { logger } from "../lib/logger";
import { Button } from "./ui";

interface State {
  failed: boolean;
}

/**
 * Catches render-time crashes so a patient sees a way forward instead of a
 * blank page, and records what happened. Error boundaries have no hooks
 * equivalent, so this stays a class.
 */
export class ErrorBoundary extends Component<{ children: ReactNode }, State> {
  state: State = { failed: false };

  static getDerivedStateFromError(): State {
    return { failed: true };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    logger.error("render crashed", {
      detail: error.message,
      componentStack: info.componentStack,
    });
  }

  render() {
    if (!this.state.failed) return this.props.children;

    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-5 bg-paper px-6 text-center">
        <div>
          <h1 className="text-2xl">Something went wrong on this page</h1>
          <p className="mt-2 max-w-sm text-sm text-ink-soft">
            Your appointment was not affected. Reload to try again, or call the hospital and the
            front desk will book you in.
          </p>
        </div>
        <div className="flex flex-wrap justify-center gap-3">
          <Button onClick={() => window.location.reload()}>Reload the page</Button>
          <Button variant="ghost" onClick={() => window.location.assign(CLINIC.phoneHref)}>
            Call {CLINIC.phoneDisplay}
          </Button>
        </div>
      </div>
    );
  }
}
