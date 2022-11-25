FROM golang:1.18-alpine

ENV GOPROXY="https://proxy.golang.org"
ENV GO111MODULE="on"
ENV NAT_ENV="production"

EXPOSE 8080
WORKDIR /go/src/github.com/icco/cacophony
COPY . .

RUN go build -o /go/bin/server .

CMD ["/go/bin/server"]
