# FROM golang:1.21-alpine as build
# # Install git for fetching dependencies
# RUN apk add --no-cache --update git

# # Set working directory inside the container
# WORKDIR /app

# COPY *.go ./
# COPY internal ./internal
# # Copy go mod and sum files
# COPY go.mod go.sum ./

# RUN go env GOCACHE 

# # Fetch dependencies
# RUN go mod download

# RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o commit-insights 

# Final stage
FROM alpine:3.14

RUN apk --no-cache --update add curl unzip git

# COPY --from=build /go/src/github.com/diegopereiraeng/commit-insights/commit-insights /bin/
COPY binary/commit-insights /bin/
WORKDIR /bin

ENTRYPOINT /bin/commit-insights
