package command

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/eviltomorrow/robber-core/pkg/mysql"
	"github.com/eviltomorrow/robber-core/pkg/pid"
	"github.com/eviltomorrow/robber-core/pkg/system"
	"github.com/eviltomorrow/robber-core/pkg/zlog"
	"github.com/eviltomorrow/robber-core/pkg/znet"
	"github.com/eviltomorrow/robber-repository/internal/config"
	"github.com/eviltomorrow/robber-repository/internal/server"
	"github.com/eviltomorrow/robber-repository/pkg/client"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	_ "net/http/pprof"
)

var rootCmd = &cobra.Command{
	Use:   "robber-repository",
	Short: "",
	Long:  "  \r\nrobber-repository server running",
	Run: func(cmd *cobra.Command, args []string) {
		if err := pid.CreatePidFile("robber-repository.pid"); err != nil {
			log.Fatalf("[Fatal] robber-repository create pid file failure, nest error: %v\r\n", err)
		}

		if pprofMode {
			go func() {
				port, err := znet.GetFreePort()
				if err != nil {
					log.Fatalf("[Fatal] Get free port failure, nest error: %v\r\n", err)
				}

				log.Printf("[Debug] Debug mode is started, pprof service is listend on http://%s:%d/debug/pprof", system.IP, port)
				if err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
					log.Fatalf("[Fatal] ListenAndServe pprof server failure, nest error: %v\r\n", err)
				}
			}()
		}

		setupCfg()
		setupVars()
		if err := mysql.Build(); err != nil {
			zlog.Fatal("Build mysql connection failure", zap.Error(err))
		}

		if err := server.StartupGRPC(); err != nil {
			zlog.Fatal("Startup GRPC service failure", zap.Error(err))
		}
		registerCleanFuncs()
		blockingUntilTermination()
	},
}

var (
	cleanFuncs []func() error
	cfg        = config.GlobalConfig
	cfgPath    = ""
	pprofMode  bool
)

func init() {
	rootCmd.CompletionOptions = cobra.CompletionOptions{
		DisableDefaultCmd: true,
	}
	rootCmd.Flags().StringVarP(&cfgPath, "config", "c", "config.toml", "robber-repository's config file")
	rootCmd.Flags().BoolVarP(&pprofMode, "pprof", "p", false, "robber-repository's pprof mode")
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func blockingUntilTermination() {
	var ch = make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	switch <-ch {
	case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
	case syscall.SIGUSR1:
	case syscall.SIGUSR2:
	default:
	}

	for _, f := range cleanFuncs {
		f()
	}
}

func registerCleanFuncs() {
	cleanFuncs = append(cleanFuncs, server.RevokeEtcdConn)
	cleanFuncs = append(cleanFuncs, server.ShutdownGRPC)
	cleanFuncs = append(cleanFuncs, pid.DestroyFile)
}

func findCfg() (string, error) {
	var possibleConf = []string{
		cfgPath,
		"./etc/config.toml",
		"../etc/config.toml",
		"/etc/config.toml",
	}
	for _, path := range possibleConf {
		if path == "" {
			continue
		}
		if _, err := os.Stat(path); err == nil {
			fp, err := filepath.Abs(path)
			if err == nil {
				return fp, nil
			}
			return path, nil
		}
	}
	return "", fmt.Errorf("not find config.toml, possible path: %v", possibleConf)
}

func setupCfg() {
	path, err := findCfg()
	if err != nil {
		log.Fatalf("[Fatal] Find config file failure, nest error: %v\r\n", err)
	}
	if err := cfg.Load(path, nil); err != nil {
		log.Fatalf("[Fatal] Load config file failure, nest error: %v\r\n", err)
	}

	global, prop, err := zlog.InitLogger(&zlog.Config{
		Level:            cfg.Log.Level,
		Format:           cfg.Log.Format,
		DisableTimestamp: cfg.Log.DisableTimestamp,
		File: zlog.FileLogConfig{
			Filename:   cfg.Log.FileName,
			MaxSize:    cfg.Log.MaxSize,
			MaxDays:    30,
			MaxBackups: 30,
		},
		DisableStacktrace: true,
	})
	if err != nil {
		log.Fatalf("[Fatal] Setup log config failure, nest error: %v\r\n", err)
	}
	zlog.ReplaceGlobals(global, prop)
	zlog.Info("Global Config info", zap.String("cfg", cfg.String()))
}

func setupVars() {
	server.Host = cfg.Server.Host
	server.Port = cfg.Server.Port
	server.Endpoints = cfg.Etcd.Endpoints

	mysql.DSN = cfg.MySQL.DSN

	client.EtcdEndpoints = cfg.Etcd.Endpoints
}
