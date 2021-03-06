package hueserver

import (
	"encoding/json"
	"net/http"
	"text/template"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var setupTemplate = template.Must(template.New("setup").Parse(`<?xml version="1.0"?>
<root xmlns="urn:schemas-upnp-org:device-1-0">
  <specVersion>
    <major>1</major>
    <minor>0</minor>
  </specVersion>
  <URLBase>http://{{.URLBase}}/</URLBase>
  <device>
    <deviceType>urn:schemas-upnp-org:device:Basic:1</deviceType>
    <friendlyName>{{.FriendlyName}}</friendlyName>
    <manufacturer>Royal Philips Electronics</manufacturer>
    <modelName>Philips hue bridge 2012</modelName>
    <modelNumber>929000226503</modelNumber>
    <UDN>uuid:{{.UUID}}</UDN>
  </device>
</root>`))

type LightState struct {
	On        bool      `json:"on"`
	Bri       int       `json:"bri"`
	Hue       int       `json:"hue"`
	Sat       int       `json:"sat"`
	Effect    string    `json:"effect"`
	Ct        int       `json:"ct"`
	Alert     string    `json:"alert"`
	Colormode string    `json:"colormode"`
	Reachable bool      `json:"reachable"`
	XY        []float64 `json:"xy"`
}

type LightStateChange struct {
	On             *bool   `json:"on,omitempty"`
	Bri            *int    `json:"bri,omitempty"`
	Hue            *int    `json:"hue,omitempty"`
	Sat            *int    `json:"sat,omitempty"`
	Effect         *string `json:"effect,omitempty"`
	Ct             *int    `json:"ct,omitempty"`
	Alert          *string `json:"alert,omitempty"`
	Colormode      *string `json:"colormode,omitempty"`
	TransitionTime int     `json:"transitiontime,omitempty"`
}

type LightStateChangeResponse []struct {
	Success map[string]interface{} `json:"success,omitempty"`
}

type Light struct {
	State            LightState `json:"state"`
	Type             string     `json:"type"`
	Name             string     `json:"name"`
	ModelID          string     `json:"modelid"`
	ManufacturerName string     `json:"manufacturername"`
	UniqueID         string     `json:"uniqueid"`
	SwVersion        string     `json:"swversion"`
	PointSymbol      struct {
		One   string `json:"1"`
		Two   string `json:"2"`
		Three string `json:"3"`
		Four  string `json:"4"`
		Five  string `json:"5"`
		Six   string `json:"6"`
		Seven string `json:"7"`
		Eight string `json:"8"`
	} `json:"pointsymbol"`
}

type LightList map[string]Light

type Server struct {
	*echo.Echo
	UUID          string
	FriendlyName  string
	URLBase       string
	GetLights     func() LightList
	GetLight      func(id string) Light
	SetLightState func(id string, state LightStateChange) LightStateChangeResponse
}

func (server *Server) Start(port string) error {
	return http.ListenAndServe(port, server)
}

func (server *Server) serveSetupXML(c echo.Context) error {
	return setupTemplate.Execute(c.Response().Writer(), server)
}

func (server *Server) getLights(c echo.Context) error {
	return c.JSON(http.StatusOK, server.GetLights())
}

func (server *Server) getLight(c echo.Context) error {
	return c.JSON(http.StatusOK, server.GetLight(c.Param("lightId")))
}

func (server *Server) setLightState(c echo.Context) error {
	defer c.Request().Body.Close()
	decoder := json.NewDecoder(c.Request().Body)
	state := &LightStateChange{}
	if err := decoder.Decode(state); err != nil {
		return err
	}
	lightID := c.Param("lightId")
	stateChangeResponse := server.SetLightState(lightID, *state)

	return c.JSON(http.StatusOK, stateChangeResponse)
}

func NewServer(uuid, urlBase, friendlyName string, getLights func() LightList, getLight func(id string) Light, setLightState func(id string, state LightStateChange) LightStateChangeResponse) (srv *Server) {
	srv = &Server{
		Echo:          echo.New(),
		UUID:          uuid,
		FriendlyName:  friendlyName,
		URLBase:       urlBase,
		GetLights:     getLights,
		GetLight:      getLight,
		SetLightState: setLightState,
	}

	srv.Use(middleware.Logger())
	srv.GET("/upnp/setup.xml", srv.serveSetupXML)
	srv.GET("/api/:userId", srv.getLights)
	srv.GET("/api/:userId/lights", srv.getLights)
	srv.PUT("/api/:userId/lights/:lightId/state", srv.setLightState)
	srv.GET("/api/:userId/lights/:lightId", srv.getLight)

	return
}
