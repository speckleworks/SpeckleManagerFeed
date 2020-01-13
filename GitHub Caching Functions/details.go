// Package p contains an HTTP Cloud Function.
package p

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"cloud.google.com/go/storage"
)
func read(client *storage.Client, bucket, object string) ([]byte, error) {
	ctx := context.Background()

	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func write(client *storage.Client, bucket, object string, r io.Reader) error {
	ctx := context.Background()

	wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
	if _, err := io.Copy(wc, r); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return nil
}

func GetDetails(w http.ResponseWriter, r *http.Request) {

	var response []byte
	var err error

	owner := r.FormValue("owner")
	name := r.FormValue("name")

	if owner == "" {
		w.Write([]byte("'owner' is required"))
		return
	}

	if name == "" {
		w.Write([]byte("'name' is required"))
		return
	}

	refresh := false
	//optional query param 'save'
	refreshStr := r.FormValue("refresh")
	if refreshStr != "" {
		refresh, err = strconv.ParseBool(refreshStr)
		if err != nil {
			log.Fatal(err)
		}
	}

	object := owner + "/" + name + "_details.json"
	bucket := "speckle-manager-storage"

	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
		return
	}

	//get object, if it doesn't exist, force refresh
	response, err = read(client, bucket, object)
	if err != nil {
		//log.Fatalf("Cannot read object: %v", err)
		refresh = true
	}

	if refresh {
		api := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, name)
		resp, err := http.Get(api)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		response, err = ioutil.ReadAll(resp.Body)

		r := bytes.NewReader(response)

		err = write(client, bucket, object, r)
		if err != nil {
			log.Fatal(err)
		}
	}

	w.Write(response)

}
