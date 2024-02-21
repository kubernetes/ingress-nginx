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
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func resetForTesting() {
	fileExists = _fileExists
	MaxmindLicenseKey = ""
	MaxmindEditionIDs = ""
	MaxmindEditionFiles = []string{}
	MaxmindMirror = ""
}

// test cases from https://github.com/maxmind/geoipupdate/blob/388c52b2ee7a7dbf784aeaaa5024d1a7718a8196/pkg/geoipupdate/database/http_reader_test.go
func TestRequestDatabase(t *testing.T) {
	testTime := time.Date(2023, 4, 10, 12, 47, 31, 0, time.UTC)

	tests := []struct {
		description    string
		checkErr       func(require.TestingT, error, ...interface{})
		requestEdition string
		requestHash    string
		responseStatus int
		responseBody   string
		responseHash   string
		responseTime   string
		result         *FetchResult
	}{
		{
			description:    "success",
			checkErr:       require.NoError,
			requestEdition: "GeoIP2-City",
			requestHash:    "fbe1786bfd80e1db9dc42ddaff868f38",
			responseStatus: http.StatusOK,
			responseBody:   "database content",
			responseHash:   "cfa36ddc8279b5483a5aa25e9a6151f4",
			responseTime:   testTime.Format(time.RFC1123),
			result: &FetchResult{
				Reader:     mockArchiveReader(t, "database content"),
				EditionID:  "GeoIP2-City",
				OldHash:    "fbe1786bfd80e1db9dc42ddaff868f38",
				NewHash:    "cfa36ddc8279b5483a5aa25e9a6151f4",
				ModifiedAt: testTime,
			},
		}, {
			description:    "no new update",
			checkErr:       require.NoError,
			requestEdition: "GeoIP2-City",
			requestHash:    "fbe1786bfd80e1db9dc42ddaff868f38",
			responseStatus: http.StatusNotModified,
			responseBody:   "",
			responseHash:   "",
			responseTime:   "",
			result: &FetchResult{
				Reader:     nil,
				EditionID:  "GeoIP2-City",
				OldHash:    "fbe1786bfd80e1db9dc42ddaff868f38",
				NewHash:    "fbe1786bfd80e1db9dc42ddaff868f38",
				ModifiedAt: time.Time{},
			},
		}, {
			description:    "bad request",
			checkErr:       require.Error,
			requestEdition: "GeoIP2-City",
			requestHash:    "fbe1786bfd80e1db9dc42ddaff868f38",
			responseStatus: http.StatusBadRequest,
			responseBody:   "",
			responseHash:   "",
			responseTime:   "",
		}, {
			description:    "missing hash header",
			checkErr:       require.Error,
			requestEdition: "GeoIP2-City",
			requestHash:    "fbe1786bfd80e1db9dc42ddaff868f38",
			responseStatus: http.StatusOK,
			responseBody:   "database content",
			responseHash:   "",
			responseTime:   testTime.Format(time.RFC1123),
		}, {
			description:    "modified time header wrong format",
			checkErr:       require.Error,
			requestEdition: "GeoIP2-City",
			requestHash:    "fbe1786bfd80e1db9dc42ddaff868f38",
			responseStatus: http.StatusOK,
			responseBody:   "database content",
			responseHash:   "fbe1786bfd80e1db9dc42ddaff868f38",
			responseTime:   testTime.Format(time.Kitchen),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, _ *http.Request) {
						if test.responseStatus != http.StatusOK {
							w.WriteHeader(test.responseStatus)
							return
						}

						w.Header().Set("X-Database-MD5", test.responseHash)
						w.Header().Set("Last-Modified", test.responseTime)

						buf := &bytes.Buffer{}
						gzWriter := gzip.NewWriter(buf)
						_, err := gzWriter.Write([]byte(test.responseBody))
						require.NoError(t, err)
						require.NoError(t, gzWriter.Flush())
						require.NoError(t, gzWriter.Close())
						_, err = w.Write(buf.Bytes())
						require.NoError(t, err)
					},
				),
			)
			defer server.Close()

			MaxmindURL = server.URL

			result, err := fetchDatabaseIfUpdated(test.requestEdition, test.requestHash)
			test.checkErr(t, err)
			if err == nil {
				require.Equal(t, result.EditionID, test.result.EditionID)
				require.Equal(t, result.OldHash, test.result.OldHash)
				require.Equal(t, result.NewHash, test.result.NewHash)
				require.Equal(t, result.ModifiedAt, test.result.ModifiedAt)

				if test.result.Reader != nil && result.Reader != nil {
					defer result.Reader.Close()
					defer test.result.Reader.Close()
					resultDatabase, err := io.ReadAll(test.result.Reader)
					require.NoError(t, err)
					expectedDatabase, err := io.ReadAll(result.Reader)
					require.NoError(t, err)
					require.Equal(t, expectedDatabase, resultDatabase)
				}
			}
		})
	}
}

func mockArchiveReader(t *testing.T, data string) io.ReadCloser {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write([]byte(data))
	require.NoError(t, err)
	require.NoError(t, gz.Close())
	require.NoError(t, gz.Flush())
	r, err := gzip.NewReader(&buf)
	require.NoError(t, err)
	return r
}

func TestGeoLite2DBExists(t *testing.T) {
	tests := []struct {
		name      string
		setup     func()
		want      bool
		wantFiles []string
	}{
		{
			name:      "empty",
			wantFiles: []string{},
		},
		{
			name: "existing files",
			setup: func() {
				MaxmindEditionIDs = "GeoLite2-City,GeoLite2-ASN"
				fileExists = func(string) bool {
					return true
				}
			},
			want:      true,
			wantFiles: []string{"GeoLite2-City.mmdb", "GeoLite2-ASN.mmdb"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetForTesting()
			// mimics assignment in flags.go
			config := &MaxmindEditionFiles

			if tt.setup != nil {
				tt.setup()
			}
			if got := GeoLite2DBExists(); got != tt.want {
				t.Errorf("GeoLite2DBExists() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(MaxmindEditionFiles, tt.wantFiles) {
				t.Errorf("nginx.MaxmindEditionFiles = %v, want %v", MaxmindEditionFiles, tt.wantFiles)
			}
			if !reflect.DeepEqual(*config, tt.wantFiles) {
				t.Errorf("config.MaxmindEditionFiles = %v, want %v", *config, tt.wantFiles)
			}
		})
	}
}
