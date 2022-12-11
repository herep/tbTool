package hystrix

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/afex/hystrix-go/hystrix"
)

type Config struct {
	Prefix string
}

func (c *Config) OnDelete(key []byte) error {
	hm.Delete(string(key))
	return nil
}

func (c *Config) OnPut(commandName []byte, cfg []byte) error {
	var commandConfig CommandConfig
	err := json.Unmarshal(cfg, &commandConfig)
	if err != nil {
		return fmt.Errorf("json unmarshal failed, error: [%s]", err.Error())
	}

	if commandConfig.Type == Client {
		notExistSetDefaultCommandConfig(&commandConfig)
	} else if commandConfig.Type == Server {
		notExistSetServerDefaultCommandConfig(&commandConfig)
	} else {
		log.Errorf("CommandConfig.Type can not be empty")
		return nil
	}

	name := JoinCommandName(strings.TrimPrefix(string(commandName), Hystrix+DotSep))
	hystrix.ConfigureCommand(name, hystrix.CommandConfig{
		Timeout:                commandConfig.Timeout,
		MaxConcurrentRequests:  commandConfig.MaxConcurrentRequests,
		RequestVolumeThreshold: commandConfig.RequestVolumeThreshold,
		SleepWindow:            commandConfig.SleepWindow,
		ErrorPercentThreshold:  commandConfig.ErrorPercentThreshold,
	})

	hm.Store(name, emptyStruct)

	log.Infof("hystrix command config is changed, name: %v, config: %+v", name, commandConfig)
	return nil
}

func (c *Config) GetPrefix() string {
	return c.Prefix
}
