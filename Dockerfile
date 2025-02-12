# Gunakan base image Go terbaru yang didukung
FROM golang:1.23.6 AS builder

# Set working directory dalam container
WORKDIR /app

# Copy semua file ke dalam container
COPY . .

# Unduh dependency
RUN go mod tidy

# Build aplikasi
RUN go build -o todo-app-backend

# Image yang lebih kecil untuk menjalankan aplikasi
FROM debian:bullseye-slim

# Set working directory
WORKDIR /root/

# Copy binary dari stage builder
COPY --from=builder /app/todo-app-backend .

# Expose port aplikasi
EXPOSE 8080

# Jalankan aplikasi
CMD ["./todo-app-backend"]