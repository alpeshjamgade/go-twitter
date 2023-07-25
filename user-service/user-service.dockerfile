FROM alpine:latest as builder

RUN mkdir /app
COPY ./build/userApp /app

CMD ["app/userApp"]