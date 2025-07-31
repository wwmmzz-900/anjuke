package main

import (
	"flag"
	"os"

	appointmentpb "anjuke/server/api/appointment/v1"
	companypb "anjuke/server/api/company/v1"
	"anjuke/server/internal/biz"
	"anjuke/server/internal/data"
	"anjuke/server/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string = "appointment-server"
	// Version is the version of the compiled software.
	Version string = "v1.0.0"
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs/config.yaml", "config path, eg: -conf config.yaml")
}

// Config 应用配置结构
type Config struct {
	Server struct {
		HTTP struct {
			Addr    string `yaml:"addr"`
			Timeout string `yaml:"timeout"`
		} `yaml:"http"`
		GRPC struct {
			Addr    string `yaml:"addr"`
			Timeout string `yaml:"timeout"`
		} `yaml:"grpc"`
	} `yaml:"server"`
	Data struct {
		MySQL data.DBConfig `yaml:"mysql"`
	} `yaml:"data"`
	Log struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
		Output string `yaml:"output"`
	} `yaml:"log"`
}

func newApp(logger log.Logger, hs *http.Server, gs *grpc.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			hs,
			gs,
		),
	)
}

func main() {
	flag.Parse()

	// 初始化日志
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
	)

	// 加载配置
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc Config
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	// 初始化数据层
	dataDB, dbCleanup, err := data.NewDBData(&bc.Data.MySQL, logger)
	if err != nil {
		panic(err)
	}
	defer dbCleanup()

	// 创建Data实例
	dataInstance, dataCleanup, err := data.NewData(dataDB, logger)
	if err != nil {
		panic(err)
	}
	defer dataCleanup()

	// 初始化仓储层
	appointmentRepo := data.NewAppointmentDBRepo(dataInstance)
	storeRepo := data.NewStoreDBRepo(dataDB, logger)
	realtorRepo := data.NewRealtorDBRepo(dataDB, logger)
	companyRepo := data.NewCompanyDBRepo(dataInstance)

	// 初始化业务逻辑层
	appointmentUC := biz.NewAppointmentUsecase(appointmentRepo, storeRepo, realtorRepo, logger)
	companyUC := biz.NewCompanyUsecase(companyRepo, storeRepo, logger)
	storeUC := biz.NewStoreUsecase(storeRepo, companyRepo, realtorRepo, logger)
	realtorUC := biz.NewRealtorUsecase(realtorRepo, storeRepo, companyRepo, logger)

	// 初始化服务层
	appointmentService := service.NewAppointmentService(appointmentUC, logger)
	companyService := service.NewCompanyService(companyUC, logger)
	storeService := service.NewStoreService(storeUC, companyUC, logger)
	realtorService := service.NewRealtorService(realtorUC, storeUC, companyUC, logger)

	// 初始化HTTP服务器
	httpSrv := http.NewServer(
		http.Address(bc.Server.HTTP.Addr),
		http.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
		),
	)

	// 初始化gRPC服务器
	grpcSrv := grpc.NewServer(
		grpc.Address(bc.Server.GRPC.Addr),
		grpc.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
		),
	)

	// 注册服务
	appointmentpb.RegisterAppointmentServiceServer(grpcSrv, appointmentService)
	appointmentpb.RegisterAppointmentServiceHTTPServer(httpSrv, appointmentService)

	companypb.RegisterCompanyServer(grpcSrv, companyService)
	companypb.RegisterCompanyHTTPServer(httpSrv, companyService)

	companypb.RegisterStoreServer(grpcSrv, storeService)
	companypb.RegisterStoreHTTPServer(httpSrv, storeService)

	companypb.RegisterRealtorServer(grpcSrv, realtorService)
	companypb.RegisterRealtorHTTPServer(httpSrv, realtorService)

	// 创建应用
	app := newApp(logger, httpSrv, grpcSrv)

	// 启动应用
	if err := app.Run(); err != nil {
		panic(err)
	}
}
