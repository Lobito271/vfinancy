import { useEffect } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { HashRouter } from 'react-router-dom';
import { TooltipProvider } from '@/components/misc';
import { useThemeStore } from '@/stores/theme';
import { ErrorBoundary } from './ErrorBoundary';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      refetchOnWindowFocus: false,
      retry: 1,
      networkMode: 'always',
    },
  },
});

function ThemeBridge({ children }: { children: React.ReactNode }) {
  const applyToDocument = useThemeStore((s) => s.applyToDocument);
  useEffect(() => {
    applyToDocument();
  }, [applyToDocument]);
  return <>{children}</>;
}

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <ThemeBridge>
          <HashRouter>
            <TooltipProvider delayDuration={300}>{children}</TooltipProvider>
          </HashRouter>
        </ThemeBridge>
      </QueryClientProvider>
    </ErrorBoundary>
  );
}
