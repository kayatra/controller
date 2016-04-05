package web

import(
  "github.com/gin-gonic/gin"
  log "github.com/Sirupsen/logrus"
  "time"
  "github.com/kayatra/controller/plugins"
)

func logger() gin.HandlerFunc{
  return func(c *gin.Context) {
    t := time.Now()
    c.Next()
    latency := time.Since(t)
    status := c.Writer.Status()
    clientIP := c.ClientIP()
    method := c.Request.Method

    uaHeader := c.Request.Header["User-Agent"]
    var userAgent string
    if uaHeader != nil && len(uaHeader) > 0{
      userAgent = string(uaHeader[0])
    }

    logFields := log.Fields{
      "Duration": latency,
      "Status": status,
      "ClientIP": clientIP,
      "Method": method,
      "Path": c.Request.URL.String(),
      "User-Agent": userAgent,
    }

    logBase := log.WithFields(logFields)
    logMsg := "HTTP Request"

    if status >= 500{
      logBase.Error(logMsg)
    } else if status >= 400 {
      logBase.Warning(logMsg)
    } else {
      logBase.Info(logMsg)
    }
  }
}


var router = gin.New()

func Start(bind string){
  router.Use(logger())
  router.Use(gin.Recovery())

  router.GET("/plugin/transport", func(c *gin.Context){
    plugins.Transport(c.Writer, c.Request)
  })

  log.WithFields(log.Fields{
    "bind":bind,
  }).Info("Starting gin server")
  router.Run(bind)
}
