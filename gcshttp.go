package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"cloud.google.com/go/storage"
)

type MyHandler struct {
	bucket *storage.BucketHandle
	ctx    context.Context
}

func main() {
	bucket_name := os.Getenv("BUCKET_NAME")
	if bucket_name == "" {
		panic("BUCKET_NAME env not set")
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Println(err)
		panic("Failed to connect to storage")
	}
	defer client.Close()

	bucket := client.Bucket(bucket_name)

	http.Handle("/", MyHandler{bucket, ctx})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func (h MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		fmt.Fprint(w, "gcshttp")
		return
	}

	oh := h.bucket.Object(r.URL.Path[1:])
	objAttrs, err := oh.Attrs(h.ctx)
	if err != nil {
		log.Println(err, r.URL)
		http.NotFound(w, r)
		return
	}
	rc, err := oh.NewReader(h.ctx)
	if err != nil {
		log.Println(err, r.URL)
		http.NotFound(w, r)
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", objAttrs.ContentType)
	w.Header().Set("Content-Encoding", objAttrs.ContentEncoding)
	w.Header().Set("Content-Length", strconv.Itoa(int(objAttrs.Size)))
	w.WriteHeader(200)
	if _, err := io.Copy(w, rc); err != nil {
		log.Println(err, r.URL)
		return
	}
}
