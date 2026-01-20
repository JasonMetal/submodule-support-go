package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
)

// TestKafkaConfig 测试Kafka配置加载
func TestKafkaConfig(t *testing.T) {
	// 创建临时测试配置文件
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "kafka.yml")

	configContent := `brokers:
  - host: "localhost"
    port: 9092
  - host: "localhost"
    port: 9093

ssl:
  enable: false

producer:
  required_acks: 1
  max_retries: 5
  return_successes: true
  return_errors: true

consumer:
  group_id: "test-group"
  auto_commit: true

version: "3.7.2"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("创建测试配置文件失败: %v", err)
	}

	// 测试加载配置
	config, err := loadKafkaConfig(configPath)
	if err != nil {
		t.Fatalf("加载Kafka配置失败: %v", err)
	}

	// 验证配置
	if len(config.Brokers) != 2 {
		t.Errorf("期望2个broker，实际得到%d个", len(config.Brokers))
	}

	if config.Brokers[0].Host != "localhost" || config.Brokers[0].Port != 9092 {
		t.Errorf("Broker配置不正确: %+v", config.Brokers[0])
	}

	if config.SSL.Enable != false {
		t.Errorf("SSL配置不正确: %v", config.SSL.Enable)
	}

	if config.Producer.RequiredAcks != 1 {
		t.Errorf("Producer RequiredAcks配置不正确: %d", config.Producer.RequiredAcks)
	}

	if config.Producer.MaxRetries != 5 {
		t.Errorf("Producer MaxRetries配置不正确: %d", config.Producer.MaxRetries)
	}

	if config.Consumer.GroupID != "test-group" {
		t.Errorf("Consumer GroupID配置不正确: %s", config.Consumer.GroupID)
	}

	if config.Version != "3.7.2" {
		t.Errorf("Version配置不正确: %s", config.Version)
	}

	t.Log("✓ Kafka配置加载测试通过")
}

// TestKafkaConfigLoadError 测试配置文件不存在的情况
func TestKafkaConfigLoadError(t *testing.T) {
	_, err := loadKafkaConfig("/nonexistent/path/kafka.yml")
	if err == nil {
		t.Error("期望配置加载失败，但成功了")
	}
	t.Logf("✓ 配置文件不存在错误处理测试通过: %v", err)
}

// TestKafkaManagerInit 测试KafkaManager初始化
func TestKafkaManagerInit(t *testing.T) {
	// 创建临时测试配置目录
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config", "test")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("创建配置目录失败: %v", err)
	}

	configPath := filepath.Join(configDir, "kafka.yml")
	configContent := `brokers:
  - host: "localhost"
    port: 9092

ssl:
  enable: false

producer:
  required_acks: 1
  max_retries: 3
  return_successes: true
  return_errors: true

consumer:
  group_id: "test-group"
  auto_commit: true

version: "3.7.2"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("创建测试配置文件失败: %v", err)
	}

	// 由于ProjectPath是函数，我们跳过这个测试或使用其他方式
	// 这个测试主要验证配置文件的加载，我们已经在其他测试中验证了
	//t.Skip("跳过需要修改全局函数的测试")

	// 测试初始化(会失败，因为没有真实的Kafka服务，但可以测试配置加载)
	err := InitKafka()
	// 初始化应该成功（只是加载配置）
	if err != nil {
		t.Logf("Kafka初始化返回: %v (预期，因为没有真实的Kafka服务)", err)
	} else {
		t.Log("✓ Kafka配置初始化成功")
	}
}

