#include <inttypes.h>
#include <chrono>
#include <cstdio>
#include <cstring>
#include "../lib/CityCapi.h"

using std::chrono::high_resolution_clock;
using std::chrono::microseconds;

int main(int argc, char *argv[]) {
    printf("=========cpp test========\n");
    const char *s1 = "hello";
    size_t l1 = strlen(s1);

    std::chrono::steady_clock::time_point start_time_ =
        std::chrono::steady_clock::now();

    // auto begin = high_resolution_clock::now();

    uint64_t res1;
    uint64_t res2;

    for (int i = 0; i < 10000; i++) {
        res1 = CityHash64(s1, l1);
        res2 = CityHash64WithSeed(s1, l1, 10);
    }

    std::chrono::steady_clock::time_point t2 = std::chrono::steady_clock::now();
    std::chrono::duration<double> time_span =
        std::chrono::duration_cast<std::chrono::duration<double>>(t2 -
                                                                  start_time_);
    double time = static_cast<double>(time_span.count() * 1000);
    // auto end = high_resolution_clock::now();

    printf(
        "str:%s\t\t\tlen:%lu\tCityHash64: %lu\tCityHash64WithSeed: "
        "%lu\telapsed:%f\n",
        s1, l1, res1, res2, time);

    const char *s2 = "CityHash64WithSeed";
    size_t l2 = strlen(s2);

    std::chrono::steady_clock::time_point start_time_1 = std::chrono::steady_clock::now();

    for (int i = 0; i < 10000; i++) {
        res1 = CityHash64(s2, l2);
        res2 = CityHash64WithSeed(s2, l2, 10);
    }

    t2 = std::chrono::steady_clock::now();
    time_span = std::chrono::duration_cast<std::chrono::duration<double>>(
        t2 - start_time_1);
    time = static_cast<double>(time_span.count() * 1000);

    printf(
        "str:%s\t\tlen:%lu\tCityHash64: %lu\tCityHash64WithSeed: "
        "%lu\t\telapsed:%f\n",
        s2, l2, res1, res2, time);
    return 0;
}
