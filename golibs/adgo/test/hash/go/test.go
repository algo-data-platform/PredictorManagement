package main

//#cgo CXXFLAGS: -std=c++11
//#cgo LDFLAGS: -L${SRCDIR}/../lib/ -lstdc++ -lcity
//#include "../lib/CityCapi.h"
//#include <stdio.h>
import "C"

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("=========go test========")

	var (
		str  string
		bys  []byte
		l    uint32
		now  time.Time
		res1 C.uint64_t
		res2 C.uint64_t
	)

	str = "hello"
	bys = []byte(str)
	l = uint32(len(bys))

	now = time.Now()

	for i := 0; i < 10000; i++ {
		res1 = C.CityHash64(C.CString(str), C.size_t(l))
		res2 = C.CityHash64WithSeed(C.CString(str), C.size_t(l), 10)
	}

	fmt.Printf("str:%s\t\t\tlen:%d\tCityHash64: %d\tCityHash64WithSeed: %d\telapsed: %s\n", str, l, res1, res2, time.Since(now))

	str = "CityHash64WithSeed"
	l = uint32(len(str))

	now = time.Now()
	for i := 0; i < 10000; i++ {
		res1 = C.CityHash64(C.CString(str), C.size_t(l))
		res2 = C.CityHash64WithSeed(C.CString(str), C.size_t(l), 10)
	}
	fmt.Printf("str:%s\t\tlen:%d\tCityHash64: %d\tCityHash64WithSeed: %d\t\telapsed: %s\n", str, l, res1, res2, time.Since(now))
}
