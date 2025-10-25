import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, Row, Col, Button, List, Switch, Table, Collapse, Tag } from 'antd';
import {
  CheckOutlined,
  CloseOutlined,
  RocketOutlined,
  CrownOutlined,
  StarOutlined,
} from '@ant-design/icons';
import './style.css';

const { Panel } = Collapse;

const Pricing = () => {
  const { t } = useTranslation();
  const [isAnnual, setIsAnnual] = useState(true);

  const plans = [
    {
      name: t('pricing.starter.name') || 'Starter',
      monthlyPrice: 0,
      annualPrice: 0,
      description: t('pricing.starter.description') || 'Perfect for getting started',
      features: t('pricing.starter.features', { returnObjects: true }) as string[] || [
        'Up to 100 transactions/month',
        '2 payment channels',
        'Basic reporting',
        'Email support',
        'Standard security',
      ],
      buttonType: 'default' as const,
      popular: false,
      icon: <RocketOutlined />,
      gradient: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    },
    {
      name: t('pricing.professional.name') || 'Professional',
      monthlyPrice: 99,
      annualPrice: 990,
      description: t('pricing.professional.description') || 'For growing businesses',
      features: t('pricing.professional.features', { returnObjects: true }) as string[] || [
        'Up to 10,000 transactions/month',
        'All payment channels',
        'Advanced analytics',
        'Priority support',
        'Risk management',
        'Custom webhooks',
        'API access',
      ],
      buttonType: 'primary' as const,
      popular: true,
      icon: <CrownOutlined />,
      gradient: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
    },
    {
      name: t('pricing.enterprise.name') || 'Enterprise',
      monthlyPrice: null,
      annualPrice: null,
      description: t('pricing.enterprise.description') || 'For large-scale operations',
      features: t('pricing.enterprise.features', { returnObjects: true }) as string[] || [
        'Unlimited transactions',
        'All channels + custom integrations',
        'Dedicated support',
        'Custom SLA',
        'Advanced fraud protection',
        'Custom contracts',
        'White-label solution',
        'Dedicated account manager',
      ],
      buttonType: 'default' as const,
      popular: false,
      icon: <StarOutlined />,
      gradient: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
    },
  ];

  const comparisonFeatures = [
    {
      category: 'Core Features',
      features: [
        { name: 'Monthly Transactions', starter: '100', professional: '10,000', enterprise: 'Unlimited' },
        { name: 'Payment Channels', starter: '2', professional: 'All', enterprise: 'All + Custom' },
        { name: 'API Access', starter: false, professional: true, enterprise: true },
        { name: 'Webhooks', starter: false, professional: true, enterprise: true },
      ]
    },
    {
      category: 'Support',
      features: [
        { name: 'Email Support', starter: true, professional: true, enterprise: true },
        { name: 'Priority Support', starter: false, professional: true, enterprise: true },
        { name: 'Dedicated Manager', starter: false, professional: false, enterprise: true },
        { name: 'Custom SLA', starter: false, professional: false, enterprise: true },
      ]
    },
    {
      category: 'Advanced Features',
      features: [
        { name: 'Advanced Analytics', starter: false, professional: true, enterprise: true },
        { name: 'Risk Management', starter: false, professional: true, enterprise: true },
        { name: 'Fraud Protection', starter: false, professional: 'Basic', enterprise: 'Advanced' },
        { name: 'White-label', starter: false, professional: false, enterprise: true },
      ]
    },
  ];

  const faqs = [
    {
      question: 'Can I switch plans at any time?',
      answer: 'Yes, you can upgrade or downgrade your plan at any time. Changes will be reflected in your next billing cycle.',
    },
    {
      question: 'What payment methods do you accept?',
      answer: 'We accept all major credit cards (Visa, Mastercard, Amex) and bank transfers for annual plans.',
    },
    {
      question: 'Is there a free trial?',
      answer: 'Yes, all paid plans come with a 14-day free trial. No credit card required.',
    },
    {
      question: 'What happens if I exceed my transaction limit?',
      answer: 'We\'ll automatically notify you when you reach 80% of your limit. You can upgrade your plan or pay for additional transactions.',
    },
    {
      question: 'Do you offer custom pricing?',
      answer: 'Yes, for Enterprise plans we offer custom pricing based on your specific needs. Contact our sales team for details.',
    },
  ];

  const getPrice = (plan: typeof plans[0]) => {
    if (plan.monthlyPrice === null) return 'Custom';
    if (plan.monthlyPrice === 0) return 'Free';
    return isAnnual
      ? `$${plan.annualPrice! / 12}/mo`
      : `$${plan.monthlyPrice}/mo`;
  };

  const getSavings = (plan: typeof plans[0]) => {
    if (plan.annualPrice === null || plan.monthlyPrice === null || plan.monthlyPrice === 0) return null;
    const monthlyCost = plan.monthlyPrice * 12;
    const savings = monthlyCost - plan.annualPrice!;
    const percentage = Math.round((savings / monthlyCost) * 100);
    return { amount: savings, percentage };
  };

  return (
    <div className="pricing-page">
      {/* Hero Section */}
      <section className="pricing-hero">
        <div className="pricing-hero-content">
          <h1 className="hero-title">{t('pricing.title') || 'Simple, Transparent Pricing'}</h1>
          <p className="hero-subtitle">{t('pricing.subtitle') || 'Choose the perfect plan for your business'}</p>

          {/* Billing Toggle */}
          <div className="billing-toggle">
            <span className={!isAnnual ? 'active' : ''}>Monthly</span>
            <Switch
              checked={isAnnual}
              onChange={setIsAnnual}
              className="toggle-switch"
            />
            <span className={isAnnual ? 'active' : ''}>
              Annual <Tag color="green">Save 17%</Tag>
            </span>
          </div>
        </div>
      </section>

      {/* Pricing Cards */}
      <section className="pricing-cards-section">
        <div className="section-container">
          <Row gutter={[32, 32]}>
            {plans.map((plan, index) => {
              const savings = getSavings(plan);
              return (
                <Col xs={24} sm={24} md={8} key={index}>
                  <Card className={`pricing-card-enhanced ${plan.popular ? 'popular' : ''}`} bordered={false}>
                    {plan.popular && <div className="popular-badge">Most Popular</div>}
                    {isAnnual && savings && (
                      <div className="savings-badge">Save ${savings.amount}/year</div>
                    )}

                    <div className="plan-icon-wrapper" style={{ background: plan.gradient }}>
                      <div className="plan-icon">{plan.icon}</div>
                    </div>

                    <h2 className="plan-name">{plan.name}</h2>
                    <div className="plan-price">
                      <span className="price-amount">{getPrice(plan)}</span>
                      {plan.monthlyPrice !== null && plan.monthlyPrice > 0 && (
                        <span className="price-period">
                          {isAnnual ? 'billed annually' : 'billed monthly'}
                        </span>
                      )}
                    </div>
                    <p className="plan-description">{plan.description}</p>

                    <Button
                      type={plan.buttonType}
                      size="large"
                      block
                      className="plan-button"
                    >
                      {plan.monthlyPrice === null ? 'Contact Sales' : plan.monthlyPrice === 0 ? 'Get Started Free' : 'Start Free Trial'}
                    </Button>

                    <List
                      className="plan-features"
                      dataSource={plan.features}
                      renderItem={(item) => (
                        <List.Item>
                          <CheckOutlined className="feature-check" />
                          {item}
                        </List.Item>
                      )}
                    />
                  </Card>
                </Col>
              );
            })}
          </Row>
        </div>
      </section>

      {/* Feature Comparison Table */}
      <section className="comparison-section">
        <div className="section-container">
          <div className="section-header">
            <h2 className="section-title">Compare Plans</h2>
            <p className="section-subtitle">Detailed feature comparison across all plans</p>
          </div>

          <div className="comparison-table-wrapper">
            {comparisonFeatures.map((category, catIndex) => (
              <div key={catIndex} className="comparison-category">
                <h3 className="category-title">{category.category}</h3>
                <div className="comparison-table">
                  <div className="comparison-header">
                    <div className="header-cell feature-name">Feature</div>
                    <div className="header-cell">Starter</div>
                    <div className="header-cell popular">Professional</div>
                    <div className="header-cell">Enterprise</div>
                  </div>
                  {category.features.map((feature, featIndex) => (
                    <div key={featIndex} className="comparison-row">
                      <div className="row-cell feature-name">{feature.name}</div>
                      <div className="row-cell">
                        {renderFeatureValue(feature.starter)}
                      </div>
                      <div className="row-cell popular">
                        {renderFeatureValue(feature.professional)}
                      </div>
                      <div className="row-cell">
                        {renderFeatureValue(feature.enterprise)}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* FAQ Section */}
      <section className="pricing-faq-section">
        <div className="section-container">
          <div className="section-header">
            <h2 className="section-title">Frequently Asked Questions</h2>
            <p className="section-subtitle">Everything you need to know about our pricing</p>
          </div>

          <div className="faq-content">
            <Collapse accordion ghost className="pricing-faq-collapse">
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
      <section className="pricing-cta">
        <div className="section-container">
          <h2 className="cta-title">Ready to Get Started?</h2>
          <p className="cta-description">
            Start your 14-day free trial. No credit card required.
          </p>
          <div className="cta-actions">
            <Button type="primary" size="large" icon={<RocketOutlined />}>
              Start Free Trial
            </Button>
            <Button size="large" href="/contact">
              Talk to Sales
            </Button>
          </div>
        </div>
      </section>
    </div>
  );

  function renderFeatureValue(value: any) {
    if (value === true) {
      return <CheckOutlined style={{ color: '#52c41a', fontSize: '18px' }} />;
    }
    if (value === false) {
      return <CloseOutlined style={{ color: '#d9d9d9', fontSize: '18px' }} />;
    }
    return <span>{value}</span>;
  }
};

export default Pricing;
