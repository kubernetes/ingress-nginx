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
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"syscall"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	klog "k8s.io/klog/v2"
)

// MaxmindLicenseKey maxmind license key to download databases
var MaxmindLicenseKey = ""

// MaxmindEditionIDs maxmind editions (GeoLite2-City, GeoLite2-Country, GeoIP2-ISP, etc)
var MaxmindEditionIDs = ""

// MaxmindEditionFiles maxmind databases on disk
var MaxmindEditionFiles []string

// MaxmindMirror maxmind database mirror url (http://geoip.local)
var MaxmindMirror = ""

// MaxmindRetriesCount number of attempts to download the GeoIP DB
var MaxmindRetriesCount = 1

// MaxmindRetriesTimeout maxmind download retries timeout in seconds, 0 - do not retry to download if something went wrong
var MaxmindRetriesTimeout = time.Second * 0

// minimumRetriesCount minimum value of the MaxmindRetriesCount parameter. If MaxmindRetriesCount less than minimumRetriesCount, it will be set to minimumRetriesCount
const minimumRetriesCount = 1

const (
	geoIPPath   = "/etc/nginx/geoip"
	dbExtension = ".mmdb"

	maxmindURL = "https://download.maxmind.com/app/geoip_download?license_key=%v&edition_id=%v&suffix=tar.gz"
)

// GeoLite2DBExists checks if the required databases for
// the GeoIP2 NGINX module are present in the filesystem
// and indexes the discovered databases for iteration in
// the config.
func GeoLite2DBExists() bool {
	files := []string{}
	for _, dbName := range strings.Split(MaxmindEditionIDs, ",") {
		filename := dbName + dbExtension
		if !fileExists(path.Join(geoIPPath, filename)) {
			klog.Error(filename, " not found")
			return false
		}
		files = append(files, filename)
	}
	MaxmindEditionFiles = files

	return true
}

// DownloadGeoLite2DB downloads the required databases by the
// GeoIP2 NGINX module using a license key from MaxMind.
func DownloadGeoLite2DB(attempts int, period time.Duration) error {
	if attempts < minimumRetriesCount {
		attempts = minimumRetriesCount
	}

	defaultRetry := wait.Backoff{
		Steps:    attempts,
		Duration: period,
		Factor:   1.5,
		Jitter:   0.1,
	}
	if period == time.Duration(0) {
		defaultRetry.Steps = minimumRetriesCount
	}

	var lastErr error
	retries := 0

	_ = wait.ExponentialBackoff(defaultRetry, func() (bool, error) {
		var dlError error
		for _, dbName := range strings.Split(MaxmindEditionIDs, ",") {
			dlError = downloadDatabase(dbName)
			if dlError != nil {
				break
			}
		}

		lastErr = dlError
		if dlError == nil {
			return true, nil
		}

		if e, ok := dlError.(*url.Error); ok {
			if e, ok := e.Err.(*net.OpError); ok {
				if e, ok := e.Err.(*os.SyscallError); ok {
					if e.Err == syscall.ECONNREFUSED {
						retries++
						klog.InfoS("download failed on attempt " + fmt.Sprint(retries))
						return false, nil
					}
				}
			}
		}
		return true, nil
	})
	return lastErr
}

func createURL(mirror, licenseKey, dbName string) string {
	if len(mirror) > 0 {
		return fmt.Sprintf("%s/%s.tar.gz", mirror, dbName)
	}
	return fmt.Sprintf(maxmindURL, licenseKey, dbName)
}

func downloadDatabase(dbName string) error {
	url := createURL(MaxmindMirror, MaxmindLicenseKey, dbName)
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

			if _, err := io.CopyN(outFile, tarReader, header.Size); err != nil {
				return err
			}

			return nil
		}
	}

	return fmt.Errorf("the URL %v does not contains the database %v",
		fmt.Sprintf(maxmindURL, "XXXXXXX", dbName), mmdbFile)
}

// ValidateGeoLite2DBEditions check provided Maxmind database editions names
func ValidateGeoLite2DBEditions() error {
	allowedEditions := map[string]bool{
		"GeoIP2-Anonymous-IP":    true,
		"GeoIP2-Country":         true,
		"GeoIP2-City":            true,
		"GeoIP2-Connection-Type": true,
		"GeoIP2-Domain":          true,
		"GeoIP2-ISP":             true,
		"GeoIP2-ASN":             true,
		"GeoLite2-ASN":           true,
		"GeoLite2-Country":       true,
		"GeoLite2-City":          true,
	}

	for _, edition := range strings.Split(MaxmindEditionIDs, ",") {
		if !allowedEditions[edition] {
			return fmt.Errorf("unknown Maxmind GeoIP2 edition name: '%s'", edition)
		}
	}
	return nil
}

func _fileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

var fileExists = _fileExists
