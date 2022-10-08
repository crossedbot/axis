ARG OS=debian:bullseye-slim
ARG GOLANG_VERSION=1.19-bullseye
ARG CGO=0
ARG GOOS=linux
ARG GOARCH=amd64

FROM golang:${GOLANG_VERSION} AS gobuilder

ARG CGO
ARG GOOS
ARG GOARCH

RUN go version
WORKDIR /go/src
COPY . .
RUN cd cmd/pins && \
    CGO_ENABLED='${CGO}' GOOS='${GOOS}' GOARCH='${GOARCH}' \
    make -f /go/src/Makefile build

#-------------------------------------------------------------------------------
FROM ${OS}

ENV PINS_HOME /usr/local/pins
ENV PATH ${PINS_HOME}/bin:$PATH
RUN mkdir -vp ${PINS_HOME}
WORKDIR ${PINS_HOME}

COPY --from=gobuilder /go/bin/pins ./bin/pins
COPY ./scripts/run.bash ./bin/run-pins

EXPOSE 7070
ENTRYPOINT [ "run-pins", "-p", "7070", "-d", "redis:6379", "-i", "/dns4/ipfs/tcp/9094", "-a", "http://simpleauth:8080/.well-known/jwks.json", "-g", "authenticated" ]
