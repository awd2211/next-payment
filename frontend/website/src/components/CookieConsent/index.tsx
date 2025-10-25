import { useState, useEffect } from 'react';
import { Button, Space } from 'antd';
import { CheckCircleOutlined, CloseOutlined } from '@ant-design/icons';
import './style.css';

const CookieConsent = () => {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    const consent = localStorage.getItem('cookieConsent');
    if (!consent) {
      // Show after 1 second delay for better UX
      setTimeout(() => setIsVisible(true), 1000);
    }
  }, []);

  const handleAccept = () => {
    localStorage.setItem('cookieConsent', 'accepted');
    localStorage.setItem('cookieConsentDate', new Date().toISOString());
    setIsVisible(false);

    // Initialize analytics after consent
    if (typeof window !== 'undefined' && (window as any).gtag) {
      (window as any).gtag('consent', 'update', {
        analytics_storage: 'granted',
      });
    }
  };

  const handleDecline = () => {
    localStorage.setItem('cookieConsent', 'declined');
    localStorage.setItem('cookieConsentDate', new Date().toISOString());
    setIsVisible(false);

    // Deny analytics
    if (typeof window !== 'undefined' && (window as any).gtag) {
      (window as any).gtag('consent', 'update', {
        analytics_storage: 'denied',
      });
    }
  };

  if (!isVisible) return null;

  return (
    <div className="cookie-consent">
      <div className="cookie-content">
        <div className="cookie-icon">
          🍪
        </div>
        <div className="cookie-text">
          <h4 className="cookie-title">We use cookies</h4>
          <p className="cookie-description">
            We use cookies to enhance your browsing experience, analyze site traffic, and personalize content.
            By clicking "Accept All", you consent to our use of cookies.{' '}
            <a href="/privacy" className="cookie-link">
              Privacy Policy
            </a>
          </p>
        </div>
        <div className="cookie-actions">
          <Space>
            <Button
              type="text"
              onClick={handleDecline}
              icon={<CloseOutlined />}
            >
              Decline
            </Button>
            <Button
              type="primary"
              onClick={handleAccept}
              icon={<CheckCircleOutlined />}
            >
              Accept All
            </Button>
          </Space>
        </div>
      </div>
    </div>
  );
};

export default CookieConsent;
