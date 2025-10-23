import { GlobalOutlined } from '@ant-design/icons';
import { Dropdown, MenuProps } from 'antd';
import { useTranslation } from 'react-i18next';

const LanguageSwitch = () => {
  const { i18n } = useTranslation();

  const items: MenuProps['items'] = [
    {
      key: 'en',
      label: 'English',
    },
    {
      key: 'zh-CN',
      label: '简体中文',
    },
  ];

  const handleMenuClick: MenuProps['onClick'] = ({ key }) => {
    i18n.changeLanguage(key);
  };

  return (
    <Dropdown menu={{ items, onClick: handleMenuClick }} placement="bottomRight">
      <a onClick={(e) => e.preventDefault()} style={{ color: 'inherit' }}>
        <GlobalOutlined style={{ fontSize: 18, marginRight: 4 }} />
        {i18n.language === 'zh-CN' ? '中文' : 'EN'}
      </a>
    </Dropdown>
  );
};

export default LanguageSwitch;