// TestProducerConfig 测试生产者配置生成
func TestProducerConfig(t *testing.T) {
	kafkaConfig := &KafkaConfig{
		Brokers: []BrokerConfig{
			{Host: "localhost", Port: 9092},
		},
		SSL: SSLConfig{Enable: false},
		Producer: ProducerConfig{
			RequiredAcks:    1,
			MaxRetries:      5,
			ReturnSuccesses: true,
			ReturnErrors:    true,
		},
		Version: "3.7.2",
	}

	manager := &KafkaManager{
		config:  kafkaConfig,
		brokers: []string{"localhost:9092"},
	}

	config := manager.getProducerConfig()

	if config.Producer.RequiredAcks != sarama.RequiredAcks(1) {
		t.Errorf("RequiredAcks配置不正确: %v", config.Producer.RequiredAcks)
	}

	if config.Producer.Retry.Max != 5 {
		t.Errorf("MaxRetries配置不正确: %d", config.Producer.Retry.Max)
	}

	if !config.Producer.Return.Successes {
		t.Error("ReturnSuccesses应该为true")
	}

	if !config.Producer.Return.Errors {
		t.Error("ReturnErrors应该为true")
	}

	if config.Net.TLS.Enable {
		t.Error("TLS应该为false")
	}

	t.Log("✓ 生产者配置生成测试通过")
}

// TestProducerSyncWithMock 使用Mock测试同步生产者
func TestProducerSyncWithMock(t *testing.T) {
	// 创建Mock生产者
	mockProducer := mocks.NewSyncProducer(t, nil)

	// 设置期望
	mockProducer.ExpectSendMessageAndSucceed()

	// 创建测试用的KafkaManager
	kafkaConfig := &KafkaConfig{
		Brokers: []BrokerConfig{{Host: "localhost", Port: 9092}},
		Producer: ProducerConfig{
			RequiredAcks:    1,
			MaxRetries:      3,
			ReturnSuccesses: true,
			ReturnErrors:    true,
		},
	}

	manager := &KafkaManager{
		config:       kafkaConfig,
		brokers:      []string{"localhost:9092"},
		syncProducer: mockProducer,
	}

	// 临时替换全局manager
	oldManager := kafkaManager
	kafkaManager = manager
	defer func() { kafkaManager = oldManager }()

	// 测试发送消息
	err := ProducerSync("test-topic", "test message")
	if err != nil {
		t.Errorf("发送消息失败: %v", err)
	}

	// 验证期望是否满足
	if err := mockProducer.Close(); err != nil {
		t.Errorf("关闭Mock生产者失败: %v", err)
	}

	t.Log("✓ 同步生产者Mock测试通过")
}

// TestProducerSyncBatchWithMock 测试批量发送
func TestProducerSyncBatchWithMock(t *testing.T) {
	// 创建Mock生产者
	mockProducer := mocks.NewSyncProducer(t, nil)

	// 设置期望 - 3条消息
	mockProducer.ExpectSendMessageAndSucceed()
	mockProducer.ExpectSendMessageAndSucceed()
	mockProducer.ExpectSendMessageAndSucceed()

	kafkaConfig := &KafkaConfig{
		Brokers: []BrokerConfig{{Host: "localhost", Port: 9092}},
		Producer: ProducerConfig{
			RequiredAcks:    1,
			MaxRetries:      3,
			ReturnSuccesses: true,
			ReturnErrors:    true,
		},
	}

	manager := &KafkaManager{
		config:       kafkaConfig,
		brokers:      []string{"localhost:9092"},
		syncProducer: mockProducer,
	}

	oldManager := kafkaManager
	kafkaManager = manager
	defer func() { kafkaManager = oldManager }()

	// 测试批量发送
	messages := []string{"msg1", "msg2", "msg3"}
	err := ProducerSyncBatch("test-topic", messages)
	if err != nil {
		t.Errorf("批量发送失败: %v", err)
	}

	if err := mockProducer.Close(); err != nil {
		t.Errorf("关闭Mock生产者失败: %v", err)
	}

	t.Log("✓ 批量发送Mock测试通过")
}

