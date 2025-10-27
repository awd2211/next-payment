# Payment Platform - Merchant Portal

Self-service merchant dashboard for payment platform merchants. Built with React 18, TypeScript, Vite, and Ant Design. Enables merchants to manage their payment operations independently.

## Features

### 🎯 Core Capabilities

- **Self-service Registration** - Merchant onboarding with KYC submission
- **API Management** - API key generation, webhook configuration, rate limits
- **Transaction Management** - Payment and order search, detailed view, export
- **Settlement Reports** - Settlement records, withdrawal requests, balance tracking
- **Analytics Dashboard** - Transaction trends, success rate, revenue statistics
- **Reconciliation Center** - Daily/monthly reconciliation, discrepancy handling
- **Channel Configuration** - Stripe/PayPal/Alipay integration settings
- **Developer Tools** - API documentation, sandbox environment, SDK download

### 🔒 Security Features

- **JWT Authentication** - Secure token-based authentication
- **Tenant Isolation** - Each merchant can only access their own data
- **API Key Management** - Secure API key generation and rotation
- **2FA Support** - Optional two-factor authentication
- **Data Masking** - Sensitive data protection
- **Webhook Signature** - Secure webhook verification

### 🌍 Multi-language Support

Internationalization support with react-i18next (configurable languages)

## Tech Stack

- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite 5 (fast HMR, optimized builds)
- **UI Library**: Ant Design 5.15 + @ant-design/icons
- **Charts**: @ant-design/charts (based on G2Plot)
- **State Management**: Zustand 4.5
- **HTTP Client**: Axios with interceptors
- **Routing**: React Router v6
- **i18n**: react-i18next
- **Code Quality**: ESLint + Prettier

## Quick Start

### Prerequisites

- **Node.js** 18+
- **npm** 9+
- **Backend services** running (Merchant BFF on port 40023)

### Installation

```bash
# Navigate to merchant-portal directory
cd frontend/merchant-portal

# Install dependencies
npm install
```

### Development

```bash
# Start development server (http://localhost:5174)
npm run dev

# The app will automatically open in your browser
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
merchant-portal/
├── src/
│   ├── pages/                  # Page components
│   │   ├── Dashboard/          # Analytics dashboard with charts
│   │   ├── Login/              # Login page
│   │   ├── Register/           # Merchant registration
│   │   ├── Payments/           # Payment management
│   │   ├── Orders/             # Order management
│   │   ├── Settlements/        # Settlement reports
│   │   ├── Withdrawals/        # Withdrawal requests
│   │   ├── Reconciliation/     # Reconciliation center
│   │   ├── APIKeys/            # API key management
│   │   ├── Webhooks/           # Webhook configuration
│   │   ├── Channels/           # Payment channel config
│   │   ├── Profile/            # Merchant profile
│   │   └── Developer/          # Developer tools & docs
│   │
│   ├── components/             # Reusable components
│   │   ├── Header/             # App header
│   │   ├── Sidebar/            # Navigation sidebar
│   │   ├── LanguageSwitch/     # Language switcher
│   │   ├── UserMenu/           # User dropdown menu
│   │   └── ...                 # Other shared components
│   │
│   ├── services/               # API services
│   │   ├── api.ts              # Axios instance with interceptors
│   │   ├── authService.ts      # Authentication APIs
│   │   ├── paymentService.ts   # Payment APIs
│   │   ├── orderService.ts     # Order APIs
│   │   ├── settlementService.ts # Settlement APIs
│   │   └── ...                 # Other service modules
│   │
│   ├── stores/                 # Zustand state stores
│   │   ├── authStore.ts        # Authentication state
│   │   ├── merchantStore.ts    # Merchant state
│   │   └── ...                 # Other stores
│   │
│   ├── hooks/                  # Custom React hooks
│   │   ├── useAuth.ts          # Authentication hook
│   │   ├── useMerchant.ts      # Merchant data hook
│   │   └── ...                 # Other custom hooks
│   │
│   ├── i18n/                   # Internationalization
│   │   ├── index.ts            # i18n configuration
│   │   └── locales/            # Translation files
│   │       ├── en.json         # English
│   │       ├── zh-CN.json      # Simplified Chinese
│   │       └── ...             # Other languages
│   │
│   ├── types/                  # TypeScript type definitions
│   │   ├── payment.ts          # Payment types
│   │   ├── order.ts            # Order types
│   │   ├── settlement.ts       # Settlement types
│   │   └── ...                 # Other type definitions
│   │
│   ├── utils/                  # Utility functions
│   │   ├── format.ts           # Formatting utilities
│   │   ├── validation.ts       # Validation utilities
│   │   └── ...                 # Other utilities
│   │
│   ├── App.tsx                 # Main app component
│   ├── main.tsx                # Entry point
│   └── index.css               # Global styles
│
├── public/                     # Static assets
├── index.html                  # HTML template
├── vite.config.ts              # Vite configuration
├── tsconfig.json               # TypeScript configuration
├── package.json                # Dependencies
└── README.md                   # This file
```

