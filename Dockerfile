### Checkout UI, Lib and Examples and Build UI
FROM node:8.12
WORKDIR /slang/

# Clone examples, libs and ui
RUN git clone https://github.com/Bitspark/slang-examples.git  examples && \
    git clone https://github.com/Bitspark/slang-lib.git  lib && \
    git clone https://github.com/Bitspark/slang-ui.git ui

# Checkout latest lib release
RUN cd lib && \
    git checkout $(git describe --tags `git rev-list --tags --max-count=1`)

# Checkout and build latest UI release
RUN cd ui && \
    git checkout $(git describe --tags `git rev-list --tags --max-count=1`) && \
    npm install && \
    ./node_modules/@angular/cli/bin/ng build --base-href /app/  --prod --output-path=dist


### Build daemon
FROM golang:1.11
WORKDIR /go/src/slang

COPY . .
RUN go get -d -v ./... && \
    env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o slangd ./cmd/slangd


### Gather UI, lib and daemon and run daemon
FROM alpine
LABEL maintainer="we@bitspark.de"
LABEL version="1"
WORKDIR /root/slang/

RUN apk --no-cache add ca-certificates
ENV USER root
ENV SLANG_PATH "/root/slang/"
COPY --from=0 /slang/examples/examples projects/examples
COPY --from=0 /slang/lib/slang lib/slang/
COPY --from=0 /slang/ui/dist   ui/
COPY --from=1 /go/src/slang/slangd .

EXPOSE 5149

ENTRYPOINT ["/root/slang/slangd", "--only-daemon", "--skip-checks"]