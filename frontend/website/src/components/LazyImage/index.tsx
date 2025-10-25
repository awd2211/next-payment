import { useState, useEffect, useRef, CSSProperties } from 'react';
import { Skeleton } from 'antd';
import './style.css';

interface LazyImageProps {
  src: string;
  alt: string;
  placeholder?: string;
  width?: number | string;
  height?: number | string;
  className?: string;
  style?: CSSProperties;
  threshold?: number;
  onLoad?: () => void;
  onError?: () => void;
}

const LazyImage: React.FC<LazyImageProps> = ({
  src,
  alt,
  placeholder,
  width,
  height,
  className = '',
  style,
  threshold = 0.1,
  onLoad,
  onError,
}) => {
  const [isLoaded, setIsLoaded] = useState(false);
  const [isInView, setIsInView] = useState(false);
  const [hasError, setHasError] = useState(false);
  const imgRef = useRef<HTMLImageElement>(null);

  useEffect(() => {
    if (!imgRef.current) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsInView(true);
          observer.disconnect();
        }
      },
      { threshold }
    );

    observer.observe(imgRef.current);

    return () => {
      observer.disconnect();
    };
  }, [threshold]);

  const handleLoad = () => {
    setIsLoaded(true);
    onLoad?.();
  };

  const handleError = () => {
    setHasError(true);
    onError?.();
  };

  return (
    <div
      className={`lazy-image-wrapper ${className}`}
      style={{ width, height, ...style }}
    >
      {!isLoaded && !hasError && (
        <Skeleton.Image
          active
          style={{
            width: width || '100%',
            height: height || '100%',
          }}
        />
      )}

      {hasError && (
        <div className="lazy-image-error">
          <span>Failed to load image</span>
        </div>
      )}

      <img
        ref={imgRef}
        src={isInView ? src : placeholder || ''}
        alt={alt}
        className={`lazy-image ${isLoaded ? 'loaded' : ''}`}
        onLoad={handleLoad}
        onError={handleError}
        style={{
          display: isLoaded && !hasError ? 'block' : 'none',
        }}
      />
    </div>
  );
};

export default LazyImage;
