# captured from Ubuntu 22.04 container, kernel 6.x via Docker Desktop on macOS
# docker run --rm ubuntu:22.04 cat /proc/1/stat
pid1_stat.ubuntu — init process (systemd)

# docker run --rm ubuntu:22.04 cat /proc/stat
proc_stat.ubuntu — system-wide CPU + btime

# docker run --rm ubuntu:22.04 sh -c 'cat /proc/self/stat; cat /proc/self/statm'
self_stat.ubuntu — cat process stat
self_statm.ubuntu — cat process statm

# docker run --rm ubuntu:22.04 cat /proc/self/auxv
self_auxv.ubuntu — auxv vector (glibc, includes AT_CLKTCK=100)

# docker run --rm alpine:3.20 cat /proc/1/stat
pid1_stat.alpine — init process (busybox init)

# docker run --rm alpine:3.20 cat /proc/stat
proc_stat.alpine — system-wide CPU + btime

# docker run --rm alpine:3.20 sh -c 'cat /proc/self/stat; cat /proc/self/statm'
self_stat.alpine — cat process stat (musl)
self_statm.alpine — cat process statm (musl)

# docker run --rm alpine:3.20 cat /proc/self/auxv
self_auxv.alpine — auxv vector (musl, includes AT_CLKTCK=100)

# captured 2026-05-30 by z3robit
