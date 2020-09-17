package util

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// reads the body and then resets the request.body reader
// runs json.unmarshal on the body into the out variable
func ParseBody(req *http.Request, out interface{}) error {
	bodybytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bodybytes, &out)
	if err != nil {
		return err
	}

	req.Body = ioutil.NopCloser(bytes.NewReader(bodybytes))

	return nil
}
