package conf

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var vp = viper.New()

func InitConfFile(confFile string) error {
	if confFile == "" {
		log.Warnf("[warning] config uninit, file is empty")
		return nil
	}
	//todo fix code
	cs := strings.Split(confFile, "/")
	if len(cs) < 2 {
		log.Warnf("[warning] config uninit, file %s\n", confFile)
		return nil
	}
	f_name := cs[len(cs)-1]

	vp.SetConfigName(f_name) // name of config file (without extension)
	vp.SetConfigType(f_name[strings.Index(f_name, ".")+1:])
	vp.AddConfigPath(strings.Join(cs[0:len(cs)-1], "/")) // optionally look for config in the working directory
	err := vp.ReadInConfig()                             // Find and read the config file
	if err != nil {                                      // Handle errors reading the config file
		return fmt.Errorf("fatal error config file: %v \n", err)
	}
	log.Infof("read from file %v success", confFile)
	return nil
}

func GetFormConfigFile(k string) interface{} {
	return vp.Get(k)
}

func SetFormConfigFile(k string, v interface{}) {
	vp.Set(k, v)
}

func SubFormConfigFile(key string) *viper.Viper {
	return vp.Sub(key)
}

// GetString returns the value associated with the key as a string.
func GetStringFormConfigFile(key string) string { return vp.GetString(key) }

// GetBool returns the value associated with the key as a boolean.
func GetBoolFormConfigFile(key string) bool { return vp.GetBool(key) }

// GetInt returns the value associated with the key as an integer.
func GetIntFormConfigFile(key string) int { return vp.GetInt(key) }

// GetInt32 returns the value associated with the key as an integer.
func GetInt32FormConfigFile(key string) int32 { return vp.GetInt32(key) }

// GetInt64 returns the value associated with the key as an integer.
func GetInt64FormConfigFile(key string) int64 { return vp.GetInt64(key) }

// GetUint returns the value associated with the key as an unsigned integer.
func GetUintFormConfigFile(key string) uint { return vp.GetUint(key) }

// GetUint32 returns the value associated with the key as an unsigned integer.
func GetUint32FormConfigFile(key string) uint32 { return vp.GetUint32(key) }

// GetUint64 returns the value associated with the key as an unsigned integer.
func GetUint64FormConfigFile(key string) uint64 { return vp.GetUint64(key) }

// GetFloat64 returns the value associated with the key as a float64.
func GetFloat64FormConfigFile(key string) float64 { return vp.GetFloat64(key) }

// GetTime returns the value associated with the key as time.
func GetTimeFormConfigFile(key string) time.Time { return vp.GetTime(key) }

// GetDuration returns the value associated with the key as a duration.
func GetDurationFormConfigFile(key string) time.Duration { return vp.GetDuration(key) }

// GetIntSlice returns the value associated with the key as a slice of int values.
func GetIntSliceFormConfigFile(key string) []int { return vp.GetIntSlice(key) }

// GetStringSlice returns the value associated with the key as a slice of strings.
func GetStringSliceFormConfigFile(key string) []string { return vp.GetStringSlice(key) }

// GetStringMap returns the value associated with the key as a map of interfaces.
func GetStringMapFormConfigFile(key string) map[string]interface{} { return vp.GetStringMap(key) }

// GetStringMapString returns the value associated with the key as a map of strings.
func GetStringMapStringFormConfigFile(key string) map[string]string {
	return vp.GetStringMapString(key)
}

// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
func GetStringMapStringSliceFormConfigFile(key string) map[string][]string {
	return vp.GetStringMapStringSlice(key)
}

// GetSizeInBytes returns the size of the value associated with the given key
// in bytes.
func GetSizeInBytesFormConfigFile(key string) uint { return vp.GetSizeInBytes(key) }
