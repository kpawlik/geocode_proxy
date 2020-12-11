package geocoder

import (
	"context"
	"fmt"

	"github.com/kpawlik/geocode_server/pkg/config"
	"googlemaps.github.io/maps"
)

const (
	workersCount = 20
)

type Request struct {
	ID      string
	Address string
}
type Response struct {
	Lat   float64
	Lng   float64
	Error error
	Request
}

func newResponse(req Request) Response {
	return Response{0.0, 0.0, nil, req}
}

func worker(reqChan chan Request, respChan chan Response, gc *maps.Client) {
	for {
		req := <-reqChan
		resp := newResponse(req)
		// fmt.Println("Address from ca: ", address)

		gReq := &maps.GeocodingRequest{
			Address: req.Address,
		}
		gResp, err := gc.Geocode(context.Background(), gReq)
		if err != nil {
			resp.Error = err
			respChan <- resp
			continue
		}
		if len(gResp) == 0 {
			resp.Error = fmt.Errorf("Address could not be geocoded")
			respChan <- resp
			continue
		}
		resp.Lat = gResp[0].Geometry.Location.Lat
		resp.Lng = gResp[0].Geometry.Location.Lng

		respChan <- resp
	}
}

// Geocode addesses
func Geocode(requests []Request, cfg config.Config) (resposes []Response, err error) {

	//options := maps.WithClientIDAndSignature(cfg.ClientID, cfg.ClientSecret)
	var (
		c *maps.Client
	)
	c, err = maps.NewClient(maps.WithAPIKey(cfg.APIKey))
	if err != nil {
		err = fmt.Errorf("Error creating Maps.Client, %v", err)
		return
	}

	cn := make(chan Response, 5)
	ca := make(chan Request, len(requests))
	cnt := 0
	for i := 0; i < cfg.WorkersNumber; i++ {
		go worker(ca, cn, c)
	}
	for _, request := range requests {
		if len(request.Address) == 0 {
			continue
		}
		cnt++
		ca <- request
	}
	fails := 0
	for i := 0; i < cnt; i++ {
		res := <-cn
		resposes = append(resposes, res)
		if res.Error != nil {
			fails++
		}
	}
	fmt.Printf("Success: %d, Fails: %d\n", cnt-fails, fails)
	return
}
