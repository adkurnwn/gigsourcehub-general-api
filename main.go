package main

import (
	"github.com/adkurnwn/gigsourcehub-general-api/app/consumer"
	http_bookmark "github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/bookmark"
	http_aichat "github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/ai_chat"
	http_cv "github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/cv"
	http_activity_log "github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/activity_log"
	http_job_role "github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/job_role"
	http_job_title "github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/job_title"
	http_kabupaten_kota "github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/kabupaten_kota"
	http_member "github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/member"
	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	http_provinsi "github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/provinsi"
	http_recruitment_status "github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/recruitment_status"
	httpdelivery_request "github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/request"
	http_search "github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/search"
	http_sector "github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/sector"
	http_system_setting "github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/system_setting"
	delivery_grpc "github.com/adkurnwn/gigsourcehub-general-api/app/delivery/grpc"
	pb "github.com/adkurnwn/gigsourcehub-general-api/proto"
	aisearchrepo "github.com/adkurnwn/gigsourcehub-general-api/app/repository/ai_search"
	gormrepo "github.com/adkurnwn/gigsourcehub-general-api/app/repository/gorm"
	rabbitmqrepo "github.com/adkurnwn/gigsourcehub-general-api/app/repository/rabbitmq"
	s3repo "github.com/adkurnwn/gigsourcehub-general-api/app/repository/s3"
	usecase_bookmark "github.com/adkurnwn/gigsourcehub-general-api/app/usecase/bookmark"
	usecase_aichat "github.com/adkurnwn/gigsourcehub-general-api/app/usecase/ai_chat"
	usecase_cv "github.com/adkurnwn/gigsourcehub-general-api/app/usecase/cv"
	usecase_activity_log "github.com/adkurnwn/gigsourcehub-general-api/app/usecase/activity_log"
	usecase_job_role "github.com/adkurnwn/gigsourcehub-general-api/app/usecase/job_role"
	usecase_job_title "github.com/adkurnwn/gigsourcehub-general-api/app/usecase/job_title"
	usecase_kabupaten_kota "github.com/adkurnwn/gigsourcehub-general-api/app/usecase/kabupaten_kota"
	usecase_member "github.com/adkurnwn/gigsourcehub-general-api/app/usecase/member"
	usecase_provinsi "github.com/adkurnwn/gigsourcehub-general-api/app/usecase/provinsi"
	usecase_recruitment_status "github.com/adkurnwn/gigsourcehub-general-api/app/usecase/recruitment_status"
	usecase_request "github.com/adkurnwn/gigsourcehub-general-api/app/usecase/request"
	usecase_search "github.com/adkurnwn/gigsourcehub-general-api/app/usecase/search"
	usecase_sector "github.com/adkurnwn/gigsourcehub-general-api/app/usecase/sector"
	usecase_system_setting "github.com/adkurnwn/gigsourcehub-general-api/app/usecase/system_setting"
	"github.com/adkurnwn/gigsourcehub-general-api/docs"

	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/gorm/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	swaggerFiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
)

func init() {
	_ = godotenv.Load()
}

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Insert JWT Token with format: **Bearer {your token}**

