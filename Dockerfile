FROM  golang:1.10 as builder

COPY . $GOPATH/src/app/
WORKDIR $GOPATH/src/app/

RUN go get -d ./...
RUN git describe --abbrev=0 --tags | xargs git checkout
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /go/bin/app

FROM scratch
COPY --from=builder /go/bin/app /go/bin/app

ENTRYPOINT ["/go/bin/app"]
CMD ["-f", "/logspout.json"]
EXPOSE 12345