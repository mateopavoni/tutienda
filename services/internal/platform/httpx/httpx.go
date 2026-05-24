// Package httpx holds small HTTP helpers shared by the services: JSON encoding,
// request decoding and a consistent error envelope.
package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
)

// Error is the JSON shape returned for every failure, so clients can rely on one format.
// It also implements the error interface, so a decoded remote error can be returned as-is.
type Error struct {
	Message string `json:"error"`
	Detail  string `json:"detail,omitempty"`
	// SKU is set when a failure is tied to a specific stock keeping unit (out of stock).
	SKU string `json:"sku,omitempty"`
}

func (e *Error) Error() string {
	if e.Detail != "" {
		return e.Message + ": " + e.Detail
	}
	return e.Message
}

// JSON writes v as JSON with the given status code.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(v)
}

// Fail writes a structured error envelope.
func Fail(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, Error{Message: msg})
}

// FailDetail writes an error envelope with an extra human-readable detail.
func FailDetail(w http.ResponseWriter, status int, msg, detail string) {
	JSON(w, status, Error{Message: msg, Detail: detail})
}

// Decode reads a JSON body into dst, rejecting empty or malformed payloads.
func Decode(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("empty request body")
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
