package rabbitmq

import (
	"encoding/json"
	"fmt"
)

type rabbitEvent struct {
	prefix string
}

func GetRabbitEvent(prefix string) *rabbitEvent {
	return &rabbitEvent{prefix: prefix}
}
func (e *rabbitEvent) OnDelete(key []byte) error {
	log.Debugf("delete key %s", string(key))
	factories.Delete(string(key))
	deleteRabbitPool(string(key))
	return nil
}
func (e *rabbitEvent) GetPrefix() string {
	return e.prefix
}

func (e *rabbitEvent) OnPut(key []byte, value []byte) error {
	log.Debugf("put key %s,%s", string(key), string(value))
	//get or update redis client
	var rabbitConfig RabbitConfig
	err := json.Unmarshal(value, &rabbitConfig)
	if err != nil {
		log.Errorf("json unmarshal, error= %s", err.Error())
		return fmt.Errorf("json unmarshal, error= %s", err.Error())
	}
	if rabbitConfig.URL == "" {
		log.Errorf("未找到url配置，prefix: %v 无效", string(key))
		return fmt.Errorf("未找到url配置，prefix: %v 无效", string(key))
	}
	rabbitConfig.Key = string(key)
	err = initRabbitPool(string(key), &rabbitConfig)
	if err != nil {
		log.Errorf("initRabbitPool, error= %s", err.Error())
		return fmt.Errorf("initRabbitPool, error= %s", err.Error())
	}
	return nil
}

func StartRabbitWithFile(rbConfigs map[string]*RabbitConfig) error {
	if len(rbConfigs) == 0 {
		return fmt.Errorf("conf rabbit key prefix is empty")
	}
	for k, v := range rbConfigs {
		err := initRabbitPool(k, v)
		if err != nil {
			log.Errorf("BuildRedisData failed redisName %v, redisConfig: %v, err: %v", k, v, err)
			return err
		}
	}
	return nil
}
