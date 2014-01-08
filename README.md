satin-go
========

CO2 Laser Saturation Intensity calculation in Go

For a similar standalone utility to the satin-c version, use ``go build`` and run
``satin-go``. Use ``satin-go -concurrent`` to use goroutines.

For a benchmark that captures profiling data, use ``go test -bench .``.

For profiling details:

    go test -bench . -cpuprofile cpu.out -memprofile mem.out
    go tool pprof satin-go cpu.out
    (pprof) top10

Performance Results
===================

Use ``satin-c`` or ``satin-go`` repositories, compile and run the resulting
binaries. Execute three times and then record the fastest execution time (as
reported in seconds during execution) below. Lower numbers indicate faster
execution.

| Config ID | GCC      | Go       | OS                           | Hardware                                                                 |
| --------- | -------- | -------- | ---------------------------- | ------------------------------------------------------------------------ |
| 1         | 4.8.2    | 1.1.2    | Linux 3.12.1-1-ARCH x86_64   | Sony Vaio Series SVZ13115GGXI (8 GB RAM; 256 GB SSD; Intel i7-3612QM)    |

| Config ID | C satin              | satin-go             | C satin -concurrent  | satin-go -concurrent |
| --------- | -------------------- | -------------------- | -------------------- | -------------------- |
| 1         | 16.97                | 14.613               | 2.06                 | 2.032                |

