FROM alpine

RUN apk update && apk add go git musl-dev && rm -rf /var/cache/apk/*

ENV GOROOT /usr/lib/go
ENV GOPATH /gopath
ENV GOBIN /usr/bin

ADD . /gopath/src/github.com/Patazerty/csf

WORKDIR /gopath/src/github.com/Patazerty/csf

RUN go get github.com/rakyll/statik  github.com/valyala/quicktemplate/qtc

RUN qtc templates && statik -f -src=./static

RUN go build -o csf main.go

FROM alpine

COPY --from=0 /gopath/src/github.com/Patazerty/csf/csf .

EXPOSE 8080 8888 8042

ENTRYPOINT ["./csf"]

CMD []
