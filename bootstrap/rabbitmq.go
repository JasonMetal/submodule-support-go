package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/JasonMetal/submodule-support-go.git/helper/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQConfig RabbitMQ配置结构
type RabbitMQConfig struct {
	Host      string                  `yaml:"host"`
	Port      int                     `yaml:"port"`
	Username  string                  `yaml:"username"`
	Password  string                  `yaml:"password"`
	Vhost     string                  `yaml:"vhost"`
	Pool      RabbitMQPoolConfig      `yaml:"pool"`
	Producer  RabbitMQProducerConfig  `yaml:"producer"`
	Consumer  RabbitMQConsumerConfig  `yaml:"consumer"`
	Reconnect RabbitMQReconnectConfig `yaml:"reconnect"`
	Queue     RabbitMQQueueConfig     `yaml:"queue"`
	Exchange  RabbitMQExchangeConfig  `yaml:"exchange"`
}

type RabbitMQPoolConfig struct {
	MaxOpen     int `yaml:"max_open"`
	MaxIdle     int `yaml:"max_idle"`
	MaxLifetime int `yaml:"max_lifetime"`
}

type RabbitMQProducerConfig struct {
	ConfirmMode bool `yaml:"confirm_mode"`
	Mandatory   bool `yaml:"mandatory"`
	Immediate   bool `yaml:"immediate"`
}

type RabbitMQConsumerConfig struct {
	AutoAck       bool `yaml:"auto_ack"`
	PrefetchCount int  `yaml:"prefetch_count"`
	PrefetchSize  int  `yaml:"prefetch_size"`
}

type RabbitMQReconnectConfig struct {
	MaxRetries int `yaml:"max_retries"`
	Interval   int `yaml:"interval"`
}

type RabbitMQQueueConfig struct {
	Durable    bool `yaml:"durable"`
	AutoDelete bool `yaml:"auto_delete"`
	Exclusive  bool `yaml:"exclusive"`
	NoWait     bool `yaml:"no_wait"`
}

type RabbitMQExchangeConfig struct {
	Durable    bool `yaml:"durable"`
	AutoDelete bool `yaml:"auto_delete"`
	Internal   bool `yaml:"internal"`
	NoWait     bool `yaml:"no_wait"`
}

// RabbitMQManager RabbitMQ管理器
type RabbitMQManager struct {
	config  *RabbitMQConfig
	conn    *amqp.Connection
	channel *amqp.Channel
	mu      sync.RWMutex
	closed  bool
}

var (
	rabbitmqManager *RabbitMQManager
	rabbitmqOnce    sync.Once
)

// InitRabbitMQ 初始化RabbitMQ连接
func InitRabbitMQ() error {
	configPath := fmt.Sprintf("%sconfig/%s/rabbitmq.yml", ProjectPath(), DevEnv)

	log.Printf("Loading RabbitMQ config from: %s", configPath)

	rmqConfig, err := loadRabbitMQConfig(configPath)
	if err != nil {
		return fmt.Errorf("加载RabbitMQ配置失败: %v", err)
	}

	manager := &RabbitMQManager{
		config: rmqConfig,
	}

	// 建立连接
	if err := manager.connect(); err != nil {
		return fmt.Errorf("连接RabbitMQ失败: %v", err)
	}

	rabbitmqManager = manager

	log.Printf("RabbitMQ初始化成功, Host: %s:%d", rmqConfig.Host, rmqConfig.Port)
	return nil
}

