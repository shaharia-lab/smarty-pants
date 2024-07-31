FROM scratch
COPY smarty-pants /app/smarty-pants
COPY migrations /app/migrations
CMD ["/app/smarty-pants"]