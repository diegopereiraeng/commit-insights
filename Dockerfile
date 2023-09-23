FROM golang:1.20-alpine as build
RUN apk add --no-cache --update git
RUN mkdir -p /go/src/github.com/diegopereiraeng/commit-insights
WORKDIR /go/src/github.com/diegopereiraeng/commit-insights 
COPY *.go ./
# Copy go mod and sum files
COPY go.mod go.sum ./

RUN go env GOCACHE 

# Fetch dependencies
RUN go mod download

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o harness-sonar

# Final stage
FROM alpine:3.14

RUN apk --no-cache --update add curl unzip git

COPY --from=build /go/src/github.com/diegopereiraeng/commit-insights/commit-insights /bin/
WORKDIR /bin

ENTRYPOINT /bin/commit-insights