// loadRabbitMQConfig 加载RabbitMQ配置
func loadRabbitMQConfig(configPath string) (*RabbitMQConfig, error) {
	cfg, err := config.GetConfig(configPath)
	if err != nil {
		return nil, err
	}

	rmqConfig := &RabbitMQConfig{}

	// 加载基本配置
	rmqConfig.Host, _ = cfg.String("host")
	rmqConfig.Port, _ = cfg.Int("port")
	rmqConfig.Username, _ = cfg.String("username")
	rmqConfig.Password, _ = cfg.String("password")
	rmqConfig.Vhost, _ = cfg.String("vhost")

	// 加载连接池配置
	rmqConfig.Pool.MaxOpen, _ = cfg.Int("pool.max_open")
	rmqConfig.Pool.MaxIdle, _ = cfg.Int("pool.max_idle")
	rmqConfig.Pool.MaxLifetime, _ = cfg.Int("pool.max_lifetime")

	// 加载生产者配置
	rmqConfig.Producer.ConfirmMode, _ = cfg.Bool("producer.confirm_mode")
	rmqConfig.Producer.Mandatory, _ = cfg.Bool("producer.mandatory")
	rmqConfig.Producer.Immediate, _ = cfg.Bool("producer.immediate")

	// 加载消费者配置
	rmqConfig.Consumer.AutoAck, _ = cfg.Bool("consumer.auto_ack")
	rmqConfig.Consumer.PrefetchCount, _ = cfg.Int("consumer.prefetch_count")
	rmqConfig.Consumer.PrefetchSize, _ = cfg.Int("consumer.prefetch_size")

	// 加载重连配置
	rmqConfig.Reconnect.MaxRetries, _ = cfg.Int("reconnect.max_retries")
	rmqConfig.Reconnect.Interval, _ = cfg.Int("reconnect.interval")

	// 加载队列配置
	rmqConfig.Queue.Durable, _ = cfg.Bool("queue.durable")
	rmqConfig.Queue.AutoDelete, _ = cfg.Bool("queue.auto_delete")
	rmqConfig.Queue.Exclusive, _ = cfg.Bool("queue.exclusive")
	rmqConfig.Queue.NoWait, _ = cfg.Bool("queue.no_wait")

	// 加载交换机配置
	rmqConfig.Exchange.Durable, _ = cfg.Bool("exchange.durable")
	rmqConfig.Exchange.AutoDelete, _ = cfg.Bool("exchange.auto_delete")
	rmqConfig.Exchange.Internal, _ = cfg.Bool("exchange.internal")
	rmqConfig.Exchange.NoWait, _ = cfg.Bool("exchange.no_wait")

	return rmqConfig, nil
}

// connect 建立连接
func (m *RabbitMQManager) connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 构建连接URL
	url := fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		m.config.Username,
		m.config.Password,
		m.config.Host,
		m.config.Port,
		m.config.Vhost,
	)

	// 建立连接
	conn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}

	// 创建Channel
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("创建Channel失败: %v", err)
	}

	// 设置QoS
	if m.config.Consumer.PrefetchCount > 0 {
		err = channel.Qos(
			m.config.Consumer.PrefetchCount,
			m.config.Consumer.PrefetchSize,
			false,
		)
		if err != nil {
			channel.Close()
			conn.Close()
			return fmt.Errorf("设置QoS失败: %v", err)
		}
	}

	m.conn = conn
	m.channel = channel
	m.closed = false

	// 监听连接关闭
	go m.handleReconnect()

	return nil
}

// handleReconnect 处理重连
func (m *RabbitMQManager) handleReconnect() {
	notifyClose := make(chan *amqp.Error)
	m.channel.NotifyClose(notifyClose)

	select {
	case err := <-notifyClose:
		if err != nil {
			log.Printf("RabbitMQ连接断开: %v, 尝试重连...", err)
			m.reconnect()
		}
	}
}

// reconnect 重新连接
func (m *RabbitMQManager) reconnect() {
	for i := 0; i < m.config.Reconnect.MaxRetries; i++ {
		time.Sleep(time.Duration(m.config.Reconnect.Interval) * time.Second)

		log.Printf("RabbitMQ重连尝试 %d/%d", i+1, m.config.Reconnect.MaxRetries)

		if err := m.connect(); err != nil {
			log.Printf("重连失败: %v", err)
			continue
		}

		log.Println("RabbitMQ重连成功")
		return
	}

	log.Println("RabbitMQ重连失败，已达到最大重试次数")
}

