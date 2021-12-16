FROM golang:1.17

WORKDIR /dcoz
COPY go.mod .
COPY go.sum .

WORKDIR /dcoz/dcoz-agent
COPY dcoz-agent/go.mod .
COPY dcoz-agent/go.sum .
RUN go mod download

WORKDIR /dcoz
COPY . .

WORKDIR /dcoz/dcoz-pause
RUN make
ENV PATH="/dcoz/dcoz-pause/bin:${PATH}"

WORKDIR /dcoz/dcoz-agent
RUN go install github.com/NicholasSpringer/dcoz/dcoz-agent

