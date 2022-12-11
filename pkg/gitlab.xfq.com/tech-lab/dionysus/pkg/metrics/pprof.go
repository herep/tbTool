package metrics

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

func RegisterPProf(r *gin.Engine, prefix string) {
	pprof.Register(r, prefix)
}
