import React, { useState, useEffect } from 'react';
import { BrowserRouter, Routes, Route, useLocation } from 'react-router-dom';
import { Navbar } from './components/Navbar';
import { Footer } from './components/Footer';
import { SearchModal } from './components/SearchModal';
import { HomePage } from './pages/HomePage';
import { DocsPage } from './pages/DocsPage';
import { TutorialsPage } from './pages/TutorialsPage';
import { TemplatesPage } from './pages/TemplatesPage';
import { ChangelogPage } from './pages/ChangelogPage';
import { NotFoundPage } from './pages/NotFoundPage';

// SEO Manager component
const SEOManager: React.FC = () => {
  const location = useLocation();

  useEffect(() => {
    // Dynamic page title
    let title = 'Zyra — Zero-Runtime-Dependency Go & React Web Framework';
    if (location.pathname.startsWith('/docs')) {
      title = 'Documentation — Zyra v1.0.0';
    } else if (location.pathname.startsWith('/tutorials')) {
      title = 'Tutorials & Guides — Zyra Framework';
    } else if (location.pathname.startsWith('/templates')) {
      title = '10 Starter Templates — Zyra Framework';
    } else if (location.pathname.startsWith('/changelog')) {
      title = 'Changelog — Zyra v1.0.0';
    }

    document.title = title;

    // Inject JSON-LD SoftwareApplication structured data
    let jsonLdScript = document.getElementById('zyra-jsonld');
    if (!jsonLdScript) {
      jsonLdScript = document.createElement('script');
      jsonLdScript.id = 'zyra-jsonld';
      jsonLdScript.setAttribute('type', 'application/ld+json');
      document.head.appendChild(jsonLdScript);
    }

    const structuredData = {
      '@context': 'https://schema.org',
      '@type': 'SoftwareApplication',
      'name': 'Zyra Web Framework',
      'operatingSystem': 'Linux, Windows, macOS, Alpine',
      'applicationCategory': 'DeveloperApplication',
      'offers': {
        '@type': 'Offer',
        'price': '0',
        'priceCurrency': 'USD'
      },
      'description': 'Zero-runtime-dependency fullstack web framework combining Go 1.23+ and React 18/19.',
      'url': 'https://zyraframework.dev'
    };

    jsonLdScript.textContent = JSON.stringify(structuredData);
  }, [location]);

  return null;
};

export const App: React.FC = () => {
  const [isSearchOpen, setIsSearchOpen] = useState(false);

  return (
    <BrowserRouter>
      <SEOManager />
      <div className="min-h-screen flex flex-col bg-[#0a0d14] text-slate-200">
        <Navbar onOpenSearch={() => setIsSearchOpen(true)} />
        <div className="flex-1">
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/docs/v1" element={<DocsPage />} />
            <Route path="/docs/v1/:docId" element={<DocsPage />} />
            <Route path="/tutorials" element={<TutorialsPage />} />
            <Route path="/tutorials/:tutorialId" element={<TutorialsPage />} />
            <Route path="/templates" element={<TemplatesPage />} />
            <Route path="/changelog" element={<ChangelogPage />} />
            <Route path="*" element={<NotFoundPage />} />
          </Routes>
        </div>
        <Footer />
        <SearchModal isOpen={isSearchOpen} onClose={() => setIsSearchOpen(false)} />
      </div>
    </BrowserRouter>
  );
};

export default App;
