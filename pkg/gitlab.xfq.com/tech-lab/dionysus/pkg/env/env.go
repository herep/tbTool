package env

import (
	"fmt"
)

type Environment string

const (
	Product Environment = "product" // 线上生成环境
	Gray    Environment = "gray"    // 灰度环境
	Test    Environment = "test"    // 测试环境
	Develop Environment = "develop" // 开发环境

	SysEnvKey = "DIONYSUS_ENV"
)

var (
	env           = Product
	availableEnvs = map[Environment]bool{Product: true, Gray: true, Test: true, Develop: true}
)

func Get() Environment {
	return env
}

func Set(e Environment) error {
	if has, ok := availableEnvs[e]; !has || !ok {
		var keys string
		for i := range availableEnvs {
			keys = fmt.Sprintf("%s %s ", keys, i)
		}
		return fmt.Errorf("Input env should be one of:[%s], but got:[%s] ", keys, e)
	}
	env = e
	return nil
}

func IsProduct() bool {
	return env == Product
}

func IsGray() bool {
	return env == Gray
}

func IsTest() bool {
	return env == Test
}

func IsDevelop() bool {
	return env == Develop
}
