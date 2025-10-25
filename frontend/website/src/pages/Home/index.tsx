import { useTranslation } from 'react-i18next';
import { Button, Card, Row, Col, Statistic, Collapse, Badge } from 'antd';
import {
  RocketOutlined,
  GlobalOutlined,
  SafetyOutlined,
  ApiOutlined,
  TeamOutlined,
  DollarOutlined,
  CheckCircleOutlined,
  ThunderboltOutlined,
  CloudOutlined,
  LockOutlined,
  QuestionCircleOutlined,
} from '@ant-design/icons';
import SEO from '../../components/SEO';
import CountUp from '../../components/CountUp';
import './style.css';

const { Panel } = Collapse;

const Home = () => {
  const { t } = useTranslation();

  const features = [
    {
      icon: <ApiOutlined />,
      title: t('home.features.items.microservices.title'),
      description: t('home.features.items.microservices.description'),
      gradient: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    },
    {
      icon: <DollarOutlined />,
      title: t('home.features.items.multiChannel.title'),
      description: t('home.features.items.multiChannel.description'),
      gradient: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
    },
    {
      icon: <RocketOutlined />,
      title: t('home.features.items.monitoring.title'),
      description: t('home.features.items.monitoring.description'),
      gradient: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
    },
    {
      icon: <SafetyOutlined />,
      title: t('home.features.items.security.title'),
      description: t('home.features.items.security.description'),
      gradient: 'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)',
    },
    {
      icon: <TeamOutlined />,
      title: t('home.features.items.multiTenant.title'),
      description: t('home.features.items.multiTenant.description'),
      gradient: 'linear-gradient(135deg, #fa709a 0%, #fee140 100%)',
    },
    {
      icon: <GlobalOutlined />,
      title: t('home.features.items.international.title'),
      description: t('home.features.items.international.description'),
      gradient: 'linear-gradient(135deg, #30cfd0 0%, #330867 100%)',
    },
  ];

  const trustBadges = [
    { icon: <CheckCircleOutlined />, text: 'PCI DSS Compliant' },
    { icon: <LockOutlined />, text: 'ISO 27001 Certified' },
    { icon: <ThunderboltOutlined />, text: '99.9% Uptime SLA' },
    { icon: <CloudOutlined />, text: 'Cloud Native' },
  ];

  const faqs = [
    {
      question: t('home.faq.items.integration.question') || 'How easy is it to integrate?',
      answer: t('home.faq.items.integration.answer') || 'Integration is simple with our RESTful APIs and comprehensive SDKs. Most merchants can go live within 1-2 weeks.',
    },
    {
      question: t('home.faq.items.fees.question') || 'What are the transaction fees?',
      answer: t('home.faq.items.fees.answer') || 'Our pricing is competitive and transparent. Transaction fees vary by payment method and volume. Contact us for custom pricing.',
    },
    {
      question: t('home.faq.items.security.question') || 'How secure is the platform?',
      answer: t('home.faq.items.security.answer') || 'We are PCI DSS Level 1 compliant and ISO 27001 certified. All data is encrypted in transit and at rest.',
    },
    {
      question: t('home.faq.items.support.question') || 'What support do you provide?',
      answer: t('home.faq.items.support.answer') || 'We offer 24/7 technical support, dedicated account managers for enterprise clients, and comprehensive documentation.',
    },
  ];

  return (
    <div className="home-page">
      <SEO
        title="Home - Payment Platform"
        description="Enterprise-grade global payment platform supporting Stripe, PayPal, and cryptocurrency. 99.9% uptime, PCI DSS compliant, processing $10B+ annually."
        keywords="payment gateway, stripe, paypal, cryptocurrency, online payments, payment processing, fintech, multi-currency"
        canonical="https://payment-platform.com/"
      />
      {/* Hero Section */}
      <section className="hero-section">
        <div className="hero-background">
          <div className="hero-particles"></div>
        </div>
        <div className="hero-container">
          <Badge.Ribbon text="Production Ready" color="purple" className="hero-badge">
            <div className="hero-content">
              <h1 className="hero-title animate-fade-in-up">{t('home.hero.title')}</h1>
              <p className="hero-subtitle animate-fade-in-up">{t('home.hero.subtitle')}</p>
              <p className="hero-description animate-fade-in-up">{t('home.hero.description')}</p>
              <div className="hero-actions animate-fade-in-up">
                <Button type="primary" size="large" href="/docs" icon={<RocketOutlined />}>
                  {t('home.hero.cta.primary')}
                </Button>
                <Button size="large" href="http://localhost:5173" target="_blank">
                  {t('home.hero.cta.secondary')}
                </Button>
              </div>
            </div>
          </Badge.Ribbon>
        </div>
      </section>

      {/* Trust Badges Section */}
      <section className="trust-section">
        <div className="section-container">
          <Row gutter={[24, 24]} justify="center">
            {trustBadges.map((badge, index) => (
              <Col xs={12} sm={6} key={index}>
                <div className="trust-badge">
                  <div className="trust-icon">{badge.icon}</div>
                  <span className="trust-text">{badge.text}</span>
                </div>
              </Col>
            ))}
          </Row>
        </div>
      </section>

      {/* Stats Section */}
      <section className="stats-section">
        <div className="section-container">
          <Row gutter={[32, 32]}>
            <Col xs={12} sm={12} md={6}>
              <div className="stat-card animate-on-scroll">
                <Statistic
                  title={t('home.stats.services')}
                  value={15}
                  prefix={<ApiOutlined />}
                  valueStyle={{ color: '#667eea' }}
                />
              </div>
            </Col>
            <Col xs={12} sm={12} md={6}>
              <div className="stat-card animate-on-scroll">
                <Statistic
                  title={t('home.stats.channels')}
                  value={4}
                  suffix="+"
                  prefix={<DollarOutlined />}
                  valueStyle={{ color: '#f093fb' }}
                />
              </div>
            </Col>
            <Col xs={12} sm={12} md={6}>
              <div className="stat-card animate-on-scroll">
                <Statistic
                  title={t('home.stats.currencies')}
                  value={32}
                  suffix="+"
                  prefix={<GlobalOutlined />}
                  valueStyle={{ color: '#4facfe' }}
                />
              </div>
            </Col>
            <Col xs={12} sm={12} md={6}>
              <div className="stat-card animate-on-scroll">
                <Statistic
                  title={t('home.stats.uptime')}
                  value={99.9}
                  suffix="%"
                  precision={1}
                  prefix={<SafetyOutlined />}
                  valueStyle={{ color: '#43e97b' }}
                />
              </div>
            </Col>
          </Row>
        </div>
      </section>

      {/* Features Section */}
      <section className="features-section">
        <div className="section-container">
          <div className="section-header">
            <h2 className="section-title">{t('home.features.title')}</h2>
            <p className="section-subtitle">{t('home.features.subtitle')}</p>
          </div>
          <Row gutter={[24, 24]}>
            {features.map((feature, index) => (
              <Col xs={24} sm={12} md={8} key={index}>
                <Card className="feature-card animate-on-scroll" bordered={false}>
                  <div
                    className="feature-icon-wrapper"
                    style={{ background: feature.gradient }}
                  >
                    <div className="feature-icon">{feature.icon}</div>
                  </div>
                  <h3 className="feature-title">{feature.title}</h3>
                  <p className="feature-description">{feature.description}</p>
                </Card>
              </Col>
            ))}
          </Row>
        </div>
      </section>

      {/* Product Demo Section */}
      <section className="demo-section">
        <div className="section-container">
          <div className="section-header">
            <h2 className="section-title">{t('home.demo.title') || 'See It In Action'}</h2>
            <p className="section-subtitle">{t('home.demo.subtitle') || 'Powerful dashboards for merchants and administrators'}</p>
          </div>
          <div className="demo-showcase">
            <Row gutter={[48, 48]} align="middle">
              <Col xs={24} md={12}>
                <div className="demo-image-wrapper">
                  <div className="demo-placeholder">
                    <ApiOutlined style={{ fontSize: 120, color: '#667eea' }} />
                    <p>Merchant Portal Dashboard</p>
                  </div>
                </div>
              </Col>
              <Col xs={24} md={12}>
                <div className="demo-features">
                  <h3>Merchant Portal</h3>
                  <ul>
                    <li><CheckCircleOutlined /> Real-time payment tracking</li>
                    <li><CheckCircleOutlined /> Multi-currency support</li>
                    <li><CheckCircleOutlined /> Automated settlement reports</li>
                    <li><CheckCircleOutlined /> API key management</li>
                  </ul>
                  <Button type="primary" href="http://localhost:5174" target="_blank">
                    Try Demo
                  </Button>
                </div>
              </Col>
            </Row>
          </div>
        </div>
      </section>

      {/* FAQ Section */}
      <section className="faq-section">
        <div className="section-container">
          <div className="section-header">
            <h2 className="section-title">
              <QuestionCircleOutlined /> {t('home.faq.title') || 'Frequently Asked Questions'}
            </h2>
            <p className="section-subtitle">{t('home.faq.subtitle') || 'Everything you need to know'}</p>
          </div>
          <div className="faq-content">
            <Collapse
              accordion
              ghost
              expandIconPosition="end"
              className="faq-collapse"
            >
              {faqs.map((faq, index) => (
                <Panel
                  header={<span className="faq-question">{faq.question}</span>}
                  key={index}
                >
                  <p className="faq-answer">{faq.answer}</p>
                </Panel>
              ))}
            </Collapse>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="cta-section">
        <div className="section-container">
          <div className="cta-content">
            <h2 className="cta-title">{t('home.cta.title')}</h2>
            <p className="cta-description">{t('home.cta.description')}</p>
            <div className="cta-actions">
              <Button type="primary" size="large" href="/docs" icon={<RocketOutlined />}>
                {t('home.cta.button')}
              </Button>
              <Button size="large" href="/pricing" ghost>
                View Pricing
              </Button>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
};

export default Home;
