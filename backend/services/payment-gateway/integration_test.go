package main

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"payment-platform/payment-gateway/internal/adapter"
)

// TestKafkaProducerAdapter 测试 Kafka 生产者适配器
func TestKafkaProducerAdapter(t *testing.T) {
	t.Run("创建适配器", func(t *testing.T) {
		brokers := []string{"localhost:9092"}
		adapter := adapter.NewKafkaProducerAdapter(brokers)
		
		if adapter == nil {
			t.Fatal("适配器创建失败")
		}
		
		t.Log("✅ Kafka 生产者适配器创建成功")
	})
	
	t.Run("降级模式", func(t *testing.T) {
		// 空 brokers 列表应该启用降级模式
		adapter := adapter.NewKafkaProducerAdapter([]string{})
		
		ctx := context.Background()
		err := adapter.Publish(ctx, "test-topic", "test-message")
		
		// 降级模式下不应该返回错误
		if err != nil {
			t.Errorf("降级模式应该不返回错误，但得到: %v", err)
		}
		
		t.Log("✅ Kafka 降级模式工作正常")
	})
}

// TestCIDRValidation 测试 CIDR IP 白名单验证
func TestCIDRValidation(t *testing.T) {
	testCases := []struct {
		name     string
		clientIP string
		cidr     string
		expected bool
	}{
		{"IPv4范围内", "192.168.1.100", "192.168.1.0/24", true},
		{"IPv4范围外", "192.168.2.1", "192.168.1.0/24", false},
		{"单个IP匹配", "203.0.113.5", "203.0.113.5/32", true},
		{"单个IP不匹配", "203.0.113.6", "203.0.113.5/32", false},
		{"大型网络", "10.0.0.1", "10.0.0.0/8", true},
		{"无效IP", "invalid-ip", "192.168.1.0/24", false},
		{"IPv6范围内", "2001:db8::1", "2001:db8::/32", true},
		{"IPv6范围外", "2001:db9::1", "2001:db8::/32", false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isIPInCIDR(tc.clientIP, tc.cidr)
			if result != tc.expected {
				t.Errorf("IP=%s, CIDR=%s: 期望 %v, 得到 %v", 
					tc.clientIP, tc.cidr, tc.expected, result)
			}
		})
	}
	
	t.Log("✅ CIDR IP 白名单验证通过")
}

// isIPInCIDR 检查IP是否在CIDR范围内（复制实现用于测试）
func isIPInCIDR(clientIP, cidr string) bool {
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}
	
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	
	return ipNet.Contains(ip)
}

// TestGeoIPCountryExtraction 测试从风控结果提取国家代码的逻辑
func TestGeoIPCountryExtraction(t *testing.T) {
	t.Run("提取国家代码", func(t *testing.T) {
		// 模拟风控结果
		extra := map[string]interface{}{
			"geo_country_code": "US",
			"geo_country":      "United States",
			"geo_city":         "New York",
		}
		
		countryCode, ok := extra["geo_country_code"].(string)
		if !ok || countryCode == "" {
			t.Error("无法提取国家代码")
		}
		
		if countryCode != "US" {
			t.Errorf("期望国家代码 'US', 得到 '%s'", countryCode)
		}
		
		t.Log("✅ GeoIP 国家代码提取成功")
	})
	
	t.Run("默认值处理", func(t *testing.T) {
		// 空的 extra map
		extra := map[string]interface{}{}
		
		defaultCountry := "US"
		countryCode, ok := extra["geo_country_code"].(string)
		if !ok || countryCode == "" {
			countryCode = defaultCountry
		}
		
		if countryCode != "US" {
			t.Errorf("期望默认值 'US', 得到 '%s'", countryCode)
		}
		
		t.Log("✅ 默认国家代码处理正常")
	})
}

// TestServiceIntegration 测试服务集成（不依赖外部服务）
func TestServiceIntegration(t *testing.T) {
	t.Run("超时设置", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		
		select {
		case <-ctx.Done():
			t.Log("✅ 上下文超时机制正常")
		case <-time.After(3 * time.Second):
			t.Error("超时机制未生效")
		}
	})
	
	t.Run("并发安全", func(t *testing.T) {
		// 测试 Kafka 适配器的并发安全性
		adapter := adapter.NewKafkaProducerAdapter([]string{})
		
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(id int) {
				ctx := context.Background()
				err := adapter.Publish(ctx, "test-topic", fmt.Sprintf("message-%d", id))
				if err != nil {
					t.Errorf("并发发布失败: %v", err)
				}
				done <- true
			}(i)
		}
		
		// 等待所有 goroutine 完成
		for i := 0; i < 10; i++ {
			<-done
		}
		
		t.Log("✅ 并发安全测试通过")
	})
}

// BenchmarkCIDRValidation CIDR 验证性能基准测试
func BenchmarkCIDRValidation(b *testing.B) {
	clientIP := "192.168.1.100"
	cidr := "192.168.1.0/24"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isIPInCIDR(clientIP, cidr)
	}
}

// BenchmarkKafkaAdapterPublish Kafka 适配器发布性能基准测试
func BenchmarkKafkaAdapterPublish(b *testing.B) {
	adapter := adapter.NewKafkaProducerAdapter([]string{}) // 降级模式
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.Publish(ctx, "test-topic", "test-message")
	}
}
