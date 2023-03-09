package shared

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type Error struct {
	status int
	code   int
	msg    string
}

func NewError(status, code int, msg string) *Error {
	return &Error{status, code, msg}
}

func (err Error) JSON(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	e := json.NewEncoder(&buf).Encode(map[string]interface{}{
		"status": err.status,
		"code":   err.code,
		"msg":    err.msg,
	})
	if e != nil {
		panic(e)
	}

	log.Printf("%s: %s", r.URL.Path, buf.String())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.status)
	w.Write(buf.Bytes())
}

// common errors
var NotFound = NewError(http.StatusNotFound, 0, "not found")