// PublishSimple 发送简单消息
func PublishSimple(queueName string, message string) error {
	return PublishSimpleWithContext(context.Background(), queueName, message)
}

// PublishSimpleWithContext 带上下文发送简单消息
func PublishSimpleWithContext(ctx context.Context, queueName string, message string) error {
	if rabbitmqManager == nil {
		return fmt.Errorf("RabbitMQ未初始化，请先调用InitRabbitMQ")
	}

	rabbitmqManager.mu.RLock()
	defer rabbitmqManager.mu.RUnlock()

	if rabbitmqManager.closed {
		return fmt.Errorf("RabbitMQ连接已关闭")
	}

	// 声明队列
	_, err := rabbitmqManager.channel.QueueDeclare(
		queueName,
		rabbitmqManager.config.Queue.Durable,
		rabbitmqManager.config.Queue.AutoDelete,
		rabbitmqManager.config.Queue.Exclusive,
		rabbitmqManager.config.Queue.NoWait,
		nil,
	)
	if err != nil {
		return fmt.Errorf("声明队列失败: %v", err)
	}

	// 发送消息
	err = rabbitmqManager.channel.PublishWithContext(
		ctx,
		"",        // exchange
		queueName, // routing key
		rabbitmqManager.config.Producer.Mandatory,
		rabbitmqManager.config.Producer.Immediate,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(message),
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("发送消息失败: %v", err)
	}

	log.Printf("消息发送成功! Queue: %s", queueName)
	return nil
}

// PublishJSON 发送JSON消息
func PublishJSON(queueName string, data interface{}) error {
	return PublishJSONWithContext(context.Background(), queueName, data)
}

// PublishJSONWithContext 带上下文发送JSON消息
func PublishJSONWithContext(ctx context.Context, queueName string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %v", err)
	}

	if rabbitmqManager == nil {
		return fmt.Errorf("RabbitMQ未初始化，请先调用InitRabbitMQ")
	}

	rabbitmqManager.mu.RLock()
	defer rabbitmqManager.mu.RUnlock()

	if rabbitmqManager.closed {
		return fmt.Errorf("RabbitMQ连接已关闭")
	}

	// 声明队列
	_, err = rabbitmqManager.channel.QueueDeclare(
		queueName,
		rabbitmqManager.config.Queue.Durable,
		rabbitmqManager.config.Queue.AutoDelete,
		rabbitmqManager.config.Queue.Exclusive,
		rabbitmqManager.config.Queue.NoWait,
		nil,
	)
	if err != nil {
		return fmt.Errorf("声明队列失败: %v", err)
	}

	// 发送消息
	err = rabbitmqManager.channel.PublishWithContext(
		ctx,
		"",        // exchange
		queueName, // routing key
		rabbitmqManager.config.Producer.Mandatory,
		rabbitmqManager.config.Producer.Immediate,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         jsonData,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("发送消息失败: %v", err)
	}

	log.Printf("JSON消息发送成功! Queue: %s", queueName)
	return nil
}

// PublishToExchange 发送消息到交换机
func PublishToExchange(exchangeName string, exchangeType string, routingKey string, message string) error {
	return PublishToExchangeWithContext(context.Background(), exchangeName, exchangeType, routingKey, message)
}

