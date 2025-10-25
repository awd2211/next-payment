import { useTranslation } from 'react-i18next';
import { Card, Row, Col, Statistic, Timeline, Avatar } from 'antd';
import {
  TeamOutlined,
  RocketOutlined,
  GlobalOutlined,
  TrophyOutlined,
  HeartOutlined,
  SafetyOutlined,
  ThunderboltOutlined,
  StarOutlined,
} from '@ant-design/icons';
import './style.css';

const About = () => {
  const { t } = useTranslation();

  const stats = [
    { icon: <TeamOutlined />, value: '500+', label: 'Team Members', color: '#667eea' },
    { icon: <GlobalOutlined />, value: '150+', label: 'Countries', color: '#43e97b' },
    { icon: <RocketOutlined />, value: '$10B+', label: 'Processed', color: '#fa709a' },
    { icon: <TrophyOutlined />, value: '10K+', label: 'Customers', color: '#4facfe' },
  ];

  const values = [
    {
      icon: <HeartOutlined />,
      title: 'Customer First',
      description: 'We put our customers at the heart of everything we do',
      gradient: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    },
    {
      icon: <SafetyOutlined />,
      title: 'Security & Trust',
      description: 'Bank-level security and compliance standards',
      gradient: 'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)',
    },
    {
      icon: <ThunderboltOutlined />,
      title: 'Innovation',
      description: 'Constantly pushing the boundaries of payment technology',
      gradient: 'linear-gradient(135deg, #fa709a 0%, #fee140 100%)',
    },
    {
      icon: <StarOutlined />,
      title: 'Excellence',
      description: 'Committed to delivering the highest quality service',
      gradient: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
    },
  ];

  const team = [
    { name: 'John Smith', role: 'CEO & Founder', avatar: 'https://i.pravatar.cc/150?img=1' },
    { name: 'Sarah Johnson', role: 'CTO', avatar: 'https://i.pravatar.cc/150?img=2' },
    { name: 'Michael Chen', role: 'Head of Product', avatar: 'https://i.pravatar.cc/150?img=3' },
    { name: 'Emily Davis', role: 'Head of Engineering', avatar: 'https://i.pravatar.cc/150?img=4' },
  ];

  const milestones = [
    { year: '2020', title: 'Company Founded', desc: 'Started with a vision to revolutionize payments' },
    { year: '2021', title: 'Series A Funding', desc: 'Raised $10M to expand globally' },
    { year: '2022', title: 'Global Expansion', desc: 'Launched in 50+ countries' },
    { year: '2023', title: 'Major Milestone', desc: 'Processed $1B in transactions' },
    { year: '2024', title: 'Market Leader', desc: 'Became a leading payment platform' },
  ];

  return (
    <div className="about-page">
      {/* Hero Section */}
      <section className="about-hero">
        <div className="about-hero-content">
          <h1 className="hero-title">About Us</h1>
          <p className="hero-subtitle">
            Building the future of global payments, one transaction at a time
          </p>
        </div>
      </section>

      {/* Mission Section */}
      <section className="mission-section">
        <div className="section-container">
          <Row gutter={[48, 48]} align="middle">
            <Col xs={24} md={12}>
              <div className="mission-content">
                <h2 className="section-title">Our Mission</h2>
                <p className="mission-text">
                  To empower businesses worldwide with seamless, secure, and innovative payment solutions
                  that enable growth and success in the digital economy.
                </p>
                <p className="mission-text">
                  We believe that every business, regardless of size or location, deserves access to
                  world-class payment infrastructure.
                </p>
              </div>
            </Col>
            <Col xs={24} md={12}>
              <div className="stats-grid">
                {stats.map((stat, index) => (
                  <Card key={index} className="stat-card-about" bordered={false}>
                    <div className="stat-icon" style={{ color: stat.color }}>
                      {stat.icon}
                    </div>
                    <Statistic
                      value={stat.value}
                      valueStyle={{ color: stat.color, fontSize: '32px', fontWeight: 700 }}
                    />
                    <div className="stat-label">{stat.label}</div>
                  </Card>
                ))}
              </div>
            </Col>
          </Row>
        </div>
      </section>

      {/* Values Section */}
      <section className="values-section">
        <div className="section-container">
          <div className="section-header">
            <h2 className="section-title">Our Values</h2>
            <p className="section-subtitle">The principles that guide everything we do</p>
          </div>
          <Row gutter={[24, 24]}>
            {values.map((value, index) => (
              <Col xs={24} sm={12} md={6} key={index}>
                <Card className="value-card" bordered={false}>
                  <div className="value-icon-wrapper" style={{ background: value.gradient }}>
                    <div className="value-icon">{value.icon}</div>
                  </div>
                  <h3 className="value-title">{value.title}</h3>
                  <p className="value-description">{value.description}</p>
                </Card>
              </Col>
            ))}
          </Row>
        </div>
      </section>

      {/* Team Section */}
      <section className="team-section">
        <div className="section-container">
          <div className="section-header">
            <h2 className="section-title">Leadership Team</h2>
            <p className="section-subtitle">Meet the people driving our vision forward</p>
          </div>
          <Row gutter={[32, 32]}>
            {team.map((member, index) => (
              <Col xs={24} sm={12} md={6} key={index}>
                <Card className="team-card" bordered={false}>
                  <Avatar
                    size={120}
                    src={member.avatar}
                    className="team-avatar"
                  />
                  <h3 className="team-name">{member.name}</h3>
                  <p className="team-role">{member.role}</p>
                </Card>
              </Col>
            ))}
          </Row>
        </div>
      </section>

      {/* Timeline Section */}
      <section className="timeline-section-about">
        <div className="section-container">
          <div className="section-header">
            <h2 className="section-title">Our Journey</h2>
            <p className="section-subtitle">Key milestones in our growth story</p>
          </div>
          <Timeline mode="alternate" className="about-timeline">
            {milestones.map((milestone, index) => (
              <Timeline.Item
                key={index}
                dot={<RocketOutlined style={{ fontSize: '20px' }} />}
                color="blue"
              >
                <div className="timeline-content">
                  <h3 className="timeline-year">{milestone.year}</h3>
                  <h4 className="timeline-title">{milestone.title}</h4>
                  <p className="timeline-desc">{milestone.desc}</p>
                </div>
              </Timeline.Item>
            ))}
          </Timeline>
        </div>
      </section>

      {/* CTA Section */}
      <section className="about-cta">
        <div className="section-container">
          <h2 className="cta-title">Join Our Mission</h2>
          <p className="cta-description">
            We're always looking for talented individuals who share our passion
          </p>
          <div className="cta-actions">
            <a href="/careers" className="cta-button">View Open Positions</a>
            <a href="/contact" className="cta-button-outline">Get in Touch</a>
          </div>
        </div>
      </section>
    </div>
  );
};

export default About;