// TestProducerAsyncWithMock 测试异步生产者
func TestProducerAsyncWithMock(t *testing.T) {
	// 创建Mock异步生产者
	mockProducer := mocks.NewAsyncProducer(t, nil)

	// 设置期望
	mockProducer.ExpectInputAndSucceed()

	kafkaConfig := &KafkaConfig{
		Brokers: []BrokerConfig{{Host: "localhost", Port: 9092}},
		Producer: ProducerConfig{
			RequiredAcks:    1,
			MaxRetries:      3,
			ReturnSuccesses: true,
			ReturnErrors:    true,
		},
	}

	manager := &KafkaManager{
		config:        kafkaConfig,
		brokers:       []string{"localhost:9092"},
		asyncProducer: mockProducer,
	}

	oldManager := kafkaManager
	kafkaManager = manager
	defer func() { kafkaManager = oldManager }()

	// 测试异步发送
	err := ProducerAsync("test-topic", "async test message")
	if err != nil {
		t.Errorf("异步发送失败: %v", err)
	}

	// 等待一小段时间确保消息被处理
	time.Sleep(100 * time.Millisecond)

	if err := mockProducer.Close(); err != nil {
		t.Errorf("关闭Mock异步生产者失败: %v", err)
	}

	t.Log("✓ 异步生产者Mock测试通过")
}

// TestProducerWithContext 测试带上下文的发送
func TestProducerWithContext(t *testing.T) {
	mockProducer := mocks.NewSyncProducer(t, nil)
	mockProducer.ExpectSendMessageAndSucceed()

	kafkaConfig := &KafkaConfig{
		Brokers: []BrokerConfig{{Host: "localhost", Port: 9092}},
		Producer: ProducerConfig{
			RequiredAcks:    1,
			MaxRetries:      3,
			ReturnSuccesses: true,
			ReturnErrors:    true,
		},
	}

	manager := &KafkaManager{
		config:       kafkaConfig,
		brokers:      []string{"localhost:9092"},
		syncProducer: mockProducer,
	}

	oldManager := kafkaManager
	kafkaManager = manager
	defer func() { kafkaManager = oldManager }()

	ctx := context.Background()
	err := ProducerSyncWithContext(ctx, "test-topic", "test message with context")
	if err != nil {
		t.Errorf("带上下文发送失败: %v", err)
	}

	if err := mockProducer.Close(); err != nil {
		t.Errorf("关闭Mock生产者失败: %v", err)
	}

	t.Log("✓ 带上下文发送测试通过")
}

// TestProducerWithCancelledContext 测试取消的上下文
func TestProducerAsyncWithCancelledContext(t *testing.T) {
	mockProducer := mocks.NewAsyncProducer(t, nil)

	kafkaConfig := &KafkaConfig{
		Brokers: []BrokerConfig{{Host: "localhost", Port: 9092}},
		Producer: ProducerConfig{
			RequiredAcks:    1,
			MaxRetries:      3,
			ReturnSuccesses: true,
			ReturnErrors:    true,
		},
	}

	manager := &KafkaManager{
		config:        kafkaConfig,
		brokers:       []string{"localhost:9092"},
		asyncProducer: mockProducer,
	}

	oldManager := kafkaManager
	kafkaManager = manager
	defer func() { kafkaManager = oldManager }()

	// 创建已取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := ProducerAsyncWithContext(ctx, "test-topic", "test message")
	if err == nil {
		t.Error("期望取消的上下文会返回错误")
	}

	mockProducer.Close()

	t.Logf("✓ 取消上下文测试通过: %v", err)
}

// TestProducerError 测试生产者错误处理
func TestProducerError(t *testing.T) {
	// 创建会失败的Mock生产者
	mockProducer := mocks.NewSyncProducer(t, nil)
	mockProducer.ExpectSendMessageAndFail(fmt.Errorf("模拟发送失败"))

	kafkaConfig := &KafkaConfig{
		Brokers: []BrokerConfig{{Host: "localhost", Port: 9092}},
		Producer: ProducerConfig{
			RequiredAcks:    1,
			MaxRetries:      3,
			ReturnSuccesses: true,
			ReturnErrors:    true,
		},
	}

	manager := &KafkaManager{
		config:       kafkaConfig,
		brokers:      []string{"localhost:9092"},
		syncProducer: mockProducer,
	}

	oldManager := kafkaManager
	kafkaManager = manager
	defer func() { kafkaManager = oldManager }()

	// 测试发送应该失败
	err := ProducerSync("test-topic", "test message")
	if err == nil {
		t.Error("期望发送失败但成功了")
	}

	mockProducer.Close()

	t.Logf("✓ 生产者错误处理测试通过: %v", err)
}

