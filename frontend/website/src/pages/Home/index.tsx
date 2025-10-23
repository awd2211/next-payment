import { useTranslation } from 'react-i18next';
import { Button, Card, Row, Col, Statistic } from 'antd';
import {
  RocketOutlined,
  GlobalOutlined,
  SafetyOutlined,
  ApiOutlined,
  TeamOutlined,
  DollarOutlined,
} from '@ant-design/icons';
import './style.css';

const Home = () => {
  const { t } = useTranslation();

  const features = [
    {
      icon: <ApiOutlined />,
      title: t('home.features.items.microservices.title'),
      description: t('home.features.items.microservices.description'),
    },
    {
      icon: <DollarOutlined />,
      title: t('home.features.items.multiChannel.title'),
      description: t('home.features.items.multiChannel.description'),
    },
    {
      icon: <RocketOutlined />,
      title: t('home.features.items.monitoring.title'),
      description: t('home.features.items.monitoring.description'),
    },
    {
      icon: <SafetyOutlined />,
      title: t('home.features.items.security.title'),
      description: t('home.features.items.security.description'),
    },
    {
      icon: <TeamOutlined />,
      title: t('home.features.items.multiTenant.title'),
      description: t('home.features.items.multiTenant.description'),
    },
    {
      icon: <GlobalOutlined />,
      title: t('home.features.items.international.title'),
      description: t('home.features.items.international.description'),
    },
  ];

  return (
    <div className="home-page">
      {/* Hero Section */}
      <section className="hero-section">
        <div className="hero-container">
          <h1 className="hero-title">{t('home.hero.title')}</h1>
          <p className="hero-subtitle">{t('home.hero.subtitle')}</p>
          <p className="hero-description">{t('home.hero.description')}</p>
          <div className="hero-actions">
            <Button type="primary" size="large" href="/docs">
              {t('home.hero.cta.primary')}
            </Button>
            <Button size="large" href="/docs">
              {t('home.hero.cta.secondary')}
            </Button>
          </div>
        </div>
      </section>

      {/* Stats Section */}
      <section className="stats-section">
        <div className="section-container">
          <Row gutter={[32, 32]}>
            <Col xs={12} sm={12} md={6}>
              <Statistic
                title={t('home.stats.services')}
                value={15}
                prefix={<ApiOutlined />}
              />
            </Col>
            <Col xs={12} sm={12} md={6}>
              <Statistic
                title={t('home.stats.channels')}
                value={3}
                suffix="+"
                prefix={<DollarOutlined />}
              />
            </Col>
            <Col xs={12} sm={12} md={6}>
              <Statistic
                title={t('home.stats.currencies')}
                value={32}
                suffix="+"
                prefix={<GlobalOutlined />}
              />
            </Col>
            <Col xs={12} sm={12} md={6}>
              <Statistic
                title={t('home.stats.uptime')}
                value={99.9}
                suffix="%"
                precision={1}
                prefix={<SafetyOutlined />}
              />
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
                <Card className="feature-card">
                  <div className="feature-icon">{feature.icon}</div>
                  <h3 className="feature-title">{feature.title}</h3>
                  <p className="feature-description">{feature.description}</p>
                </Card>
              </Col>
            ))}
          </Row>
        </div>
      </section>

      {/* CTA Section */}
      <section className="cta-section">
        <div className="section-container">
          <h2 className="cta-title">{t('home.cta.title')}</h2>
          <p className="cta-description">{t('home.cta.description')}</p>
          <Button type="primary" size="large" href="/docs">
            {t('home.cta.button')}
          </Button>
        </div>
      </section>
    </div>
  );
};

export default Home;
