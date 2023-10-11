FROM scratch
COPY ./octi-sync-server /usr/local/bin/octi-sync-server
ENTRYPOINT ["/usr/local/bin/octi-sync-server"]
