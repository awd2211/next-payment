import { Helmet } from 'react-helmet-async';

interface SEOProps {
  title?: string;
  description?: string;
  keywords?: string;
  author?: string;
  ogType?: string;
  ogImage?: string;
  ogUrl?: string;
  twitterCard?: string;
  canonical?: string;
}

const SEO: React.FC<SEOProps> = ({
  title = 'Payment Platform - Global Payment Solutions',
  description = 'Enterprise-grade payment platform supporting multiple payment channels including Stripe, PayPal, and cryptocurrency. Process payments globally with 99.9% uptime.',
  keywords = 'payment gateway, stripe integration, paypal, multi-currency, payment processing, online payments, payment platform, fintech, cryptocurrency payments',
  author = 'Payment Platform Team',
  ogType = 'website',
  ogImage = '/og-image.png',
  ogUrl = 'https://payment-platform.com',
  twitterCard = 'summary_large_image',
  canonical,
}) => {
  const siteTitle = title.includes('Payment Platform')
    ? title
    : `${title} | Payment Platform`;

  return (
    <Helmet>
      {/* Primary Meta Tags */}
      <title>{siteTitle}</title>
      <meta name="title" content={siteTitle} />
      <meta name="description" content={description} />
      <meta name="keywords" content={keywords} />
      <meta name="author" content={author} />
      <meta name="robots" content="index, follow" />
      <meta name="language" content="English" />
      <meta name="revisit-after" content="7 days" />

      {/* Canonical URL */}
      {canonical && <link rel="canonical" href={canonical} />}

      {/* Open Graph / Facebook */}
      <meta property="og:type" content={ogType} />
      <meta property="og:url" content={ogUrl} />
      <meta property="og:title" content={siteTitle} />
      <meta property="og:description" content={description} />
      <meta property="og:image" content={ogImage} />
      <meta property="og:site_name" content="Payment Platform" />

      {/* Twitter */}
      <meta property="twitter:card" content={twitterCard} />
      <meta property="twitter:url" content={ogUrl} />
      <meta property="twitter:title" content={siteTitle} />
      <meta property="twitter:description" content={description} />
      <meta property="twitter:image" content={ogImage} />

      {/* Additional Meta Tags */}
      <meta name="theme-color" content="#667eea" />
      <meta name="mobile-web-app-capable" content="yes" />
      <meta name="apple-mobile-web-app-capable" content="yes" />
      <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent" />
    </Helmet>
  );
};

export default SEO;
