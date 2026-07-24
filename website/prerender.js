import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const routesToPrerender = [
  '/',
  '/docs/v1/01-getting-started',
  '/docs/v1/02-core-architecture',
  '/docs/v1/03-go-actions-rpc',
  '/docs/v1/04-dx-helpers',
  '/docs/v1/05-rendering-modes',
  '/docs/v1/06-auth-module',
  '/docs/v1/07-ui-components',
  '/docs/v1/08-security-audit',
  '/docs/v1/09-deployment',
  '/tutorials',
  '/tutorials/01-first-zyra-app-10-min',
  '/tutorials/02-building-saas-with-zyra',
  '/tutorials/03-migrating-from-nextjs-to-zyra',
  '/tutorials/04-migrating-from-express-react',
  '/templates',
  '/changelog',
];

const distDir = path.resolve(__dirname, 'dist');
const templateHtml = fs.readFileSync(path.join(distDir, 'index.html'), 'utf-8');

console.log('⚡ Prerendering static HTML for Cloudflare Pages SEO optimization...');

routesToPrerender.forEach((route) => {
  if (route === '/') return; // index.html already exists

  const routePath = route.startsWith('/') ? route.slice(1) : route;
  const targetDir = path.join(distDir, routePath);
  
  if (!fs.existsSync(targetDir)) {
    fs.mkdirSync(targetDir, { recursive: true });
  }

  // Generate personalized title for static HTML prerender
  let title = 'Zyra — Zero-Runtime-Dependency Go & React Web Framework';
  if (route.includes('/docs/v1/01-getting-started')) title = 'Getting Started — Zyra Docs v1.0.0';
  else if (route.includes('/docs/v1/02-core-architecture')) title = 'Core Architecture — Zyra Docs v1.0.0';
  else if (route.includes('/docs/v1/03-go-actions-rpc')) title = 'Go Actions RPC Protocol — Zyra Docs v1.0.0';
  else if (route.includes('/docs/v1/04-dx-helpers')) title = '45 DX Helpers — Zyra Docs v1.0.0';
  else if (route.includes('/docs/v1/05-rendering-modes')) title = 'Rendering Modes — Zyra Docs v1.0.0';
  else if (route.includes('/docs/v1/06-auth-module')) title = 'Official Auth Module — Zyra Docs v1.0.0';
  else if (route.includes('/docs/v1/07-ui-components')) title = '40+ UI Components — Zyra Docs v1.0.0';
  else if (route.includes('/docs/v1/08-security-audit')) title = 'Security Audit — Zyra Docs v1.0.0';
  else if (route.includes('/docs/v1/09-deployment')) title = 'Production Deployment — Zyra Docs v1.0.0';
  else if (route.includes('/tutorials')) title = 'Tutorials & Guides — Zyra Framework';
  else if (route.includes('/templates')) title = '10 Starter Templates — Zyra Framework';
  else if (route.includes('/changelog')) title = 'Changelog — Zyra v1.0.0';

  const htmlContent = templateHtml.replace(
    '<title>Zyra — Zero-Runtime-Dependency Go & React Web Framework</title>',
    `<title>${title}</title>`
  );

  fs.writeFileSync(path.join(targetDir, 'index.html'), htmlContent, 'utf-8');
  console.log(`  ✓ Prerendered ${route} -> dist/${routePath}/index.html`);
});

console.log('✅ Static pre-rendering completed successfully!');
