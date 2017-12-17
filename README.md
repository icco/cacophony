# cacophony

[![Build Status](https://travis-ci.org/icco/cacophony.svg?branch=master)](https://travis-ci.org/icco/cacophony) [![GoDoc](https://godoc.org/github.com/icco/cacophony?status.svg)](https://godoc.org/github.com/icco/cacophony)

Provide an rss feed of urls tweeted in my feed

## Documentation

```
< data.json jq '. | map([(.TweetIds | length | tostring), .Link] | join(", ")) | join("\n")' | sed 's/\\n/\
/g' | sort -n > urls.txt
```
