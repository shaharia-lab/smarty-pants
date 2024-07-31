FROM scratch
COPY smarty-pants /app/smarty-pants
COPY migrations /app/migrations
ENTRYPOINT ["/app/smarty-pants", "start"]