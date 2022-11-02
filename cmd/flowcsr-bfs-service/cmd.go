package flowcsr_bfs_service

import (
	"bigrule/common/db"
	"bigrule/common/etcd"
	"bigrule/common/global"
	"bigrule/common/logger"
	"bigrule/pkg/format"
	"bigrule/services/flowcsr-bfs-service/config"
	"bigrule/services/flowcsr-bfs-service/middleware/queue"
	"bigrule/services/flowcsr-bfs-service/router"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/web"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	configYml string
	port      string
	mode      string
	StartCmd  = &cobra.Command{
		Use:          "flowcsr-bfs-service",
		Short:        "Start API flowcsr-bfs-service",
		Example:      "bigrule flowcsr-bfs-service -c config/settings.yml",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			setup()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}
)

func init() {
	StartCmd.PersistentFlags().StringVarP(&configYml, "config", "c", "config/settings.yml", "Start server with provided configuration file")
	StartCmd.PersistentFlags().StringVarP(&port, "port", "p", "8000", "Tcp port server listening on")
	StartCmd.PersistentFlags().StringVarP(&mode, "mode", "m", "dev", "server mode ; eg:debug,release")
}

func setup() {

	//1. 读取配置
	config.Setup(configYml)
	//2. 设置日志
	logger.InitLogger(config.LoggerConfig.Path, config.LoggerConfig.Level, config.LoggerConfig.Stdout)
	//3. 初始化数据库链接
	db.DBSetUp(config.DbMysqlConfig.Addr, config.DbMysqlConfig.Loglevel)

	usageStr := `starting api server`
	logger.Info(usageStr)

	queue.InitQ()
}

func run() error {
	if viper.GetString("bigrule.flowcsr-bfs-service.application.mode") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	//router register setup
	router.RouterSetup()
	//etcd setup
	etcd.Setup()
	// server init
	srv := &http.Server{
		Addr:    config.ApplicationConfig.Host + ":" + config.ApplicationConfig.Port,
		Handler: global.GinEngine,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		//注册服务
		microService := web.NewService(
			web.Name("repo-bfs-service"),
			//web.RegisterTTL(time.Second*30),//设置注册服务的过期时间
			//web.RegisterInterval(time.Second*20),//设置间隔多久再次注册服务
			web.Address(config.ApplicationConfig.Host+":"+config.ApplicationConfig.Port),
			web.Handler(global.GinEngine),
			web.Registry(global.EtcdReg),
		)
		if err := microService.Run(); err != nil {
			logger.Fatal("listen: ", err)
		}

	}()
	fmt.Println(format.Red(string(global.Logo)))
	tip()
	fmt.Println(format.Green("Server run at:"))
	fmt.Printf("-  Local:   http://localhost:%s/ \r\n", config.ApplicationConfig.Port)
	fmt.Printf("-  Network: http://%s:%s/ \r\n", format.GetLocaHonst(), config.ApplicationConfig.Port)
	fmt.Printf("%s Enter Control + C Shutdown Server \r\n", format.GetCurrentTimeStr())
	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	fmt.Printf("%s Shutdown Server ... \r\n", format.GetCurrentTimeStr())

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server Shutdown:", err)
	}
	logger.Info("Server exiting")

	return nil
}

func tip() {
	usageStr := `欢迎使用 ` + format.Green(global.ProjectName+" "+global.Version) + ` 可以使用 ` + format.Red(`-h`) + ` 查看命令`
	fmt.Printf("%s \n\n", usageStr)
}
