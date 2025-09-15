FROM golang:1.24-alpine
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o blog-api .
EXPOSE 8080
CMD ["./blog-api"]
