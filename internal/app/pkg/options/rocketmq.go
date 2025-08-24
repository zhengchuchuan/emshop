package options

import (
	"fmt"
	"github.com/spf13/pflag"
)

// RocketMQOptions RocketMQ配置选项
type RocketMQOptions struct {
	NameServers   []string `json:"nameservers" mapstructure:"nameservers"`
	ConsumerGroup string   `json:"consumer_group" mapstructure:"consumer_group"`
	Topic         string   `json:"topic" mapstructure:"topic"`
	MaxReconsume  int32    `json:"max_reconsume" mapstructure:"max_reconsume"`
}

// NewRocketMQOptions 创建默认RocketMQ配置
func NewRocketMQOptions() *RocketMQOptions {
	return &RocketMQOptions{
		NameServers:   []string{"localhost:9876"},
		ConsumerGroup: "goods-sync-consumer-group",
		Topic:         "goods-binlog-topic",
		MaxReconsume:  3,
	}
}

// Validate 验证配置
func (o *RocketMQOptions) Validate() []error {
	var errors []error
	
	if len(o.NameServers) == 0 {
		errors = append(errors, fmt.Errorf("nameservers cannot be empty"))
	}
	
	if o.ConsumerGroup == "" {
		errors = append(errors, fmt.Errorf("consumer_group cannot be empty"))
	}
	
	if o.Topic == "" {
		errors = append(errors, fmt.Errorf("topic cannot be empty"))
	}
	
	if o.MaxReconsume <= 0 {
		errors = append(errors, fmt.Errorf("max_reconsume must be greater than 0"))
	}
	
	return errors
}

// AddFlags 添加命令行参数
func (o *RocketMQOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&o.NameServers, "rocketmq.nameservers", o.NameServers,
		"RocketMQ name servers list")
	
	fs.StringVar(&o.ConsumerGroup, "rocketmq.consumer-group", o.ConsumerGroup,
		"RocketMQ consumer group name")
	
	fs.StringVar(&o.Topic, "rocketmq.topic", o.Topic,
		"RocketMQ topic name")
	
	fs.Int32Var(&o.MaxReconsume, "rocketmq.max-reconsume", o.MaxReconsume,
		"Maximum number of reconsume attempts")
}