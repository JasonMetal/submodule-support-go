package bootstrap

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/JasonMetal/submodule-support-go.git/helper/config"
)

// KafkaConfig Kafka配置结构
type KafkaConfig struct {
	Brokers  []BrokerConfig `yaml:"brokers"`
	SSL      SSLConfig      `yaml:"ssl"`
	Producer ProducerConfig `yaml:"producer"`
	Consumer ConsumerConfig `yaml:"consumer"`
	Version  string         `yaml:"version"`
}

type BrokerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type SSLConfig struct {
	Enable bool `yaml:"enable"`
}

type ProducerConfig struct {
	RequiredAcks    int  `yaml:"required_acks"`
	MaxRetries      int  `yaml:"max_retries"`
	ReturnSuccesses bool `yaml:"return_successes"`
	ReturnErrors    bool `yaml:"return_errors"`
}

type ConsumerConfig struct {
	GroupID    string `yaml:"group_id"`
	AutoCommit bool   `yaml:"auto_commit"`
}

// KafkaManager Kafka管理器
type KafkaManager struct {
	config        *KafkaConfig
	brokers       []string
	syncProducer  sarama.SyncProducer
	asyncProducer sarama.AsyncProducer
}

var kafkaManager *KafkaManager

// InitKafka 初始化Kafka连接
func InitKafka() error {
	configPath := fmt.Sprintf("%sconfig/%s/kafka.yml", ProjectPath(), DevEnv)
	log.Printf("InitKafka DevEnv %s", DevEnv)
	kafkaConfig, err := loadKafkaConfig(configPath)
	if err != nil {
		return fmt.Errorf("加载Kafka配置失败: %v", err)
	}

	// 构建broker地址列表
	brokers := make([]string, 0, len(kafkaConfig.Brokers))
	for _, broker := range kafkaConfig.Brokers {
		brokers = append(brokers, fmt.Sprintf("%s:%d", broker.Host, broker.Port))
	}

	kafkaManager = &KafkaManager{
		config:  kafkaConfig,
		brokers: brokers,
	}

	log.Printf("Kafka初始化成功, Brokers: %v", brokers)
	return nil
}

// loadKafkaConfig 加载Kafka配置
func loadKafkaConfig(configPath string) (*KafkaConfig, error) {
	cfg, err := config.GetConfig(configPath)
	if err != nil {
		return nil, err
	}

	kafkaConfig := &KafkaConfig{}

	// 加载brokers
	brokersList, err := cfg.List("brokers")
	if err != nil {
		return nil, fmt.Errorf("读取brokers配置失败: %v", err)
	}

	for _, item := range brokersList {
		var broker BrokerConfig
		var host string
		var port int

		// 尝试多种类型转换
		if brokerMap, ok := item.(map[interface{}]interface{}); ok {
			host = brokerMap["host"].(string)
			// port可能是int或float64
			switch v := brokerMap["port"].(type) {
			case int:
				port = v
			case float64:
				port = int(v)
			default:
				return nil, fmt.Errorf("无法解析broker port: %v", v)
			}
		} else if brokerMap, ok := item.(map[string]interface{}); ok {
			host = brokerMap["host"].(string)
			// port可能是int或float64
			switch v := brokerMap["port"].(type) {
			case int:
				port = v
			case float64:
				port = int(v)
			default:
				return nil, fmt.Errorf("无法解析broker port: %v", v)
			}
		} else {
			return nil, fmt.Errorf("无法解析broker配置: %v", item)
		}

		broker = BrokerConfig{Host: host, Port: port}
		kafkaConfig.Brokers = append(kafkaConfig.Brokers, broker)
	}

	// 加载SSL配置
	sslEnable, _ := cfg.Bool("ssl.enable")
	kafkaConfig.SSL.Enable = sslEnable

	// 加载Producer配置
	requiredAcks, _ := cfg.Int("producer.required_acks")
	maxRetries, _ := cfg.Int("producer.max_retries")
	returnSuccesses, _ := cfg.Bool("producer.return_successes")
	returnErrors, _ := cfg.Bool("producer.return_errors")

	kafkaConfig.Producer = ProducerConfig{
		RequiredAcks:    requiredAcks,
		MaxRetries:      maxRetries,
		ReturnSuccesses: returnSuccesses,
		ReturnErrors:    returnErrors,
	}

	// 加载Consumer配置
	groupID, _ := cfg.String("consumer.group_id")
	autoCommit, _ := cfg.Bool("consumer.auto_commit")

	kafkaConfig.Consumer = ConsumerConfig{
		GroupID:    groupID,
		AutoCommit: autoCommit,
	}

	// 加载版本
	version, _ := cfg.String("version")
	kafkaConfig.Version = version

	return kafkaConfig, nil
}

