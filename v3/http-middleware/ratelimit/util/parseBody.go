package util

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// reads the body and then resets the request.body reader
// runs json.unmarshal on the body into the out variable
func ParseBody(req *http.Request, out interface{}) error {
	bodybytes, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bodybytes, &out)
	if err != nil {
		return err
	}

	req.Body = io.NopCloser(bytes.NewReader(bodybytes))

	return nil
}
