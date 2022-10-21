# Builder
FROM golang:1.19-alpine3.16 as builder

RUN apk update && apk upgrade && \
    apk --update add git make

WORKDIR /app

COPY . .

RUN make build

# Distribution
FROM alpine:latest

RUN apk update && apk upgrade && \
    apk --update --no-cache add tzdata && \
    mkdir /app

WORKDIR /app

EXPOSE 9090

COPY --from=builder /app/engine /app
COPY --from=builder /app/config/config.yaml  /app/config/config.yaml
CMD /app/engine