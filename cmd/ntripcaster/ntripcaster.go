package main

import (
	"flag"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/go-gnss/ntrip/internal/caster"
)

// Config exists to demonstrate that this responsibility lies with the
// application implementing the ntrip.caster module and not the module itself
type Config struct {
	HTTP struct {
		Port string
	}
}

func main() {
	ntripcaster := caster.Caster{
		Mounts:  make(map[string]*caster.Mountpoint),
		Timeout: 5 * time.Second,
	} // TODO: Hide behind NewCaster which can include a DefaultAuthenticator
	log.SetFormatter(&log.JSONFormatter{})

	configFile := flag.String("config", "configs/caster.yml", "Path to config file")
	flag.Parse()

	viper.SetConfigFile(*configFile)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	conf := Config{}
	err = viper.Unmarshal(&conf)
	if err != nil {
		panic(err)
	}

	log.Info("NTRIP Caster started")
	panic(ntripcaster.ListenHTTP(conf.HTTP.Port))
}
