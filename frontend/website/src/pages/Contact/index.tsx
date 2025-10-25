import { useState } from 'react';
import { Form, Input, Button, Card, Row, Col, message } from 'antd';
import {
  MailOutlined,
  PhoneOutlined,
  EnvironmentOutlined,
  ClockCircleOutlined,
  SendOutlined,
} from '@ant-design/icons';
import './style.css';

const { TextArea } = Input;

const Contact = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  const contactInfo = [
    {
      icon: <MailOutlined />,
      title: 'Email',
      content: 'support@payment-platform.com',
      link: 'mailto:support@payment-platform.com',
      gradient: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    },
    {
      icon: <PhoneOutlined />,
      title: 'Phone',
      content: '+1 (555) 123-4567',
      link: 'tel:+15551234567',
      gradient: 'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)',
    },
    {
      icon: <EnvironmentOutlined />,
      title: 'Address',
      content: '123 Payment Street, San Francisco, CA 94102',
      link: 'https://maps.google.com',
      gradient: 'linear-gradient(135deg, #fa709a 0%, #fee140 100%)',
    },
    {
      icon: <ClockCircleOutlined />,
      title: 'Business Hours',
      content: 'Mon-Fri: 9:00 AM - 6:00 PM PST',
      link: null,
      gradient: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
    },
  ];

  const departments = [
    { name: 'Sales', email: 'sales@payment-platform.com', desc: 'New business inquiries' },
    { name: 'Support', email: 'support@payment-platform.com', desc: 'Technical assistance' },
    { name: 'Partnerships', email: 'partners@payment-platform.com', desc: 'Strategic partnerships' },
    { name: 'Media', email: 'press@payment-platform.com', desc: 'Press and media inquiries' },
  ];

  const handleSubmit = async (values: any) => {
    setLoading(true);
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1500));
      message.success('Thank you for contacting us! We\'ll get back to you soon.');
      form.resetFields();
    } catch (error) {
      message.error('Something went wrong. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="contact-page">
      {/* Hero Section */}
      <section className="contact-hero">
        <div className="contact-hero-content">
          <h1 className="hero-title">Get in Touch</h1>
          <p className="hero-subtitle">
            Have questions? We'd love to hear from you. Send us a message and we'll respond as soon as possible.
          </p>
        </div>
      </section>

      {/* Contact Info Cards */}
      <section className="contact-info-section">
        <div className="section-container">
          <Row gutter={[24, 24]}>
            {contactInfo.map((info, index) => (
              <Col xs={24} sm={12} md={6} key={index}>
                <Card
                  className="contact-info-card"
                  bordered={false}
                  onClick={() => info.link && window.open(info.link, '_blank')}
                  style={{ cursor: info.link ? 'pointer' : 'default' }}
                >
                  <div className="contact-icon-wrapper" style={{ background: info.gradient }}>
                    <div className="contact-icon">{info.icon}</div>
                  </div>
                  <h3 className="contact-info-title">{info.title}</h3>
                  <p className="contact-info-content">{info.content}</p>
                </Card>
              </Col>
            ))}
          </Row>
        </div>
      </section>

      {/* Contact Form Section */}
      <section className="contact-form-section">
        <div className="section-container">
          <Row gutter={[48, 48]}>
            <Col xs={24} md={12}>
              <div className="form-intro">
                <h2 className="section-title">Send us a Message</h2>
                <p className="section-subtitle">
                  Fill out the form and our team will get back to you within 24 hours.
                </p>

                <div className="departments-list">
                  <h3>Or reach out directly:</h3>
                  {departments.map((dept, index) => (
                    <div key={index} className="department-item">
                      <h4>{dept.name}</h4>
                      <a href={`mailto:${dept.email}`}>{dept.email}</a>
                      <p>{dept.desc}</p>
                    </div>
                  ))}
                </div>
              </div>
            </Col>
            <Col xs={24} md={12}>
              <Card className="contact-form-card" bordered={false}>
                <Form
                  form={form}
                  layout="vertical"
                  onFinish={handleSubmit}
                  requiredMark={false}
                >
                  <Form.Item
                    name="name"
                    label="Full Name"
                    rules={[{ required: true, message: 'Please enter your name' }]}
                  >
                    <Input size="large" placeholder="John Doe" />
                  </Form.Item>

                  <Form.Item
                    name="email"
                    label="Email Address"
                    rules={[
                      { required: true, message: 'Please enter your email' },
                      { type: 'email', message: 'Please enter a valid email' },
                    ]}
                  >
                    <Input size="large" placeholder="john@example.com" />
                  </Form.Item>

                  <Form.Item
                    name="company"
                    label="Company"
                  >
                    <Input size="large" placeholder="Your Company" />
                  </Form.Item>

                  <Form.Item
                    name="subject"
                    label="Subject"
                    rules={[{ required: true, message: 'Please enter a subject' }]}
                  >
                    <Input size="large" placeholder="How can we help?" />
                  </Form.Item>

                  <Form.Item
                    name="message"
                    label="Message"
                    rules={[{ required: true, message: 'Please enter your message' }]}
                  >
                    <TextArea
                      rows={6}
                      placeholder="Tell us more about your inquiry..."
                    />
                  </Form.Item>

                  <Form.Item>
                    <Button
                      type="primary"
                      size="large"
                      htmlType="submit"
                      loading={loading}
                      icon={<SendOutlined />}
                      block
                    >
                      Send Message
                    </Button>
                  </Form.Item>
                </Form>
              </Card>
            </Col>
          </Row>
        </div>
      </section>

      {/* Map Section (Placeholder) */}
      <section className="map-section">
        <div className="map-placeholder">
          <EnvironmentOutlined style={{ fontSize: 64, color: '#667eea' }} />
          <p>123 Payment Street, San Francisco, CA 94102</p>
          <Button type="primary" href="https://maps.google.com" target="_blank">
            View on Map
          </Button>
        </div>
      </section>
    </div>
  );
};

export default Contact;
