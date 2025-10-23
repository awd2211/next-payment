import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { GithubOutlined, TwitterOutlined, LinkedinOutlined } from '@ant-design/icons';
import './style.css';

const Footer = () => {
  const { t } = useTranslation();

  const footerLinks = {
    company: [
      { label: t('footer.company.about'), href: '#' },
      { label: t('footer.company.careers'), href: '#' },
      { label: t('footer.company.contact'), href: '#' },
    ],
    product: [
      { label: t('footer.product.features'), href: '/products' },
      { label: t('footer.product.pricing'), href: '/pricing' },
      { label: t('footer.product.security'), href: '#' },
    ],
    resources: [
      { label: t('footer.resources.docs'), href: '/docs' },
      { label: t('footer.resources.api'), href: '/docs' },
      { label: t('footer.resources.support'), href: '#' },
    ],
    legal: [
      { label: t('footer.legal.privacy'), href: '#' },
      { label: t('footer.legal.terms'), href: '#' },
    ],
  };

  return (
    <footer className="website-footer">
      <div className="footer-container">
        <div className="footer-content">
          <div className="footer-column footer-brand">
            <h3 className="footer-logo">Payment Platform</h3>
            <p className="footer-description">
              Enterprise-grade global payment platform
            </p>
            <div className="footer-social">
              <a href="#" target="_blank" rel="noopener noreferrer">
                <GithubOutlined />
              </a>
              <a href="#" target="_blank" rel="noopener noreferrer">
                <TwitterOutlined />
              </a>
              <a href="#" target="_blank" rel="noopener noreferrer">
                <LinkedinOutlined />
              </a>
            </div>
          </div>

          <div className="footer-column">
            <h4 className="footer-title">{t('footer.company.title')}</h4>
            <ul className="footer-links">
              {footerLinks.company.map((link, index) => (
                <li key={index}>
                  <a href={link.href}>{link.label}</a>
                </li>
              ))}
            </ul>
          </div>

          <div className="footer-column">
            <h4 className="footer-title">{t('footer.product.title')}</h4>
            <ul className="footer-links">
              {footerLinks.product.map((link, index) => (
                <li key={index}>
                  <Link to={link.href}>{link.label}</Link>
                </li>
              ))}
            </ul>
          </div>

          <div className="footer-column">
            <h4 className="footer-title">{t('footer.resources.title')}</h4>
            <ul className="footer-links">
              {footerLinks.resources.map((link, index) => (
                <li key={index}>
                  <Link to={link.href}>{link.label}</Link>
                </li>
              ))}
            </ul>
          </div>

          <div className="footer-column">
            <h4 className="footer-title">{t('footer.legal.title')}</h4>
            <ul className="footer-links">
              {footerLinks.legal.map((link, index) => (
                <li key={index}>
                  <a href={link.href}>{link.label}</a>
                </li>
              ))}
            </ul>
          </div>
        </div>

        <div className="footer-bottom">
          <p>{t('footer.copyright')}</p>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
