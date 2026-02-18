package main

import (
	"encoding/json"
	"fmt"
	"github.com/fmnx/cftun/client"
	"github.com/fmnx/cftun/log"
	"github.com/spf13/pflag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

type ClientConfig struct {
	Client *client.Config `yaml:"client" json:"client"`
}

func parseConfig(configFile string) (*client.Config, error) {
	if configFile == "" {
		currentDir, _ := os.Getwd()
		configFile = filepath.Join(currentDir, "config.json")
	}
	buf, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	if len(buf) == 0 {
		return nil, fmt.Errorf("configuration file %s is empty", configFile)
	}
	rawCfg := &ClientConfig{}
	if err := json.Unmarshal(buf, rawCfg); err != nil {
		return nil, err
	}
	return rawCfg.Client, nil
}

var (
	configFile  string
	Version     = "unknown"
	BuildDate   = "unknown"
	BuildType   = "DEV"
	showVersion bool
	tunName     string
)

func init() {
	pflag.StringVarP(&configFile, "config", "c", "./config.json", "")
	pflag.BoolVarP(&showVersion, "version", "v", false, "")

	pflag.Usage = func() {
		fmt.Println("Cftun Client - Cloudflare Tunnel Client")
		fmt.Println("Usage:")
		fmt.Printf("  -c,--config\tSpecify the path to the config file.(default: \"./config.json\")\n")
		fmt.Printf("  -v,--version\tDisplay the current binary file version.\n")
	}
	pflag.Parse()
}

func main() {
	if showVersion {
		printVersion()
		return
	}

	cfg, err := parseConfig(configFile)
	if err != nil {
		log.Fatalln("Failed to parse config file: ", err.Error())
	}
	if cfg == nil {
		log.Fatalln("Client configuration is empty")
	}

	if cfg.Tun != nil && cfg.Tun.Enable {
		tunName = cfg.Tun.Name
		if tunName == "" {
			tunName = "cftun0"
			cfg.Tun.Name = tunName
		}
	}

	cfg.Run()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-sigCh:
			if tunName != "" {
				client.DeleteTunDevice(tunName)
			}
			return
		}
	}
}

func printVersion() {
	fmt.Printf("BuildType: %s\nCftunVersion: %s\nBuildDate: %s\n",
		BuildType, Version, BuildDate)
}
