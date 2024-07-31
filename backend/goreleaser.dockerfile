FROM scratch
COPY smarty-pants /app/smarty-pants
COPY backend/migrations /app/migrations
CMD ["/app/smarty-pants"]