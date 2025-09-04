FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

#install dependencies, and build mosquitto
RUN apt-get update
RUN apt-get install curl -y

RUN apt install software-properties-common -y
RUN add-apt-repository ppa:deadsnakes/ppa

RUN apt-get update

RUN apt-get install python3.10 -y

# Install uv
RUN curl -LsSf https://astral.sh/uv/install.sh | sh

WORKDIR /app