package geocoder

import (
	"context"
	"errors"
	"strings"

	"github.com/kpawlik/geocode_server/pkg/config"
	log "github.com/sirupsen/logrus"

	"googlemaps.github.io/maps"
)

const (
	queryLimitPrefix = "maps: OVER_QUERY_LIMIT"
)

var (
	// ErrUnableToGeocode address cannot be geocoded
	ErrUnableToGeocode = errors.New("UNABLE_TO_GEOCODE")
	// ErrGoogleLimit limit query from Google API
	ErrGoogleLimit = errors.New("GOOGLE_OVER_QUERY_LIMIT")
	// ErrQuotaLimit internal query limit
	ErrQuotaLimit = errors.New("SERVER_QUERY_LIMIT")
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
	Error string
	Request
}

func newResponse(req Request) Response {
	return Response{Lat: 0.0, Lng: 0.0, Error: "", Request: req}
}

// Geocoder geocoder struct
type Geocoder struct {
	client *maps.Client
	cfg    *config.Config
	Err    error
}

// NewGeocoder returns a new Geocoder instance
func NewGeocoder(cfg *config.Config) *Geocoder {
	c, err := NewClient(cfg)
	g := &Geocoder{cfg: cfg,
		client: c,
		Err:    err}
	return g
}

// Geocode request
func (g *Geocoder) Geocode(req Request) (resp Response) {
	if !g.isAviableQuota() {
		resp = newResponse(req)
		resp.Error = ErrQuotaLimit.Error()
		return
	}
	g.IncQuota()
	resp = Geocode(g.client, req)
	return
}

//IncQuota increment used quota
func (g *Geocoder) IncQuota() {
	g.cfg.IncQuota()
}

func (g *Geocoder) isAviableQuota() bool {
	return g.cfg.IsAviableQuota()
}

//Geocode address
func Geocode(c *maps.Client, req Request) (resp Response) {
	var (
		gResp []maps.GeocodingResult
		err   error
	)
	resp = newResponse(req)
	log.Debugf("Request from request channel: %s", req)
	gReq := &maps.GeocodingRequest{
		Address: req.Address,
	}
	gResp, err = c.Geocode(context.Background(), gReq)
	if err != nil {
		log.Errorf("Error geocoding address '%s' (%s). %v", req.Address, req.ID, err)
		if IsGoogleOverQueryLimit(err) {
			resp.Error = ErrGoogleLimit.Error()
		} else {
			resp.Error = err.Error()
		}
		return
	}
	if len(gResp) == 0 {
		log.Errorf("Address '%s' (%s) could not be geocoded", req.Address, req.ID)
		resp.Error = ErrUnableToGeocode.Error()
		return
	}
	resp.Lat = gResp[0].Geometry.Location.Lat
	resp.Lng = gResp[0].Geometry.Location.Lng
	log.Debug("Address '%s' (%s) geocoded {%f, %f}", resp.Address, resp.ID, resp.Lat, resp.Lng)
	return
}

func worker(reqChan chan Request, respChan chan Response, closeChan chan bool, g *Geocoder) {
	log.Debugf("Creating worker")
	for {
		select {
		case req := <-reqChan:
			resp := g.Geocode(req)
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
	auth := cfg.Geocoder
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
	responses := make(chan Response, total)
	requests := make(chan Request, total)
	close := make(chan bool)
	return requests, responses, close
}

// StartWorkers start geocoding workers
func StartWorkers(c *Geocoder, n int, total int) (chan Request, chan Response, chan bool) {
	requests, responses, close := Channels(total)
	for i := 0; i < n; i++ {
		go worker(requests, responses, close, c)
	}
	return requests, responses, close
}

// IsGoogleOverQueryLimit check if this is google limit error
func IsGoogleOverQueryLimit(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), queryLimitPrefix)
}
