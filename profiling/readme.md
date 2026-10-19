* run bench with mem and cpu output
```
go test -run='^$' -bench=BuildFanOutKeys -benchtime=3s -cpuprofile cpu.out -memprofile mem.out
```

* run bench with mem and cpu output
```shell
go tool pprof -sample_index=alloc_objects -top mem.out
go tool pprof -sample_index=alloc_objects -list 'BuildFanOutKeys' mem.out
```