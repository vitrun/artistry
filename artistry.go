/**
 * Copyright ©2014-04-08 Alex <zhirun.yu@duitang.com>
 */
package main

import (
	"net/http"
	"strconv"
	"strings"
	"log"
	"github.com/vitrun/qart"
	"github.com/go-martini/martini"
	"github.com/vitrun/artistry/urlshortener"
	"os"
	"mime/multipart"
)

const SIZE_LIMIT = 1024 * 1024

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
			c <- ""
		} else{
			c <- nurl
		}
	}
	c <- url
}

func readImage(files []*multipart.FileHeader) ([]byte, int){
			// limit to 1M
	img := make([]byte, SIZE_LIMIT+1)
	var size int
	if len(files) > 0{
		file, err := files[0].Open()
		defer file.Close()
		if err != nil {
			return img, 0
		}
		s, err := file.Read(img)
		size = s
	}else{
		file, err := os.OpenFile("public/in.png", os.O_RDONLY, (os.FileMode)(0644))
		defer file.Close()
		if err != nil {
			return img, 0
		}
		s, err := file.Read(img)
		size = s
	}
	return img, size
}

func main() {
	m := martini.Classic()
	m.Map(log.New(os.Stdout, "[martini] ", log.LstdFlags))
	m.Use(martini.Static("public", martini.StaticOptions{"/qr/static/", true, "", nil}))

	m.Get("/qr/", func() (int, string) {
		return 200, `<html><title> Put pictures in QR codes</title>
			<body>
			<h4>Artistry engineers the encoded values to create the picture in a code
			 with no inherent errors.
			</h4>
			<form action='/qr/gen/' method='POST' enctype='multipart/form-data'>
				image: <input name='files' type='file' multiple='multiple' />
				version: <select name='version' type='text' />
				<option value="4">4</option>
				<option value="6">6</option>
				<option value="8">8</option>
				</select>
				shorturl: <input name='short' type='checkbox' /><br />
				url: <input name='url' type='text' size="80" />
				<input type='submit' />
			</form>
			<div>
			default picture: <img src="/qr/static/in.png" /><br />
			github: <a href="https://github.com/vitrun/artistry">Artistry</a>&nbsp;
			document: <a href="http://research.swtch.com/qart">Qart</a>
			</div>
			</body></html>`
	})

	m.Post("/qr/gen/", func(r *http.Request, log *log.Logger) (int, string) {
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

		img, size := readImage(files)
		if size > SIZE_LIMIT {
			return http.StatusInternalServerError, "Image too large"
		}

		qrImg := qart.InitImage(img, 879633355, version, 4, 2, 4, 4,
			false, false, false, false)

		url = <- c
		// error, timeout maybe
		log.Println("Got ur:", url, " shorten:", shorturl)
		if url == "" {
			return http.StatusInternalServerError, "Can not get shortened url"
		}
		qrData := qart.EncodeUrl(url, qrImg)

		// use only the first file
		return 200, (string)(qrData)
	})
	http.ListenAndServe(":8080", m)
}
