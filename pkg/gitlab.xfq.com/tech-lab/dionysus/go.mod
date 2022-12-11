module gitlab.xfq.com/tech-lab/dionysus

go 1.14

require (
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/coreos/etcd v3.3.22+incompatible
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/gin-contrib/pprof v1.3.0
	github.com/gin-gonic/gin v1.6.2
	github.com/go-redis/redis/v7 v7.2.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/protobuf v1.3.5 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20190430165422-3e4dfb77656c // indirect
	github.com/jinzhu/gorm v1.9.13
	github.com/onsi/ginkgo v1.12.0 // indirect
	github.com/onsi/gomega v1.9.0 // indirect
	github.com/panjf2000/ants/v2 v2.4.1
	github.com/prometheus/client_golang v1.3.0
	github.com/prometheus/common v0.9.1 // indirect
	github.com/smartystreets/assertions v1.0.0 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.3
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	go.uber.org/zap v1.15.0 // indirect
	golang.org/x/crypto v0.0.0-20200323165209-0ec3e9974c59 // indirect
	golang.org/x/tools v0.0.0-20191216052735-49a3e744a425 // indirect
	google.golang.org/grpc v1.26.0
	sigs.k8s.io/yaml v1.2.0 // indirect
	vitess.io/vitess v3.0.0-rc.3+incompatible
	gitlab.xfq.com/tech-lab/ngkit v0.0.0-00010101000000-000000000000
	gitlab.xfq.com/tech-lab/utils v0.0.0-00010101000000-000000000000
	gitlab.xfq.com/tech-lab/watcher v0.0.0-00010101000000-000000000000
	github.com/prometheus/client_golang v1.3.0
)

replace (
	gitlab.xfq.com/tech-lab/ngkit => ../ngkit/log
    gitlab.xfq.com/tech-lab/utils => ../utils
    gitlab.xfq.com/tech-lab/watcher => ../watcher
)
