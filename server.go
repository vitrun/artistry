/**
 * Copyright ©2014-04-08 Alex <zhirun.yu@duitang.com>
 */
package main

import (
	"github.com/go-martini/martini"
	"github.com/vitrun/qart"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	m := martini.Classic()
	SIZE_LIMIT := 1024 * 1024
	m.Get("/", func() (int, string) {
		return 200, `<html><body>
			<form action='/qr/gen/' method='POST' enctype='multipart/form-data'>
				image: <input name='files' type='file' multiple='multiple' />
				url: <input name='url' type='text' />
				version: <input name='version' type='text' />
				<input type='submit' />
			</form>
			</body></html>`
	})

	m.Post("/qr/gen/", func(r *http.Request) (int, string) {
		url := r.FormValue("url")
		versionStr := r.FormValue("version")
		files := r.MultipartForm.File["files"]

		err := r.ParseMultipartForm(100000)
		if err != nil {
			return http.StatusInternalServerError, err.Error()
		}
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			return http.StatusInternalServerError, err.Error()
		}
		// should remove #
		url = strings.Split(url, "#")[0]

		for i, _ := range files {
			file, err := files[i].Open()
			defer file.Close()
			if err != nil {
				return http.StatusInternalServerError, err.Error()
			}
			// limit to 1M
			img := make([]byte, SIZE_LIMIT+1)
			size, err := file.Read(img)
			if size > SIZE_LIMIT {
				return http.StatusInternalServerError, "Image too large"
			}
			if err != nil {
				return http.StatusInternalServerError, err.Error()
			}
			qrImg := qart.Encode(url, img, 879633355, version, 4, 2, 4, 4,
				false, false, false, false)

			// use only the fist file
			return 200, (string)(qrImg)
		}
		return 200, "ok"
	})
	http.ListenAndServe(":8080", m)
}
