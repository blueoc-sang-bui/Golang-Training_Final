# Blog API
```
## Các endpoint chính
- `GET /posts/search-by-tag?tag=<tag_name>`: Tìm bài viết theo tag
- `POST /posts`: Tạo bài viết mới (transaction log)
- `GET /posts/:id`: Lấy chi tiết bài viết (cache-aside)
- `PUT /posts/:id`: Cập nhật bài viết (cache invalidation)
- `GET /posts/search?q=<query_string>`: Tìm kiếm full-text

## Kiểm tra API
```bash
# Tìm bài viết theo tag
curl 'http://localhost:8080/posts/search-by-tag?tag=golang'

# Tạo bài viết mới
curl -X POST 'http://localhost:8080/posts' -H 'Content-Type: application/json' -d '{"title":"Test","content":"Hello","tags":["golang","api"]}'

# Lấy chi tiết bài viết
curl 'http://localhost:8080/posts/1'

# Cập nhật bài viết
curl -X PUT 'http://localhost:8080/posts/1' -H 'Content-Type: application/json' -d '{"title":"Updated","content":"New content","tags":["golang"]}'

# Tìm kiếm full-text
curl 'http://localhost:8080/posts/search?q=Hello'
```

## Cấu trúc thư mục
- `main.go`: Entrypoint API
- `internal/handlers`: Xử lý endpoint
- `internal/models`: DB logic
- `internal/cache`: Redis
- `internal/search`: Elasticsearch
