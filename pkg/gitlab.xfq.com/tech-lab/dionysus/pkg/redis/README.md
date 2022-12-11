# pkg说明

### redis

主要通过watch etcd中的redis配置，动态实时更新redis client的配置。```dotPrefix = "gapi.redis."  conf.RegisterEtcdWatch(&redis.RedisEvent{Prefix: dotPrefix})```
这种方式下通过etcdClient.Put(context.Background(), "/projectname/gapi/redis/+"redisWatch", "{\"addr\":\"127.0.0.1:6379\", \"db\":0, \"password\":\"foobared\"}")往
etcd存入redis数据后，调用```redis.GetClient(context.Background(), dotPrefix+"redisWatch")```可以获取到对应的redis client。