// getProducerConfig 获取生产者配置
func (km *KafkaManager) getProducerConfig() *sarama.Config {
	config := sarama.NewConfig()

	// 设置版本
	if km.config.Version != "" {
		if version, err := sarama.ParseKafkaVersion(km.config.Version); err == nil {
			config.Version = version
		}
	}

	// Producer配置
	config.Producer.RequiredAcks = sarama.RequiredAcks(km.config.Producer.RequiredAcks)
	config.Producer.Retry.Max = km.config.Producer.MaxRetries
	config.Producer.Return.Successes = km.config.Producer.ReturnSuccesses
	config.Producer.Return.Errors = km.config.Producer.ReturnErrors
	config.Producer.Partitioner = sarama.NewRandomPartitioner

	// SSL配置
	config.Net.TLS.Enable = km.config.SSL.Enable

	return config
}

// getSyncProducer 获取同步生产者(单例模式)
func (km *KafkaManager) getSyncProducer() (sarama.SyncProducer, error) {
	if km.syncProducer != nil {
		return km.syncProducer, nil
	}

	config := km.getProducerConfig()
	producer, err := sarama.NewSyncProducer(km.brokers, config)
	if err != nil {
		return nil, fmt.Errorf("创建同步生产者失败: %v", err)
	}

	km.syncProducer = producer
	return producer, nil
}

type AsyncErrorHandler func(topic string, message string, err error)

var asyncErrorHandler AsyncErrorHandler

// SetAsyncErrorHandler 设置异步错误处理器
func SetAsyncErrorHandler(handler AsyncErrorHandler) {
	asyncErrorHandler = handler
}

// getAsyncProducer 获取异步生产者(单例模式)
func (km *KafkaManager) getAsyncProducer() (sarama.AsyncProducer, error) {
	if km.asyncProducer != nil {
		return km.asyncProducer, nil
	}

	config := km.getProducerConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal // 异步模式使用更高的吞吐量

	producer, err := sarama.NewAsyncProducer(km.brokers, config)
	if err != nil {
		return nil, fmt.Errorf("创建异步生产者失败: %v", err)
	}

	km.asyncProducer = producer

	// 启动监听goroutine
	go func() {
		for {
			select {
			case success := <-producer.Successes():
				if success != nil {
					log.Printf("Kafka消息发送成功: Topic=%s, Partition=%d, Offset=%d",
						success.Topic, success.Partition, success.Offset)
				}
			case errMsg := <-producer.Errors():
				if errMsg != nil {
					log.Printf("Kafka消息发送失败: %v", errMsg.Err)

					// ✅ 调用错误处理器
					if asyncErrorHandler != nil {
						asyncErrorHandler(
							errMsg.Msg.Topic,
							string(errMsg.Msg.Value.(sarama.StringEncoder)),
							errMsg.Err,
						)
					}
				}
			}
		}
	}()

	return producer, nil
}

// ProducerSync 同步发送消息
func ProducerSync(topic string, message string) error {
	return ProducerSyncWithContext(context.Background(), topic, message)
}

// ProducerSyncWithContext 带上下文的同步发送消息
func ProducerSyncWithContext(ctx context.Context, topic string, message string) error {
	if kafkaManager == nil {
		return fmt.Errorf("Kafka未初始化，请先调用InitKafka")
	}

	producer, err := kafkaManager.getSyncProducer()
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("发送消息失败: %v", err)
	}

	log.Printf("消息发送成功! Topic: %s, Partition: %d, Offset: %d", topic, partition, offset)
	return nil
}

// ProducerAsync 异步发送消息
func ProducerAsync(topic string, message string) error {
	return ProducerAsyncWithContext(context.Background(), topic, message)
}

// ProducerAsyncWithContext 带上下文的异步发送消息
func ProducerAsyncWithContext(ctx context.Context, topic string, message string) error {
	if kafkaManager == nil {
		return fmt.Errorf("Kafka未初始化，请先调用InitKafka")
	}

	producer, err := kafkaManager.getAsyncProducer()
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	select {
	case producer.Input() <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return fmt.Errorf("发送消息超时")
	}
}

// ProducerSyncBatch 批量同步发送消息
func ProducerSyncBatch(topic string, messages []string) error {
	if kafkaManager == nil {
		return fmt.Errorf("Kafka未初始化，请先调用InitKafka")
	}

	producer, err := kafkaManager.getSyncProducer()
	if err != nil {
		return err
	}

	var sendErrors []error
	for _, message := range messages {
		msg := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.StringEncoder(message),
		}

		_, _, err := producer.SendMessage(msg)
		if err != nil {
			sendErrors = append(sendErrors, err)
		}
	}

	if len(sendErrors) > 0 {
		return fmt.Errorf("批量发送失败, 错误数: %d, 首个错误: %v", len(sendErrors), sendErrors[0])
	}

	log.Printf("批量发送成功! Topic: %s, 消息数: %d", topic, len(messages))
	return nil
}

