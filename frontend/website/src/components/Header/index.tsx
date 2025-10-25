import { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Button, Drawer, Space } from 'antd';
import { MenuOutlined } from '@ant-design/icons';
import LanguageSwitch from '../LanguageSwitch';
import ThemeSwitch from '../ThemeSwitch';
import './style.css';

const Header = () => {
  const { t } = useTranslation();
  const location = useLocation();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const navItems = [
    { key: 'home', path: '/', label: t('nav.home') || 'Home' },
    { key: 'products', path: '/products', label: t('nav.products') || 'Products' },
    { key: 'pricing', path: '/pricing', label: t('nav.pricing') || 'Pricing' },
    { key: 'docs', path: '/docs', label: t('nav.docs') || 'Docs' },
    { key: 'about', path: '/about', label: t('nav.about') || 'About' },
    { key: 'contact', path: '/contact', label: t('nav.contact') || 'Contact' },
  ];

  const isActive = (path: string) => {
    if (path === '/') {
      return location.pathname === '/';
    }
    return location.pathname.startsWith(path);
  };

  return (
    <header className="website-header">
      <div className="header-container">
        <div className="header-logo">
          <Link to="/">
            <span className="logo-text">Payment Platform</span>
          </Link>
        </div>

        {/* Desktop Navigation */}
        <nav className="header-nav desktop-nav">
          {navItems.map((item) => (
            <Link
              key={item.key}
              to={item.path}
              className={`nav-link ${isActive(item.path) ? 'active' : ''}`}
            >
              {item.label}
            </Link>
          ))}
        </nav>

        <div className="header-actions">
          <Space size="middle">
            <ThemeSwitch />
            <LanguageSwitch />
            <Button type="link" href="http://localhost:5173" target="_blank">
              {t('nav.login')}
            </Button>
            <Button type="primary" href="http://localhost:5173" target="_blank">
              {t('nav.register')}
            </Button>

            {/* Mobile Menu Button */}
            <Button
              className="mobile-menu-btn"
              type="text"
              icon={<MenuOutlined />}
              onClick={() => setMobileMenuOpen(true)}
            />
          </Space>
        </div>
      </div>

      {/* Mobile Drawer */}
      <Drawer
        title="Menu"
        placement="right"
        onClose={() => setMobileMenuOpen(false)}
        open={mobileMenuOpen}
      >
        <nav className="mobile-nav">
          {navItems.map((item) => (
            <Link
              key={item.key}
              to={item.path}
              className={`mobile-nav-link ${isActive(item.path) ? 'active' : ''}`}
              onClick={() => setMobileMenuOpen(false)}
            >
              {item.label}
            </Link>
          ))}
        </nav>
      </Drawer>
    </header>
  );
};

export default Header;
