FROM scratch
COPY smarty-pants /app/smarty-pants
ENTRYPOINT ["/app/smarty-pants", "start"]