package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/kpawlik/geocode_server/pkg/config"
	"github.com/kpawlik/geocode_server/pkg/geocoder"
	log "github.com/sirupsen/logrus"
	"googlemaps.github.io/maps"
)

var (

	// ErrQuotaLimit internal query limit
	ErrQuotaLimit = errors.New("SERVER_QUERY_LIMIT")
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
		err               error
		gRequests         []geocoder.Request
		noOfRequests      int
		geocodedAddresses []geocodeAddressResponse
	)
	// Check internal quota limit
	if !cfg.CheckQuotaLimit() {
		log.Info("Quota limit exceeded")
		resp.Error = ErrQuotaLimit.Error()
		return
	}
	// Run geocoding

	if gRequests, err = prepareRequests(request.Body); err != nil {
		log.Error(err)
		resp.Error = fmt.Sprintf("Error decoding request body, %v", err)
		return
	}
	if noOfRequests = len(gRequests); noOfRequests == 0 {
		log.Info("No requests to geocode")
		return
	}
	// Check first address to make sure that we didn't exceeded Google Query Limit
	first := gRequests[0]
	gRequests = gRequests[1:]
	cfg.IncQuota()
	geocodedAddresses, err = checkFirstAddress(c, first)
	if err != nil {
		resp.Error = err.Error()
		return
	}
	// if ok process rest
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
				Error:   ErrQuotaLimit.Error(),
			}
			geocodedAddresses = append(geocodedAddresses, a)
		}
	}
	// collect results
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
	// send signal to close all workers
	close(closeCh)
	resp.Addresses = geocodedAddresses
	log.Infof("Remaining quota: %d", cfg.GetRemainingQuota())
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
	config.StartQuotaTimer(cfg)
	http.HandleFunc("/geocode", newGeocodeHandler(cfg, c))
	err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil)
	return

}

func checkFirstAddress(c *maps.Client, request geocoder.Request) (responses []geocodeAddressResponse, err error) {
	responses = []geocodeAddressResponse{}
	response := geocoder.Geocode(c, request)
	if response.Error != nil && errors.Is(response.Error, geocoder.ErrGoogleLimit) {
		err = geocoder.ErrGoogleLimit
		return
	}
	a := geocodeAddressResponse{
		ID:      response.ID,
		Address: response.Address,
		Lat:     response.Lat,
		Lng:     response.Lng,
	}
	if response.Error != nil {
		a.Error = response.Error.Error()
	}
	responses = append(responses, a)

	return
}
