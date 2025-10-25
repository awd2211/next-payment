// Performance monitoring and analytics utilities

interface PerformanceMetrics {
  pageLoadTime: number;
  domContentLoaded: number;
  firstPaint: number;
  firstContentfulPaint: number;
  timeToInteractive: number;
}

export class Analytics {
  private static instance: Analytics;
  private performanceData: PerformanceMetrics | null = null;

  private constructor() {
    if (typeof window !== 'undefined') {
      this.init();
    }
  }

  public static getInstance(): Analytics {
    if (!Analytics.instance) {
      Analytics.instance = new Analytics();
    }
    return Analytics.instance;
  }

  private init() {
    // Wait for page to fully load
    if (document.readyState === 'complete') {
      this.collectPerformanceMetrics();
    } else {
      window.addEventListener('load', () => {
        this.collectPerformanceMetrics();
      });
    }
  }

  private collectPerformanceMetrics() {
    if (!window.performance || !window.performance.timing) {
      console.warn('Performance API not supported');
      return;
    }

    const perfData = window.performance.timing;
    const navigation = perfData.navigationStart;

    this.performanceData = {
      pageLoadTime: perfData.loadEventEnd - navigation,
      domContentLoaded: perfData.domContentLoadedEventEnd - navigation,
      firstPaint: this.getFirstPaint(),
      firstContentfulPaint: this.getFirstContentfulPaint(),
      timeToInteractive: perfData.domInteractive - navigation,
    };

    // Log performance metrics (in production, send to analytics service)
    this.logMetrics();
  }

  private getFirstPaint(): number {
    const paint = window.performance.getEntriesByType('paint');
    const fp = paint.find((entry) => entry.name === 'first-paint');
    return fp ? fp.startTime : 0;
  }

  private getFirstContentfulPaint(): number {
    const paint = window.performance.getEntriesByType('paint');
    const fcp = paint.find((entry) => entry.name === 'first-contentful-paint');
    return fcp ? fcp.startTime : 0;
  }

  private logMetrics() {
    if (!this.performanceData) return;

    const metrics = this.performanceData;

    console.group('📊 Performance Metrics');
    console.log('Page Load Time:', `${metrics.pageLoadTime}ms`);
    console.log('DOM Content Loaded:', `${metrics.domContentLoaded}ms`);
    console.log('First Paint:', `${metrics.firstPaint}ms`);
    console.log('First Contentful Paint:', `${metrics.firstContentfulPaint}ms`);
    console.log('Time to Interactive:', `${metrics.timeToInteractive}ms`);
    console.groupEnd();

    // Performance grading
    this.gradePerformance();
  }

  private gradePerformance() {
    if (!this.performanceData) return;

    const { pageLoadTime, firstContentfulPaint } = this.performanceData;
    let grade = 'A';
    let color = '#52c41a';

    if (pageLoadTime > 5000 || firstContentfulPaint > 3000) {
      grade = 'F';
      color = '#f5222d';
    } else if (pageLoadTime > 4000 || firstContentfulPaint > 2500) {
      grade = 'D';
      color = '#fa8c16';
    } else if (pageLoadTime > 3000 || firstContentfulPaint > 2000) {
      grade = 'C';
      color = '#faad14';
    } else if (pageLoadTime > 2000 || firstContentfulPaint > 1500) {
      grade = 'B';
      color = '#52c41a';
    }

    console.log(
      `%c Performance Grade: ${grade}`,
      `font-size: 16px; font-weight: bold; color: ${color}`
    );
  }

  // Track page views
  public trackPageView(pageName: string, path: string) {
    console.log(`📄 Page View: ${pageName} (${path})`);

    // In production, send to analytics service (e.g., Google Analytics)
    if (typeof window !== 'undefined' && (window as any).gtag) {
      (window as any).gtag('config', 'GA_MEASUREMENT_ID', {
        page_path: path,
        page_title: pageName,
      });
    }
  }

  // Track custom events
  public trackEvent(category: string, action: string, label?: string, value?: number) {
    console.log(`📌 Event: ${category} - ${action}`, label, value);

    // In production, send to analytics service
    if (typeof window !== 'undefined' && (window as any).gtag) {
      (window as any).gtag('event', action, {
        event_category: category,
        event_label: label,
        value: value,
      });
    }
  }

  // Track errors
  public trackError(error: Error, errorInfo?: any) {
    console.error('❌ Error tracked:', error, errorInfo);

    // In production, send to error tracking service (e.g., Sentry)
    if (typeof window !== 'undefined' && (window as any).Sentry) {
      (window as any).Sentry.captureException(error, { extra: errorInfo });
    }
  }

  // Get current performance data
  public getMetrics(): PerformanceMetrics | null {
    return this.performanceData;
  }

  // Track user interactions
  public trackClick(element: string, location: string) {
    this.trackEvent('User Interaction', 'Click', `${element} (${location})`);
  }

  public trackFormSubmit(formName: string, success: boolean) {
    this.trackEvent('Form', success ? 'Submit Success' : 'Submit Failed', formName);
  }

  public trackScroll(depth: number) {
    const depths = [25, 50, 75, 100];
    const milestone = depths.find(d => Math.abs(d - depth) < 5);

    if (milestone) {
      this.trackEvent('Scroll', 'Scroll Depth', `${milestone}%`, milestone);
    }
  }
}

// Export singleton instance
export const analytics = Analytics.getInstance();

// Hook for tracking scroll depth
export const useScrollTracking = () => {
  if (typeof window === 'undefined') return;

  let ticking = false;

  const trackScroll = () => {
    const winHeight = window.innerHeight;
    const docHeight = document.documentElement.scrollHeight;
    const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
    const scrollPercent = (scrollTop / (docHeight - winHeight)) * 100;

    analytics.trackScroll(scrollPercent);
    ticking = false;
  };

  window.addEventListener('scroll', () => {
    if (!ticking) {
      window.requestAnimationFrame(trackScroll);
      ticking = true;
    }
  });
};

export default analytics;
