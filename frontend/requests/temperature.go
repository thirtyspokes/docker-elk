package requests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const temperatureServiceHost = "http://temperature:8080/temperature"

// TemperatureResponse represents the temperature data portion of a response
// from the temperature microservice (both celsius and fahrenheit).
type TemperatureResponse struct {
	TemperatureCelsius    float64 `json:"temp_celsius"`
	TemperatureFahrenheit float64 `json:"temp_fahrenheit"`
}

// GetTemperatureByLocation uses the provided latitude and longitude to query for temperature
// information for that area via the temperature microservice.
func GetTemperatureByLocation(requestID, latitude, longitude string) (*TemperatureResponse, error) {
	tempResponse := &TemperatureResponse{}
	url := fmt.Sprintf("%s?latitude=%s&longitude=%s", temperatureServiceHost, latitude, longitude)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return tempResponse, err
	}

	req.Header.Set("X-Request-Id", requestID)
	resp, err := client.Do(req)
	if err != nil {
		return tempResponse, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return tempResponse, err
	}
	defer resp.Body.Close()

	err = json.Unmarshal(body, tempResponse)
	return tempResponse, err
}
