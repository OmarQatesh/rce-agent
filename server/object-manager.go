package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"context"

	"github.com/minio/minio-go/v6"
)

// func main() {
// 	getClientReleases("clients", "wasfaty", "uat")
// }

func getClientReleases(clientsDir string, clientsID string, clientsEnv string) minio.ObjectInfo {
	// Note: YOUR-ACCESSKEYID, YOUR-SECRETACCESSKEY, my-bucketname and my-prefixname
	// are dummy values, please replace them with original values.

	// Requests are always secure (HTTPS) by default. Set secure=false to enable insecure (HTTP) access.
	// This boolean value is the last argument for New().

	// New returns an Amazon S3 compatible client object. API compatibility (v2 or v4) is automatically
	// determined based on the Endpoint value.
	// var minioURL string
	// var miniokey string
	// var minioSecret string
	minioURL := os.Getenv("MINIO_URL")
	if empty(minioURL) {
		minioURL = "10.162.193.225:32320"
	}
	miniokey := os.Getenv("MINIO_KEY")
	if empty(miniokey) {
		miniokey = "myaccesskey"
	}
	minioSecret := os.Getenv("MINIO_SECRET")
	if empty(minioSecret) {
		minioSecret = "mysecretkey"
	}

	s3Client, err := minio.New(minioURL, miniokey, minioSecret, false)
	if err != nil {
		fmt.Println(err)
		return minio.ObjectInfo{Err: err}
	}

	found, err := s3Client.BucketExists(clientsDir)
	if err != nil {
		log.Fatalln(err)
		return minio.ObjectInfo{Err: err}
	}

	if found {
		log.Println("Bucket found.")
		// Create a done channel to control 'ListObjects' go routine.
		doneCh := make(chan struct{})

		// Indicate to our routine to exit cleanly upon return.
		defer close(doneCh)

		// List all objects from a bucket-name with a matching prefix.
		for object := range s3Client.ListObjectsV2(clientsDir, clientsID+"/"+clientsEnv, true, doneCh) {
			if object.Err != nil {
				fmt.Println(object.Err)
				return object
			}
			fmt.Println(object)

			// stat, err := s3Client.StatObject(clientsDir, object.Key, minio.StatObjectOptions{})
			// if err != nil {
			// 	log.Fatalln(err)
			// }
			// log.Println(stat)

			object = getReleaseDataFile(object, *s3Client, clientsDir, object.Key)
			if object.Err != nil {
				fmt.Println(object.Err)
				return object
			}
			return object
		}

		log.Println("Client or env not found.")
		return minio.ObjectInfo{Err: errors.New("client or env not found")}
	} else {
		log.Println("Bucket not found.")
		return minio.ObjectInfo{Err: errors.New("bucket not found")}
	}
	log.Fatalln("Shit hit the fan!!!")
	return minio.ObjectInfo{Err: errors.New("shit hit the fan!!!")}
}

func getReleaseDataFile(object minio.ObjectInfo, s3Client minio.Client, clientsDir string, releaseFileName string) minio.ObjectInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	opts := minio.GetObjectOptions{}
	//opts.SetMatchETagExcept("downloaded")
	opts.SetUnmodified(time.Now().Round(100000 * time.Hour)) // get object if was unmodified within the last 10 minutes
	//opts.SetModified(time.Now().Round(10 * time.Minute))
	reader, err := s3Client.GetObjectWithContext(ctx, clientsDir, object.Key, opts)
	if err != nil {
		log.Fatalln(err)
		return minio.ObjectInfo{Err: err}
	}
	defer reader.Close()

	var fileName []string = strings.Split(releaseFileName, "/")
	var name string = fileName[len(fileName)-1]

	stat, err := reader.Stat()
	if err != nil {
		log.Fatalln(err)
		return minio.ObjectInfo{Err: err}
	}
	localFile, err := os.Create(name)
	if err != nil {
		log.Fatalln(err)
		return minio.ObjectInfo{Err: err}
	}
	defer localFile.Close()
	if _, err := io.CopyN(localFile, reader, stat.Size); err != nil {
		log.Fatalln(err)
		return minio.ObjectInfo{Err: err}
	}
	if _, err := s3Client.FPutObjectWithContext(ctx, clientsDir, object.Key, name, minio.PutObjectOptions{}); err != nil {
		log.Fatalln(err)
		return minio.ObjectInfo{Err: err}
	}
	log.Println("Successfully updated " + object.Key)
	return object
}

func empty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}
