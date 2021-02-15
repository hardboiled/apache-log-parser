# Apache Log Aggregating Client

This application is meant to emulate a client that a user would install to provide observability metrics for their apache server.

## How it works

There is one main routine and two subroutines.

The main routine bootstraps an input and output channel, then passes these to the channels that will each respectively handle this I/O.

The input channel is handled by a simple routine that calls an OSS library `"github.com/gocarina/gocsv"` to serialize the data.

The output routine takes the real-time data and minor calculations done by the main routine and handles additional processing (like interval calculations) and prints any alarms triggered.

The primary reason for this split is to avoid blocking on I/O as much as possible. However, the output routine is probably unnecessary and could be done in the main thread. I found the extra output routine made the code more complex than I wanted it to be (especially with the need to add an extra wait group just for this purpose), but it was fun to debug and would be interesting to try in production code to see if there's any difference in performance, so I left it in.

Once both channels are setup, I run the main loops which schedules the printing of alarms and interval stats. After the CSV file is completely read, I flush any data remaining, wait for the output thread to finish processing it, and then exit.

## How to install

- Install go.
- run `go get`


I ran this on the following go version, but if you have an older one installed, you can update the go.mod file to use something prior to go 15.

```bash
$ go version
> go version go1.15.8 darwin/amd64
```

## How to run application

Running is simple. Here's how to run with example parameters. The input file defaults to "input_files/sample_csv.txt"

```golang
go run main.go -interval=10 -window-retention=290 -alarm-threshold=10 -input-filepath=<your-filepath>
```

## How run tests

```golang
go test ./...
```

### Caveats

#### Interval Printing

Interval printing in a way violates some best practices in GO around sending data instead of sharing memory. The idea is that we should send messages in between processes to avoid data corruption due to race conditions. In my case, I am copying a slice of the length of the interval supplied by the user and sending it over to be processed for output. This slice contains shallow copies (i.e. pointers) instead of being a complete copy.

That being said, the buffer is treated as read-only, and this is a standard pattern in producer consumer models like kafka, which I'm trying to emulate. The window is large enough so that slices being processed should never be overwritten, since the main go routing blocks until the output process is done with the previous slice.

### Potential Improvements

#### Output Buffer

At the moment, I'm just printing everything to standard out. This is fine because in production, most of the time we just redirect standard out to the buffer of our choosing. However, I configued the code to use a writer interface, so that if we wanted to replace it with a different write buffer in the future it would be straightfoward.

#### 2min warning

I left the 2 min warning as unconfigurable. It might be nice to allow users to configure it as they see fit... (e.g. 15 minutes). This would be useful when people has a bursty traffic profile and want to amortize over a longer period.
