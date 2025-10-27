# Payment Platform - Admin Portal

Enterprise admin dashboard for payment platform operators. Built with React 18, TypeScript, Vite, and Ant Design. Supports 12 languages for global operations.

## Features

### 🎯 Core Capabilities

- **Merchant Management** - Register, approve, KYC verification, freeze/unfreeze merchants
- **Payment Monitoring** - Real-time transaction flow, payment search, anomaly alerts
- **Risk Management** - Risk rule configuration, blacklist management, fraud scoring
- **Order Management** - Order search, status tracking, detailed information
- **Settlement Management** - Settlement approval, records, reconciliation
- **Financial Analytics** - Revenue statistics, transaction trends, channel distribution
- **System Administration** - User management, role permissions (6 roles), audit logs
- **Configuration Management** - System settings, email templates, feature flags

### 🌍 Multi-language Support

Supports **12 languages** with react-i18next:
- **English** (en)
- **简体中文** (zh-CN)
- **繁體中文** (zh-TW)
- **日本語** (ja)
- **한국어** (ko)
- **Español** (es)
- **Français** (fr)
- **Deutsch** (de)
- **Português** (pt)
- **Русский** (ru)
- **العربية** (ar)
- **हिन्दी** (hi)

### 🔒 Security Features

- **JWT Authentication** - Secure token-based authentication
- **Role-Based Access Control (RBAC)** - 6 role types (super_admin, operator, finance, risk_manager, support, auditor)
- **2FA/TOTP Support** - Two-factor authentication for sensitive operations
- **Audit Logging** - Complete operation trail
- **Data Masking** - PII protection (phone, email, ID card, bank card)

## Tech Stack

- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite 5 (fast HMR, optimized builds)
- **UI Library**: Ant Design 5.15 + @ant-design/icons
- **Charts**: @ant-design/charts (based on G2Plot)
- **State Management**: Zustand 4.5
- **HTTP Client**: Axios with interceptors
- **Routing**: React Router v6
- **i18n**: react-i18next (12 languages)
- **Code Quality**: ESLint + Prettier

## Quick Start

### Prerequisites

- **Node.js** 18+
- **npm** 9+
- **Backend services** running (Admin BFF on port 40001)

### Installation

```bash
# Navigate to admin-portal directory
cd frontend/admin-portal

# Install dependencies
npm install
```

### Development

```bash
# Start development server (http://localhost:5173)
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
admin-portal/
├── src/
│   ├── pages/                  # Page components
│   │   ├── Dashboard/          # Analytics dashboard with charts
│   │   ├── Login/              # Login page with 2FA
│   │   ├── Merchants/          # Merchant management
│   │   ├── Payments/           # Payment monitoring
│   │   ├── Orders/             # Order management
│   │   ├── RiskControl/        # Risk management
│   │   ├── Settlements/        # Settlement management
│   │   ├── Users/              # User management
│   │   ├── Roles/              # Role & permission management
│   │   ├── AuditLogs/          # Audit log viewer
│   │   ├── SystemConfig/       # System configuration
│   │   └── EmailTemplates/     # Email template management
│   │
│   ├── components/             # Reusable components
│   │   ├── Header/             # App header with language switch
│   │   ├── Sidebar/            # Navigation sidebar
│   │   ├── LanguageSwitch/     # Language switcher
│   │   ├── UserMenu/           # User dropdown menu
│   │   └── ...                 # Other shared components
│   │
│   ├── services/               # API services
│   │   ├── api.ts              # Axios instance with interceptors
│   │   ├── authService.ts      # Authentication APIs
│   │   ├── merchantService.ts  # Merchant APIs
│   │   ├── paymentService.ts   # Payment APIs
│   │   └── ...                 # Other service modules
│   │
│   ├── stores/                 # Zustand state stores
│   │   ├── authStore.ts        # Authentication state
│   │   ├── userStore.ts        # User state
│   │   └── ...                 # Other stores
│   │
│   ├── hooks/                  # Custom React hooks
│   │   ├── useAuth.ts          # Authentication hook
│   │   ├── usePagination.ts    # Pagination hook
│   │   └── ...                 # Other custom hooks
│   │
│   ├── i18n/                   # Internationalization
│   │   ├── index.ts            # i18n configuration
│   │   └── locales/            # Translation files
│   │       ├── en.json         # English
│   │       ├── zh-CN.json      # Simplified Chinese
│   │       ├── zh-TW.json      # Traditional Chinese
│   │       ├── ja.json         # Japanese
│   │       ├── ko.json         # Korean
│   │       ├── es.json         # Spanish
│   │       ├── fr.json         # French
│   │       ├── de.json         # German
│   │       ├── pt.json         # Portuguese
│   │       ├── ru.json         # Russian
│   │       ├── ar.json         # Arabic
│   │       └── hi.json         # Hindi
│   │
│   ├── types/                  # TypeScript type definitions
│   │   ├── merchant.ts         # Merchant types
│   │   ├── payment.ts          # Payment types
│   │   ├── order.ts            # Order types
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
- Real-time analytics with interactive charts
- GMV, transaction count, success rate
- Channel distribution, payment trends
- Top merchants, recent transactions

### Merchant Management
- Merchant list with search and filters
- Merchant approval workflow
- KYC verification and document review
- Merchant freeze/unfreeze operations
- Merchant detail view with complete information

### Payment Monitoring
- Real-time payment transaction list
- Advanced search (merchant, order, amount, status, date)
- Payment detail view
- Refund operations
- Transaction export

### Risk Management
- Risk rule configuration
- Blacklist management (IP, device, card, user)
- Risk scoring and fraud detection
- Alert configuration

### Settlement Management
- Settlement approval workflow
- Settlement records with search
- Reconciliation reports
- Settlement detail view

### User & Role Management
- User creation and management
- 6 role types (super_admin, operator, finance, risk_manager, support, auditor)
- Permission assignment
- Role-based access control

### Audit Logs
- Complete operation trail
- Search by user, operation, time
- Detailed action logs
- Export capabilities

## API Integration

### Base URL Configuration

```typescript
// src/services/api.ts
const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:40001/api/v1',
  timeout: 10000,
});
```

### Environment Variables

Create `.env` file in project root:

```bash
# API Base URL (Admin BFF Service)
VITE_API_BASE_URL=http://localhost:40001/api/v1

