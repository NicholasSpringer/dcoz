FROM golang:1.17

WORKDIR /dcoz
COPY ./dcoz-pause ./dcoz-pause
COPY ./dcoz-agent ./dcoz-agent

WORKDIR /dcoz/dcoz-pause
RUN make


WORKDIR /dcoz/dcoz-agent
RUN go get -d -v ./...
RUN go build -v -o ./bin/dcoz-agent ./...
ENV PATH="/dcoz/dcoz-agent/bin:/dcoz/dcoz-pause/bin:${PATH}"

WORKDIR /go/src/app