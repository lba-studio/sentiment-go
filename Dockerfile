FROM golang:1.15.6-alpine AS build
WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download
ARG GOOS=linux
ARG GOARCH=amd64
# ARG CGO_ENABLED=0
COPY . .
RUN go build -o /out/main main.go

FROM alpine AS bin
COPY --from=build /out/main /go/bin/main
EXPOSE 8080
CMD ["/go/bin/main"]
