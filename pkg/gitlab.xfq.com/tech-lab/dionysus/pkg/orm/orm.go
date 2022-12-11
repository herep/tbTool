package orm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

const (
	MaxLifetime  = 2 * 3600 // 单位  time.Second
	MaxOpenConns = 100      // 设置数据库连接池最大连接数
	MaxIdleConns = 5        // 连接池最大允许的空闲连接数，如果没有sql任务需要执行的连接数大于5，超过的连接会被连接池关闭
)

var ormPoolMap sync.Map

// orm链接配置
type ormConf struct {
	OrmName string `json:"ormName"`
	DbURL   string `json:"dbURL"`

	MaxLifetime  int `json:"maxLifetime"`
	MaxIdleConns int `json:"maxIdleConns"`
	MaxOpenConns int `json:"maxOpenConns"`
}

func newOrmConf() *ormConf {
	return &ormConf{
		MaxLifetime:  MaxLifetime,
		MaxIdleConns: MaxIdleConns,
		MaxOpenConns: MaxOpenConns,
	}
}

type Client struct {
	*gorm.DB
}

// 从资源池获取资源
func pickupOrmDB(key string) (*gorm.DB, error) {
	lowerName := strings.ToLower(key)
	result, ok := ormPoolMap.Load(lowerName)
	if ok {
		return result.(*gorm.DB), nil
	}
	log.Errorf("can not get gorm db ,ormName= %s", key)
	ormPoolMap.Range(func(key, value interface{}) bool {
		log.Debugf("db pool has orm key(%v)", key)
		return true
	})
	return nil, errors.New("can not get gorm db ,key=" + key)
}

// 判断是否存在
func getDbMap(key string) *gorm.DB {
	lowerName := strings.ToLower(key)
	result, ok := ormPoolMap.Load(lowerName)
	if ok {
		return result.(*gorm.DB)
	}
	return nil
}

// 关闭资源池
func defCloseDB(db *gorm.DB, key string) {

	if db != nil {
		time.AfterFunc(time.Second*30, func() {
			if err := closeDB(db, key); err != nil {
				log.Errorf("deferCloseOrmPool err:%v", err)
			}

		})
	}
}

func closeDB(db *gorm.DB, key string) error {
	if db != nil {
		if e := db.Close(); e != nil {
			return fmt.Errorf("close ormDB fail,error=%v, key: %s", e, key)
		}
	}
	return nil
}

// 组装转换数据，初始化资源池
func initOrmPool(key string, ormConf *ormConf) error {
	db, err := newDB(ormConf)
	if err != nil {
		return fmt.Errorf("gorm client(%+v) 连接失败, error=%+v", key, err)
	}

	lowerName := strings.ToLower(key)

	hisOrmDB := getDbMap(lowerName)
	defCloseDB(hisOrmDB, lowerName) //  防止内存泄漏

	// 设置 db存活期；空闲最大链接数；最大链接数；
	db.DB().SetConnMaxLifetime(time.Duration(ormConf.MaxLifetime) * time.Second)
	db.DB().SetMaxIdleConns(ormConf.MaxIdleConns)
	db.DB().SetMaxOpenConns(ormConf.MaxOpenConns)
	// 新加入或替换
	ormPoolMap.Store(lowerName, db)

	log.Infof("gorm rebuild client pool done - %v", lowerName)
	return nil
}

/*
	获取orm客户端
	ctx
	ormName 配置中的信息
*/
// Caution!!! This function return a db-client, you should not close it after use.
func GetClient(ctx context.Context, key string) (*Client, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context nil")
	}
	ormDB, err := pickupOrmDB(key)
	if err != nil {
		return nil, fmt.Errorf("convert orm fail, initOrmPoolMap.Load(%s) = %+v", key, ormDB)
	}

	return &Client{
		ormDB,
	}, nil

}

// new orm pool from config center by watch
func newDB(oc *ormConf) (*gorm.DB, error) {

	db, err := gorm.Open(oc.OrmName, oc.DbURL)
	if err != nil {
		return nil, fmt.Errorf("gorm open mysql connect error: %v", err)
	}

	return db, nil
}

// 删除orm资源
func clearOrmResource(key string) {
	lowerName := strings.ToLower(key)
	pool := getDbMap(lowerName)
	ormPoolMap.Delete(lowerName)
	log.Infof("delete orm client pool done - %v", lowerName)
	defCloseDB(pool, lowerName)
}
