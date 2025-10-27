# Payment Platform - Official Website

Public-facing marketing website for the Payment Platform. Showcases features, products, documentation, and pricing to potential customers. Built with React 18, TypeScript, Vite, and Ant Design.

## Features

### 🏠 Home Page
- **Hero Section** - Eye-catching hero with call-to-action buttons
- **Platform Statistics** - Real-time metrics (100+ merchants, $50M+ processed, 99.9% uptime)
- **Feature Highlights** - Key platform capabilities with icons
- **Customer Testimonials** - Social proof (optional)
- **Call-to-Action** - Links to Admin Portal and Merchant Portal

### 📦 Products Page
- **Payment Gateway** - Unified payment API, multi-channel support, intelligent routing
- **Risk Management** - Real-time fraud detection, rule engine, blacklist management
- **Settlement System** - Automated settlement, multi-currency, reconciliation
- **Monitoring & Analytics** - Real-time dashboards, transaction insights, performance metrics

### 📚 Documentation Page
- **Quick Start Guide** - Getting started tutorial for new merchants
- **API Reference** - Complete API documentation with examples
- **SDKs** - Client libraries (Node.js, Python, PHP, Java, Go)
- **Webhooks** - Event notification system documentation

### 💰 Pricing Page
Three-tier pricing plans:
1. **Starter** - Free for testing, 100 transactions/month, Sandbox access
2. **Professional** - $99/month, Unlimited transactions, 2.9% + $0.30 per transaction
3. **Enterprise** - Custom pricing, Volume discounts, Dedicated support, SLA

### 🌍 Bilingual Support
- **English** - Default language
- **简体中文** - Simplified Chinese

### 📱 Responsive Design
- Mobile-friendly layout
- Adaptive navigation
- Touch-optimized interactions

## Tech Stack

- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite 5 (fast HMR, optimized builds)
- **UI Library**: Ant Design 5.15 + @ant-design/icons
- **Routing**: React Router v6
- **i18n**: react-i18n ext (English & 简体中文)
- **Code Quality**: ESLint + Prettier

## Quick Start

### Prerequisites

- **Node.js** 18+
- **npm** 9+

### Installation

```bash
# Navigate to website directory
cd frontend/website

# Install dependencies
npm install
```

### Development

```bash
# Start development server (http://localhost:5175)
npm run dev

# The website will automatically open in your browser
# Hot Module Replacement (HMR) is enabled
```

### Build

```bash
# Build for production
npm run build

# Output will be in dist/ directory
# Optimized with code splitting and tree shaking

# Preview production build
npm run preview
```

### Code Quality

```bash
# Run ESLint
npm run lint

# Format code
npm run format
```

## Project Structure

```
website/
├── src/
│   ├── pages/              # Page components
│   │   ├── Home/           # Landing page with hero section
│   │   │   ├── index.tsx   # Main home component
│   │   │   └── styles.css  # Home page styles
│   │   ├── Products/       # Product showcase page
│   │   │   ├── index.tsx   # Products component
│   │   │   └── styles.css  # Product styles
│   │   ├── Docs/           # Documentation hub
│   │   │   ├── index.tsx   # Documentation component
│   │   │   └── styles.css  # Docs styles
│   │   └── Pricing/        # Pricing plans page
│   │       ├── index.tsx   # Pricing component
│   │       └── styles.css  # Pricing styles
│   │
│   ├── components/         # Shared components
│   │   ├── Header/         # Site navigation header
│   │   │   ├── index.tsx   # Header component
│   │   │   └── styles.css  # Header styles
│   │   ├── Footer/         # Site footer
│   │   │   ├── index.tsx   # Footer component
│   │   │   └── styles.css  # Footer styles
│   │   └── LanguageSwitch/ # Language switcher
│   │       └── index.tsx   # Language switch component
│   │
│   ├── i18n/               # Internationalization
│   │   ├── index.ts        # i18n configuration
│   │   └── locales/        # Translation files
│   │       ├── en.json     # English translations
│   │       └── zh-CN.json  # Simplified Chinese translations
│   │
│   ├── App.tsx             # Main app with routing
│   ├── main.tsx            # Entry point
│   └── index.css           # Global styles
│
├── public/                 # Static assets
│   ├── logo.png           # Platform logo
│   └── ...                # Other public assets
│
├── index.html              # HTML template
├── vite.config.ts          # Vite configuration
├── tsconfig.json           # TypeScript configuration
├── package.json            # Dependencies
└── README.md              # This file
```

