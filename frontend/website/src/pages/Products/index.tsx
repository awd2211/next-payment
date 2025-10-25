import { useTranslation } from 'react-i18next';
import { Card, Row, Col, List, Tabs, Button, Timeline, Badge } from 'antd';
import {
  ApiOutlined,
  SafetyOutlined,
  DollarOutlined,
  DashboardOutlined,
  CheckCircleOutlined,
  ThunderboltOutlined,
  GlobalOutlined,
  LockOutlined,
  CloudOutlined,
  RocketOutlined,
  TeamOutlined,
} from '@ant-design/icons';
import './style.css';

const { TabPane } = Tabs;

const Products = () => {
  const { t } = useTranslation();

  const products = [
    {
      icon: <ApiOutlined />,
      title: t('products.gateway.title') || 'Payment Gateway',
      description: t('products.gateway.description') || 'Complete payment processing solution',
      features: t('products.gateway.features', { returnObjects: true }) as string[] || [
        'Multi-channel payment processing',
        'Real-time transaction monitoring',
        'Automatic reconciliation',
        'Fraud prevention',
      ],
      gradient: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      benefits: ['99.99% Uptime', 'PCI DSS Compliant', 'Global Coverage'],
    },
    {
      icon: <SafetyOutlined />,
      title: t('products.risk.title') || 'Risk Management',
      description: t('products.risk.description') || 'Advanced fraud detection and prevention',
      features: t('products.risk.features', { returnObjects: true }) as string[] || [
        'AI-powered fraud detection',
        'Real-time risk scoring',
        'Customizable rules engine',
        'Chargeback prevention',
      ],
      gradient: 'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)',
      benefits: ['Reduce Fraud 80%', 'Real-time Analysis', 'ML-Powered'],
    },
    {
      icon: <DollarOutlined />,
      title: t('products.settlement.title') || 'Settlement System',
      description: t('products.settlement.description') || 'Automated settlement and reconciliation',
      features: t('products.settlement.features', { returnObjects: true }) as string[] || [
        'Automated settlement processing',
        'Multi-currency support',
        'Flexible payout schedules',
        'Detailed financial reports',
      ],
      gradient: 'linear-gradient(135deg, #fa709a 0%, #fee140 100%)',
      benefits: ['T+1 Settlement', '32+ Currencies', 'Auto Reconciliation'],
    },
    {
      icon: <DashboardOutlined />,
      title: t('products.monitoring.title') || 'Monitoring & Analytics',
      description: t('products.monitoring.description') || 'Real-time monitoring and business intelligence',
      features: t('products.monitoring.features', { returnObjects: true }) as string[] || [
        'Real-time dashboards',
        'Custom reports',
        'Performance analytics',
        'Alert notifications',
      ],
      gradient: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
      benefits: ['Real-time Data', 'Custom Dashboards', 'Smart Alerts'],
    },
  ];

  const techStack = [
    { icon: <CloudOutlined />, name: 'Microservices Architecture', desc: '15 independent services' },
    { icon: <ThunderboltOutlined />, name: 'High Performance', desc: '10,000+ TPS' },
    { icon: <GlobalOutlined />, name: 'Global Coverage', desc: '150+ countries' },
    { icon: <LockOutlined />, name: 'Enterprise Security', desc: 'Bank-level encryption' },
  ];

  const integrations = [
    'Stripe', 'PayPal', 'Alipay', 'WeChat Pay',
    'Visa', 'Mastercard', 'UnionPay', 'JCB',
  ];

  return (
    <div className="products-page">
      <SEO
        title="Products - Payment Platform Solutions"
        description="Comprehensive payment solutions including Payment Gateway, Risk Management, Settlement System, and Real-time Monitoring. Enterprise-grade features for businesses of all sizes."
        keywords="payment products, payment gateway, risk management, settlement system, payment monitoring, payment solutions"
        canonical="https://payment-platform.com/products"
      />
      {/* Hero Section */}
      <div className="products-hero">
        <div className="products-hero-content">
          <h1 className="hero-title">{t('products.title') || 'Our Products'}</h1>
          <p className="hero-subtitle">{t('products.subtitle') || 'Enterprise-grade payment solutions for modern businesses'}</p>
          <Button type="primary" size="large" icon={<RocketOutlined />} href="/docs">
            Get Started
          </Button>
        </div>
      </div>

      {/* Core Products */}
      <section className="products-section">
        <div className="section-container">
          <div className="section-header">
            <h2 className="section-title">Core Products</h2>
            <p className="section-subtitle">Comprehensive payment platform for every business need</p>
          </div>
          <Row gutter={[32, 32]}>
            {products.map((product, index) => (
              <Col xs={24} sm={24} md={12} key={index}>
                <Card className="product-card-enhanced" bordered={false}>
                  <div className="product-icon-wrapper" style={{ background: product.gradient }}>
                    <div className="product-icon">{product.icon}</div>
                  </div>
                  <h2 className="product-title">{product.title}</h2>
                  <p className="product-description">{product.description}</p>

                  <div className="product-benefits">
                    {product.benefits.map((benefit, idx) => (
                      <Badge key={idx} count={benefit} style={{ backgroundColor: '#52c41a' }} />
                    ))}
                  </div>

                  <List
                    className="product-features"
                    dataSource={product.features}
                    renderItem={(item) => (
                      <List.Item>
                        <CheckCircleOutlined style={{ color: '#52c41a', marginRight: 8 }} />
                        {item}
                      </List.Item>
                    )}
                  />

                  <Button type="link" block className="learn-more-btn">
                    Learn More →
                  </Button>
                </Card>
              </Col>
            ))}
          </Row>
        </div>
      </section>

      {/* Technology Stack */}
      <section className="tech-section">
        <div className="section-container">
          <div className="section-header">
            <h2 className="section-title">Built on Modern Technology</h2>
            <p className="section-subtitle">Scalable, secure, and reliable infrastructure</p>
          </div>
          <Row gutter={[24, 24]}>
            {techStack.map((tech, index) => (
              <Col xs={12} sm={12} md={6} key={index}>
                <Card className="tech-card" bordered={false}>
                  <div className="tech-icon">{tech.icon}</div>
                  <h3 className="tech-name">{tech.name}</h3>
                  <p className="tech-desc">{tech.desc}</p>
                </Card>
              </Col>
            ))}
          </Row>
        </div>
      </section>

      {/* Integrations */}
      <section className="integrations-section">
        <div className="section-container">
          <div className="section-header">
            <h2 className="section-title">Payment Integrations</h2>
            <p className="section-subtitle">Connect with leading payment providers worldwide</p>
          </div>
          <div className="integration-grid">
            {integrations.map((integration, index) => (
              <div key={index} className="integration-item">
                <Card className="integration-card" bordered={false}>
                  <h4>{integration}</h4>
                </Card>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Implementation Timeline */}
      <section className="timeline-section">
        <div className="section-container">
          <div className="section-header">
            <h2 className="section-title">Quick Implementation</h2>
            <p className="section-subtitle">Go live in weeks, not months</p>
          </div>
          <Timeline mode="alternate" className="implementation-timeline">
            <Timeline.Item dot={<RocketOutlined />} color="blue">
              <h3>Week 1: Setup</h3>
              <p>Account creation and API credentials</p>
            </Timeline.Item>
            <Timeline.Item dot={<ApiOutlined />} color="green">
              <h3>Week 2: Integration</h3>
              <p>API integration and testing</p>
            </Timeline.Item>
            <Timeline.Item dot={<SafetyOutlined />} color="orange">
              <h3>Week 3: Testing</h3>
              <p>Security audit and compliance check</p>
            </Timeline.Item>
            <Timeline.Item dot={<CheckCircleOutlined />} color="purple">
              <h3>Week 4: Go Live</h3>
              <p>Production deployment and monitoring</p>
            </Timeline.Item>
          </Timeline>
        </div>
      </section>

      {/* CTA Section */}
      <section className="products-cta">
        <div className="section-container">
          <h2 className="cta-title">Ready to Get Started?</h2>
          <p className="cta-description">Join thousands of businesses using our platform</p>
          <div className="cta-actions">
            <Button type="primary" size="large" icon={<RocketOutlined />}>
              Start Free Trial
            </Button>
            <Button size="large" href="/contact">
              Contact Sales
            </Button>
          </div>
        </div>
      </section>
    </div>
  );
};

export default Products;