## Key Pages

### Dashboard
- Real-time merchant analytics
- Transaction volume, revenue trends
- Success rate, channel performance
- Top products, recent transactions
- Interactive charts and graphs

### Merchant Registration
- Self-service merchant onboarding
- Business information form
- KYC document upload
- Company verification
- Registration status tracking

### Payment Management
- Payment transaction list with search
- Payment detail view
- Refund initiation
- Transaction export (CSV, Excel)
- Real-time status updates

### Order Management
- Order list with filters
- Order detail view
- Order status tracking
- Bulk export capabilities

### Settlement Reports
- Settlement record list
- Settlement detail view
- Balance tracking
- Withdrawal request creation
- Settlement history

### API Key Management
- API key generation
- API key rotation
- Key permission configuration
- Usage statistics
- Sandbox/Production keys

### Webhook Configuration
- Webhook URL setup
- Event subscription
- Webhook signature verification
- Retry configuration
- Webhook logs and debugging

### Channel Configuration
- Stripe integration setup
- PayPal account linking
- Alipay configuration
- Channel fee rates
- Channel status monitoring

### Developer Tools
- API documentation viewer
- SDK download links
- Code examples
- Sandbox environment
- API testing tools

## API Integration

### Base URL Configuration

```typescript
// src/services/api.ts
const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:40023/api/v1',
  timeout: 10000,
});
```

### Environment Variables

Create `.env` file in project root:

```bash
# API Base URL (Merchant BFF Service)
VITE_API_BASE_URL=http://localhost:40023/api/v1

# App Configuration
VITE_APP_TITLE=Payment Platform Merchant Portal
VITE_APP_VERSION=1.0.0

# Feature Flags
VITE_ENABLE_2FA=true
VITE_ENABLE_SANDBOX=true
```

### Authentication Flow

```typescript
// Login
const response = await authService.login(email, password);
localStorage.setItem('token', response.data.token);
localStorage.setItem('merchantId', response.data.merchantId);

// Add token to requests
api.interceptors.request.use(config => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle 401 responses
api.interceptors.response.use(
  response => response,
  error => {
    if (error.response?.status === 401) {
      localStorage.clear();
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
```

### Tenant Isolation

All API requests automatically include merchant context from JWT token. The backend enforces tenant isolation:

```typescript
// Backend automatically extracts merchant_id from JWT
// Merchants can only access their own data
const response = await paymentService.getList({
  page: 1,
  pageSize: 20,
  // merchant_id automatically injected by backend
});
```

## Payment Integration

### Creating a Payment

```typescript
import { paymentService } from '@/services/paymentService';

const createPayment = async () => {
  const response = await paymentService.create({
    order_no: 'ORDER-123',
    amount: 10000,  // Amount in cents (100.00 USD)
    currency: 'USD',
    channel: 'stripe',
    description: 'Payment for order #123',
    return_url: 'https://merchant.com/payment/return',
    notify_url: 'https://merchant.com/payment/notify',
  });

  // Redirect user to payment URL
  window.location.href = response.data.payment_url;
};
```

### Handling Webhooks

```typescript
// Configure webhook URL in Webhook settings page
const webhookUrl = 'https://merchant.com/webhooks/payment';

// Backend will send webhook notifications for:
// - payment.succeeded
// - payment.failed
// - refund.succeeded
// - refund.failed

// Webhook payload includes signature for verification
```

## Charts and Visualization

Using `@ant-design/charts`:

```typescript
import { Line, Column, Pie } from '@ant-design/charts';

// Transaction trend
<Line
  data={transactionData}
  xField="date"
  yField="amount"
  smooth
/>

// Revenue by channel
<Column
  data={channelRevenue}
  xField="channel"
  yField="revenue"
  seriesField="status"
/>

// Payment method distribution
<Pie
  data={methodDistribution}
  angleField="value"
  colorField="method"
  label={{ type: 'outer' }}
/>
```

