package config

import (
	"github.com/spf13/viper"
)

var AppConfig Config

type Config struct {
	Feed
	Kafka
	Elastic
	Retry
}

type Feed struct {
	Url string `mapstructure:"url"`
}

type Kafka struct {
	Topic             string `mapstructure:"topic"`
	NumPartitions     int    `mapstructure:"num_partitions"`
	ReplicationFactor int    `mapstructure:"replication_factor"`
	BootstrapServers  string `mapstructure:"bootstrap_servers"`
	SchemaRegistryUrl string `mapstructure:"schema_registry_url"`
	AvroSchemaFile    string `mapstructure:"avro_schema_file"`
	Consumer
}

type Consumer struct {
	ConsumerGroupId string `mapstructure:"consumer_group_id"`
	AutoOffsetReset string `mapstructure:"auto_offset_reset"`
}

type Elastic struct {
	Url         string `mapstructure:"url"`
	Index       string `mapstructure:"index"`
	MappingFile string `mapstructure:"mapping_file"`
}

type Retry struct {
	SleepTime int `mapstructure:"sleep_time"`
	Timeout   int `mapstructure:"timeout"`
}

func init() {
	viper.AddConfigPath("./../shared/config")
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		panic(err)
	}
}
