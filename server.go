package calypso

import (
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/workfoxes/calypso/pkg/client/db"
	"gorm.io/gorm"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	_logger "github.com/sirupsen/logrus"
	"go.uber.org/dig"

	"github.com/workfoxes/calypso/pkg/config"
	"github.com/workfoxes/calypso/pkg/log"
)

type ApplicationConfig struct {
	Name string
	Port int
}

// New : Will create New Server the Need as default for the Workfoxes Application
// 		 Also will add all the default provider to this server
func New(config *ApplicationConfig) *ApplicationServer {
	app := fiber.New()
	_server := &ApplicationServer{Name: config.Name, Port: config.Port, Server: app, container: dig.New()}
	DefaultProviders(_server)
	return _server
}

// DefaultProviders : will provide all the default provider in the server start
func DefaultProviders(app *ApplicationServer) {
	app.AddProvider(config.GetConfig)
	app.AddProvider(log.Init)
	app.AddProvider(db.Init)
	app.Invoker(func(l *_logger.Logger) {
		log.Info("logger is setup")
	})
	app.Invoker(func(_config *config.Config) {
		config.C = _config
	})
	app.Invoker(func(DB *gorm.DB) {
		db.DB = DB
	})
}

// AddProvider : This will add new provider to the server container
func (app *ApplicationServer) AddProvider(constructor interface{}, opts ...dig.ProvideOption) {
	err := app.container.Provide(constructor, opts...)
	if err != nil {
		panic(err)
	}
}

// Invoker : This will add new provider to the server container
func (app *ApplicationServer) Invoker(function interface{}, opts ...dig.ProvideOption) {
	err := app.container.Invoke(function)
	if err != nil {
		panic(err)
	}
}

// ApplicationServer : Application server will hold the service object for the application
type ApplicationServer struct {
	Server    *fiber.App
	Name      string
	Port      int
	container *dig.Container
	// config    *config.Config
}

// CreateAppServer : func to create Application server object to Manage the application server
func CreateAppServer(Name string, Port int) *ApplicationServer {
	app := fiber.New()
	_server := &ApplicationServer{Name: Name, Port: Port, Server: app, container: dig.New()}
	return _server
}

// LoadDefaultMiddleware : this function will load all the middleware that are need for application
func (app *ApplicationServer) LoadDefaultMiddleware() {
	app.Use(logger.New())
	app.Use(limiter.New())
	app.Use(etag.New())
	app.Use(csrf.New())
	app.Use(pprof.New())
	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(compress.New(compress.Config{Level: compress.LevelBestCompression}))
	app.Use(cors.New(cors.Config{
		Next:             nil,
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowCredentials: false,
		ExposeHeaders:    "",
		MaxAge:           0,
	}))

}

// Use : This function will allow us to add the middleware into the web application
func (app *ApplicationServer) Use(args ...interface{}) {
	app.Server.Use(args...)
}

// Start : Will Start the Application service for the Calypso
func (app *ApplicationServer) Start() {
	_port := strconv.Itoa(app.Port)
	err := app.Server.Listen(":" + _port)
	log.Debug(err.Error())
}
