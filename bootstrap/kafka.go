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
func InitKafka(env string) error {
	configPath := fmt.Sprintf("%sconfig/%s/kafka.yml", ProjectPath(), env)
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
			case err := <-producer.Errors():
				if err != nil {
					log.Printf("Kafka消息发送失败: %v", err)
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
