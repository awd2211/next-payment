# Payment Platform Website

Official website for the Payment Platform - showcasing features, products, documentation, and pricing.

## Tech Stack

- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite 5
- **UI Library**: Ant Design 5.15
- **Routing**: React Router v6
- **i18n**: react-i18next (English & 简体中文)

## Features

- 🏠 **Home Page**: Hero section, platform statistics, feature highlights
- 📦 **Products**: Payment gateway, risk management, settlement, monitoring
- 📚 **Documentation**: Quick start guide, API reference, SDKs, webhooks
- 💰 **Pricing**: Three-tier pricing plans (Starter, Professional, Enterprise)
- 🌍 **Bilingual**: English and Simplified Chinese support
- 📱 **Responsive**: Mobile-friendly design

## Development

```bash
# Install dependencies
npm install

# Start development server (http://localhost:5175)
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Lint code
npm run lint
```

## Project Structure

```
src/
├── components/         # Reusable components
│   ├── Header/        # Site navigation
│   ├── Footer/        # Site footer
│   └── LanguageSwitch/ # Language switcher
├── pages/             # Page components
│   ├── Home/          # Landing page
│   ├── Products/      # Product features
│   ├── Docs/          # Documentation
│   └── Pricing/       # Pricing plans
├── i18n/              # Translation files
│   ├── index.ts       # i18n configuration
│   └── locales/       # Language files
│       ├── en.json    # English translations
│       └── zh-CN.json # Chinese translations
├── App.tsx            # Main app component
└── main.tsx           # Entry point
```

## Environment

- Node.js 18+
- npm 9+

## Links

- Admin Portal: http://localhost:5173
- Merchant Portal: http://localhost:5174
- Website: http://localhost:5175

## License

Commercial