## Common Tasks

### Add New Page

1. Create page component:
```bash
mkdir src/pages/NewFeature
touch src/pages/NewFeature/index.tsx
```

2. Add route in `App.tsx`:
```typescript
import NewFeature from './pages/NewFeature';

<Route path="/new-feature" element={<NewFeature />} />
```

3. Add menu item in `Sidebar` component

### Add New API Service

```typescript
// src/services/newService.ts
import api from './api';

export const newService = {
  getList: (params: any) => api.get('/new-endpoint', { params }),
  create: (data: any) => api.post('/new-endpoint', data),
  update: (id: string, data: any) => api.put(`/new-endpoint/${id}`, data),
  delete: (id: string) => api.delete(`/new-endpoint/${id}`),
};
```

### Format Currency

```typescript
// src/utils/format.ts
export const formatCurrency = (amount: number, currency: string = 'USD') => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: currency,
  }).format(amount / 100); // Convert cents to dollars
};

// Usage
formatCurrency(10000, 'USD'); // "$100.00"
```

### Export Data

```typescript
import { exportService } from '@/services/exportService';

const exportTransactions = async () => {
  const response = await exportService.exportPayments({
    start_date: '2025-01-01',
    end_date: '2025-01-31',
    format: 'csv', // or 'excel'
  });

  // Download file
  const url = window.URL.createObjectURL(new Blob([response.data]));
  const link = document.createElement('a');
  link.href = url;
  link.download = 'transactions.csv';
  link.click();
};
```

## Security Best Practices

### API Key Management

- **Never expose API keys** in frontend code
- Use API keys only for server-to-server communication
- Rotate keys regularly
- Use sandbox keys for testing

### Webhook Verification

```typescript
// Backend verifies webhook signatures
// Frontend only displays webhook logs and configuration
```

### Data Masking

Sensitive data is automatically masked by backend:
- Card numbers: `4242 **** **** 1234`
- Email: `u****r@example.com`
- Phone: `138****5678`

## Troubleshooting

### Port Already in Use

```bash
# Kill process on port 5174
lsof -ti:5174 | xargs kill -9

# Or use different port
npm run dev -- --port 5175
```

### Build Errors

```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install

# Clear Vite cache
rm -rf node_modules/.vite
npm run dev
```

### API Connection Issues

1. Check Merchant BFF service is running:
```bash
curl http://localhost:40023/health
```

2. Verify API base URL in `.env`

3. Check browser console for CORS errors

4. Verify JWT token is valid and not expired

### Tenant Isolation Issues

If you see data from other merchants:
1. Clear localStorage and login again
2. Check JWT token contains correct merchant_id
3. Verify backend enforces tenant isolation

## Development Tips

### Hot Module Replacement (HMR)

Vite provides fast HMR. Your changes will reflect immediately without full page reload.

### TypeScript

- Define types for all API responses
- Use interfaces for merchant data models
- Avoid `any` type when possible

### State Management

```typescript
// Use Zustand for global state
import create from 'zustand';

interface MerchantStore {
  merchant: Merchant | null;
  setMerchant: (merchant: Merchant) => void;
}

export const useMerchantStore = create<MerchantStore>((set) => ({
  merchant: null,
  setMerchant: (merchant) => set({ merchant }),
}));
```

### Performance Optimization

- Implement pagination for large lists
- Lazy load heavy components
- Use React.memo() for expensive renders
- Cache API responses when appropriate

## Testing

### Unit Tests

```bash
# Run tests
npm test

# Watch mode
npm test -- --watch

# Coverage report
npm test -- --coverage
```

### E2E Tests

```bash
# Using Playwright or Cypress
npm run test:e2e
```

## Deployment

### Production Build

```bash
# Build for production
npm run build

# Output in dist/ directory
# Optimized, minified, code-split
```

### Deploy to Server

```bash
# Using nginx
cp -r dist/* /var/www/html/merchant-portal/

# Configure nginx
server {
  listen 80;
  server_name merchant.example.com;
  root /var/www/html/merchant-portal;

  location / {
    try_files $uri $uri/ /index.html;
  }
}
```

### Docker Deployment

```dockerfile
FROM node:18-alpine as build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

## License

Commercial License

## Support

- **Backend API**: http://localhost:40023/swagger/index.html
- **Developer Docs**: Available in the Developer Tools page
- **Documentation**: See project root `README.md` and `CLAUDE.md`
- **Issues**: Report in GitHub issue tracker
