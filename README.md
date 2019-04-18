# tsbWriter

This library provides a read/write thread-safe buffered writer in Golang.  Includes some basic monitoring functionality. 


**Example**
```
//slowWriter could be an HTTP connection, multiplexer, etc.
slowWriter := getSlowWriter()

//Wraps slowWriter with an unbounded thread-safe buffer, allowing a much slower downstream connection to write at its leisure.  By adding a name, prompts writer to report on buffer size and total written periodically.
tsbWriter := NewTSBWriter(slowWriter, 1024, "HTTP to external cache")

buf := make([]byte, 1024)
var n int

for {
	n, err = source.Read(buf)
	if err != nil {
		if err == io.EOF {
			return
		}
		panic(err)
	}
	tsbWriter.Write(buf[:n])
}
```

## FAQ
- **Why not just use channels?**
	Channels should be one of your first options.  However, channels have a byte throughput of roughly 5MB/sec. For some applications, this is too slow. A channel of pointers could have much higher throughput, but adds additional complexity/lock-in. 

## TODO
- Provide other buffer backends such (e.g., cycle buffers, byte pools, channels of buffers)
