package orm

import (
	"encoding/json"
	"fmt"
)

type ormEvent struct {
	Prefix string
}

func NewOrmEvent(prefix string) *ormEvent {
	return &ormEvent{Prefix: prefix}
}

func (e *ormEvent) OnDelete(key []byte) error {
	log.Debugf("ormEvent delete key = %s", string(key))
	clearOrmResource(string(key))
	return nil
}
func (e *ormEvent) GetPrefix() string {
	return e.Prefix
}

func (e *ormEvent) OnPut(key []byte, value []byte) error {

	log.Debugf("ormEvent put key:%s,value:%s", string(key), string(value))

	// get or update orm client
	oc := newOrmConf()
	err := json.Unmarshal(value, oc)
	if err != nil {
		return fmt.Errorf("ormEvent json unmarshal, error= %s", err.Error())
	}

	// 验证orm传入的参数
	if oc.DbURL == "" || oc.MaxOpenConns <= 0 || oc.MaxIdleConns <= 0 || oc.MaxLifetime <= 0 {
		return fmt.Errorf("ormEvent检测gorm配置错误，prefix: %s 无效", string(key))
	}

	if err = initOrmPool(string(key), oc); err != nil {
		return fmt.Errorf("ormEvent initOrmPool, error= %s", err.Error())
	}
	return nil
}
