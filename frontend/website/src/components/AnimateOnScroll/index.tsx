import { useEffect, useRef, useState, ReactNode } from 'react';
import './style.css';

interface AnimateOnScrollProps {
  children: ReactNode;
  animation?: 'fade-up' | 'fade-down' | 'fade-left' | 'fade-right' | 'zoom-in' | 'flip' | 'slide-up';
  delay?: number;
  duration?: number;
  threshold?: number;
  once?: boolean;
  className?: string;
}

const AnimateOnScroll: React.FC<AnimateOnScrollProps> = ({
  children,
  animation = 'fade-up',
  delay = 0,
  duration = 600,
  threshold = 0.1,
  once = true,
  className = '',
}) => {
  const [isVisible, setIsVisible] = useState(false);
  const elementRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const element = elementRef.current;
    if (!element) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsVisible(true);
          if (once) {
            observer.disconnect();
          }
        } else if (!once) {
          setIsVisible(false);
        }
      },
      { threshold }
    );

    observer.observe(element);

    return () => {
      observer.disconnect();
    };
  }, [threshold, once]);

  return (
    <div
      ref={elementRef}
      className={`animate-on-scroll ${animation} ${isVisible ? 'visible' : ''} ${className}`}
      style={{
        animationDelay: `${delay}ms`,
        animationDuration: `${duration}ms`,
      }}
    >
      {children}
    </div>
  );
};

export default AnimateOnScroll;
