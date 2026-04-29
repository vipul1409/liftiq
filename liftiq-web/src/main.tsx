import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router';
import { InspectionProvider } from './context/InspectionContext';
import { App } from './App';
import './index.css';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <InspectionProvider>
        <App />
      </InspectionProvider>
    </BrowserRouter>
  </StrictMode>,
);
