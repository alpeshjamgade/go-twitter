FROM alpine:latest as builder

RUN mkdir /app

COPY ./build/authApp /app

CMD ["app/authApp"]