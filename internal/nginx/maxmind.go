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
	"crypto/md5" //nolint:gosec // md5 is used for file integrity check
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

// MaxmindURL maxmind download url base
var MaxmindURL = ""

// MaxmindLicenseKey maxmind license key to download databases
var MaxmindLicenseKey = ""

// MaxmindAccountID maxmind account id
var MaxmindAccountID = 0

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

// MaxmindEnableSync enable periodic sync of maxmind databases
var MaxmindEnableSync = false

// MaxmindSyncPeriod maxmind databases sync period
var MaxmindSyncPeriod = time.Hour * 24

const (
	geoIPPath   = "/etc/ingress-controller/geoip"
	dbExtension = ".mmdb"

	maxmindURLFormat = "%s/geoip/databases/%s/update?db_md5=%s"
)

type FetchResult struct {
	Reader     io.ReadCloser
	EditionID  string
	OldHash    string
	NewHash    string
	ModifiedAt time.Time
}

// GeoLite2DBExists checks if the required databases for
// the GeoIP2 NGINX module are present in the filesystem
// and indexes the discovered databases for iteration in
// the config.
func GeoLite2DBExists() bool {
	files := []string{}
	for _, editionID := range strings.Split(MaxmindEditionIDs, ",") {
		filename := editionID + dbExtension
		if !fileExists(path.Join(geoIPPath, filename)) {
			klog.Error(filename, " not found")
			return false
		}
		files = append(files, filename)
	}
	MaxmindEditionFiles = files

	return true
}

// DownloadGeoLite2DBPeriodically starts a goroutine to periodically sync the GeoIP databases
func DownloadGeoLite2DBPeriodically(syncPeriod time.Duration) {
	go func() {
		for range time.Tick(syncPeriod) {
			if err := DownloadGeoLite2DB(MaxmindRetriesCount, MaxmindRetriesTimeout); err != nil {
				klog.ErrorS(err, "error syncing GeoIP databases")
			}
		}
	}()
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

	lastErr = wait.ExponentialBackoff(defaultRetry, func() (bool, error) {
		var dlError error
		for _, editionID := range strings.Split(MaxmindEditionIDs, ",") {
			dlError = downloadDatabase(editionID)
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

func downloadDatabase(editionID string) error {
	mmdbFile := editionID + dbExtension

	result, err := fetchDatabase(editionID)
	if err != nil {
		return err
	}
	defer func() {
		if result.Reader != nil {
			result.Reader.Close()
		}
	}()

	if MaxmindMirror == "" && strings.EqualFold(result.OldHash, result.NewHash) {
		return nil
	}

	tarReader := tar.NewReader(result.Reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if header.Typeflag == tar.TypeReg {
			if !strings.HasSuffix(header.Name, mmdbFile) {
				continue
			}
			return func() error {
				outFile, err := os.Create(path.Join(geoIPPath, mmdbFile))
				if err != nil {
					return err
				}

				defer outFile.Close()

				if _, err := io.CopyN(outFile, tarReader, header.Size); err != nil {
					return err
				}
				return nil
			}()
		}
	}

	return fmt.Errorf("no %s file in the db archive", mmdbFile)
}

func fetchDatabase(editionID string) (*FetchResult, error) {
	if MaxmindMirror != "" {
		return fetchDatabaseFromMirror(editionID)
	}
	return fetchDatabaseIfUpdated(editionID, calculateMD5Hash(editionID))
}

// backwards compatibility support - fetch directly from mirror without checking md5
// without md5 check and no auth
func fetchDatabaseFromMirror(editionID string) (result *FetchResult, err error) {
	mirrorURL := fmt.Sprintf("%s/%s.tar.gz", MaxmindMirror, editionID)
	resp, err := http.Get(mirrorURL) //nolint:gosec // URL is based on flag value
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %v", resp.Status)
	}

	archive, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return &FetchResult{
		Reader:    archive,
		EditionID: editionID,
	}, nil
}

func fetchDatabaseIfUpdated(editionID, md5hash string) (result *FetchResult, err error) {
	updateDatabaseURL := fmt.Sprintf(maxmindURLFormat, MaxmindURL, editionID, md5hash)
	req, err := http.NewRequest(http.MethodGet, updateDatabaseURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.SetBasicAuth(strconv.Itoa(MaxmindAccountID), MaxmindLicenseKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request to fetch db %v: %w", editionID, err)
	}

	defer func() {
		if err != nil {
			resp.Body.Close()
		}
	}()

	switch resp.StatusCode {
	case http.StatusNotModified:
		klog.InfoS("database is up to date", "db", editionID)
		return &FetchResult{
			EditionID: editionID,
			OldHash:   md5hash,
			NewHash:   md5hash,
		}, nil
	case http.StatusOK:
		klog.InfoS("downloading database", "db", editionID)
	default:
		return nil, fmt.Errorf("unexpected status code: %v", resp.Status)
	}

	newHash := resp.Header.Get("X-Database-MD5")
	if newHash == "" {
		return nil, fmt.Errorf("no md5 hash in response")
	}

	lastModified, err := time.ParseInLocation(time.RFC1123, resp.Header.Get("Last-Modified"), time.UTC)
	if err != nil {
		return nil, fmt.Errorf("parse last modified: %w", err)
	}

	archive, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("create gzip reader: %w", err)
	}

	return &FetchResult{
		Reader:     archive,
		EditionID:  editionID,
		OldHash:    md5hash,
		NewHash:    newHash,
		ModifiedAt: lastModified,
	}, nil
}

func calculateMD5Hash(editionID string) string {
	file, err := os.Open(path.Join(geoIPPath, editionID+dbExtension))
	if err != nil {
		return ""
	}

	defer file.Close()

	hash := md5.New() //nolint:gosec  // md5 is used for file integrity check
	if _, err := io.Copy(hash, file); err != nil {
		klog.ErrorS(err, "error calculating md5 hash")
		return ""
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
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
