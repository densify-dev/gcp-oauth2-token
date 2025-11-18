FROM golang:bookworm AS builder
# Enable Docker BuildKit automatic platform ARGs
ARG TARGETARCH
ARG VERSION
ADD . /github.com/densify-dev/gcp-oauth2-token
WORKDIR /github.com/densify-dev/gcp-oauth2-token
RUN echo -n "v${VERSION}" > cmd/version.txt
RUN make build

FROM alpine:latest
ARG VERSION
ARG RELEASE
# Enable Docker BuildKit automatic platform ARGs for runtime stage
ARG TARGETARCH

LABEL name="gcp-oauth2-token" \
      vendor="Densify" \
      maintainer="support@densify.com" \
      version="${VERSION}" \
      release="${RELEASE}" \
      summary="Google Cloud Platform OAuth2 API token" \
      description="Gets a OAuth2 bearer token for usage with Google Cloud Platform REST API"

RUN addgroup -g 3000 densify && \
	adduser -h /home/densify -s /bin/sh -u 3000 -G densify -g "" -D densify && \
	chmod 755 /home/densify && \
	rm -f /sbin/apk

# Copy the compiled Go binary
WORKDIR /home/densify
COPY --chown=densify:densify --chmod=755 --from=builder /github.com/densify-dev/gcp-oauth2-token/build/$TARGETARCH/gcp-oauth2-token bin/
USER 3000

# Run the webhook server
ENTRYPOINT ["/home/densify/bin/gcp-oauth2-token"]
