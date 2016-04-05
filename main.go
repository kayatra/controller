package main

import(
  log "github.com/Sirupsen/logrus"
  "github.com/voxelbrain/goptions"
  "github.com/home-control/controller/config"
  "github.com/home-control/controller/web"
)

type options struct {
  Config    string          `goptions:"-c, --config, description='Path to config file folders'"`
  Verbose   bool            `goptions:"-v, --verbose, description='Log verbosely'"`
  Bind      string          `goptions:"--bind, description='Bind address'"`
  Help      goptions.Help   `goptions:"-h, --help, description='Show help'"`
}

func main() {
  parsedOptions := options{}
  parsedOptions.Bind = ":9211"

  goptions.ParseAndFail(&parsedOptions)

  if parsedOptions.Verbose{
    log.SetLevel(log.DebugLevel)
    log.Debug("Logging verbosely")
  } else {
    log.SetLevel(log.InfoLevel)
  }

  log.SetFormatter(&log.TextFormatter{FullTimestamp:true})

  config.Load(parsedOptions.Config)

  web.Start(parsedOptions.Bind)
}
