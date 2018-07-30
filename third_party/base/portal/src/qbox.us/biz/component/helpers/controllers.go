package helpers

import (
	"io"
	"net/http"

	"github.com/teapots/teapot"
)

func CSRFHandler(rw http.ResponseWriter, log teapot.Logger) {
	log.Info("csrf detected")

	rw.WriteHeader(http.StatusForbidden)
	io.WriteString(rw, "csrf detected")
}
