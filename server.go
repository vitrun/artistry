/**
 * Copyright ©2014-04-08 Alex <zhirun.yu@duitang.com>
 */
package main

import (
//    "github.com/vitrun/qart"
	"github.com/go-martini/martini"
//	"fmt"
)

func main(){
    m := martini.Classic()
    m.Get("/", func () string {
        return "Hello World"
    })
    m.Run()
}
