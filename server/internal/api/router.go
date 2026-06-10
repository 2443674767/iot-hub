package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/user/can-server/config"
	"github.com/user/can-server/internal/api/handler"
	"github.com/user/can-server/internal/api/middleware"
	"github.com/user/can-server/internal/api/ws"
	"github.com/user/can-server/internal/db/repository"
	"github.com/user/can-server/internal/influx"
	mqttsub "github.com/user/can-server/internal/mqtt"
	"github.com/user/can-server/internal/service"
)

func NewServer(cfg *config.Config) *Server {
	return &Server{
		cfg:    cfg,
		engine: gin.Default(),
	}
}

type Server struct {
	cfg            *config.Config
	engine         *gin.Engine
	mqttSubscriber *mqttsub.Subscriber
	influxWriter   *influx.Writer
}

func (s *Server) Start() error {
	canFrames, iotSvc := s.registerRoutes()
	s.mqttSubscriber = mqttsub.NewSubscriber(s.cfg.MQTT, canFrames, iotSvc)
	if err := s.mqttSubscriber.Start(); err != nil {
		log.Printf("mqtt subscriber start failed: %v", err)
	}
	return s.engine.Run(s.cfg.ServerAddr())
}

func (s *Server) registerRoutes() (*service.CANFrameService, *service.IoTService) {
	s.engine.Use(middleware.CORS())

	s.influxWriter = influx.NewWriter(s.cfg.InfluxDB)
	svc := service.NewDeviceService(s.cfg)
	h := handler.NewDeviceHandler(svc)
	tcpSvc := service.NewTCPConfigService(&repository.TCPConfigRepo{})
	tcpHandler := handler.NewTCPConfigHandler(tcpSvc)
	canHub := ws.NewHub()
	canFrameSvc := service.NewCANFrameService(&repository.LogRepo{}, canHub)
	canFrameHandler := handler.NewCANFrameHandler(canFrameSvc)
	iotSvc := service.NewIoTService(&repository.IoTHostRepo{}, &repository.IoTChannelRepo{}, &repository.IoTChannelDataRepo{}, s.influxWriter)
	iotHandler := handler.NewIoTHandler(iotSvc)

	api := s.engine.Group("/api/v1")
	{
		api.GET("/devices", h.ListDevices)
		api.GET("/devices/:id/data", h.GetDeviceData)
		api.POST("/devices/:id/command", h.SendCommand)
		api.POST("/can/frames", canFrameHandler.Receive)
		api.GET("/ws/can", canHub.Handle)
		api.GET("/tcp-configs", tcpHandler.List)
		api.POST("/tcp-configs", tcpHandler.Create)
		api.PUT("/tcp-configs/:id", tcpHandler.Update)
		api.DELETE("/tcp-configs/:id", tcpHandler.Delete)
		api.GET("/iot/hosts", iotHandler.ListHosts)
		api.POST("/iot/hosts", iotHandler.CreateHost)
		api.PUT("/iot/hosts/:id", iotHandler.UpdateHost)
		api.DELETE("/iot/hosts/:id", iotHandler.DeleteHost)
		api.GET("/iot/channels", iotHandler.ListChannels)
		api.POST("/iot/channels", iotHandler.CreateChannel)
		api.PUT("/iot/channels/:id", iotHandler.UpdateChannel)
		api.DELETE("/iot/channels/:id", iotHandler.DeleteChannel)
		api.GET("/iot/channel-data", iotHandler.ListChannelData)
	}
	return canFrameSvc, iotSvc
}
