ARG ARCH=amd64
ARG OS=linux
ARG CGO_ENABLED=0
# Build the manager binary
FROM golang:1.23 AS builder

WORKDIR /workspace

# Copy source files
COPY . .

# Build
RUN ARCH=${ARCH} OS=${OS} make build


FROM alpine

WORKDIR /
COPY --from=builder /workspace/bin/modelxd-${OS}-${ARCH} /bin/modelxd
COPY --from=builder /workspace/bin/modelx-${OS}-${ARCH} /bin/modelx

USER nobody:nobody

LABEL org.opencontainers.image.source="https://github.com/kubeservice-stack/modelx" \
    org.opencontainers.image.url="https://stack.kubeservice.cn/" \
    org.opencontainers.image.documentation="https://stack.kubeservice.cn/" \
    org.opencontainers.image.licenses="Apache-2.0"
    
WORKDIR /app
ENTRYPOINT ["/bin/modelxd"]