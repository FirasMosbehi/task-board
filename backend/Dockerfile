# Runtime-only image (no build stage needed)
FROM alpine:3.20

# Install CA certificates (important for HTTPS calls)
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy the binary produced by the GitHub Actions CI
COPY taskboard-backend /app/taskboard-backend

# Optional: make sure the binary is executable
RUN chmod +x /app/taskboard-backend

# Environment variables
ENV PORT=8080

EXPOSE 8080

CMD ["/app/taskboard-backend"]
