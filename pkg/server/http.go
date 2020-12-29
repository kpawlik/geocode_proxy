package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kpawlik/geocode_server/pkg/config"
	"github.com/kpawlik/geocode_server/pkg/geocoder"
	log "github.com/sirupsen/logrus"
)

type address struct {
	ID      string `json:"id"`
	Address string `json:"address"`
}

type request struct {
	Addresses []address `json:"addresses"`
}

type geocodedAddress struct {
	ID      string  `json:"id"`
	Address string  `json:"address"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Error   string  `json:"error"`
}

type response struct {
	Addresses []geocodedAddress `json:"addresses"`
	Error     string            `json:"error"`
}

func newGeocodeHandler(cfg *config.Config, c *geocoder.Geocoder) http.HandlerFunc {
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
	gr := request{}
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

func geocode(rw http.ResponseWriter, request *http.Request, cfg *config.Config, c *geocoder.Geocoder) (resp response) {
	var (
		err               error
		requests          []geocoder.Request
		noOfRequests      int
		geocodedAddresses []geocodedAddress
	)
	if requests, err = prepareRequests(request.Body); err != nil {
		log.Error(err)
		resp.Error = fmt.Sprintf("Error decoding request body, %v", err)
		return
	}
	noOfRequests = len(requests)
	reqCh, respCh, closeCh := geocoder.StartWorkers(c, cfg.WorkersNumber, noOfRequests)
	// no of responses to collect from channel

	for _, request := range requests {
		reqCh <- request
	}
	log.Debugf("Wait for all responses (%d)", noOfRequests)
	// collect results from channel/workers
	for i := 0; i < noOfRequests; i++ {
		res := <-respCh
		a := geocodedAddress{
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
		c *geocoder.Geocoder
	)
	if c = geocoder.NewGeocoder(cfg); c.Err != nil {
		err = c.Err
		return
	}
	config.StartQuotaTimer(cfg)
	http.HandleFunc("/geocode", newGeocodeHandler(cfg, c))
	err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), nil)
	return
}