// PublishToExchangeWithContext 带上下文发送消息到交换机
func PublishToExchangeWithContext(ctx context.Context, exchangeName string, exchangeType string, routingKey string, message string) error {
	if rabbitmqManager == nil {
		return fmt.Errorf("RabbitMQ未初始化，请先调用InitRabbitMQ")
	}

	rabbitmqManager.mu.RLock()
	defer rabbitmqManager.mu.RUnlock()

	if rabbitmqManager.closed {
		return fmt.Errorf("RabbitMQ连接已关闭")
	}

	// 声明交换机
	err := rabbitmqManager.channel.ExchangeDeclare(
		exchangeName,
		exchangeType, // direct, fanout, topic, headers
		rabbitmqManager.config.Exchange.Durable,
		rabbitmqManager.config.Exchange.AutoDelete,
		rabbitmqManager.config.Exchange.Internal,
		rabbitmqManager.config.Exchange.NoWait,
		nil,
	)
	if err != nil {
		return fmt.Errorf("声明交换机失败: %v", err)
	}

	// 发送消息
	err = rabbitmqManager.channel.PublishWithContext(
		ctx,
		exchangeName, // exchange
		routingKey,   // routing key
		rabbitmqManager.config.Producer.Mandatory,
		rabbitmqManager.config.Producer.Immediate,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(message),
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("发送消息失败: %v", err)
	}

	log.Printf("消息发送成功! Exchange: %s, RoutingKey: %s", exchangeName, routingKey)
	return nil
}

// CloseRabbitMQ 关闭RabbitMQ连接
func CloseRabbitMQ() error {
	if rabbitmqManager == nil {
		return nil
	}

	rabbitmqManager.mu.Lock()
	defer rabbitmqManager.mu.Unlock()

	rabbitmqManager.closed = true

	if rabbitmqManager.channel != nil {
		if err := rabbitmqManager.channel.Close(); err != nil {
			log.Printf("关闭Channel失败: %v", err)
		}
	}

	if rabbitmqManager.conn != nil {
		if err := rabbitmqManager.conn.Close(); err != nil {
			log.Printf("关闭连接失败: %v", err)
		}
	}

	log.Println("RabbitMQ连接已关闭")
	return nil
}

// GetRabbitMQManager 获取RabbitMQ管理器实例(用于测试)
func GetRabbitMQManager() *RabbitMQManager {
	return rabbitmqManager
}

// ============= 查询相关功能 =============

// QueueInfo 队列信息
type QueueInfo struct {
	Name       string `json:"name"`
	Messages   int    `json:"messages"`    // 队列中的消息数
	Consumers  int    `json:"consumers"`   // 消费者数量
	Durable    bool   `json:"durable"`     // 是否持久化
	AutoDelete bool   `json:"auto_delete"` // 是否自动删除
	Exclusive  bool   `json:"exclusive"`   // 是否排他
}

// QueueMessage 队列消息
type QueueMessage struct {
	Body          string            `json:"body"`           // 消息体
	ContentType   string            `json:"content_type"`   // 内容类型
	DeliveryMode  uint8             `json:"delivery_mode"`  // 投递模式：1=非持久化, 2=持久化
	Priority      uint8             `json:"priority"`       // 优先级
	CorrelationId string            `json:"correlation_id"` // 关联ID
	ReplyTo       string            `json:"reply_to"`       // 回复队列
	Expiration    string            `json:"expiration"`     // 过期时间
	MessageId     string            `json:"message_id"`     // 消息ID
	Timestamp     time.Time         `json:"timestamp"`      // 时间戳
	Type          string            `json:"type"`           // 消息类型
	UserId        string            `json:"user_id"`        // 用户ID
	AppId         string            `json:"app_id"`         // 应用ID
	Headers       map[string]string `json:"headers"`        // 消息头
	DeliveryTag   uint64            `json:"delivery_tag"`   // 投递标签
	Redelivered   bool              `json:"redelivered"`    // 是否重新投递
	Exchange      string            `json:"exchange"`       // 交换机
	RoutingKey    string            `json:"routing_key"`    // 路由键
}