## Key Pages

### Home (`/`)
The landing page featuring:
```typescript
// Hero section with statistics
<Hero
  title="Global Payment Platform"
  subtitle="Enterprise payment gateway for modern businesses"
  stats={[
    { label: "Merchants", value: "100+" },
    { label: "Processed", value: "$50M+" },
    { label: "Uptime", value: "99.9%" }
  ]}
/>

// Feature highlights
<Features features={[
  { icon: <RocketOutlined />, title: "Fast Integration", description: "..." },
  { icon: <SafetyOutlined />, title: "Secure & Compliant", description: "..." },
  // ...
]} />
```

### Products (`/products`)
Showcases platform capabilities:
- Payment Gateway features
- Risk Management system
- Settlement capabilities
- Monitoring & Analytics

### Documentation (`/docs`)
Documentation hub with:
- Quick Start Guide
- API Reference links
- SDK download links
- Webhook documentation

### Pricing (`/pricing`)
Three-tier pricing plans in card layout:
```typescript
<PricingCard
  plan="Professional"
  price="$99"
  period="/month"
  features={[
    "Unlimited transactions",
    "2.9% + $0.30 per transaction",
    "Multi-channel support",
    "24/7 support"
  ]}
/>
```

## Internationalization

### Adding New Translation

1. Add translations to `src/i18n/locales/{lang}.json`:

```json
{
  "home": {
    "hero": {
      "title": "Global Payment Platform",
      "subtitle": "Enterprise payment gateway for modern businesses"
    }
  }
}
```

2. Use in components:

```typescript
import { useTranslation } from 'react-i18next';

function Home() {
  const { t } = useTranslation();

  return (
    <h1>{t('home.hero.title')}</h1>
  );
}
```

### Language Switching

```typescript
// LanguageSwitch component
const { i18n } = useTranslation();

const changeLanguage = (lang: 'en' | 'zh-CN') => {
  i18n.changeLanguage(lang);
  localStorage.setItem('language', lang);
};
```

## Links to Other Applications

The website includes navigation to:

- **Admin Portal**: http://localhost:5173 (for platform operators)
- **Merchant Portal**: http://localhost:5174 (for merchants)
- **API Documentation**: Links to Swagger UI endpoints

Configured in Header component:

```typescript
<Button onClick={() => window.location.href = 'http://localhost:5173'}>
  Admin Login
</Button>
<Button onClick={() => window.location.href = 'http://localhost:5174'}>
  Merchant Login
</Button>
```

## Customization

### Update Platform Statistics

Edit `src/pages/Home/index.tsx`:

```typescript
const stats = [
  { label: "Active Merchants", value: "100+", icon: <TeamOutlined /> },
  { label: "Total Processed", value: "$50M+", icon: <DollarOutlined /> },
  { label: "Uptime", value: "99.9%", icon: <CheckCircleOutlined /> },
  { label: "Countries", value: "50+", icon: <GlobalOutlined /> },
];
```

### Update Pricing Plans

Edit `src/pages/Pricing/index.tsx`:

```typescript
const plans = [
  {
    name: "Starter",
    price: "Free",
    features: [
      "100 transactions/month",
      "Sandbox environment",
      "Email support"
    ]
  },
  // ...
];
```

### Add New Page

1. Create page component:
```bash
mkdir src/pages/NewPage
touch src/pages/NewPage/index.tsx
```

