FROM golang:1.11.1-alpine3.8 as build-env
# All these steps will be cached
RUN mkdir /hello
WORKDIR /hello
COPY go.mod . # <- COPY go.mod and go.sum files to the workspace
COPY go.sum .

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download
# COPY the source code as the last step
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/hello
FROM scratch # <- Second step to build minimal image
COPY --from=build-env /go/bin/hello /go/bin/hello
ENTRYPOINT ["/go/bin/hello"]