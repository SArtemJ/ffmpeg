package main

import (
	"fmt"
	"github.com/asticode/goav/avformat"
)

func main() {

	// Alloc ctx
	ctxFormat := avformat.AvformatAllocContext()

	// Open input
	// We need to create an intermediate variable to avoid "cgo argument has Go pointer to Go pointer" errors
	if ret := avformat.AvformatOpenInput(&ctxFormat, "sample.mp4", nil, nil); ret < 0 {
		fmt.Sprintf("astilibav: avformat.AvformatOpenInput on %+v failed", ret)
		return
	} else {
		fmt.Printf("context %+v", ctxFormat)
		fmt.Printf("ret %+v", ret)
	}

	fmt.Println("Test ok")
}
