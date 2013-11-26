satin-go
========

CO2 Laser Saturation Intensity calculation in Go

This is a port of Alan Stewart's C-based code to golang.

For a similar standalone utility to the C version, use ``go build`` and run
``satin-go``. Use ``satin-go -concurrent`` to use goroutines.

For a benchmark that captures profiling data, use ``go test -bench .``.

For profiling details:

    go test -bench . -cpuprofile cpu.out -memprofile mem.out
    go tool pprof satin-go cpu.out
    (pprof) top10

Results
=======

Test machine:

* Sony Vaio Series SVZ13115GGXI (8 GB RAM; 256 GB SSD; Intel i7-3612QM)
* Linux 3.12.0-1-ARCH x86_64
* GCC 4.8.2
* Go 1.1.2

C version (best of 3, run with ``make bench`` in the ``satin-c`` repository):

* ``./satin``: 2.43 seconds
* ``./satinSingleThread``: 16.79 seconds

golang version (best of 3, run with ``satin-go`` or ``satin-go -concurrent``):

* ``satin-go -concurrent``: 3.109 seconds
* ``satin-go``: 14.589 seconds
