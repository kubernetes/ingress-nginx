/*
Copyright 2017 The Kubernetes Authors.

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

package errors

import "testing"

func TestIsLocationDenied(t *testing.T) {
	err := NewLocationDenied("demo")
	if !IsLocationDenied(err) {
		t.Error("expected true")
	}
	if IsLocationDenied(nil) {
		t.Error("expected false")
	}
}

func TestIsMissingAnnotations(t *testing.T) {
	if !IsMissingAnnotations(ErrMissingAnnotations) {
		t.Error("expected true")
	}
}

func TestInvalidContent(t *testing.T) {
	if IsInvalidContent(ErrMissingAnnotations) {
		t.Error("expected false")
	}
	err := NewInvalidAnnotationContent("demo", "")
	if !IsInvalidContent(err) {
		t.Error("expected true")
	}
	if IsInvalidContent(nil) {
		t.Error("expected false")
	}
	err = NewLocationDenied("demo")
	if IsInvalidContent(err) {
		t.Error("expected false")
	}
}
