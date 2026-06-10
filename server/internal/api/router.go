package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/user/can-server/config"
	"github.com/user/can-server/internal/api/handler"
	"github.com/user/can-server/internal/api/middleware"
	"github.com/user/can-server/internal/api/ws"
	"github.com/user/can-server/internal/db/repository"
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
}

func (s *Server) Start() error {
	canFrames := s.registerRoutes()
	s.mqttSubscriber = mqttsub.NewSubscriber(s.cfg.MQTT, canFrames)
	if err := s.mqttSubscriber.Start(); err != nil {
		log.Printf("mqtt subscriber start failed: %v", err)
	}
	return s.engine.Run(s.cfg.ServerAddr())
}

func (s *Server) registerRoutes() *service.CANFrameService {
	s.engine.Use(middleware.CORS())

	svc := service.NewDeviceService(s.cfg)
	h := handler.NewDeviceHandler(svc)
	tcpSvc := service.NewTCPConfigService(&repository.TCPConfigRepo{})
	tcpHandler := handler.NewTCPConfigHandler(tcpSvc)
	canHub := ws.NewHub()
	canFrameSvc := service.NewCANFrameService(&repository.LogRepo{}, canHub)
	canFrameHandler := handler.NewCANFrameHandler(canFrameSvc)

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
	}
	return canFrameSvc
}
