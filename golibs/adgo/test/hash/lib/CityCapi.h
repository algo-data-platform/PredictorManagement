#ifndef CITY_H_NGAHRTJN
#define CITY_H_NGAHRTJN

#include <stdint.h>
#include <stdlib.h>

#ifdef __cplusplus 
extern "C" {
#endif

// Hash function for a byte array.
uint64_t CityHash64(const char *buf, size_t len);

// Hash function for a byte array.  For convenience, a 64-bit seed is also
// hashed into the result.
uint64_t CityHash64WithSeed(const char *buf, size_t len, uint64_t seed);

// Hash function for a byte array.  For convenience, two seeds are also
// hashed into the result.
uint64_t CityHash64WithSeeds(const char *buf, size_t len,
                           uint64_t seed0, uint64_t seed1);

#ifdef __cplusplus
}
#endif

#endif /* end of include guard: CITY_H_NGAHRTJN */
