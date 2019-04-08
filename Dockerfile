# FROM golang:latest
FROM golang:1.10.3

LABEL maintainer "ericotieno99@gmail.com"
LABEL vendor="Ekas Technologies"

# Copy the project in to the container
ADD . /go/src/github.com/ekas-data-portal

# Go get the project deps
RUN go get github.com/ekas-data-portal

# Go install the project
RUN go install github.com/ekas-data-portal

# Set the working environment.
ENV GO_ENV production

# Run the ekas-data-portal command by default when the container starts.
ENTRYPOINT /go/bin/ekas-data-portal

#Expose the port specific to the ekas API Application.
EXPOSE 8082


# FROM golang as builder
# WORKDIR /go/src/github.com/habibridho/simple-go/
# COPY . ./
# RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix .

# FROM alpine:latest
# WORKDIR /app/
# COPY --from=builder /go/src/github.com/habibridho/simple-go/simple-go /app/simple-go
# EXPOSE 8888
# ENTRYPOINT ./simple-go

