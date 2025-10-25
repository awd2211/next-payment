package service

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/payment-platform/pkg/kafka"
	"github.com/payment-platform/pkg/logger"
	"go.uber.org/zap"
)

// ConfigChangeEvent 配置变更事件
type ConfigChangeEvent struct {
	EventID     string    `json:"event_id"`
	ConfigID    uuid.UUID `json:"config_id"`
	ServiceName string    `json:"service_name"`
	ConfigKey   string    `json:"config_key"`
	Environment string    `json:"environment"`
	OldValue    string    `json:"old_value,omitempty"`
	NewValue    string    `json:"new_value"`
	ChangeType  string    `json:"change_type"` // created, updated, deleted, rollback
	ChangedBy   string    `json:"changed_by"`
	Timestamp   time.Time `json:"timestamp"`
}

// ConfigNotifier 配置变更通知服务
type ConfigNotifier interface {
	// 发布配置变更事件（通过 Kafka）
	PublishConfigChange(ctx context.Context, event *ConfigChangeEvent) error

	// WebSocket 订阅管理
	Subscribe(clientID string, filters map[string]string) chan *ConfigChangeEvent
	Unsubscribe(clientID string)

	// 关闭通知服务
	Close() error
}

type configNotifier struct {
	kafkaProducer *kafka.Producer
	subscribers   map[string]*subscriber
	mu            sync.RWMutex
	eventBus      chan *ConfigChangeEvent
	stopChan      chan struct{}
}

type subscriber struct {
	clientID string
	filters  map[string]string // serviceName, environment, configKey
	eventCh  chan *ConfigChangeEvent
}

// NewConfigNotifier 创建配置通知服务
func NewConfigNotifier(kafkaBrokers []string) (ConfigNotifier, error) {
	// 初始化 Kafka Producer
	var producer *kafka.Producer
	if len(kafkaBrokers) > 0 && kafkaBrokers[0] != "" {
		producer = kafka.NewProducer(kafka.ProducerConfig{
			Brokers: kafkaBrokers,
			Topic:   "config-changes",
		})
		logger.Info("Kafka producer initialized for config notifier", zap.Strings("brokers", kafkaBrokers))
	} else {
		logger.Warn("Kafka brokers not configured, notifications will use WebSocket only")
	}

	notifier := &configNotifier{
		kafkaProducer: producer,
		subscribers:   make(map[string]*subscriber),
		eventBus:      make(chan *ConfigChangeEvent, 100),
		stopChan:      make(chan struct{}),
	}

	// 启动事件分发协程
	go notifier.dispatchEvents()

	return notifier, nil
}

// PublishConfigChange 发布配置变更事件
func (n *configNotifier) PublishConfigChange(ctx context.Context, event *ConfigChangeEvent) error {
	event.EventID = uuid.New().String()
	event.Timestamp = time.Now()

	// 1. 发送到 Kafka（异步通知其他服务）
	if n.kafkaProducer != nil {
		if err := n.kafkaProducer.Publish(ctx, event.ConfigID.String(), event); err != nil {
			logger.Error("Failed to send config change to Kafka", zap.Error(err))
		} else {
			logger.Info("Config change published to Kafka",
				zap.String("event_id", event.EventID),
				zap.String("service_name", event.ServiceName),
				zap.String("config_key", event.ConfigKey))
		}
	}

	// 2. 发送到 WebSocket 订阅者（实时推送）
	select {
	case n.eventBus <- event:
	case <-time.After(1 * time.Second):
		logger.Warn("Event bus full, dropping event", zap.String("event_id", event.EventID))
	}

	return nil
}

// Subscribe 订阅配置变更（WebSocket 客户端）
func (n *configNotifier) Subscribe(clientID string, filters map[string]string) chan *ConfigChangeEvent {
	n.mu.Lock()
	defer n.mu.Unlock()

	// 如果已经订阅，先取消旧订阅
	if old, exists := n.subscribers[clientID]; exists {
		close(old.eventCh)
	}

	sub := &subscriber{
		clientID: clientID,
		filters:  filters,
		eventCh:  make(chan *ConfigChangeEvent, 10),
	}
	n.subscribers[clientID] = sub

	logger.Info("Client subscribed to config changes",
		zap.String("client_id", clientID),
		zap.Any("filters", filters))

	return sub.eventCh
}

// Unsubscribe 取消订阅
func (n *configNotifier) Unsubscribe(clientID string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if sub, exists := n.subscribers[clientID]; exists {
		close(sub.eventCh)
		delete(n.subscribers, clientID)
		logger.Info("Client unsubscribed", zap.String("client_id", clientID))
	}
}

// dispatchEvents 分发事件到所有订阅者
func (n *configNotifier) dispatchEvents() {
	for {
		select {
		case event := <-n.eventBus:
			n.mu.RLock()
			for _, sub := range n.subscribers {
				// 检查过滤条件
				if n.matchFilters(event, sub.filters) {
					select {
					case sub.eventCh <- event:
					case <-time.After(100 * time.Millisecond):
						logger.Warn("Subscriber channel full, skipping event",
							zap.String("client_id", sub.clientID),
							zap.String("event_id", event.EventID))
					}
				}
			}
			n.mu.RUnlock()

		case <-n.stopChan:
			logger.Info("Config notifier stopped")
			return
		}
	}
}

// matchFilters 检查事件是否匹配订阅过滤条件
func (n *configNotifier) matchFilters(event *ConfigChangeEvent, filters map[string]string) bool {
	if len(filters) == 0 {
		return true // 无过滤条件，匹配所有
	}

	if serviceName, ok := filters["service_name"]; ok && serviceName != "" {
		if event.ServiceName != serviceName {
			return false
		}
	}

	if environment, ok := filters["environment"]; ok && environment != "" {
		if event.Environment != environment {
			return false
		}
	}

	if configKey, ok := filters["config_key"]; ok && configKey != "" {
		if event.ConfigKey != configKey {
			return false
		}
	}

	return true
}

// Close 关闭通知服务
func (n *configNotifier) Close() error {
	close(n.stopChan)

	// 关闭所有订阅者
	n.mu.Lock()
	for _, sub := range n.subscribers {
		close(sub.eventCh)
	}
	n.subscribers = nil
	n.mu.Unlock()

	// 关闭 Kafka Producer
	if n.kafkaProducer != nil {
		return n.kafkaProducer.Close()
	}

	return nil
}
