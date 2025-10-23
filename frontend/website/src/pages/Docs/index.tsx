import { useTranslation } from 'react-i18next';
import { Card, Row, Col } from 'antd';
import {
  RocketOutlined,
  ApiOutlined,
  CodeOutlined,
  BellOutlined,
} from '@ant-design/icons';
import './style.css';

const Docs = () => {
  const { t } = useTranslation();

  const docSections = [
    {
      icon: <RocketOutlined />,
      title: t('docs.quickStart.title'),
      description: t('docs.quickStart.description'),
      link: '#quick-start',
    },
    {
      icon: <ApiOutlined />,
      title: t('docs.apiReference.title'),
      description: t('docs.apiReference.description'),
      link: '#api-reference',
    },
    {
      icon: <CodeOutlined />,
      title: t('docs.sdks.title'),
      description: t('docs.sdks.description'),
      link: '#sdks',
    },
    {
      icon: <BellOutlined />,
      title: t('docs.webhooks.title'),
      description: t('docs.webhooks.description'),
      link: '#webhooks',
    },
  ];

  return (
    <div className="docs-page">
      <div className="docs-header">
        <h1 className="page-title">{t('docs.title')}</h1>
        <p className="page-subtitle">{t('docs.subtitle')}</p>
      </div>

      <div className="docs-container">
        <Row gutter={[24, 24]}>
          {docSections.map((section, index) => (
            <Col xs={24} sm={12} md={6} key={index}>
              <Card
                className="doc-card"
                hoverable
                onClick={() => (window.location.hash = section.link)}
              >
                <div className="doc-icon">{section.icon}</div>
                <h3 className="doc-title">{section.title}</h3>
                <p className="doc-description">{section.description}</p>
              </Card>
            </Col>
          ))}
        </Row>

        <div className="doc-content">
          <Card>
            <h2 id="quick-start">Quick Start</h2>
            <p>Get started with our API in just a few minutes:</p>
            <pre className="code-block">
              {`// Install SDK
npm install @payment-platform/sdk

// Initialize client
import { PaymentClient } from '@payment-platform/sdk';

const client = new PaymentClient({
  apiKey: 'your_api_key',
  apiSecret: 'your_api_secret'
});

// Create payment
const payment = await client.payments.create({
  amount: 10000, // Amount in cents
  currency: 'USD',
  orderNo: 'ORDER-001',
  notifyUrl: 'https://your-domain.com/webhook',
  returnUrl: 'https://your-domain.com/return'
});

console.log('Payment URL:', payment.paymentUrl);`}
            </pre>

            <h2 id="api-reference">API Reference</h2>
            <p>
              Complete API documentation is available in our{' '}
              <a href="#" target="_blank" rel="noopener noreferrer">
                API Reference
              </a>
              .
            </p>

            <h2 id="sdks">SDKs & Libraries</h2>
            <p>Official SDKs available for:</p>
            <ul>
              <li>Node.js / JavaScript</li>
              <li>Python</li>
              <li>PHP</li>
              <li>Java</li>
              <li>Go</li>
            </ul>

            <h2 id="webhooks">Webhooks</h2>
            <p>
              Receive real-time notifications about payment events. Configure your webhook
              endpoint in the merchant portal.
            </p>
          </Card>
        </div>
      </div>
    </div>
  );
};

export default Docs;
