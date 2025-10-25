import { Button, Result } from 'antd';
import { useNavigate } from 'react-router-dom';
import { HomeOutlined, SearchOutlined, RocketOutlined } from '@ant-design/icons';
import './style.css';

const NotFound = () => {
  const navigate = useNavigate();

  return (
    <div className="notfound-page">
      <div className="notfound-container">
        {/* Animated 404 */}
        <div className="notfound-number">
          <span className="digit">4</span>
          <span className="digit rotating">0</span>
          <span className="digit">4</span>
        </div>

        {/* Main Content */}
        <div className="notfound-content">
          <h1 className="notfound-title">Page Not Found</h1>
          <p className="notfound-description">
            Oops! The page you're looking for doesn't exist or has been moved.
          </p>

          {/* Action Buttons */}
          <div className="notfound-actions">
            <Button
              type="primary"
              size="large"
              icon={<HomeOutlined />}
              onClick={() => navigate('/')}
            >
              Back to Home
            </Button>
            <Button
              size="large"
              icon={<SearchOutlined />}
              onClick={() => navigate('/docs')}
            >
              View Documentation
            </Button>
          </div>

          {/* Quick Links */}
          <div className="notfound-links">
            <h3>Quick Links</h3>
            <div className="links-grid">
              <a href="/" className="link-card">
                <HomeOutlined />
                <span>Home</span>
              </a>
              <a href="/products" className="link-card">
                <RocketOutlined />
                <span>Products</span>
              </a>
              <a href="/pricing" className="link-card">
                <span className="icon">💰</span>
                <span>Pricing</span>
              </a>
              <a href="/docs" className="link-card">
                <SearchOutlined />
                <span>Docs</span>
              </a>
              <a href="/about" className="link-card">
                <span className="icon">ℹ️</span>
                <span>About</span>
              </a>
              <a href="/contact" className="link-card">
                <span className="icon">📧</span>
                <span>Contact</span>
              </a>
            </div>
          </div>
        </div>

        {/* Floating Elements */}
        <div className="floating-elements">
          <div className="float-element element-1"></div>
          <div className="float-element element-2"></div>
          <div className="float-element element-3"></div>
          <div className="float-element element-4"></div>
        </div>
      </div>
    </div>
  );
};

export default NotFound;
