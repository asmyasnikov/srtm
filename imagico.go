package srtm

import (
	"archive/zip"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var client = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
	Timeout: time.Second * 10,
}

func parse(r io.Reader) (string, error) {
	var v []interface{}
	if err := json.NewDecoder(r).Decode(&v); err != nil {
		fmt.Println("error on decode json", err)
		return "", err
	}
	if len(v) == 0 {
		return "", fmt.Errorf("No tiles found (%+v)", v)
	}
	fields, ok := v[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Cannot cast (%+v) to map[string]interface{}", v[0])
	}
	field, ok := fields["link"]
	if !ok {
		return "", fmt.Errorf("Not found 'link' field in %+v", fields)
	}
	link, ok := field.(string)
	if !ok {
		return "", fmt.Errorf("Cannot cast (%+v) to string", field)
	}
	return link, nil
}

func search(ll LatLng) (string, error) {
	r, err := client.Get(fmt.Sprintf("http://www.imagico.de/map/dem_json.php?date=&lon=%0.7f&lat=%0.7f&lonE=%0.7f&latE=%0.7f&vf=1", ll.Longitude, ll.Latitude, ll.Longitude, ll.Latitude))
	if err != nil {
		fmt.Println("error on GET", err)
		return "", err
	}
	if r.StatusCode != http.StatusOK {
		err = fmt.Errorf("status code for request '%s' is not Ok (%d)", r.Request.RequestURI, r.StatusCode)
		fmt.Println("error on GET", err)
		return "", err
	}
	defer r.Body.Close()
	return parse(r.Body)
}

func download(tileDir, key string, ll LatLng) (string, error) {
	url, err := search(ll)
	if err != nil {
		return "", err
	}
	targetDir := path.Join(os.TempDir(), key)
	err = os.Mkdir(targetDir, 0755)
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(targetDir)
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	b, err := ioutil.ReadAll(response.Body)
	zipReader, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return "", err
	}
	extracted := make([]string, 0)
	for _, file := range zipReader.File {
		zippedFile, err := file.Open()
		if err != nil {
			log.Fatal(err)
		}
		extractedFilePath := filepath.Join(
			targetDir,
			file.Name,
		)
		if file.FileInfo().IsDir() {
			os.MkdirAll(extractedFilePath, file.Mode())
		} else {
			outputFile, err := os.OpenFile(
				extractedFilePath,
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				file.Mode(),
			)
			if err != nil {
				fmt.Println("unzip:", err)
			}
			_, err = io.Copy(outputFile, zippedFile)
			if err != nil {
				fmt.Println("unzip:", err)
			}
			outputFile.Close()
			if err == nil {
				hgt := strings.ReplaceAll(extractedFilePath, targetDir, tileDir)
				extracted = append(extracted, hgt)
				if err := os.Rename(extractedFilePath, hgt); err != nil {
					fmt.Println("move: ", err)
				}
			}
		}
		zippedFile.Close()
	}
	for _, hgt := range extracted {
		if strings.Contains(hgt, key) {
			return hgt, nil
		}
	}
	return "", fmt.Errorf("tile file for key = %s is not exists", key)
}