FROM alpine:latest

WORKDIR /app

ARG VERSION=0.1.1

# Use the VERSION argument in the COPY and ENTRYPOINT instructions
COPY ./releases/go-reversi-${VERSION}-linux-x86 /app/go-reversi

RUN chmod +x /app/go-reversi

ENV VERSION=${VERSION}
ENTRYPOINT ["/app/go-reversi"]

# Accept Command line options
CMD []