2. Add route in `App.tsx`:
```typescript
import NewPage from './pages/NewPage';

<Route path="/new-page" element={<NewPage />} />
```

3. Add navigation link in Header component

## SEO Optimization

### Update Page Titles

```typescript
// In each page component
import { useEffect } from 'react';

function Home() {
  useEffect(() => {
    document.title = 'Payment Platform - Enterprise Payment Gateway';
  }, []);

  return <div>...</div>;
}
```

### Add Meta Tags

Update `index.html`:

```html
<head>
  <meta name="description" content="Enterprise payment gateway for modern businesses">
  <meta name="keywords" content="payment, gateway, api, stripe, paypal">
  <meta property="og:title" content="Payment Platform">
  <meta property="og:description" content="Enterprise payment gateway">
</head>
```

## Deployment

### Production Build

```bash
# Build for production
npm run build

# Output in dist/ directory
# Optimized, minified, code-split
```

### Deploy to Static Hosting

```bash
# Deploy to Netlify
netlify deploy --prod --dir=dist

# Deploy to Vercel
vercel --prod

# Deploy to AWS S3
aws s3 sync dist/ s3://your-bucket-name/
```

### Configure nginx

```nginx
server {
  listen 80;
  server_name www.payment-platform.com;
  root /var/www/html/website;

  location / {
    try_files $uri $uri/ /index.html;
  }

  # Gzip compression
  gzip on;
  gzip_types text/plain text/css application/json application/javascript;

  # Cache static assets
  location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
  }
}
```

## Performance Optimization

### Code Splitting

Vite automatically code splits by route. For manual splitting:

```typescript
import { lazy } from 'react';

const Products = lazy(() => import('./pages/Products'));
const Pricing = lazy(() => import('./pages/Pricing'));
```

### Image Optimization

```typescript
// Use WebP format
<img src="/images/hero.webp" alt="Hero" loading="lazy" />

// Responsive images
<picture>
  <source srcset="/images/hero-mobile.webp" media="(max-width: 768px)" />
  <img src="/images/hero-desktop.webp" alt="Hero" />
</picture>
```

### Lazy Loading

```typescript
import { Suspense } from 'react';

<Suspense fallback={<Spin />}>
  <Products />
</Suspense>
```

## Analytics Integration

### Google Analytics

```html
<!-- In index.html -->
<script async src="https://www.googletagmanager.com/gtag/js?id=GA_MEASUREMENT_ID"></script>
<script>
  window.dataLayer = window.dataLayer || [];
  function gtag(){dataLayer.push(arguments);}
  gtag('js', new Date());
  gtag('config', 'GA_MEASUREMENT_ID');
</script>
```

### Track Page Views

```typescript
import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';

function App() {
  const location = useLocation();

  useEffect(() => {
    // Track page view
    gtag('config', 'GA_MEASUREMENT_ID', {
      page_path: location.pathname
    });
  }, [location]);

  return <div>...</div>;
}
```

## Troubleshooting

### Port Already in Use

```bash
# Kill process on port 5175
lsof -ti:5175 | xargs kill -9

# Or use different port
npm run dev -- --port 5176
```

### Build Errors

```bash
# Clear cache and rebuild
rm -rf node_modules/.vite dist
npm install
npm run build
```

### Routing Issues in Production

Ensure server is configured for SPA routing (serve `index.html` for all routes).

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)
- Mobile browsers (iOS Safari, Chrome Mobile)

## Accessibility

- Semantic HTML structure
- ARIA labels for interactive elements
- Keyboard navigation support
- Sufficient color contrast ratios
- Alt text for all images

## License

Commercial License

## Related Links

- **Admin Portal**: http://localhost:5173 (Platform operators)
- **Merchant Portal**: http://localhost:5174 (Merchants)
- **API Documentation**: See backend Swagger endpoints
- **Project Documentation**: See root `README.md` and `CLAUDE.md`

## Support

- **Documentation**: See project documentation
- **Issues**: Report in GitHub issue tracker
- **Email**: support@payment-platform.com
