package main

import (
	"encoding/json"
	"fmt"
	"github.com/fmnx/cftun/log"
	"github.com/fmnx/cftun/server"
	"github.com/spf13/pflag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

type ServerConfig struct {
	Server *server.Config `yaml:"server" json:"server"`
}

func parseConfig(configFile string) (*server.Config, error) {
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
	rawCfg := &ServerConfig{}
	if err := json.Unmarshal(buf, rawCfg); err != nil {
		return nil, err
	}
	return rawCfg.Server, nil
}

var (
	configFile         string
	token              string
	isQuick            bool
	proxy4             bool
	proxy6             bool
	port               int
	Version            = "unknown"
	BuildDate          = "unknown"
	BuildType          = "DEV"
	CloudflaredVersion = "2025.4.1"
	showVersion        bool
	quickData          = &server.QuickData{}
)

func init() {
	pflag.StringVarP(&configFile, "config", "c", "./config.json", "")
	pflag.StringVarP(&token, "token", "t", "", "")
	pflag.BoolVarP(&isQuick, "quick", "q", false, "")
	pflag.BoolVarP(&proxy4, "proxy4", "4", false, "")
	pflag.BoolVarP(&proxy6, "proxy6", "6", false, "")
	pflag.IntVarP(&port, "port", "p", 51280, "")
	pflag.BoolVarP(&showVersion, "version", "v", false, "")

	pflag.Usage = func() {
		fmt.Println("Cftun Server - Cloudflare Tunnel Server")
		fmt.Println("Usage:")
		fmt.Printf("  -c,--config\tSpecify the path to the config file.(default: \"./config.json\")\n")
		fmt.Printf("  -t,--token\tWhen a token is provided, the configuration file will be ignored.\n")
		fmt.Printf("  -q,--quick\tTemporary server, no Cloudflare account required, based on try.cloudflare.com.\n")
		fmt.Printf("  -4,--proxy4\tUse the WARP proxy for IPv4 traffic; Ignored when using a configuration file.\n")
		fmt.Printf("  -6,--proxy6\tUse the WARP proxy for IPv6 traffic; Ignored when using a configuration file.\n")
		fmt.Printf("  -p,--port\tSet the local port for WARP; Ignored when using a configuration file.\n")
		fmt.Printf("  -v,--version\tDisplay the current binary file version.\n")
	}
	pflag.Parse()
}

func main() {
	bInfo := server.GetBuildInfo(BuildType, CloudflaredVersion)
	if showVersion {
		printVersion(bInfo)
		return
	}
	
	var srv *server.Config
	if token != "" || isQuick {
		var warp *server.Warp
		if proxy4 || proxy6 {
			warp = &server.Warp{
				Auto:   true,
				Port:   uint16(port),
				Proxy4: proxy4,
				Proxy6: proxy6,
			}
		}
		if isQuick {
			token = "quick"
		} else if token == "quick" {
			isQuick = true
		}
		srv = &server.Config{
			Token:  token,
			HaConn: 4,
			Warp:   warp,
		}
	} else {
		var err error
		srv, err = parseConfig(configFile)
		if err != nil {
			log.Fatalln("Failed to parse config file: ", err.Error())
		}
		if srv == nil {
			log.Fatalln("Server configuration is empty")
		}
		if srv.Token == "quick" {
			isQuick = true
		}
	}

	go srv.Run(bInfo, quickData)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-sigCh:
			if isQuick {
				quickData.Save()
			}
			return
		}
	}
}

func printVersion(buildInfo *server.BuildInfo) {
	fmt.Printf("GoOS: %s\nGoArch: %s\nGoVersion: %s\nBuildType: %s\nCftunVersion: %s\nBuildDate: %s\n",
		buildInfo.GoOS, buildInfo.GoArch, buildInfo.GoVersion, buildInfo.BuildType, Version, BuildDate)
}
