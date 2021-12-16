FROM golang:1.17

# These are necessary for go mod download to accept 'replace' of dcoz with local
WORKDIR /dcoz
COPY go.mod .
COPY go.sum .

WORKDIR /dcoz/dcoz-controller
COPY dcoz-controller/go.mod .
COPY dcoz-controller/go.sum .
RUN go mod download

WORKDIR /dcoz
COPY . .
WORKDIR /dcoz/dcoz-controller 
RUN go install github.com/NicholasSpringer/dcoz/dcoz-controller