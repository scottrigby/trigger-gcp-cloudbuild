package storage

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"cloud.google.com/go/storage"
	"github.com/mholt/archiver"
	"google.golang.org/api/googleapi"
)

// MakeTarball makes a context directory tarball.
func MakeTarball(dir string, fileName string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	// Convert to []string for archiver.TarGz.Make().
	filePaths := make([]string, len(files))
	for i, file := range files {
		filePaths[i] = dir + "/" + file.Name()
	}
	err2 := archiver.TarGz.Make(fileName, filePaths)
	if err2 != nil {
		return err2
	}
	return nil
}

// CreateBucket creates a storage bucket.
func CreateBucket(bucket *storage.BucketHandle, name string, attrs *storage.BucketAttrs) error {
	ctx := context.Background()
	err := bucket.Create(ctx, name, attrs)

	// If the bucket exists, return without error, but give a friendly message.
	// > Error 409: You already own this bucket. Please select another name., conflict
	if e, ok := err.(*googleapi.Error); ok && e.Code == http.StatusConflict {
		fmt.Println("storage bucket already exists")
		return nil
	}

	// If any other error return it.
	if err != nil {
		return err
	}
	return nil
}

// GetBucketHandle gets a storage bucket handle.
func GetBucketHandle(name string) (*storage.BucketHandle, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.Bucket(name), nil
}

// WriteToStorage writes a file to storage.
// ref: https://github.com/GoogleCloudPlatform/golang-samples/blob/master/getting-started/bookshelf/app/app.go#L218
func WriteToStorage(storageBucket *storage.BucketHandle, bucketName string, file string) (string, error) {
	ctx := context.Background()
	w := storageBucket.Object(file).NewWriter(ctx)
	w.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	w.ContentType = "application/gzip"
	w.CacheControl = "public, max-age=86400"

	in, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer in.Close()

	if _, err := io.Copy(w, in); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}

	const url = "https://storage.googleapis.com/%s/%s"
	return fmt.Sprintf(url, bucketName, file), nil
}
