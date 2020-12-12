package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kpawlik/geocode_server/pkg/config"
	"github.com/kpawlik/geocode_server/pkg/geocoder"
	"github.com/sirupsen/logrus"
)

var (
	logger = logrus.New()
)

type geocodeAddressRequest struct {
	ID      string `json:"id"`
	Address string `json:"address"`
}

type geocodeRequest struct {
	Addresses []geocodeAddressRequest `json:"addresses"`
}

type geocodeAddressResponse struct {
	ID      string  `json:"id"`
	Address string  `json:"address"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Error   string  `json:"error"`
}

type geocodeResponse struct {
	Addresses []geocodeAddressResponse `json:"addresses"`
	Error     string                   `json:"error"`
}

func newGeocodeHandler(cfg config.Config) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			bytes []byte
			err   error
		)
		w.Header().Add("Content-Type", "application/json")
		resp := geocode(w, r, cfg)
		if bytes, err = json.Marshal(resp); err != nil {
			bytes = []byte(fmt.Sprintf("Error encoding response, %v", err))
		}
		w.Write(bytes)
	})
}

func geocode(rw http.ResponseWriter, request *http.Request, cfg config.Config) (resp geocodeResponse) {
	var (
		err        error
		gRequests  []geocoder.Request
		gResponses []geocoder.Response
	)

	decoder := json.NewDecoder(request.Body)

	gr := geocodeRequest{}
	if err = decoder.Decode(&gr); err != nil {
		resp.Error = fmt.Sprintf("Error decoding request body, %v", err)
		return
	}

	gRequests = make([]geocoder.Request, len(gr.Addresses))
	for i, req := range gr.Addresses {
		gRequest := geocoder.Request{ID: req.ID,
			Address: req.Address}
		gRequests[i] = gRequest
	}
	if gResponses, err = geocoder.Geocode(gRequests, cfg, logger); err != nil {
		resp.Error = fmt.Sprintf("Error geocoding , %v", err)
		return
	}
	geocodedAddresses := []geocodeAddressResponse{}
	for _, gResp := range gResponses {

		a := geocodeAddressResponse{
			ID:      gResp.ID,
			Address: gResp.Address,
			Lat:     gResp.Lat,
			Lng:     gResp.Lng,
		}
		if gResp.Error != nil {
			a.Error = gResp.Error.Error()
		}
		geocodedAddresses = append(geocodedAddresses, a)
	}
	resp.Addresses = geocodedAddresses
	return
}

func setLogLevel(cfg config.Config) {
	switch cfg.LogLevel {

	case "warn":
		logger.SetLevel(logrus.WarnLevel)
		break
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
		break
	default:
		logger.SetLevel(logrus.InfoLevel)
		break
	}
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: false,
	})
}

// Serve serve http server
func Serve(cfg config.Config) (err error) {
	setLogLevel(cfg)
	http.HandleFunc("/geocode", newGeocodeHandler(cfg))
	err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil)
	return

}
