FROM alpine:3.13.4

EXPOSE 8444
WORKDIR /app

COPY dlog /app

CMD ["/app/dlog"]