// Close 关闭Kafka连接
func CloseKafka() error {
	if kafkaManager == nil {
		return nil
	}

	var errs []error

	if kafkaManager.syncProducer != nil {
		if err := kafkaManager.syncProducer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("关闭同步生产者失败: %v", err))
		}
	}

	if kafkaManager.asyncProducer != nil {
		if err := kafkaManager.asyncProducer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("关闭异步生产者失败: %v", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("关闭Kafka连接时出现错误: %v", errs)
	}

	log.Println("Kafka连接已关闭")
	return nil
}

// GetKafkaManager 获取Kafka管理器实例(用于测试)
func GetKafkaManager() *KafkaManager {
	return kafkaManager
}

// SetKafkaManager 设置Kafka管理器实例(用于测试)
func SetKafkaManager(manager *KafkaManager) {
	kafkaManager = manager
}

// KafkaMessage Kafka消息结构
type KafkaMessage struct {
	Topic     string `json:"topic"`
	Partition int32  `json:"partition"`
	Offset    int64  `json:"offset"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
}

// getConsumerConfig 获取消费者配置
func (km *KafkaManager) getConsumerConfig() *sarama.Config {
	config := sarama.NewConfig()

	// 设置版本
	if km.config.Version != "" {
		if version, err := sarama.ParseKafkaVersion(km.config.Version); err == nil {
			config.Version = version
		}
	}

	// Consumer配置
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest // 从最早的消息开始读取

	// SSL配置
	config.Net.TLS.Enable = km.config.SSL.Enable

	return config
}

// FetchMessages 获取主题消息(支持分页)
func FetchMessages(topic string, partition int32, offset int64, limit int) ([]*KafkaMessage, error) {
	if kafkaManager == nil {
		return nil, fmt.Errorf("Kafka未初始化，请先调用InitKafka")
	}

	config := kafkaManager.getConsumerConfig()

	// 创建 Client 用于获取 offset 信息
	client, err := sarama.NewClient(kafkaManager.brokers, config)
	if err != nil {
		return nil, fmt.Errorf("创建客户端失败: %v", err)
	}
	defer client.Close()

	// 创建Consumer
	consumer, err := sarama.NewConsumer(kafkaManager.brokers, config)
	if err != nil {
		return nil, fmt.Errorf("创建消费者失败: %v", err)
	}
	defer consumer.Close()

	// 获取主题的所有分区
	partitions, err := consumer.Partitions(topic)
	if err != nil {
		return nil, fmt.Errorf("获取主题分区失败: %v", err)
	}

	// 如果没有指定分区，默认使用第一个分区
	if partition < 0 && len(partitions) > 0 {
		partition = partitions[0]
	}

	// 先检查分区是否有消息可读
	newestOffset, err := client.GetOffset(topic, partition, sarama.OffsetNewest)
	if err != nil {
		return nil, fmt.Errorf("获取最新偏移量失败: %v", err)
	}

	// 如果请求的 offset 已经超过最新 offset，直接返回空结果
	if offset >= newestOffset {
		log.Printf("请求的 offset %d 已经到达或超过最新 offset %d，返回空结果", offset, newestOffset)
		return []*KafkaMessage{}, nil
	}

	// 计算实际可读取的消息数量
	availableMessages := newestOffset - offset
	if int64(limit) > availableMessages {
		limit = int(availableMessages)
	}

	// 如果没有消息可读，直接返回
	if limit <= 0 {
		return []*KafkaMessage{}, nil
	}

	// 创建分区消费者
	partitionConsumer, err := consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		return nil, fmt.Errorf("创建分区消费者失败: %v", err)
	}
	defer partitionConsumer.Close()

	// 读取消息
	messages := make([]*KafkaMessage, 0, limit)
	totalTimeout := time.After(3 * time.Second)          // 总超时时间 3 秒
	idleTimeout := time.NewTimer(300 * time.Millisecond) // 空闲超时 300ms
	defer idleTimeout.Stop()

	for len(messages) < limit {
		select {
		case msg := <-partitionConsumer.Messages():
			if msg != nil {
				kafkaMsg := &KafkaMessage{
					Topic:     msg.Topic,
					Partition: msg.Partition,
					Offset:    msg.Offset,
					Key:       string(msg.Key),
					Value:     string(msg.Value),
					Timestamp: msg.Timestamp.Unix(),
				}
				messages = append(messages, kafkaMsg)

				// 重置空闲超时
				if !idleTimeout.Stop() {
					select {
					case <-idleTimeout.C:
					default:
					}
				}
				idleTimeout.Reset(300 * time.Millisecond)
			}
		case err := <-partitionConsumer.Errors():
			if err != nil {
				log.Printf("消费消息错误: %v", err)
			}
		case <-idleTimeout.C:
			// 空闲超时：300ms 内没有新消息，立即返回已读取的消息
			if len(messages) > 0 {
				log.Printf("空闲超时，已读取 %d 条消息", len(messages))
				return messages, nil
			}
			// 如果一条消息都没读到，继续等待
			idleTimeout.Reset(300 * time.Millisecond)
		case <-totalTimeout:
			// 总超时：3 秒后强制返回
			log.Printf("总超时，已读取 %d 条消息", len(messages))
			return messages, nil
		}
	}

	return messages, nil
}

// FetchMessagesFromAllPartitions 从所有分区获取消息
func FetchMessagesFromAllPartitions(topic string, offset int64, limit int) ([]*KafkaMessage, error) {
	if kafkaManager == nil {
		return nil, fmt.Errorf("Kafka未初始化，请先调用InitKafka")
	}

	config := kafkaManager.getConsumerConfig()

	// 创建Consumer
	consumer, err := sarama.NewConsumer(kafkaManager.brokers, config)
	if err != nil {
		return nil, fmt.Errorf("创建消费者失败: %v", err)
	}
	defer consumer.Close()

	// 获取主题的所有分区
	partitions, err := consumer.Partitions(topic)
	if err != nil {
		return nil, fmt.Errorf("获取主题分区失败: %v", err)
	}

	if len(partitions) == 0 {
		return []*KafkaMessage{}, nil
	}

	// 从所有分区读取消息
	messages := make([]*KafkaMessage, 0, limit)
	perPartitionLimit := limit / len(partitions)
	if perPartitionLimit == 0 {
		perPartitionLimit = 1
	}

	for _, partition := range partitions {
		partitionConsumer, err := consumer.ConsumePartition(topic, partition, offset)
		if err != nil {
			log.Printf("创建分区 %d 消费者失败: %v", partition, err)
			continue
		}

		// 读取该分区的消息
		partitionMessages := 0
		timeout := time.After(2 * time.Second) // 每个分区2秒超时

	PartitionLoop:
		for partitionMessages < perPartitionLimit && len(messages) < limit {
			select {
			case msg := <-partitionConsumer.Messages():
				if msg != nil {
					kafkaMsg := &KafkaMessage{
						Topic:     msg.Topic,
						Partition: msg.Partition,
						Offset:    msg.Offset,
						Key:       string(msg.Key),
						Value:     string(msg.Value),
						Timestamp: msg.Timestamp.Unix(),
					}
					messages = append(messages, kafkaMsg)
					partitionMessages++
				}
			case err := <-partitionConsumer.Errors():
				if err != nil {
					log.Printf("分区 %d 消费消息错误: %v", partition, err)
				}
			case <-timeout:
				break PartitionLoop
			}
		}

		partitionConsumer.Close()
	}

	return messages, nil
}

// GetTopicPartitions 获取主题的分区信息
func GetTopicPartitions(topic string) ([]int32, error) {
	if kafkaManager == nil {
		return nil, fmt.Errorf("Kafka未初始化，请先调用InitKafka")
	}

	config := kafkaManager.getConsumerConfig()

	consumer, err := sarama.NewConsumer(kafkaManager.brokers, config)
	if err != nil {
		return nil, fmt.Errorf("创建消费者失败: %v", err)
	}
	defer consumer.Close()

	partitions, err := consumer.Partitions(topic)
	if err != nil {
		return nil, fmt.Errorf("获取主题分区失败: %v", err)
	}

	return partitions, nil
}

// GetPartitionOffset 获取分区的最新和最早偏移量
func GetPartitionOffset(topic string, partition int32) (oldest int64, newest int64, err error) {
	if kafkaManager == nil {
		return 0, 0, fmt.Errorf("Kafka未初始化，请先调用InitKafka")
	}

	config := kafkaManager.getConsumerConfig()

	// 使用 Client 而不是 Consumer 来获取 offset
	client, err := sarama.NewClient(kafkaManager.brokers, config)
	if err != nil {
		return 0, 0, fmt.Errorf("创建客户端失败: %v", err)
	}
	defer client.Close()

	// 获取最早的 offset
	oldest, err = client.GetOffset(topic, partition, sarama.OffsetOldest)
	if err != nil {
		return 0, 0, fmt.Errorf("获取最早偏移量失败: %v", err)
	}

	// 获取最新的 offset
	newest, err = client.GetOffset(topic, partition, sarama.OffsetNewest)
	if err != nil {
		return 0, 0, fmt.Errorf("获取最新偏移量失败: %v", err)
	}

	return oldest, newest, nil
}
