package srtm

import (
	"archive/zip"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func client() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: time.Second * 10,
	}
}

func parse(r io.Reader) ([]string, error) {
	var v []interface{}
	if err := json.NewDecoder(r).Decode(&v); err != nil {
		log.Error().Caller().Err(err).Msg("decode json")
		return nil, err
	}
	if len(v) == 0 {
		return nil, fmt.Errorf("No tiles found (%+v)", v)
	}
	urls := make([]string, 0, len(v))
	for _, f := range v {
		fields, ok := f.(map[string]interface{})
		if !ok {
			continue
		}
		field, ok := fields["link"]
		if !ok {
			continue
		}
		link, ok := field.(string)
		if !ok {
			continue
		}
		urls = append(urls, link)
	}
	return urls, nil
}

func search(ll LatLng) ([]string, error) {
	r, err := client().Get(fmt.Sprintf("http://www.imagico.de/map/dem_json.php?date=&lon=%0.7f&lat=%0.7f&lonE=%0.7f&latE=%0.7f&vf=1", ll.Longitude, ll.Latitude, ll.Longitude, ll.Latitude))
	if err != nil {
		log.Error().Caller().Err(err).Msg("GET")
		return nil, err
	}
	if r.StatusCode != http.StatusOK {
		err = fmt.Errorf("status code for request '%s' is not Ok (%d)", r.Request.RequestURI, r.StatusCode)
		log.Error().Caller().Err(err).Msg("GET")
		return nil, err
	}
	defer r.Body.Close()
	return parse(r.Body)
}

func moveHgt(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("Couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("Writing to output file failed: %s", err)
	}
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("Failed removing original file: %s", err)
	}
	return nil
}

func downloadByURL(tileDir, url string) (extracted []string) {
	targetDir := path.Join(os.TempDir(), "srtm-" + strconv.Itoa(rand.Int()))
	err := os.Mkdir(targetDir, 0755)
	if err != nil {
		log.Error().Caller().Err(err).Msg("")
		return extracted
	}
	defer os.RemoveAll(targetDir)
	response, err := http.Get(url)
	if err != nil {
		log.Error().Caller().Err(err).Msg("")
		return extracted
	}
	defer response.Body.Close()
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error().Caller().Err(err).Msg("")
		return extracted
	}
	zipReader, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		log.Error().Caller().Err(err).Msg("")
		return extracted
	}
	for _, file := range zipReader.File {
		zippedFile, err := file.Open()
		if err != nil {
			log.Error().Caller().Err(err).Msg("unzip")
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
				log.Error().Caller().Err(err).Msg("open file")
			}
			_, err = io.Copy(outputFile, zippedFile)
			if err != nil {
				log.Error().Caller().Err(err).Msg("copy")
			}
			outputFile.Close()
			if err == nil {
				_, file := path.Split(extractedFilePath)
				hgt := path.Join(tileDir, file)
				extracted = append(extracted, hgt)
				if err := moveHgt(extractedFilePath, hgt); err != nil {
					log.Error().Caller().Err(err).Msg("move")
				}
			}
		}
		zippedFile.Close()
	}
	return extracted
}

func download(tileDir string, ll LatLng) (string, os.FileInfo, error) {
	key := tileKey(ll)
	urls, err := search(ll)
	if err != nil {
		return "", nil, err
	}
	extracted := make([]string, 0)
	for _, url := range urls {
		extracted = append(extracted, downloadByURL(tileDir, url)...)
	}
	for _, hgt := range extracted {
		if strings.Contains(hgt, key) {
			info, err := os.Stat(hgt)
			if err != nil {
				return "", nil, nil
			}
			return hgt, info, nil
		}
	}
	return "", nil, fmt.Errorf("tile file for key = %s is not exists (urls %+v -> %+v)", key, urls, extracted)
}
