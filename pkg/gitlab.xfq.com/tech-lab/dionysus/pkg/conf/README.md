# pkg说明

### conf

主要提供从配置文件读取和从etcd读取配置文件两种方式。从配置文件读取时通过-c, --config string, 或者用环境变量GAPI_CONFIG配置指定配置文件的路径。使用conf.GetXXXFormConfigFile函数读取。
通过etcd读取配置时通过```StartWatchConfig("test.pre")```动态读取etcd中/projectname/test/pre路径下的配置。其中projectname由环境变量GAPI_PROJECT_NAME指定。conf.LoadXXX()读取配置。

example:
读取本目录下的t_conf.yaml的配置```conf.GetStringFormConfigFile("test_config.child_config.name")="child_config"```
在etcd中读取配置，首先需要调用```etcdpath := "test.pre", StartWatchConfig(etcdpath)```注册动态获取etcd配置的路径，
在etcd中存入配置后```etcdClient.Put(context.Background(), "/dionysus/test/pre/before", "1.2345"), 通过conf.LoadBytes("test.pre.before")=1.2345```
