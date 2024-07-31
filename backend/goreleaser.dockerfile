# This dockerfile is only used to build the backend image for the application using goreleaser.
FROM scratch
COPY smarty-pants /app/smarty-pants
COPY migrations /app/migrations
ENTRYPOINT ["/app/smarty-pants", "start"]
