FROM scratch

WORKDIR /app

# This value should be passed as a build argument
ARG VERSION=0

# Use the VERSION argument in the COPY and ENTRYPOINT instructions
COPY ./releases/go-reversi-${VERSION}-linux-x86 /app/go-reversi

ENTRYPOINT ["/app/go-reversi"]

# Accept Command line options
CMD []
