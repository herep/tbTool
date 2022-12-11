## rabbitmq

### 基本功能

rabbitmq的简单封装,已经屏蔽掉了相关的Exchange、Queue、QueueBind等操作

获取rabbitmq客户端channel通过 PickupRabbitClient 方法
发送消息通过 DioPublish 方法
订阅消息通过 DioSubscribe 方法
关闭rabbitmq客户端channel通过 Close 方法，注意：用完如果不用后一定要关掉哦！！！链接池由pool_resource维护，close会归还channel

对于topic的使用各个业务方来自己通过标示保证不冲突，可以借鉴redis key的设计规范
例子如下
```go

func main() {
	err := conf.InitEtcdClient([]string{"127.0.0.1:2379"}, "dionysus")
	if err != nil {
		log.Fatal(" init etcd failed, error: ", err)
	}

	RegisterDemoRouter()


	srv := activate.StartEngine()

	err = RegisterSubscribe()
	if err != nil {
		log.Fatal("init subscribe failed, error: ",err)
	}

	activate.ShutDown(srv)
}

func RegisterDemoRouter() {
	activate.RegActionSteps("rabbit",2, func() error {
		// etcdctl put /dionysus/rabbit/rabbit "{\"url\":\"amqp://guest:guest@127.0.0.1:5672/\"}"
		return conf.RegisterEtcdWatch(rabbitmq.GetRabbitEvent("rabbit"))
	})

	// 建立路由组
	group := activate.NewGroup("/rabbitmq")

	// rabbit publish 使用例子
	// publish，简单用法
	group.RegRouterPost("/demo/publish", DemoPublishRabbitMQ)
}

func RegisterSubscribe() error{
	// 从资源池中获取channel
	ch, err := rabbitmq.PickupRabbitClient(context.Background(), "rabbit.rabbit")
	if err != nil {
		logger.Errorf("pickup rabbit client error: %v",err)
		return err
	}
	// 建议autoAck 是false，由业务处理情况来回复ack,handler 为具体要处理的业务逻辑
	err = ch.DioSubscribe("example-topic", false, handler)
	if err != nil {
		return err
	}
	return nil
}

// 业务处理流程
func handler(delivery amqp.Delivery) error {
	logger.Infof("subscribe %s",string(delivery.Body))
	// 1、2、3 各种业务

	// autoAck 为false，处理成功要回复ACK
	return delivery.Ack(true)
}



func DemoPublishRabbitMQ(c *gin.Context) pkg.Render {
	data, err := ioutil.ReadAll(bufio.NewReader(c.Request.Body))
	if err != nil {
		logger.Errorf("io read error: %v",err)
		return pkg.JSON{Data: gin.H{"code": 20001, "err": err.Error(), "msg": "err"}}
	}

	// 从资源池中获取channel
	ch, err := rabbitmq.PickupRabbitClient(c.Request.Context(), "rabbit.rabbit")
	if err != nil {
		logger.Errorf("pickup rabbit client error: %v",err)
		return pkg.JSON{Data: gin.H{"code": 20002, "err": err.Error(), "msg": "err"}}
	}

	// 向rabbitmq中发送消息
	if err := ch.DioPublish("example-topic", data); err != nil {
		logger.Errorf("ch dio publish error: %v",err)
		return pkg.JSON{Data: gin.H{"code": 20003, "err": err.Error(), "msg": "err"}}
	}
	logger.Infof("pub %s", string(data))

	// 关闭使用的ch，注意：拿了ch必须归还ch，不然会导致ch被耗尽
	if err := ch.Close();err != nil {
		logger.Errorf("close ch error: %v",err)
	}

	return pkg.JSON{Data: gin.H{"code": 20000, "data": string(data), "msg": "OK"}}
}
```

### 原生方法使用
如果不想使用封装的方法，可以使用原生的方法来操作

```go
func main() {
	err := logger.Setup()
	if err != nil {
		panic(err)
	}
	err = conf.Setup()
	if err != nil {
		panic(err)
	}

	PutRabbitConfigToEtcd()

	// 1、获取client channel
	ch, err := rabbitmq.PickupRabbitClient(context.TODO(), rabbitName)
	if err != nil {
		panic(err)
	}

	// 2、绑定 exchange
	err = ch.DeclareExchange(exchangeName, string(rabbitmq.ExchangeKindDirect), true, false)
	if err != nil {
		panic(err)
	}

	// 3、绑定 queue
	q, err := ch.DeclareQueue(topicName, true, false)
	if err != nil {
		panic(err)
	}

	err = ch.Bind(q.Name, topicName, exchangeName)
	if err != nil {
		panic(err)
	}

	// 4、订阅 queue
	msgChan, err := ch.Consume(q.Name, true)
	if err != nil {
		panic(err)
	}

	// 5、读取数据
	go func() {
		for msg := range msgChan {
			count++
			logger.Debugf("consum msg %s", string(msg.Body))
		}
	}()

	// 打印消费情况
	sigChan := make(chan os.Signal, 1)
	exitChan := make(chan struct{})
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	logger.Debug("signal", <-sigChan)

	go func() {
		logger.Warn("stopping ")
		logger.Debugf("consumer count %d", count)
		exitChan <- struct{}{}
	}()

	select {
	case <-exitChan:
	case s := <-sigChan:
		logger.Debugf("signal %s", s)
	}
}
```