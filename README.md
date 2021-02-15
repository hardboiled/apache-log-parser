# Apache Log Parser



### Caveats

#### Interval Printing

Interval printing in a way violates some best practices in GO around sending data instead of sharing memory. The idea is that we should send messages in between processes to avoid data corruption due to race conditions. In my case, I am copying a slice of the length of the interval supplied by the user and sending it over to be processed for output. This slice contains shallow copies (i.e. pointers) instead of being a complete copy.

That being said, the buffer is treated as read-only, and this is a standard pattern in producer consumer models like kafka, which I'm trying to emulate. The window is large enough so that slices being processed should never be overwritten, since the main go routing blocks until the output process is done with the previous slice.
