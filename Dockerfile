FROM golang:1.21.4-alpine3.18 AS BUILD_IMAGE

WORKDIR /opt/juju-build/
COPY . .

RUN go build -o ./juju-server cmd/juju-server/main.go

FROM golang:1.21.4-alpine3.18 AS RUNNER_IMAGE

WORKDIR /opt/juju
COPY --from=BUILD_IMAGE /opt/juju-build/juju-server .

ENV DOCKER="true"

CMD ["/opt/juju/juju-server"]
