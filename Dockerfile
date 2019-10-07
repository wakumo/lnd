FROM golang:1.13-alpine as builder

# Force Go to use the cgo based DNS resolver. This is required to ensure DNS
# queries required to connect to linked containers succeed.
ENV GODEBUG netdns=cgo

# Install dependencies and build the binaries.
RUN apk add --no-cache --update alpine-sdk \
    git \
    make \
    gcc

COPY . /go/src/github.com/lightningnetwork/lnd/

WORKDIR /go/src/github.com/lightningnetwork/lnd

RUN make \
 && make install tags="signrpc walletrpc chainrpc invoicesrpc routerrpc"

COPY lnrpc/rpc.pb.go lnrpc_client/lnrpc/lnrpc/

# Expose lnd ports (p2p, rpc).
EXPOSE 9736 10008

CMD ["lnd"]