func main() {
	// programmatically set swagger info
	docs.SwaggerInfo.Title = "Swagger Golang API"
	docs.SwaggerInfo.Description = "Documentations"
	docs.SwaggerInfo.Version = "1.0"

	// Use APP_URL if set, otherwise fallback to HOST:PORT
	appURL := os.Getenv("APP_URL")
	if appURL != "" {
		// Parse APP_URL to extract scheme and host
		u, err := url.Parse(appURL)
		if err == nil {
			docs.SwaggerInfo.Host = u.Host
			docs.SwaggerInfo.Schemes = []string{u.Scheme}
		} else {
			logrus.Warnf("Failed to parse APP_URL: %v, falling back to HOST:PORT", err)
			docs.SwaggerInfo.Host = fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT"))
			docs.SwaggerInfo.Schemes = []string{"http", "https"}
		}
	} else {
		// Fallback to HOST:PORT
		docs.SwaggerInfo.Host = fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT"))
		docs.SwaggerInfo.Schemes = []string{"http", "https"}
	}

	docs.SwaggerInfo.BasePath = "/api"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	timeoutStr := os.Getenv("TIMEOUT")
	if timeoutStr == "" {
		timeoutStr = "5"
	}
	timeout, _ := strconv.Atoi(timeoutStr)
	timeoutContext := time.Duration(timeout) * time.Second

	aiSearchTimeoutStr := os.Getenv("AI_SEARCH_TIMEOUT")
	if aiSearchTimeoutStr == "" {
		aiSearchTimeoutStr = "120" // Default for local LLMs
	}
	aiSearchTimeoutVal, _ := strconv.Atoi(aiSearchTimeoutStr)
	aiSearchTimeout := time.Duration(aiSearchTimeoutVal) * time.Second

	// logger
	writers := make([]io.Writer, 0)
	if logSTDOUT, _ := strconv.ParseBool(os.Getenv("LOG_TO_STDOUT")); logSTDOUT {
		writers = append(writers, os.Stdout)
	}

	if logFILE, _ := strconv.ParseBool(os.Getenv("LOG_TO_FILE")); logFILE {
		logMaxSize, _ := strconv.Atoi(os.Getenv("LOG_MAX_SIZE"))
		if logMaxSize == 0 {
			logMaxSize = 50 //default 50 megabytes
		}

		logFilename := os.Getenv("LOG_FILENAME")
		if logFilename == "" {
			logFilename = "server.log"
		}

		lg := &lumberjack.Logger{
			Filename:   logFilename,
			MaxSize:    logMaxSize,
			MaxBackups: 1,
			LocalTime:  true,
		}

		writers = append(writers, lg)
	}

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(io.MultiWriter(writers...))

	// set gin writer to logrus
	gin.DefaultWriter = logrus.StandardLogger().Writer()

	// init potgresql database (use sqlx + pgx stdlib)
	db, err := sqlx.ConnectContext(context.Background(), "pgx", os.Getenv("POSTGRES_URL"))
	if err != nil {
		logrus.Fatalf("failed to connect to postgres: %v", err)
	}
	// psqlPrep previously came from yureka_sql; use db directly
	psqlPrep := db

	// init repo
	repo := gormrepo.NewGormRepo(psqlPrep, logger.Default)

	// init storage repo
	storageRepo := s3repo.NewS3Repo()

	// init ai search repo
	aiRepo, err := aisearchrepo.NewAISearchRepository(os.Getenv("AI_API_URL"))
	if err != nil {
		logrus.Errorf("failed to init ai search repo: %v", err)
	}

	// init usecase
	ucMember := usecase_member.NewAppUsecase(usecase_member.RepoInjection{
		GormDbRepo:   repo,
		StorageRepo:  storageRepo,
		AISearchRepo: aiRepo,
	}, timeoutContext)

	// init job role usecase
	ucJobRole := usecase_job_role.NewAppUsecase(usecase_job_role.RepoInjection{
		GormDbRepo: repo,
	}, timeoutContext)

	// init job title usecase
	ucJobTitle := usecase_job_title.NewAppUsecase(usecase_job_title.RepoInjection{
		GormDbRepo: repo,
	}, timeoutContext)

	// init sector usecase
	ucSector := usecase_sector.NewAppUsecase(usecase_sector.RepoInjection{
		GormDbRepo: repo,
	}, timeoutContext)

	// init recruitment status usecase
	ucRecruitmentStatus := usecase_recruitment_status.NewAppUsecase(usecase_recruitment_status.RepoInjection{
		GormDbRepo: repo,
	}, timeoutContext)

	// init kabupaten kota usecase
	ucKabupatenKota := usecase_kabupaten_kota.NewAppUsecase(usecase_kabupaten_kota.RepoInjection{
		GormDbRepo: repo,
	}, timeoutContext)

	// init provinsi usecase
	ucProvinsi := usecase_provinsi.NewAppUsecase(usecase_provinsi.RepoInjection{
		GormDbRepo: repo,
	}, timeoutContext)

	// init bookmark usecase
	ucBookmark := usecase_bookmark.NewAppUsecase(usecase_bookmark.RepoInjection{
		GormDbRepo: repo,
	}, timeoutContext)

	// init ai chat usecase
	ucAIChat := usecase_aichat.NewAIChatUsecase(repo, timeoutContext)

	// init request usecase
	ucRequest := usecase_request.NewRequestAppUsecase(repo, timeoutContext)

	// init mq repo
	mqRepo, err := rabbitmqrepo.NewRabbitMQRepo(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		logrus.Errorf("failed to init rabbitmq: %v", err)
		// Don't fatal, just log, so app can start without rabbitmq if needed
		// or handle gracefully. For now, we proceed but publishing will fail.
	} else {
		defer mqRepo.Close()
	}

	// init cv usecase
	ucCV := usecase_cv.NewCVUsecase(repo, storageRepo, mqRepo, timeoutContext)

	// init activity log usecase
	ucActivityLog := usecase_activity_log.NewAppUsecase(usecase_activity_log.RepoInjection{
		GormDbRepo: repo,
	}, timeoutContext)

	// init system setting usecase
	ucSystemSetting := usecase_system_setting.NewAppUsecase(usecase_system_setting.RepoInjection{
		GormDbRepo: repo,
	}, timeoutContext)

	// start consumer
	if mqRepo != nil {
		cvConsumer := consumer.NewCVParserConsumer(mqRepo, repo)
		go func() {
			if err := cvConsumer.Start(context.Background()); err != nil {
				logrus.Errorf("CVParserConsumer exited with error: %v", err)
			}
		}()
	}

	// init middleware — pass nil redis client and the actual gorm repo
	mdl := middleware.NewMiddleware(nil, repo)

	// gin mode realease when go env is production
	if os.Getenv("GO_ENV") == "production" || os.Getenv("GO_ENV") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// init gin
	ginEngine := gin.New()

	// add exception handler
	ginEngine.Use(mdl.Recovery())

	// add logger
	ginEngine.Use(mdl.Logger(io.MultiWriter(writers...)))

	// cors
	ginEngine.Use(mdl.Cors())

	// Context Enricher (Real IP, User-Agent, etc.)
	ginEngine.Use(mdl.ContextEnricher())

	// default route
	ginEngine.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, map[string]any{
			"message": "It works",
		})
	})

	// swagger route
	ginEngine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// init route
	apiGroup := ginEngine.Group("/api")
	http_member.NewRouteHandler(apiGroup, mdl, ucMember)
	http_cv.NewCVHandler(apiGroup, mdl, ucCV)
	http_job_role.NewJobRoleHandler(apiGroup, mdl, ucJobRole)
	httpdelivery_request.NewRequestHandler(apiGroup, mdl, ucRequest)
	http_job_title.NewJobTitleHandler(apiGroup, mdl, ucJobTitle)
	http_sector.NewSectorHandler(apiGroup, mdl, ucSector)
	http_kabupaten_kota.NewKabupatenKotaHandler(apiGroup, ucKabupatenKota)
	http_provinsi.NewProvinsiHandler(apiGroup, ucProvinsi)
	http_recruitment_status.NewRecruitmentStatusHandler(apiGroup, mdl, ucRecruitmentStatus)
	http_bookmark.NewBookmarkHandler(apiGroup, mdl, ucBookmark)
	http_aichat.NewAIChatHandler(apiGroup, mdl, ucAIChat)
	http_activity_log.NewActivityLogHandler(apiGroup, mdl, ucActivityLog)
	http_system_setting.NewSystemSettingHandler(apiGroup, mdl, ucSystemSetting)

	// init search (AI)
	if aiRepo != nil {
		// Use a dedicated timeout for AI Search as it involves slow LLM evaluations
		ucSearch := usecase_search.NewSearchUsecase(aiRepo, repo, aiSearchTimeout)
		http_search.NewSearchHandler(apiGroup, mdl, ucSearch)
	}

	// init grpc handler
	candidateGrpcHandler := delivery_grpc.NewCandidateHandler(repo)

	// start grpc server
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50052"
	}
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		logrus.Errorf("failed to listen for grpc: %v", err)
	} else {
		s := grpc.NewServer()
		pb.RegisterCandidateServiceServer(s, candidateGrpcHandler)
		logrus.Infof("gRPC server running on port %s", grpcPort)
		go func() {
			if err := s.Serve(lis); err != nil {
				logrus.Errorf("failed to serve grpc: %v", err)
			}
		}()
	}

	port := os.Getenv("PORT")

	logrus.Infof("Service running on port %s", port)
	ginEngine.Run(":" + port)
}
