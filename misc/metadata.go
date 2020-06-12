package misc

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	// MetadataServer address
	MetadataServer string
	// VerifyHeader is MetadataServer VerifyHeader header name
	VerifyHeader string
	// VerifyValue is MetadataServer VerifyHeader value
	VerifyValue string
)

// Metadata get metadata from server
func Metadata(metadata string) (interface{}, error) {
	var value interface{}
	var result []byte
	client := &http.Client{}
	req, err := http.NewRequest("GET", MetadataServer+"/"+metadata, nil)
	if err != nil {
		log.Print(err)
	}
	req.Header.Add(VerifyHeader, VerifyValue)
	resp, err := client.Do(req)
	if err != nil {
		log.Print(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Print(err)
		}
		result = bodyBytes
	} else {
		result = []byte{}
	}
	err = json.Unmarshal(result, &value)
	if err != nil {
		log.Print(err)
	}
	return value, err
}
