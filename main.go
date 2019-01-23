package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/schema"

	"github.com/dylanmacdonald/image-service/api"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.WithField("service", "image-service")
	r := mux.NewRouter()

	r.Handle("/images", handlers.MethodHandler{
		"GET": images(logger),
	})

	logger.Info("Listening on port 8080")
	logger.Fatal(http.ListenAndServe(":8080", r))
}

func images(logger logrus.FieldLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("New images request")
		req, err := decodeImageRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.WithError(err).Error()
			return
		}
		// could use ioutil.Open
		// decided to use os.Open because the response implements io.Reader
		imgReader, err := os.Open(fmt.Sprintf("./images/%s", req.Path))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			logger.WithError(err).Error()
			return
		}
		// get the image encoding scheme
		img, t, err := image.Decode(imgReader)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.WithError(err).Error()
			return
		}

		imgBuff := &bytes.Buffer{}
		switch t {
		case "jpeg":
			w.Header().Set("Content-Type", "image/jpeg")
			err = jpeg.Encode(imgBuff, img, nil)
		case "png":
			w.Header().Set("Content-Type", "image/png")
			err = png.Encode(imgBuff, img)
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.WithError(err).Error()
			return
		}

		w.Header().Set("Content-Length", strconv.Itoa(imgBuff.Len()))
		_, err = w.Write(imgBuff.Bytes())
		if err != nil {
			// possibly status partial content but partial content is kinda useless with no images
			w.WriteHeader(http.StatusInternalServerError)
			logger.WithError(err).Error()
			return
		}
	})
}

func decodeImageRequest(r *http.Request) (*api.ImageRequest, error) {
	req := &api.ImageRequest{}
	return req, schema.NewDecoder().Decode(req, r.URL.Query())
}
