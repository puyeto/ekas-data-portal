# FROM golang:1.10.3
FROM golang:alpine AS build-env

LABEL maintainer "ericotieno99@gmail.com"
LABEL vendor="Ekas Technologies"

RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

# Create appuser
RUN adduser -D -g '' appuser

WORKDIR /go/src/github.com/ekas-data-portal

ENV GOOS=linux
ENV GOARCH=386
ENV CGO_ENABLED=0

# Copy the project in to the container
ADD . /go/src/github.com/ekas-data-portal

# Go get the project deps
RUN go get github.com/ekas-data-portal

# Go install the project
# RUN go install github.com/ekas-data-portal
RUN go build

# Set the working environment.
ENV GO_ENV production

# Run the ekas-data-portal command by default when the container starts.
# ENTRYPOINT /go/bin/ekas-data-portal

# Run the ekas-portal-api command by default when the container starts.
# ENTRYPOINT /go/bin/ekas-portal-api

FROM alpine:latest
WORKDIR /go/

COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-env /etc/passwd /etc/passwd
COPY --from=build-env /go/src/github.com/ekas-data-portal/ekas-data-portal /go/ekas-data-portal
COPY --from=build-env /go/src/github.com/ekas-data-portal/logs/data.json /go/logs/data.json
RUN chown -R appuser:appuser /go/logs
RUN chmod -R 666 /go/logs

# Use an unprivileged user.
USER appuser

# Set the working environment.
ENV GO_ENV production

ENTRYPOINT ./ekas-data-portal

#Expose the port specific to the ekas API Application.
EXPOSE 8083
EXPOSE 7001


