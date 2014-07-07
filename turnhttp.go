// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package turnhttp

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func credentials(username, secret string) (user, pass string) {
	timestamp := time.Now().Unix()
	user = fmt.Sprintf("%d:%s", timestamp, username)
	h := hmac.New(sha1.New, []byte(secret))
	h.Write([]byte(user))
	pass = base64.StdEncoding.EncodeToString(h.Sum(nil))
	return
}

type TurnResponse struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Uris     []string `json:"uris"`
}

type Service struct {
	Domains []string
	Secret  string
	Uris    []string
}

func (self *Service) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	if username == "" {
		http.Error(rw, "Must supply a username.", http.StatusBadRequest)
		return
	}

	origin := r.Host
	for _, domain := range self.Domains {
		if origin == domain {
			user, pass := credentials(username, self.Secret)
			resp := TurnResponse{
				Username: user,
				Password: pass,
				Uris:     self.Uris,
			}

			rw.Header().Set("Content-Type", "application/json")
			rw.Header().Set("Access-Control-Allow-Origin", "*")
			rw.Header().Set("Access-Control-Allow-Methods", "GET")
			rw.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token")
			rw.Header().Set("Access-Control-Allow-Credentials", "true")
			enc := json.NewEncoder(rw)
			enc.Encode(resp)
			return
		}
	}
	http.Error(rw, "Invalid host.", http.StatusBadRequest)
	return

}
