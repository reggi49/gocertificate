import { Component, type ErrorInfo, type ReactNode } from 'react';
import { AlertCircle, RefreshCw } from 'lucide-react';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
  componentStack?: string | null;
}

export class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
    componentStack: null,
  };

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Uncaught component error:', error, errorInfo);
    this.setState({ componentStack: errorInfo.componentStack });
  }

  public handleReset = () => {
    this.setState({ hasError: false, error: null, componentStack: null });
  };

  public render() {
    if (this.state.hasError) {
      return (
        <div className="flex flex-col items-center justify-center p-8 bg-red-50 border border-red-200 rounded-xl m-4 text-center">
          <AlertCircle className="w-12 h-12 text-red-500 mb-3" />
          <h2 className="text-lg font-bold text-red-800 mb-1">
            Terjadi Kesalahan Tampilan
          </h2>
          <p className="text-sm text-red-600 max-w-md mb-3 font-medium">
            {this.state.error?.message || 'Terjadi kesalahan yang tidak terduga pada antarmuka.'}
          </p>
          {(this.state.error?.stack || this.state.componentStack) && (
            <details className="text-[11px] text-slate-600 bg-white p-3 rounded-xl border border-red-200 text-left w-full max-w-xl mb-4 overflow-auto max-h-40">
              <summary className="cursor-pointer font-semibold text-slate-700 mb-1 select-none">
                Detail Teknis (Klik untuk melihat info debug)
              </summary>
              <pre className="font-mono text-[10px] whitespace-pre-wrap text-slate-600">
                {this.state.error?.stack}
                {this.state.componentStack}
              </pre>
            </details>
          )}
          <button
            onClick={this.handleReset}
            className="flex items-center gap-2 px-4 py-2 bg-red-600 text-white rounded-lg text-sm font-medium hover:bg-red-700 transition cursor-pointer shadow-xs"
          >
            <RefreshCw className="w-4 h-4" />
            Coba Lagi
          </button>
        </div>
      );
    }

    return this.props.children;
  }
}
