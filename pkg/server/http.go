package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kpawlik/geocode_server/pkg/config"
	"github.com/kpawlik/geocode_server/pkg/geocoder"
	log "github.com/sirupsen/logrus"
	"googlemaps.github.io/maps"
)

const (
	QuotaLimit = "QuotaLimit"
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

func newGeocodeHandler(cfg *config.Config, c *maps.Client) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			bytes []byte
			err   error
		)
		w.Header().Add("Content-Type", "application/json")
		resp := geocode(w, r, cfg, c)
		if bytes, err = json.Marshal(resp); err != nil {
			bytes = []byte(fmt.Sprintf("Error encoding response, %v", err))
		}
		w.Write(bytes)
	})
}

func prepareRequests(data io.Reader) (requests []geocoder.Request, err error) {
	decoder := json.NewDecoder(data)
	gr := geocodeRequest{}
	if err = decoder.Decode(&gr); err != nil {
		return
	}
	requests = make([]geocoder.Request, len(gr.Addresses))
	for i, req := range gr.Addresses {
		gRequest := geocoder.Request{ID: req.ID,
			Address: req.Address}
		requests[i] = gRequest
	}
	return

}

func geocode(rw http.ResponseWriter, request *http.Request, cfg *config.Config, c *maps.Client) (resp geocodeResponse) {
	var (
		err       error
		gRequests []geocoder.Request
	)
	if !cfg.CheckQuotaLimit() {
		log.Info("Quota limit exceeded")
		resp.Error = QuotaLimit
		return
	}
	geocodedAddresses := []geocodeAddressResponse{}
	if gRequests, err = prepareRequests(request.Body); err != nil {
		log.Error(err)
		resp.Error = fmt.Sprintf("Error decoding request body, %v", err)
		return
	}
	noOfRequests := len(gRequests)

	reqCh, respCh, closeCh := geocoder.StartWorkers(c, cfg.WorkersNumber, noOfRequests)
	_, _, _ = reqCh, respCh, closeCh
	noOfResponses := 0
	for _, request := range gRequests {
		if cfg.CheckQuotaLimit() {
			reqCh <- request
			cfg.IncQuota()
			noOfResponses++
		} else {
			a := geocodeAddressResponse{
				ID:      request.ID,
				Address: request.Address,
				Error:   QuotaLimit,
			}
			geocodedAddresses = append(geocodedAddresses, a)
		}
	}
	// if gResponses, err = geocoder.Geocode(gRequests, cfg, logger); err != nil {
	// 	logger.Error(err)
	// 	resp.Error = fmt.Sprintf("Error geocoding , %v", err)
	// 	return
	// }

	for i := 0; i < noOfResponses; i++ {
		res := <-respCh

		a := geocodeAddressResponse{
			ID:      res.ID,
			Address: res.Address,
			Lat:     res.Lat,
			Lng:     res.Lng,
		}
		if res.Error != nil {
			a.Error = res.Error.Error()
		}
		geocodedAddresses = append(geocodedAddresses, a)
	}
	close(closeCh)

	resp.Addresses = geocodedAddresses
	return
}

// Serve serve http server
func Serve(cfg *config.Config) (err error) {
	var (
		c *maps.Client
	)
	if c, err = geocoder.NewClient(cfg); err != nil {
		return
	}
	http.HandleFunc("/geocode", newGeocodeHandler(cfg, c))
	err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil)
	return

}
