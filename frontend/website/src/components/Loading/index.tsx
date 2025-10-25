import { Spin } from 'antd';
import { LoadingOutlined } from '@ant-design/icons';
import './style.css';

interface LoadingProps {
  fullscreen?: boolean;
  tip?: string;
  size?: 'small' | 'default' | 'large';
}

const Loading: React.FC<LoadingProps> = ({
  fullscreen = true,
  tip = 'Loading...',
  size = 'large'
}) => {
  const loadingIcon = <LoadingOutlined style={{ fontSize: size === 'large' ? 48 : 24 }} spin />;

  if (fullscreen) {
    return (
      <div className="loading-fullscreen">
        <div className="loading-content">
          <Spin indicator={loadingIcon} size={size} />
          {tip && <div className="loading-tip">{tip}</div>}
        </div>
      </div>
    );
  }

  return (
    <div className="loading-inline">
      <Spin indicator={loadingIcon} tip={tip} size={size} />
    </div>
  );
};

export default Loading;
