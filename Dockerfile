# Stage 1: Build aplikasi dengan Go
FROM golang:1.23.6 AS builder

# Set working directory dalam container
WORKDIR /app

# Copy semua file ke dalam container
COPY . .

# Unduh dependency
RUN go mod tidy

# Build aplikasi
RUN go build -o todo-app-backend

# Stage 2: Gunakan Alpine dengan glibc
FROM frolvlad/alpine-glibc

# Set working directory
WORKDIR /root/

# Install dependencies yang diperlukan
RUN apk --no-cache add ca-certificates

# Copy binary dari stage builder
COPY --from=builder /app/todo-app-backend .

# Expose port aplikasi
EXPOSE 8080

# Jalankan aplikasi
CMD ["./todo-app-backend"]