FROM golang:1.11
ARG slVersion
ARG slBuildTime

LABEL maintainer="we@bitspark.de"

LABEL version="1"

WORKDIR /go/src/slang
COPY . .

RUN go get -d -v ./...

ENV ldFlagVersion "-X main.Version=${slVersion}"
ENV ldFlagBuildTime "-X main.BuildTime=${slBuildTime}"
ENV ldFlags "${ldFlagVersion} ${ldFlagBuildTime}"

RUN echo $ldFlags

RUN env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "${ldFlags}" -o slangd ./cmd/slangd

FROM alpine

RUN apk --no-cache add ca-certificates
WORKDIR "/root/slang/"
ENV USER root
ENV SLANG_PATH "/root/slang/"
COPY --from=0 /go/src/slang/slangd .

EXPOSE 5149
EXPOSE 50001-50099

ENTRYPOINT ["/root/slang/slangd", "--only-daemon"]