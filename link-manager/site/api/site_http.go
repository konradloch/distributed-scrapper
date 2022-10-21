package api

import (
	"github.com/konradloch/distributed-scrapper/link-manager/site/usecases"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

type HttpServer struct {
	e       *echo.Echo
	service *usecases.Service
}

func NewHttpServer(service *usecases.Service) *HttpServer {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	return &HttpServer{e: e, service: service}
}

func (h *HttpServer) StartServer() {
	h.e.POST("/start", h.PostSite)
	h.e.Logger.Fatal(h.e.Start(":9998"))
}

func (h *HttpServer) PostSite(c echo.Context) error {
	err := h.service.Publish(c.FormValue("url"))
	if err != nil {
		return err
	}
	return c.String(http.StatusOK, "ok")
}
