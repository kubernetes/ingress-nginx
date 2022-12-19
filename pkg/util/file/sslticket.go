/*
Copyright 2022 The Kubernetes Authors.

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

package file

import (
	"encoding/base64"
	"fmt"
	"os"
)

func WriteSSLTicketFile(ticket string, filename string) error {
	if ticket == "" || filename == "" {
		return fmt.Errorf("ticket and filename cannot be empty")
	}
	ticketBytes := base64.StdEncoding.WithPadding(base64.StdPadding).DecodedLen(len(ticket))

	// 81 used instead of 80 because of padding
	if !(ticketBytes == 48 || ticketBytes == 81) {
		return fmt.Errorf("ssl-session-ticket-key must contain either 48 or 80 bytes")
	}

	decodedTicket, err := base64.StdEncoding.DecodeString(ticket)
	if err != nil {
		return fmt.Errorf("unexpected error decoding ssl-session-ticket-key: %v", err)
	}

	err = os.WriteFile(filename, decodedTicket, ReadWriteByUser)
	if err != nil {
		return fmt.Errorf("unexpected error writing ssl-session-ticket-key to %s: %v", filename, err)
	}
	return nil
}
