package redis

import (
	"encoding/json"
	"fmt"
)

func StartRedisFile(rdConfigs map[string]*Rdconfig) {
	for k, v := range rdConfigs {
		err := InitClient(k, v)
		if err != nil {
			log.Errorf("BuildRedisData failed redisName %v, redisConfig: %v, err: %v", k, v, err)
			continue
		}
	}
}

type RedisEvent struct {
	Prefix string
}

func (e *RedisEvent) OnDelete(key []byte) error {
	deleteClientsData(string(key))
	return nil
}
func (e *RedisEvent) GetPrefix() string {
	return e.Prefix
}

func (e *RedisEvent) OnPut(key []byte, value []byte) error {
	//get or update redis client
	var redisConfig Rdconfig
	err := json.Unmarshal(value, &redisConfig)
	if err != nil {
		return fmt.Errorf("json unmarshal, error= %s", err.Error())
	}
	if redisConfig.Addr == "" {
		return fmt.Errorf("未找到addr配置，prefix: %v 无效", string(key))
	}

	err = InitClient(string(key), &redisConfig)
	if err != nil {
		return fmt.Errorf("buildRedisData, error= %s", err.Error())
	}
	return nil
}

/*
func StartRedisWithWatch(redisPrefix string) error {
	if redisPrefix == "" {
		return errors.New("watch redis key prefix is empty")
	}
	err := conf.RegisterEtcdWatch(&RedisEvent{prefix: redisPrefix})
	return err
}
*/
