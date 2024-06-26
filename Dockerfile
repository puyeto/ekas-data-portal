# FROM golang:1.10.3
# FROM golang:alpine AS build-env
FROM golang:latest AS build-env

LABEL maintainer "ericotieno99@gmail.com"
LABEL vendor="Ekas Technologies"

# RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

# Create appuser
# RUN adduser -D -g '' appuser

WORKDIR /go/ekas-data-portal

ENV GOOS=linux
ENV GOARCH=386
ENV CGO_ENABLED=0

# Copy the project in to the container
ADD . /go/ekas-data-portal

RUN go mod download 

# Go get the project deps
RUN go get github.com/ekas-data-portal

# Set the working environment.
ENV GO_ENV production

# Go install the project
# RUN go install github.com/ekas-data-portal
RUN go build


# Run the ekas-data-portal command by default when the container starts.
# ENTRYPOINT /go/bin/ekas-data-portal

# Run the ekas-portal-api command by default when the container starts.
# ENTRYPOINT /go/bin/ekas-portal-api

# FROM alpine:latest
FROM golang:latest
WORKDIR /go/

COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-env /etc/passwd /etc/passwd
COPY --from=build-env /go/ekas-data-portal/ekas-data-portal /go/ekas-data-portal
RUN mkdir p logs  

# Use an unprivileged user.
# USER appuser

# Set the working environment.
ENV GO_ENV production

ENTRYPOINT ./ekas-data-portal

#Expose the port specific to the ekas API Application.
EXPOSE 8083
EXPOSE 7001


