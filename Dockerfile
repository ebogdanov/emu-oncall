FROM golang:1.20.1-alpine as builder

ARG LDFLAGS=""

ENV GIT_TERMINAL_PROMPT=1

WORKDIR /emu-oncall
COPY go.mod go.sum ./

RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -a -o bin/emu-oncall cmd/main.go


FROM ubuntu:20.04

ENV TZ=Europe/Moscow
RUN echo $TZ > /etc/timezone && \
    apt-get update && apt-get install -y tzdata && \
    apt-get install -y ca-certificates && \
    rm /etc/localtime && \
    ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && \
    dpkg-reconfigure -f noninteractive tzdata && \
    apt-get clean

WORKDIR /opt/
COPY --from=builder /emu-oncall/bin/emu-oncall ./
RUN chmod +x ./emu-oncall

COPY ./config/config.yaml /opt/config/config.yaml
CMD ["/opt/emu-oncall"]
