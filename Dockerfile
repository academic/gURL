# parent image
FROM golang:latest AS builder

# workspace directory
WORKDIR /app

# copy `go.mod` and `go.sum`
ADD go.mod go.sum ./

# install dependencies
RUN go mod download

# copy source code
COPY . .

# build executable
RUN go build -o ./bin/gurl .

##################################

# parent image
FROM alpine:3.12.2

# workspace directory
WORKDIR /app

# copy binary file from the `builder` stage
COPY --from=builder /app/bin/gurl ./

# set entrypoint
ENTRYPOINT [ "./gurl" ]