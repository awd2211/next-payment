import { useTranslation } from 'react-i18next';
import { Card, Row, Col, Button, List } from 'antd';
import { CheckOutlined } from '@ant-design/icons';
import './style.css';

const Pricing = () => {
  const { t } = useTranslation();

  const plans = [
    {
      name: t('pricing.starter.name'),
      price: t('pricing.starter.price'),
      period: '',
      description: t('pricing.starter.description'),
      features: t('pricing.starter.features', { returnObjects: true }) as string[],
      buttonType: 'default' as const,
      popular: false,
    },
    {
      name: t('pricing.professional.name'),
      price: t('pricing.professional.price'),
      period: t('pricing.professional.period'),
      description: t('pricing.professional.description'),
      features: t('pricing.professional.features', { returnObjects: true }) as string[],
      buttonType: 'primary' as const,
      popular: true,
    },
    {
      name: t('pricing.enterprise.name'),
      price: t('pricing.enterprise.price'),
      period: '',
      description: t('pricing.enterprise.description'),
      features: t('pricing.enterprise.features', { returnObjects: true }) as string[],
      buttonType: 'default' as const,
      popular: false,
    },
  ];

  return (
    <div className="pricing-page">
      <div className="pricing-header">
        <h1 className="page-title">{t('pricing.title')}</h1>
        <p className="page-subtitle">{t('pricing.subtitle')}</p>
      </div>

      <div className="pricing-container">
        <Row gutter={[24, 24]}>
          {plans.map((plan, index) => (
            <Col xs={24} sm={24} md={8} key={index}>
              <Card className={`pricing-card ${plan.popular ? 'popular' : ''}`}>
                {plan.popular && <div className="popular-badge">Popular</div>}
                <h2 className="plan-name">{plan.name}</h2>
                <div className="plan-price">
                  <span className="price-amount">{plan.price}</span>
                  {plan.period && <span className="price-period">{plan.period}</span>}
                </div>
                <p className="plan-description">{plan.description}</p>
                <Button
                  type={plan.buttonType}
                  size="large"
                  block
                  className="plan-button"
                >
                  {t('pricing.cta')}
                </Button>
                <List
                  className="plan-features"
                  dataSource={plan.features}
                  renderItem={(item) => (
                    <List.Item>
                      <CheckOutlined style={{ color: '#52c41a', marginRight: 8 }} />
                      {item}
                    </List.Item>
                  )}
                />
              </Card>
            </Col>
          ))}
        </Row>
      </div>
    </div>
  );
};

export default Pricing;
