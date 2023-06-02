FROM golang:1.20 as builder

WORKDIR /opt/

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

ENV GOOS linux
ENV CGO_ENABLED=0

ARG VERSION
RUN go build -v -ldflags "-X main.version=$VERSION" -o app ./cmd

FROM alpine:3.17.3

COPY ./config.yml /tmp/config.yml

EXPOSE 8181
ENTRYPOINT ["app"]
CMD ["--config.file=/tmp/config.yml"]
RUN addgroup -g 1000 app && \
    adduser -u 1000 -D -G app app -h /app

WORKDIR /app/


RUN apk --no-cache add ca-certificates
COPY --from=builder /opt/app /usr/local/bin/app
COPY ./static/email.tpl.html .
USER app
