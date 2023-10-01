# Build golang backend 
FROM golang:alpine AS build-backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go build -o lostsonstv

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=build-backend /app/lostsonstv ./
COPY static ./static
EXPOSE 80
CMD ["./lostsonstv"]