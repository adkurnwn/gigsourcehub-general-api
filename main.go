package main

import (
	http_cv "github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/cv"
	http_member "github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/member"
	"github.com/adkurnwn/gigsourcehub-general-api/app/delivery/http/middleware"
	gormrepo "github.com/adkurnwn/gigsourcehub-general-api/app/repository/gorm"
	rabbitmqrepo "github.com/adkurnwn/gigsourcehub-general-api/app/repository/rabbitmq"
	s3repo "github.com/adkurnwn/gigsourcehub-general-api/app/repository/s3"
	usecase_cv "github.com/adkurnwn/gigsourcehub-general-api/app/usecase/cv"
	usecase_member "github.com/adkurnwn/gigsourcehub-general-api/app/usecase/member"
	"github.com/adkurnwn/gigsourcehub-general-api/docs"

	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

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
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%s", os.Getenv("HOST"), os.Getenv("PORT"))
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	timeoutStr := os.Getenv("TIMEOUT")
	if timeoutStr == "" {
		timeoutStr = "5"
	}
	timeout, _ := strconv.Atoi(timeoutStr)
	timeoutContext := time.Duration(timeout) * time.Second

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

	// init usecase
	ucMember := usecase_member.NewAppUsecase(usecase_member.RepoInjection{
		GormDbRepo: repo,
	}, timeoutContext)

	// init storage repo
	storageRepo := s3repo.NewS3Repo()

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

	// init middleware — pass nil redis client
	mdl := middleware.NewMiddleware(nil)

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

	// default route
	ginEngine.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, map[string]any{
			"message": "It works",
		})
	})

	// swagger route
	ginEngine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// init route
	http_member.NewRouteHandler(ginEngine.Group(""), mdl, ucMember)
	http_cv.NewCVHandler(ginEngine.Group(""), mdl, ucCV)

	port := os.Getenv("PORT")

	logrus.Infof("Service running on port %s", port)
	ginEngine.Run(":" + port)
}
