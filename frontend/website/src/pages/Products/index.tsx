import { useTranslation } from 'react-i18next';
import { Card, Row, Col, List } from 'antd';
import {
  ApiOutlined,
  SafetyOutlined,
  DollarOutlined,
  DashboardOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons';
import './style.css';

const Products = () => {
  const { t } = useTranslation();

  const products = [
    {
      icon: <ApiOutlined />,
      title: t('products.gateway.title'),
      description: t('products.gateway.description'),
      features: t('products.gateway.features', { returnObjects: true }) as string[],
      color: '#1890ff',
    },
    {
      icon: <SafetyOutlined />,
      title: t('products.risk.title'),
      description: t('products.risk.description'),
      features: t('products.risk.features', { returnObjects: true }) as string[],
      color: '#52c41a',
    },
    {
      icon: <DollarOutlined />,
      title: t('products.settlement.title'),
      description: t('products.settlement.description'),
      features: t('products.settlement.features', { returnObjects: true }) as string[],
      color: '#faad14',
    },
    {
      icon: <DashboardOutlined />,
      title: t('products.monitoring.title'),
      description: t('products.monitoring.description'),
      features: t('products.monitoring.features', { returnObjects: true }) as string[],
      color: '#722ed1',
    },
  ];

  return (
    <div className="products-page">
      <div className="products-header">
        <h1 className="page-title">{t('products.title')}</h1>
        <p className="page-subtitle">{t('products.subtitle')}</p>
      </div>

      <div className="products-container">
        <Row gutter={[24, 24]}>
          {products.map((product, index) => (
            <Col xs={24} sm={24} md={12} key={index}>
              <Card className="product-card">
                <div className="product-icon" style={{ color: product.color }}>
                  {product.icon}
                </div>
                <h2 className="product-title">{product.title}</h2>
                <p className="product-description">{product.description}</p>
                <List
                  className="product-features"
                  dataSource={product.features}
                  renderItem={(item) => (
                    <List.Item>
                      <CheckCircleOutlined style={{ color: product.color, marginRight: 8 }} />
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

export default Products;
