import { useTranslation } from 'react-i18next';
import { Card, Row, Col, Tabs, Tag, Button, Collapse, Timeline, Steps } from 'antd';
import {
  RocketOutlined,
  ApiOutlined,
  CodeOutlined,
  BellOutlined,
  CheckCircleOutlined,
  CopyOutlined,
  GithubOutlined,
  BookOutlined,
  ToolOutlined,
  SafetyOutlined,
  ThunderboltOutlined,
  GlobalOutlined,
} from '@ant-design/icons';
import { useState } from 'react';
import './style.css';

const { TabPane } = Tabs;
const { Panel } = Collapse;

const Docs = () => {
  const { t } = useTranslation();
  const [copiedCode, setCopiedCode] = useState<string | null>(null);

  const copyToClipboard = (code: string, id: string) => {
    navigator.clipboard.writeText(code);
    setCopiedCode(id);
    setTimeout(() => setCopiedCode(null), 2000);
  };

  const docSections = [
    {
      icon: <RocketOutlined />,
      title: t('docs.quickStart.title') || 'Quick Start',
      description: t('docs.quickStart.description') || 'Get started in 5 minutes',
      link: '#quick-start',
      gradient: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    },
    {
      icon: <ApiOutlined />,
      title: t('docs.apiReference.title') || 'API Reference',
      description: t('docs.apiReference.description') || 'Complete API documentation',
      link: '#api-reference',
      gradient: 'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)',
    },
    {
      icon: <CodeOutlined />,
      title: t('docs.sdks.title') || 'SDKs & Libraries',
      description: t('docs.sdks.description') || 'Official SDKs for all languages',
      link: '#sdks',
      gradient: 'linear-gradient(135deg, #fa709a 0%, #fee140 100%)',
    },
    {
      icon: <BellOutlined />,
      title: t('docs.webhooks.title') || 'Webhooks',
      description: t('docs.webhooks.description') || 'Real-time event notifications',
      link: '#webhooks',
      gradient: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
    },
  ];

  const features = [
    { icon: <ThunderboltOutlined />, text: 'High Performance', color: '#667eea' },
    { icon: <SafetyOutlined />, text: 'Secure by Default', color: '#43e97b' },
    { icon: <GlobalOutlined />, text: 'Global Coverage', color: '#fa709a' },
    { icon: <ToolOutlined />, text: 'Easy Integration', color: '#4facfe' },
  ];

  const sdkLanguages = [
    {
      name: 'Node.js',
      icon: '🟢',
      install: 'npm install @payment-platform/sdk',
      version: 'v2.3.0',
      docs: '#nodejs',
    },
    {
      name: 'Python',
      icon: '🐍',
      install: 'pip install payment-platform',
      version: 'v2.1.5',
      docs: '#python',
    },
    {
      name: 'PHP',
      icon: '🐘',
      install: 'composer require payment-platform/sdk',
      version: 'v2.0.8',
      docs: '#php',
    },
    {
      name: 'Java',
      icon: '☕',
      install: 'implementation "com.payment-platform:sdk:2.2.1"',
      version: 'v2.2.1',
      docs: '#java',
    },
    {
      name: 'Go',
      icon: '🔷',
      install: 'go get github.com/payment-platform/sdk-go',
      version: 'v2.1.0',
      docs: '#go',
    },
    {
      name: 'Ruby',
      icon: '💎',
      install: 'gem install payment_platform',
      version: 'v2.0.3',
      docs: '#ruby',
    },
  ];

  const apiEndpoints = [
    {
      method: 'POST',
      path: '/api/v1/payments',
      description: 'Create a new payment',
      auth: 'Required',
    },
    {
      method: 'GET',
      path: '/api/v1/payments/:id',
      description: 'Get payment details',
      auth: 'Required',
    },
    {
      method: 'POST',
      path: '/api/v1/payments/:id/refund',
      description: 'Refund a payment',
      auth: 'Required',
    },
    {
      method: 'GET',
      path: '/api/v1/orders',
      description: 'List all orders',
      auth: 'Required',
    },
  ];

  const webhookEvents = [
    { event: 'payment.created', description: 'Payment created successfully' },
    { event: 'payment.succeeded', description: 'Payment completed' },
    { event: 'payment.failed', description: 'Payment failed' },
    { event: 'payment.refunded', description: 'Payment refunded' },
    { event: 'order.created', description: 'Order created' },
    { event: 'order.completed', description: 'Order completed' },
  ];

  const quickStartCode = `// Install SDK
npm install @payment-platform/sdk

// Initialize client
import { PaymentClient } from '@payment-platform/sdk';

const client = new PaymentClient({
  apiKey: 'your_api_key',
  apiSecret: 'your_api_secret',
  environment: 'production' // or 'sandbox'
});

// Create payment
const payment = await client.payments.create({
  amount: 10000, // Amount in cents ($100.00)
  currency: 'USD',
  orderNo: 'ORDER-' + Date.now(),
  notifyUrl: 'https://your-domain.com/webhook',
  returnUrl: 'https://your-domain.com/return',
  description: 'Product purchase'
});

console.log('Payment URL:', payment.paymentUrl);
console.log('Payment ID:', payment.id);`;

  const webhookCode = `// Express.js example
const express = require('express');
const app = express();

app.post('/webhook', express.raw({type: 'application/json'}), (req, res) => {
  const signature = req.headers['x-payment-signature'];

  // Verify webhook signature
  const isValid = client.webhooks.verify(
    req.body,
    signature,
    process.env.WEBHOOK_SECRET
  );

  if (!isValid) {
    return res.status(400).send('Invalid signature');
  }

  const event = JSON.parse(req.body);

  // Handle the event
  switch (event.type) {
    case 'payment.succeeded':
      console.log('Payment succeeded:', event.data.id);
      // Update your database
      break;
    case 'payment.failed':
      console.log('Payment failed:', event.data.id);
      // Notify customer
      break;
    default:
      console.log('Unhandled event type:', event.type);
  }

  res.json({received: true});
});`;

  return (
    <div className="docs-page">
      {/* Hero Section */}
      <div className="docs-hero">
        <div className="docs-hero-content">
          <BookOutlined className="hero-icon" />
          <h1 className="hero-title">Developer Documentation</h1>
          <p className="hero-subtitle">
            Everything you need to integrate our payment platform
          </p>
          <div className="hero-features">
            {features.map((feature, index) => (
              <div key={index} className="hero-feature-item">
                <span className="hero-feature-icon" style={{ color: feature.color }}>
                  {feature.icon}
                </span>
                <span>{feature.text}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Quick Navigation */}
      <section className="docs-nav-section">
        <div className="section-container">
          <Row gutter={[24, 24]}>
            {docSections.map((section, index) => (
              <Col xs={24} sm={12} md={6} key={index}>
                <Card
                  className="doc-nav-card"
                  hoverable
                  onClick={() => (window.location.hash = section.link)}
                >
                  <div className="doc-nav-icon-wrapper" style={{ background: section.gradient }}>
                    <div className="doc-nav-icon">{section.icon}</div>
                  </div>
                  <h3 className="doc-nav-title">{section.title}</h3>
                  <p className="doc-nav-description">{section.description}</p>
                </Card>
              </Col>
            ))}
          </Row>
        </div>
      </section>

      {/* Main Content */}
      <section className="docs-content-section">
        <div className="section-container">
          {/* Quick Start */}
          <div id="quick-start" className="doc-section">
            <h2 className="section-title">
              <RocketOutlined /> Quick Start Guide
            </h2>
            <Card className="content-card">
              <Steps
                direction="vertical"
                current={-1}
                items={[
                  {
                    title: 'Create Account',
                    description: 'Sign up for a merchant account and get your API credentials',
                    icon: <CheckCircleOutlined />,
                  },
                  {
                    title: 'Install SDK',
                    description: 'Choose your preferred language and install our SDK',
                    icon: <CodeOutlined />,
                  },
                  {
                    title: 'Make First Payment',
                    description: 'Create your first payment with just a few lines of code',
                    icon: <ThunderboltOutlined />,
                  },
                  {
                    title: 'Test & Go Live',
                    description: 'Test in sandbox mode, then switch to production',
                    icon: <RocketOutlined />,
                  },
                ]}
              />

              <div className="code-example">
                <div className="code-header">
                  <span className="code-title">Example: Create Payment</span>
                  <Button
                    type="text"
                    size="small"
                    icon={copiedCode === 'quickstart' ? <CheckCircleOutlined /> : <CopyOutlined />}
                    onClick={() => copyToClipboard(quickStartCode, 'quickstart')}
                  >
                    {copiedCode === 'quickstart' ? 'Copied!' : 'Copy'}
                  </Button>
                </div>
                <pre className="code-block">{quickStartCode}</pre>
              </div>
            </Card>
          </div>

          {/* API Reference */}
          <div id="api-reference" className="doc-section">
            <h2 className="section-title">
              <ApiOutlined /> API Reference
            </h2>
            <Card className="content-card">
              <p className="section-intro">
                Our RESTful API uses standard HTTP methods and returns JSON responses.
                All API requests require authentication using API keys.
              </p>

              <h3>Base URL</h3>
              <div className="api-base-url">
                <Tag color="blue">Production</Tag>
                <code>https://api.payment-platform.com</code>
              </div>
              <div className="api-base-url">
                <Tag color="orange">Sandbox</Tag>
                <code>https://sandbox-api.payment-platform.com</code>
              </div>

              <h3>Common Endpoints</h3>
              <div className="api-endpoints-table">
                {apiEndpoints.map((endpoint, index) => (
                  <div key={index} className="api-endpoint-row">
                    <Tag color={
                      endpoint.method === 'GET' ? 'blue' :
                      endpoint.method === 'POST' ? 'green' :
                      endpoint.method === 'PUT' ? 'orange' : 'red'
                    }>
                      {endpoint.method}
                    </Tag>
                    <code className="api-path">{endpoint.path}</code>
                    <span className="api-description">{endpoint.description}</span>
                    <Tag>{endpoint.auth}</Tag>
                  </div>
                ))}
              </div>

              <Button type="primary" icon={<BookOutlined />} size="large" style={{ marginTop: 24 }}>
                View Full API Reference
              </Button>
            </Card>
          </div>

          {/* SDKs */}
          <div id="sdks" className="doc-section">
            <h2 className="section-title">
              <CodeOutlined /> SDKs & Libraries
            </h2>
            <Card className="content-card">
              <p className="section-intro">
                Official SDKs for all major programming languages with full TypeScript support,
                automatic retries, and built-in error handling.
              </p>

              <Row gutter={[16, 16]}>
                {sdkLanguages.map((sdk, index) => (
                  <Col xs={24} sm={12} md={8} key={index}>
                    <Card className="sdk-card" hoverable>
                      <div className="sdk-header">
                        <span className="sdk-icon">{sdk.icon}</span>
                        <div>
                          <h4 className="sdk-name">{sdk.name}</h4>
                          <Tag color="blue">{sdk.version}</Tag>
                        </div>
                      </div>
                      <div className="sdk-install">
                        <code>{sdk.install}</code>
                        <Button
                          type="text"
                          size="small"
                          icon={<CopyOutlined />}
                          onClick={() => copyToClipboard(sdk.install, sdk.name)}
                        />
                      </div>
                      <Button type="link" block>
                        View Documentation →
                      </Button>
                    </Card>
                  </Col>
                ))}
              </Row>

              <div style={{ marginTop: 32, textAlign: 'center' }}>
                <Button icon={<GithubOutlined />} size="large">
                  View on GitHub
                </Button>
              </div>
            </Card>
          </div>

          {/* Webhooks */}
          <div id="webhooks" className="doc-section">
            <h2 className="section-title">
              <BellOutlined /> Webhooks
            </h2>
            <Card className="content-card">
              <p className="section-intro">
                Webhooks allow you to receive real-time notifications about events in your account.
                Configure webhook endpoints in your merchant dashboard.
              </p>

              <h3>Webhook Events</h3>
              <Collapse accordion className="webhook-events">
                {webhookEvents.map((webhook, index) => (
                  <Panel
                    key={index}
                    header={
                      <div className="webhook-header">
                        <code>{webhook.event}</code>
                        <span className="webhook-desc">{webhook.description}</span>
                      </div>
                    }
                  >
                    <p>This event is triggered when {webhook.description.toLowerCase()}.</p>
                    <h4>Example Payload:</h4>
                    <pre className="code-block">
{`{
  "type": "${webhook.event}",
  "data": {
    "id": "pay_1234567890",
    "amount": 10000,
    "currency": "USD",
    "status": "${webhook.event.split('.')[1]}",
    "created_at": "2024-01-15T10:30:00Z"
  }
}`}
                    </pre>
                  </Panel>
                ))}
              </Collapse>

              <h3 style={{ marginTop: 32 }}>Webhook Implementation</h3>
              <div className="code-example">
                <div className="code-header">
                  <span className="code-title">Example: Webhook Handler</span>
                  <Button
                    type="text"
                    size="small"
                    icon={copiedCode === 'webhook' ? <CheckCircleOutlined /> : <CopyOutlined />}
                    onClick={() => copyToClipboard(webhookCode, 'webhook')}
                  >
                    {copiedCode === 'webhook' ? 'Copied!' : 'Copy'}
                  </Button>
                </div>
                <pre className="code-block">{webhookCode}</pre>
              </div>

              <div className="webhook-security">
                <SafetyOutlined style={{ fontSize: 24, color: '#52c41a' }} />
                <div>
                  <h4>Security Best Practices</h4>
                  <ul>
                    <li>Always verify webhook signatures before processing</li>
                    <li>Use HTTPS endpoints only</li>
                    <li>Implement idempotency to handle duplicate events</li>
                    <li>Return 200 OK quickly, process asynchronously</li>
                  </ul>
                </div>
              </div>
            </Card>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="docs-cta">
        <div className="section-container">
          <h2 className="cta-title">Ready to Start Building?</h2>
          <p className="cta-description">
            Get your API credentials and start integrating in minutes
          </p>
          <div className="cta-actions">
            <Button type="primary" size="large" icon={<RocketOutlined />}>
              Get API Keys
            </Button>
            <Button size="large" icon={<BookOutlined />}>
              Explore Examples
            </Button>
          </div>
        </div>
      </section>
    </div>
  );
};

export default Docs;
