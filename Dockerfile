FROM golang:1.13 AS builder

RUN apk add bash ca-certificates git gcc g++ libc-dev
WORKDIR /build
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . /build/

ARG APP=dumpbeat
ARG RELEASE=0.0.0-dev
ARG COMMIT=HEAD

RUN BUILD_TIME="$(date -u '+%Y-%m-%dT%H:%M:%SZ')" ; \
    CGO_ENABLED=0 go build \
    -ldflags "-s -w -X $APP/pkg/version.version=$RELEASE -X $APP/pkg/version.commit=$COMMIT -X $APP/pkg/version.buildTime=$BUILD_TIME" \
    -o bin/$APP

FROM alpine:latest as alpine
RUN apk --no-cache add tzdata zip ca-certificates
WORKDIR /usr/share/zoneinfo
RUN zip -r -0 /zoneinfo.zip .

FROM scratch
COPY --from=build /app/bin/app /app
ENV ZONEINFO /zoneinfo.zip
COPY --from=alpine /zoneinfo.zip /
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/$APP"]