// GetQueueInfo 获取指定队列的详细信息
func GetQueueInfo(queueName string) (*QueueInfo, error) {
	if rabbitmqManager == nil {
		return nil, fmt.Errorf("RabbitMQ未初始化")
	}

	rabbitmqManager.mu.RLock()
	defer rabbitmqManager.mu.RUnlock()

	if rabbitmqManager.closed {
		return nil, fmt.Errorf("RabbitMQ连接已关闭")
	}

	// 使用 QueueInspect 被动声明队列以获取信息
	queue, err := rabbitmqManager.channel.QueueInspect(queueName)
	if err != nil {
		return nil, fmt.Errorf("获取队列信息失败: %v", err)
	}

	return &QueueInfo{
		Name:      queue.Name,
		Messages:  queue.Messages,
		Consumers: queue.Consumers,
	}, nil
}

// DeclareAndGetQueueInfo 声明并获取队列信息（如果队列不存在则创建）
func DeclareAndGetQueueInfo(queueName string) (*QueueInfo, error) {
	if rabbitmqManager == nil {
		return nil, fmt.Errorf("RabbitMQ未初始化")
	}

	rabbitmqManager.mu.RLock()
	defer rabbitmqManager.mu.RUnlock()

	if rabbitmqManager.closed {
		return nil, fmt.Errorf("RabbitMQ连接已关闭")
	}

	// 声明队列
	queue, err := rabbitmqManager.channel.QueueDeclare(
		queueName,                               // 队列名称
		rabbitmqManager.config.Queue.Durable,    // 持久化
		rabbitmqManager.config.Queue.AutoDelete, // 自动删除
		rabbitmqManager.config.Queue.Exclusive,  // 排他
		rabbitmqManager.config.Queue.NoWait,     // 不等待
		nil,                                     // 参数
	)
	if err != nil {
		return nil, fmt.Errorf("声明队列失败: %v", err)
	}

	return &QueueInfo{
		Name:       queue.Name,
		Messages:   queue.Messages,
		Consumers:  queue.Consumers,
		Durable:    rabbitmqManager.config.Queue.Durable,
		AutoDelete: rabbitmqManager.config.Queue.AutoDelete,
		Exclusive:  rabbitmqManager.config.Queue.Exclusive,
	}, nil
}

// PeekMessages 查看队列消息（不消费，查看后重新入队）
// queueName: 队列名称
// count: 要查看的消息数量（最多100条）
func PeekMessages(queueName string, count int) ([]QueueMessage, error) {
	if rabbitmqManager == nil {
		return nil, fmt.Errorf("RabbitMQ未初始化")
	}

	if count <= 0 || count > 100 {
		return nil, fmt.Errorf("count 必须在 1-100 之间")
	}

	rabbitmqManager.mu.RLock()
	defer rabbitmqManager.mu.RUnlock()

	if rabbitmqManager.closed {
		return nil, fmt.Errorf("RabbitMQ连接已关闭")
	}

	var messages []QueueMessage

	for i := 0; i < count; i++ {
		// 使用 Get 方法获取单条消息（autoAck=false，不自动确认）
		delivery, ok, err := rabbitmqManager.channel.Get(queueName, false)
		if err != nil {
			return messages, fmt.Errorf("获取消息失败: %v", err)
		}

		// 没有更多消息
		if !ok {
			break
		}

		// 转换消息头
		headers := make(map[string]string)
		if delivery.Headers != nil {
			for k, v := range delivery.Headers {
				headers[k] = fmt.Sprintf("%v", v)
			}
		}

		msg := QueueMessage{
			Body:          string(delivery.Body),
			ContentType:   delivery.ContentType,
			DeliveryMode:  delivery.DeliveryMode,
			Priority:      delivery.Priority,
			CorrelationId: delivery.CorrelationId,
			ReplyTo:       delivery.ReplyTo,
			Expiration:    delivery.Expiration,
			MessageId:     delivery.MessageId,
			Timestamp:     delivery.Timestamp,
			Type:          delivery.Type,
			UserId:        delivery.UserId,
			AppId:         delivery.AppId,
			Headers:       headers,
			DeliveryTag:   delivery.DeliveryTag,
			Redelivered:   delivery.Redelivered,
			Exchange:      delivery.Exchange,
			RoutingKey:    delivery.RoutingKey,
		}

		messages = append(messages, msg)

		// Nack 并重新入队，这样消息不会丢失（查看模式）
		err = rabbitmqManager.channel.Nack(delivery.DeliveryTag, false, true)
		if err != nil {
			log.Printf("拒绝消息失败: %v", err)
		}
	}

	return messages, nil
}

