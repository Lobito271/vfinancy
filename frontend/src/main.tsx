import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { App } from '@/app/App';
import { Toaster } from '@/components/feedback';
import '@fontsource/figtree/400.css';
import '@fontsource/figtree/500.css';
import '@fontsource/figtree/600.css';
import '@fontsource/figtree/700.css';
import '@fontsource/montserrat/500.css';
import '@fontsource/montserrat/600.css';
import '@fontsource/montserrat/700.css';
import '@fontsource/montserrat/800.css';
import '@/index.css';

const container = document.getElementById('app');
if (!container) {
  throw new Error('Root element #app not found');
}

createRoot(container).render(
  <StrictMode>
    <App />
    <Toaster />
  </StrictMode>,
);
