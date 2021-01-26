package cityhash

//#cgo LDFLAGS: -L${SRCDIR} -lstdc++ -lcity
//#include "CityCapi.h"
//#include <stdio.h>
//#include <stdlib.h>
import "C"
import "unsafe"

func CityHash64WithSeed(buf string, length uint64, seed uint64) uint64 {
	cstr := C.CString(buf)
	hash := C.CityHash64WithSeed(cstr, C.size_t(length), C.uint64_t(seed))
	C.free(unsafe.Pointer(cstr))
	return uint64(hash)
}
