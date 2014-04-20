/**
 * Copyright ©2014-04-08 Alex <zhirun.yu@duitang.com>
 */
package main

import (
	"net/http"
	"strconv"
	"strings"
	"github.com/vitrun/qart"
	"github.com/go-martini/martini"
	"github.com/vitrun/artistry/urlshortener"
)


func shorten(url string) (string, error){
	// 使用http中的default client创建一个新的 urlshortener 实例
	svc, _ := urlshortener.New(http.DefaultClient)
	res, err := svc.Url.Insert(&urlshortener.Url { LongUrl: url, }).Do()
	if err != nil {
		return "", err
	}
	return res.Id, err
}

func lengthen(url string) (string, error) {
	svc, _ := urlshortener.New(http.DefaultClient)
	res, err := svc.Url.Get(url).Do()
	if err != nil {
		return "", err
	}
	return res.LongUrl, err
}

func prepareUrl(c chan string, toShort bool) {
	url := <- c
	// should remove #
	url = strings.Split(url, "#")[0]
	if toShort {
		nurl, err := shorten(url)
		if err != nil {
			c <- nurl
		} else{
			c <- ""
		}
	}
	c <- url
}

func main() {
	m := martini.Classic()
	SIZE_LIMIT := 1024 * 1024
	m.Get("/", func() (int, string) {
		return 200, `<html><body>
			<form action='/qr/gen/' method='POST' enctype='multipart/form-data'>
				image: <input name='files' type='file' multiple='multiple' />
				url: <input name='url' type='text' />
				version: <input name='version' type='text' />
				shorturl: <input name='short' type='checkbox' />
				<input type='submit' />
			</form>
			</body></html>`
	})

	m.Post("/qr/gen/", func(r *http.Request) (int, string) {
		url := r.FormValue("url")
		versionStr := r.FormValue("version")
		shorturl := r.FormValue("short")
		files := r.MultipartForm.File["files"]

		err := r.ParseMultipartForm(100000)
		if err != nil {
			return http.StatusInternalServerError, err.Error()
		}
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			return http.StatusInternalServerError, err.Error()
		}

		c := make(chan string, 1)
		c <- url
		go prepareUrl(c, shorturl=="on")
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
			qrImg := qart.InitImage(img, 879633355, version, 4, 2, 4, 4,
				false, false, false, false)

			url := <- c
			// error, timeout maybe
			if url == "" {
				return http.StatusInternalServerError, "Can not get shortened url"
			}
			qrData := qart.EncodeUrl(url, qrImg)

			// use only the first file
			return 200, (string)(qrData)
		}
		return 200, "ok"
	})
	http.ListenAndServe(":8080", m)
}
