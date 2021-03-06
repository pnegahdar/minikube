/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package kubernetes_versions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type URLHandlerCorrect struct {
	K8sReleases k8sReleases
}

func (h *URLHandlerCorrect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(h.K8sReleases)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/javascript")
	fmt.Fprintf(w, string(b))
}

func TestGetK8sVersionsFromURLCorrect(t *testing.T) {
	// test that the version is correctly parsed if returned if valid JSON is returned the url endpoint
	version0 := "0.0.0"
	version1 := "1.0.0"
	handler := &URLHandlerCorrect{
		K8sReleases: []k8sRelease{{Version: version0}, {Version: version1}},
	}
	server := httptest.NewServer(handler)

	k8sVersions, err := getK8sVersionsFromURL(server.URL)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if len(k8sVersions) != 2 { // TODO(aprindle) change to len(handler....)
		//Check values here as well?  Write eq method?
		t.Fatalf("Expected two kubernetes versions from URL to be %s, it was instead %s", 2, len(k8sVersions))
	}
}

type URLHandlerNone struct{}

func (h *URLHandlerNone) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func TestGetK8sVersionsFromURLNone(t *testing.T) {
	// test that an error is returned if nothing is returned at the url endpoint
	handler := &URLHandlerNone{}
	server := httptest.NewServer(handler)

	_, err := getK8sVersionsFromURL(server.URL)
	if err == nil {
		t.Fatalf("No kubernetes versions were returned from URL but no error was thrown")
	}
}

type URLHandlerMalformed struct{}

func (h *URLHandlerMalformed) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	fmt.Fprintf(w, "Malformed JSON")
}

func TestGetK8sVersionsFromURLMalformed(t *testing.T) {
	// test that an error is returned if malformed JSON is at the url endpoint
	handler := &URLHandlerMalformed{}
	server := httptest.NewServer(handler)

	_, err := getK8sVersionsFromURL(server.URL)
	if err == nil {
		t.Fatalf("Malformed version value was returned from URL but no error was thrown")
	}
}

func TestPrintKubernetesVersions(t *testing.T) {
	// test that no kubernetes version text is printed if there are no versions being served
	// TODO(aprindle) or should this be an error?!?!
	handlerNone := &URLHandlerNone{}
	server := httptest.NewServer(handlerNone)

	var outputBuffer bytes.Buffer
	PrintKubernetesVersions(&outputBuffer, server.URL)
	if len(outputBuffer.String()) != 0 {
		t.Fatalf("Expected PrintKubernetesVersions to not output text as there are no versioned served at the current URL but output was [%s]", outputBuffer.String())
	}

	// test that update text is printed if the latest version is greater than the current version
	// k8sVersionsFromURL = "100.0.0-dev"
	version0 := "0.0.0"
	version1 := "1.0.0"
	handlerCorrect := &URLHandlerCorrect{
		K8sReleases: []k8sRelease{{Version: version0}, {Version: version1}},
	}
	server = httptest.NewServer(handlerCorrect)

	PrintKubernetesVersions(&outputBuffer, server.URL)
	if len(outputBuffer.String()) == 0 {
		t.Fatalf("Expected PrintKubernetesVersion to output text as %s versions were served from URL but output was [%s]",
			2, outputBuffer.String()) //TODO(aprindle) change the 2
	}
}
