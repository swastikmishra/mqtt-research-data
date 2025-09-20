FROM ubuntu:22.04

ARG GO_VERSION
ENV GO_VERSION=1.24.0

RUN apt-get update
RUN apt-get install -y wget git gcc

RUN wget -P /tmp "https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz"

RUN tar -C /usr/local -xzf "/tmp/go${GO_VERSION}.linux-amd64.tar.gz"
RUN rm "/tmp/go${GO_VERSION}.linux-amd64.tar.gz"

ENV GOPATH /app
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

RUN wget https://deb.nodesource.com/setup_20.x 
RUN chmod +x setup_20.x
RUN ./setup_20.x
RUN apt-get install -y nodejs

RUN npm --global add nodemon

WORKDIR $GOPATH