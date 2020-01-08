/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nginx

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

// MaxmindLicenseKey maxmind license key to download databases
var MaxmindLicenseKey = ""

const (
	geoIPPath = "/etc/nginx/geoip"

	geoLiteCityDB = "GeoLite2-City"
	geoLiteASNDB  = "GeoLite2-ASN"

	dbExtension = ".mmdb"

	maxmindURL = "https://download.maxmind.com/app/geoip_download?license_key=%v&edition_id=%v&suffix=tar.gz"
)

// GeoLite2DBExists checks if the required databases for
// the GeoIP2 NGINX module are present in the filesystem
func GeoLite2DBExists() bool {
	if !fileExists(path.Join(geoIPPath, geoLiteASNDB+dbExtension)) {
		return false
	}

	if !fileExists(path.Join(geoIPPath, geoLiteCityDB+dbExtension)) {
		return false
	}

	return true
}

// DownloadGeoLite2DB downloads the required databases by the
// GeoIP2 NGINX module using a license key from MaxMind.
func DownloadGeoLite2DB() error {
	err := downloadDatabase(geoLiteCityDB)
	if err != nil {
		return err
	}

	err = downloadDatabase(geoLiteASNDB)
	if err != nil {
		return err
	}

	return nil
}

func downloadDatabase(dbName string) error {
	url := fmt.Sprintf(maxmindURL, MaxmindLicenseKey, dbName)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status %v", resp.Status)
	}

	archive, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer archive.Close()

	mmdbFile := dbName + dbExtension

	tarReader := tar.NewReader(archive)
	for true {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeReg:
			if !strings.HasSuffix(header.Name, mmdbFile) {
				continue
			}

			outFile, err := os.Create(path.Join(geoIPPath, mmdbFile))
			if err != nil {
				return err
			}

			defer outFile.Close()

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}

			return nil
		}
	}

	return fmt.Errorf("the URL %v does not contains the database %v",
		fmt.Sprintf(maxmindURL, "XXXXXXX", dbName), mmdbFile)
}

func fileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}
