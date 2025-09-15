package handlers

import (
	"net/http"
	"encoding/json"
	"database/sql"
	"os"
	"fmt"
	"bytes"
	"context"
	"io"
	"github.com/lib/pq"
	"github.com/gorilla/mux"
	"blogapi/internal/cache"
	"blogapi/internal/search"
)

func SearchByTag(w http.ResponseWriter, r *http.Request) {
	   vars := mux.Vars(r)
	   tag, ok := vars["tag"]
	   if !ok || tag == "" {
		   http.Error(w, "Missing tag parameter", http.StatusBadRequest)
		   return
	   }

	dsn := os.Getenv("POSTGRES_DSN")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		http.Error(w, "DB connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query(`SELECT id, title, content, tags, created_at FROM posts WHERE tags @> ARRAY[$1]::text[]`, tag)
	if err != nil {
		http.Error(w, "Query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id int
		var title, content string
		var tags []string
		var createdAt string
		err := rows.Scan(&id, &title, &content, pq.Array(&tags), &createdAt)
		if err != nil {
			http.Error(w, "Scan error", http.StatusInternalServerError)
			return
		}
		results = append(results, map[string]interface{}{
			"id": id,
			"title": title,
			"content": content,
			"tags": tags,
			"created_at": createdAt,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Title   string   `json:"title"`
		Content string   `json:"content"`
		Tags    []string `json:"tags"`
	}
	var body reqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	dsn := os.Getenv("POSTGRES_DSN")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		http.Error(w, "DB connection error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Transaction error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var postID int
	err = tx.QueryRow(`INSERT INTO posts (title, content, tags) VALUES ($1, $2, $3) RETURNING id`, body.Title, body.Content, pq.Array(body.Tags)).Scan(&postID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Insert post error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`INSERT INTO activity_logs (action, post_id) VALUES ($1, $2)`, "new_post", postID)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Insert log error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Commit error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Index vào Elasticsearch
	esClient := search.GetESClient()
	doc := map[string]interface{}{
		"id": postID,
		"title": body.Title,
		"content": body.Content,
	}
	docJSON, _ := json.Marshal(doc)
	esClient.Index("posts", bytes.NewReader(docJSON))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": postID,
		"message": "Post created and logged successfully",
	})
}

func GetPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		id = r.URL.Query().Get("id")
	}
	if id == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	redisClient := cache.GetRedisClient()
	ctx := context.Background()
	cacheKey := "post:" + id
	cached, err := redisClient.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}

	dsn := os.Getenv("POSTGRES_DSN")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		http.Error(w, "DB connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var post struct {
		ID        int      `json:"id"`
		Title     string   `json:"title"`
		Content   string   `json:"content"`
		Tags      []string `json:"tags"`
		CreatedAt string   `json:"created_at"`
	}
	err = db.QueryRow(`SELECT id, title, content, tags, created_at FROM posts WHERE id = $1`, id).Scan(&post.ID, &post.Title, &post.Content, pq.Array(&post.Tags), &post.CreatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Query error", http.StatusInternalServerError)
		return
	}
	resp, _ := json.Marshal(post)
	// Set cache with TTL 5 phút
	redisClient.Set(ctx, cacheKey, string(resp), 300000000000).Err() // 5 phút = 300s = 300_000_000_000ns
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		id = r.URL.Query().Get("id")
	}
	if id == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}
	type reqBody struct {
		Title   string   `json:"title"`
		Content string   `json:"content"`
		Tags    []string `json:"tags"`
	}
	var body reqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	dsn := os.Getenv("POSTGRES_DSN")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		http.Error(w, "DB connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	res, err := db.Exec(`UPDATE posts SET title = $1, content = $2, tags = $3 WHERE id = $4`, body.Title, body.Content, pq.Array(body.Tags), id)
	if err != nil {
		http.Error(w, "Update error", http.StatusInternalServerError)
		return
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}
	redisClient := cache.GetRedisClient()
	ctx := context.Background()
	cacheKey := "post:" + id
	redisClient.Del(ctx, cacheKey)
	esClient := search.GetESClient()
	doc := map[string]interface{}{
		"id": id,
		"title": body.Title,
		"content": body.Content,
	}
	docJSON, _ := json.Marshal(doc)
	esClient.Index("posts", bytes.NewReader(docJSON))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": id,
		"message": "Post updated successfully",
	})
}

func SearchPosts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "Missing q parameter", http.StatusBadRequest)
		return
	}

	esClient := search.GetESClient()
	var buf bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query": q,
				"fields": []string{"title", "content"},
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		http.Error(w, "Query encoding error", http.StatusInternalServerError)
		return
	}
	   res, err := esClient.Search(
		   esClient.Search.WithIndex("posts"),
		   esClient.Search.WithBody(&buf),
		   esClient.Search.WithTrackTotalHits(true),
	   )
	   if err != nil {
		   http.Error(w, "Elasticsearch error", http.StatusInternalServerError)
		   return
	   }
	   defer res.Body.Close()
	   if res.IsError() {
		   body, _ := io.ReadAll(res.Body)
		   if bytes.Contains(body, []byte("index_not_found_exception")) {
			   esClient := search.GetESClient()
			   esClient.Indices.Create("posts")
		   }
		   http.Error(w, "Query error: "+string(body), http.StatusInternalServerError)
		   return
	   }
	var esResp struct {
		Hits struct {
			Hits []struct {
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		http.Error(w, "Decode error", http.StatusInternalServerError)
		return
	}
	var results []map[string]interface{}
	for _, hit := range esResp.Hits.Hits {
		results = append(results, hit.Source)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
