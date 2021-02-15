# Apache Log Parser

### Caveats

#### Interval Printing
When printing results per interval, I print them as soon as a new time comes in that is newer than the latest time previously recorded and if the new time is % the interval. This is almost correct, except that the log occasionally will receive entries that are slightly out of time order, so in some cases, entries from the past will not be recorded in the interval printouts.