# App Configuration
VITE_APP_TITLE=Payment Platform Admin
VITE_APP_VERSION=1.0.0
```

### Authentication Flow

```typescript
// Login
const response = await authService.login(username, password);
localStorage.setItem('token', response.data.token);

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
      // Redirect to login
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
```

## Role-Based Access Control

### 6 Role Types

1. **super_admin** - Full access (wildcard `*` permission)
2. **operator** - Merchant & order management, KYC approval
3. **finance** - Accounting, settlements, withdrawals
4. **risk_manager** - Risk control, disputes, fraud detection
5. **support** - Read-only access (customer support)
6. **auditor** - Audit logs and analytics viewing

### Permission Check

```typescript
// Check if user has permission
const hasPermission = (requiredPermission: string) => {
  const user = useUserStore(state => state.user);
  if (user.role === 'super_admin') return true;

  return user.permissions.some(p =>
    p === requiredPermission || p.endsWith('*')
  );
};

// Usage in components
{hasPermission('merchant:approve') && (
  <Button onClick={handleApprove}>Approve</Button>
)}
```

## Internationalization

### Add New Translation

1. Add key-value pairs to `src/i18n/locales/{lang}.json`:

```json
{
  "dashboard": {
    "title": "Dashboard",
    "totalRevenue": "Total Revenue",
    "totalTransactions": "Total Transactions"
  }
}
```

2. Use in components:

```typescript
import { useTranslation } from 'react-i18next';

function Dashboard() {
  const { t } = useTranslation();

  return (
    <div>
      <h1>{t('dashboard.title')}</h1>
      <p>{t('dashboard.totalRevenue')}</p>
    </div>
  );
}
```

### Change Language

```typescript
import { useTranslation } from 'react-i18next';

function LanguageSwitch() {
  const { i18n } = useTranslation();

  const changeLanguage = (lang: string) => {
    i18n.changeLanguage(lang);
    localStorage.setItem('language', lang);
  };

  return (
    <Select defaultValue={i18n.language} onChange={changeLanguage}>
      <Option value="en">English</Option>
      <Option value="zh-CN">简体中文</Option>
      {/* ... other languages */}
    </Select>
  );
}
```

## Charts and Visualization

Using `@ant-design/charts` (based on G2Plot):

```typescript
import { Line, Column, Pie } from '@ant-design/charts';

// Line chart
<Line
  data={transactionData}
  xField="date"
  yField="amount"
  seriesField="channel"
/>

// Column chart
<Column
  data={revenueData}
  xField="month"
  yField="revenue"
  label={{ position: 'top' }}
/>

// Pie chart
<Pie
  data={channelDistribution}
  angleField="value"
  colorField="channel"
  radius={0.8}
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

1. Create service file:
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

2. Use in components:
```typescript
import { newService } from '@/services/newService';

const fetchData = async () => {
  const response = await newService.getList({ page: 1 });
  setData(response.data);
};
```

## Troubleshooting

### Port Already in Use

```bash
# Kill process on port 5173
lsof -ti:5173 | xargs kill -9

# Or use different port
npm run dev -- --port 5174
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

1. Check backend services are running:
```bash
curl http://localhost:40001/health
```

2. Verify API base URL in `.env`

3. Check browser console for CORS errors

### Language Not Switching

1. Check language file exists in `src/i18n/locales/`
2. Verify language code in i18n configuration
3. Clear browser cache and localStorage

## Development Tips

### Hot Module Replacement (HMR)

Vite provides fast HMR. Your changes will reflect immediately without full page reload.

### TypeScript

- Always define types for API responses
- Use interfaces for complex objects
- Avoid `any` type when possible

### Code Splitting

Vite automatically code splits by route. For manual splitting:

```typescript
const Dashboard = lazy(() => import('./pages/Dashboard'));
```

### Performance Optimization

- Use `React.memo()` for expensive components
- Implement pagination for large lists
- Lazy load images and heavy components
- Use Zustand for global state (faster than Redux)

## Deployment

### Production Build

```bash
# Build for production
npm run build

# Output in dist/ directory
# Optimized, minified, code-split

# Preview build locally
npm run preview
```

### Deploy to Server

```bash
# Using nginx
cp -r dist/* /var/www/html/admin-portal/

# Using Docker
docker build -t admin-portal .
docker run -p 80:80 admin-portal
```

### Environment-Specific Builds

```bash
# Production build
npm run build

# Staging build
VITE_API_BASE_URL=https://staging-api.example.com npm run build
```

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

## License

Commercial License

## Support

- **Backend API**: http://localhost:40001/swagger/index.html
- **Documentation**: See project root `README.md` and `CLAUDE.md`
- **Issues**: Report in GitHub issue tracker
