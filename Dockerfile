FROM golang:1.10 as build

COPY . /go/src/github.com/johananl/heating-control
WORKDIR /go/src/github.com/johananl/heating-control
RUN go get -v github.com/golang/dep/cmd/dep
RUN dep ensure -v
RUN CGO_ENABLED=0 GOOS=linux go build -a

FROM alpine
COPY --from=build /go/src/github.com/johananl/heating-control/heating-control /
CMD /heating-control