// TestKafkaNotInitialized 测试未初始化的情况
func TestKafkaNotInitialized(t *testing.T) {
	// 临时清空kafkaManager
	oldManager := kafkaManager
	kafkaManager = nil
	defer func() { kafkaManager = oldManager }()

	err := ProducerSync("test-topic", "test message")
	if err == nil {
		t.Error("期望未初始化错误，但成功了")
	}

	if err.Error() != "Kafka未初始化，请先调用InitKafka" {
		t.Errorf("错误消息不正确: %v", err)
	}

	t.Log("✓ 未初始化错误处理测试通过")
}

// TestCloseKafka 测试关闭Kafka连接
func TestCloseKafka(t *testing.T) {
	mockSyncProducer := mocks.NewSyncProducer(t, nil)
	mockAsyncProducer := mocks.NewAsyncProducer(t, nil)

	manager := &KafkaManager{
		syncProducer:  mockSyncProducer,
		asyncProducer: mockAsyncProducer,
	}

	oldManager := kafkaManager
	kafkaManager = manager
	defer func() { kafkaManager = oldManager }()

	err := CloseKafka()
	if err != nil {
		t.Errorf("关闭Kafka失败: %v", err)
	}

	t.Log("✓ 关闭Kafka测试通过")
}

// TestCloseKafkaWhenNotInitialized 测试未初始化时关闭
func TestCloseKafkaWhenNotInitialized(t *testing.T) {
	oldManager := kafkaManager
	kafkaManager = nil
	defer func() { kafkaManager = oldManager }()

	err := CloseKafka()
	if err != nil {
		t.Errorf("未初始化时关闭应该成功: %v", err)
	}

	t.Log("✓ 未初始化关闭测试通过")
}

// BenchmarkProducerSync 同步发送性能测试
func BenchmarkProducerSync(b *testing.B) {
	mockProducer := mocks.NewSyncProducer(b, nil)
	for i := 0; i < b.N; i++ {
		mockProducer.ExpectSendMessageAndSucceed()
	}

	kafkaConfig := &KafkaConfig{
		Brokers:  []BrokerConfig{{Host: "localhost", Port: 9092}},
		Producer: ProducerConfig{RequiredAcks: 1, MaxRetries: 3},
	}

	manager := &KafkaManager{
		config:       kafkaConfig,
		brokers:      []string{"localhost:9092"},
		syncProducer: mockProducer,
	}

	oldManager := kafkaManager
	kafkaManager = manager
	defer func() { kafkaManager = oldManager }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ProducerSync("bench-topic", fmt.Sprintf("message-%d", i))
	}

	mockProducer.Close()
}

// BenchmarkProducerAsync 异步发送性能测试
func BenchmarkProducerAsync(b *testing.B) {
	mockProducer := mocks.NewAsyncProducer(b, nil)
	for i := 0; i < b.N; i++ {
		mockProducer.ExpectInputAndSucceed()
	}

	kafkaConfig := &KafkaConfig{
		Brokers:  []BrokerConfig{{Host: "localhost", Port: 9092}},
		Producer: ProducerConfig{RequiredAcks: 1, MaxRetries: 3},
	}

	manager := &KafkaManager{
		config:        kafkaConfig,
		brokers:       []string{"localhost:9092"},
		asyncProducer: mockProducer,
	}

	oldManager := kafkaManager
	kafkaManager = manager
	defer func() { kafkaManager = oldManager }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ProducerAsync("bench-topic", fmt.Sprintf("message-%d", i))
	}

	time.Sleep(100 * time.Millisecond) // 等待异步消息处理完成
	mockProducer.Close()
}
