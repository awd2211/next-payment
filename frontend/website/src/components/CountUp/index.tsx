import { useState, useEffect, useRef } from 'react';

interface CountUpProps {
  end: number;
  duration?: number;
  suffix?: string;
  prefix?: string;
  decimals?: number;
  separator?: string;
  onEnd?: () => void;
  startOnView?: boolean;
}

const CountUp: React.FC<CountUpProps> = ({
  end,
  duration = 2000,
  suffix = '',
  prefix = '',
  decimals = 0,
  separator = ',',
  onEnd,
  startOnView = true,
}) => {
  const [count, setCount] = useState(0);
  const [hasStarted, setHasStarted] = useState(!startOnView);
  const elementRef = useRef<HTMLSpanElement>(null);

  useEffect(() => {
    if (!startOnView) {
      startCounting();
      return;
    }

    if (!elementRef.current) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting && !hasStarted) {
          setHasStarted(true);
        }
      },
      { threshold: 0.5 }
    );

    observer.observe(elementRef.current);

    return () => {
      observer.disconnect();
    };
  }, [startOnView, hasStarted]);

  useEffect(() => {
    if (hasStarted) {
      startCounting();
    }
  }, [hasStarted]);

  const startCounting = () => {
    const startTime = Date.now();
    const startValue = 0;
    const endValue = end;

    const easeOutQuart = (t: number): number => {
      return 1 - Math.pow(1 - t, 4);
    };

    const animate = () => {
      const now = Date.now();
      const progress = Math.min((now - startTime) / duration, 1);
      const easedProgress = easeOutQuart(progress);
      const currentValue = startValue + (endValue - startValue) * easedProgress;

      setCount(currentValue);

      if (progress < 1) {
        requestAnimationFrame(animate);
      } else {
        setCount(endValue);
        onEnd?.();
      }
    };

    requestAnimationFrame(animate);
  };

  const formatNumber = (num: number): string => {
    const fixed = num.toFixed(decimals);
    const parts = fixed.split('.');
    const integerPart = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, separator);
    return decimals > 0 ? `${integerPart}.${parts[1]}` : integerPart;
  };

  return (
    <span ref={elementRef}>
      {prefix}
      {formatNumber(count)}
      {suffix}
    </span>
  );
};

export default CountUp;
