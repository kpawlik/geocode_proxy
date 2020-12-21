package geocoder

import (
	"context"
	"fmt"

	"github.com/kpawlik/geocode_server/pkg/config"
	log "github.com/sirupsen/logrus"

	"googlemaps.github.io/maps"
)

const (
	workersCount = 20
)

// Request to geocoder
type Request struct {
	ID      string
	Address string
}

// Response from geocoder
type Response struct {
	Lat   float64
	Lng   float64
	Error error
	Request
}

func newResponse(req Request) Response {
	return Response{0.0, 0.0, nil, req}
}

func worker(reqChan chan Request, respChan chan Response, closeChan chan bool, gc *maps.Client) {
	log.Debugf("Creating worker")
	for {
		select {
		case req := <-reqChan:

			resp := newResponse(req)
			log.Debugf("Request from request channel: %s", req)
			gReq := &maps.GeocodingRequest{
				Address: req.Address,
			}
			gResp, err := gc.Geocode(context.Background(), gReq)
			if err != nil {
				resp.Error = fmt.Errorf("%v. Error geocoding address '%s' (%s)", err, req.Address, req.ID)
				respChan <- resp
				continue
			}

			if len(gResp) == 0 {
				resp.Error = fmt.Errorf("Address '%s' (%s) could not be geocoded", req.Address, req.ID)
				respChan <- resp
				continue
			}
			resp.Lat = gResp[0].Geometry.Location.Lat
			resp.Lng = gResp[0].Geometry.Location.Lng
			log.Infof("Address '%s' (%s) geocoded {%f, %f}", resp.Address, resp.ID, resp.Lat, resp.Lng)
			respChan <- resp
		case _, close := <-closeChan:
			if !close {
				log.Debugf("Closing worker")
				return
			}

		}
	}
}

// NewClient returns new Google API client
func NewClient(cfg *config.Config) (client *maps.Client, err error) {
	options := make([]maps.ClientOption, 0)
	auth := cfg.Authentication
	if len(auth.ClientID) > 0 && len(auth.ClientSecret) > 0 {
		options = append(options, maps.WithClientIDAndSignature(auth.ClientID, auth.ClientSecret))
	}
	if len(auth.APIKey) > 0 {
		options = append(options, maps.WithAPIKey(auth.APIKey))
	}
	if len(auth.ClientID) > 0 && len(auth.ClientSecret) > 0 {
		options = append(options, maps.WithAPIKeyAndSignature(auth.APIKey, auth.ClientSecret))
	}
	if len(auth.Channel) > 0 {
		options = append(options, maps.WithChannel(auth.Channel))
	}
	client, err = maps.NewClient(options...)
	return

}

// Channels returns channels required to comunication
func Channels(total int) (chan Request, chan Response, chan bool) {
	responses := make(chan Response, 5)
	requests := make(chan Request, total)
	close := make(chan bool)
	return requests, responses, close
}

// StartWorkers start geocoding workers
func StartWorkers(c *maps.Client, n int, total int) (chan Request, chan Response, chan bool) {
	requests, responses, close := Channels(total)
	for i := 0; i < n; i++ {
		go worker(requests, responses, close, c)
	}
	return requests, responses, close
}

// // Geocode addesses
// func Geocode(requests []Request, cfg *config.Config) (resposes []Response, err error) {
// 	//options := maps.WithClientIDAndSignature(cfg.ClientID, cfg.ClientSecret)
// 	var (
// 		gc *maps.Client
// 	)
// 	total := len(requests)
// 	cn := make(chan Response, 5)
// 	ca := make(chan Request, total)
// 	cc := make(chan bool)
// 	for i := 0; i < cfg.WorkersNumber; i++ {
// 		if gc, err = NewClient(cfg); err != nil {
// 			err = fmt.Errorf("Error creating Maps.Client, %v", err)
// 			return
// 		}
// 		go worker(ca, cn, cc, gc)
// 	}
// 	for _, request := range requests {
// 		ca <- request
// 	}
// 	fails := 0
// 	for i := 0; i < total; i++ {
// 		res := <-cn
// 		resposes = append(resposes, res)
// 		if res.Error != nil {
// 			fails++
// 			log.Error(res.Error)
// 		}
// 	}
// 	log.Infof("Set of %d addresses processed. Success: %d, Fails: %d", total, total-fails, fails)
// 	close(cc)
// 	return
// }