// ConsumeMessages 消费队列消息（会从队列中删除）
// queueName: 队列名称
// count: 要消费的消息数量（最多100条）
func ConsumeMessages(queueName string, count int) ([]QueueMessage, error) {
	if rabbitmqManager == nil {
		return nil, fmt.Errorf("RabbitMQ未初始化")
	}

	if count <= 0 || count > 100 {
		return nil, fmt.Errorf("count 必须在 1-100 之间")
	}

	rabbitmqManager.mu.RLock()
	defer rabbitmqManager.mu.RUnlock()

	if rabbitmqManager.closed {
		return nil, fmt.Errorf("RabbitMQ连接已关闭")
	}

	var messages []QueueMessage

	for i := 0; i < count; i++ {
		// 使用 Get 方法获取单条消息（autoAck=true，自动确认并删除）
		delivery, ok, err := rabbitmqManager.channel.Get(queueName, true)
		if err != nil {
			return messages, fmt.Errorf("获取消息失败: %v", err)
		}

		// 没有更多消息
		if !ok {
			break
		}

		// 转换消息头
		headers := make(map[string]string)
		if delivery.Headers != nil {
			for k, v := range delivery.Headers {
				headers[k] = fmt.Sprintf("%v", v)
			}
		}

		msg := QueueMessage{
			Body:          string(delivery.Body),
			ContentType:   delivery.ContentType,
			DeliveryMode:  delivery.DeliveryMode,
			Priority:      delivery.Priority,
			CorrelationId: delivery.CorrelationId,
			ReplyTo:       delivery.ReplyTo,
			Expiration:    delivery.Expiration,
			MessageId:     delivery.MessageId,
			Timestamp:     delivery.Timestamp,
			Type:          delivery.Type,
			UserId:        delivery.UserId,
			AppId:         delivery.AppId,
			Headers:       headers,
			DeliveryTag:   delivery.DeliveryTag,
			Redelivered:   delivery.Redelivered,
			Exchange:      delivery.Exchange,
			RoutingKey:    delivery.RoutingKey,
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// PurgeQueue 清空队列中的所有消息
func PurgeQueue(queueName string) (int, error) {
	if rabbitmqManager == nil {
		return 0, fmt.Errorf("RabbitMQ未初始化")
	}

	rabbitmqManager.mu.RLock()
	defer rabbitmqManager.mu.RUnlock()

	if rabbitmqManager.closed {
		return 0, fmt.Errorf("RabbitMQ连接已关闭")
	}

	count, err := rabbitmqManager.channel.QueuePurge(queueName, false)
	if err != nil {
		return 0, fmt.Errorf("清空队列失败: %v", err)
	}

	return count, nil
}

// DeleteQueue 删除队列
func DeleteQueue(queueName string, ifUnused, ifEmpty bool) (int, error) {
	if rabbitmqManager == nil {
		return 0, fmt.Errorf("RabbitMQ未初始化")
	}

	rabbitmqManager.mu.RLock()
	defer rabbitmqManager.mu.RUnlock()

	if rabbitmqManager.closed {
		return 0, fmt.Errorf("RabbitMQ连接已关闭")
	}

	count, err := rabbitmqManager.channel.QueueDelete(queueName, ifUnused, ifEmpty, false)
	if err != nil {
		return 0, fmt.Errorf("删除队列失败: %v", err)
	}

	return count, nil
}
