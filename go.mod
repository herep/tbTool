module tbTool

go 1.16

require (
	github.com/gin-gonic/gin v1.8.1
	gitlab.xfq.com/tech-lab/dionysus v0.0.0-00010101000000-000000000000
	go.uber.org/dig v1.14.1
)

replace (
	gitlab.xfq.com/tech-lab/dionysus => ./pkg/gitlab.xfq.com/tech-lab/dionysus
	gitlab.xfq.com/tech-lab/ngkit => ./pkg/gitlab.xfq.com/tech-lab/ngkit/log
	gitlab.xfq.com/tech-lab/utils => ./pkg/gitlab.xfq.com/tech-lab/utils
	gitlab.xfq.com/tech-lab/watcher => ./pkg/gitlab.xfq.com/tech-lab/watcher
)
