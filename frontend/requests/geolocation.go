package requests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const geolocationServiceHost = "http://geolocation:8080/geolocate"

// GeolocateResponse represents the latitude and longitude portions
// of the response from from the Geolocation microservice.
type GeolocateResponse struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

// GeolocateAddress takes a street address as a string and queries the
// geolocation microservice to fetch the specific latitude and longitude of
// the address.
func GeolocateAddress(requestID string, address string) (*GeolocateResponse, error) {
	locationResponse := &GeolocateResponse{}
	url := fmt.Sprintf("%s?address=%s", geolocationServiceHost, url.QueryEscape(address))

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return locationResponse, err
	}

	req.Header.Set("X-Request-Id", requestID)

	resp, err := client.Do(req)
	if err != nil {
		return locationResponse, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return locationResponse, err
	}
	defer resp.Body.Close()

	err = json.Unmarshal(body, locationResponse)
	return locationResponse, err
}
