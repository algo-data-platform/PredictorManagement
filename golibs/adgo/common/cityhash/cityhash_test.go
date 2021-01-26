package cityhash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCityHash64WithSeed(t *testing.T) {
	assert.Equal(t, uint64(14952328244738539343), CityHash64WithSeed("test_service", uint64(len("test_service")), 0), "error")
	assert.Equal(t, uint64(13185074222889755016), CityHash64WithSeed("thrift", uint64(len("thrift")), uint64(14952328244738539343)), "error")
}

func BenchmarkCityHash64WithSeed(b *testing.B) {
	var seed uint64 = 0
	str := "BenchmarkCityHash64WithSeed"
	length := uint64(len(str))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		seed = CityHash64WithSeed(str, length, seed)
	}
}
