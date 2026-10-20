* run bench with mem and cpu output
```
go test -run='^$' -bench=BuildFanOutKeys -benchtime=3s -cpuprofile cpu.out -memprofile mem.out
```
* run bench with mem and cpu output
```shell
go tool pprof -sample_index=alloc_objects -top mem.out
go tool pprof -sample_index=alloc_objects -list 'BuildFanOutKeys' mem.out
```
* flags
```
inuse_space	Current memory footprint in bytes.	Diagnosing Memory Leaks and discovering what is consuming RAM right now.
inuse_objects	Current count of live objects.	Finding out if your program is maintaining an excessive number of tiny data structures (like maps or structs).
alloc_space	Cumulative memory footprint in bytes.	Reducing the total volume of memory churn and spotting oversized temporary allocations.
alloc_objects	Cumulative count of allocated objects.	Optimizing CPU / Garbage Collection performance by identifying high-frequency allocation hotspots.
```

