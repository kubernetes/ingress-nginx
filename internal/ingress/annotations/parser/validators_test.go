/*
Copyright 2023 The Kubernetes Authors.

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

package parser

import (
	"fmt"
	"testing"

	networking "k8s.io/api/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidateArrayOfServerName(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "should accept common name",
			value:   "something.com,anything.com",
			wantErr: false,
		},
		{
			name:    "should accept wildcard name",
			value:   "*.something.com,otherthing.com",
			wantErr: false,
		},
		{
			name:    "should allow names with spaces between array and some regexes",
			value:   `~^www\d+\.example\.com$,something.com`,
			wantErr: false,
		},
		{
			name:    "should allow names with regexes",
			value:   `http://some.test.env.com:2121/$someparam=1&$someotherparam=2`,
			wantErr: false,
		},
		{
			name:    "should allow names with wildcard in middle common name",
			value:   "*.so*mething.com,bla.com",
			wantErr: false,
		},
		{
			name:    "should allow comma separated query params",
			value:   "https://oauth.example/oauth2/auth?allowed_groups=gid1,gid2",
			wantErr: false,
		},
		{
			name:    "should deny names with weird characters",
			value:   "something.com,lolo;xpto.com,nothing.com",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateArrayOfServerName(tt.value); (err != nil) != tt.wantErr {
				t.Errorf("ValidateArrayOfServerName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_checkAnnotation(t *testing.T) {
	type args struct {
		name   string
		ing    *networking.Ingress
		fields AnnotationFields
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "null ingress should error",
			want: "",
			args: args{
				name: "some-random-annotation",
			},
			wantErr: true,
		},
		{
			name: "not having a validator for a specific annotation is a bug",
			want: "",
			args: args{
				name: "some-new-invalid-annotation",
				ing: &networking.Ingress{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							GetAnnotationWithPrefix("some-new-invalid-annotation"): "xpto",
						},
					},
				},
				fields: AnnotationFields{
					"otherannotation": AnnotationConfig{
						Validator: func(_ string) error { return nil },
					},
				},
			},
			wantErr: true,
		},
		{
			name: "annotationconfig found and no validation func defined on annotation is a bug",
			want: "",
			args: args{
				name: "some-new-invalid-annotation",
				ing: &networking.Ingress{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							GetAnnotationWithPrefix("some-new-invalid-annotation"): "xpto",
						},
					},
				},
				fields: AnnotationFields{
					"some-new-invalid-annotation": AnnotationConfig{},
				},
			},
			wantErr: true,
		},
		{
			name: "no annotation can turn into a null pointer and should fail",
			want: "",
			args: args{
				name: "some-new-invalid-annotation",
				ing: &networking.Ingress{
					ObjectMeta: v1.ObjectMeta{},
				},
				fields: AnnotationFields{
					"some-new-invalid-annotation": AnnotationConfig{},
				},
			},
			wantErr: true,
		},
		{
			name: "no AnnotationField config should bypass validations",
			want: GetAnnotationWithPrefix("some-valid-annotation"),
			args: args{
				name: "some-valid-annotation",
				ing: &networking.Ingress{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							GetAnnotationWithPrefix("some-valid-annotation"): "xpto",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "annotation with invalid value should fail",
			want: "",
			args: args{
				name: "some-new-annotation",
				ing: &networking.Ingress{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							GetAnnotationWithPrefix("some-new-annotation"): "xpto1",
						},
					},
				},
				fields: AnnotationFields{
					"some-new-annotation": AnnotationConfig{
						Validator: func(value string) error {
							if value != "xpto" {
								return fmt.Errorf("this is an error")
							}
							return nil
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "annotation with valid value should pass",
			want: GetAnnotationWithPrefix("some-other-annotation"),
			args: args{
				name: "some-other-annotation",
				ing: &networking.Ingress{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							GetAnnotationWithPrefix("some-other-annotation"): "xpto",
						},
					},
				},
				fields: AnnotationFields{
					"some-other-annotation": AnnotationConfig{
						Validator: func(value string) error {
							if value != "xpto" {
								return fmt.Errorf("this is an error")
							}
							return nil
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkAnnotation(tt.args.name, tt.args.ing, tt.args.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkAnnotation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("checkAnnotation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckAnnotationRisk(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		maxrisk     AnnotationRisk
		config      AnnotationFields
		wantErr     bool
	}{
		{
			name:    "high risk should not be accepted with maximum medium",
			maxrisk: AnnotationRiskMedium,
			annotations: map[string]string{
				"nginx.ingress.kubernetes.io/bla": "blo",
				"nginx.ingress.kubernetes.io/bli": "bl3",
			},
			config: AnnotationFields{
				"bla": {
					Risk: AnnotationRiskHigh,
				},
				"bli": {
					Risk: AnnotationRiskMedium,
				},
			},
			wantErr: true,
		},
		{
			name:    "high risk should  be accepted with maximum critical",
			maxrisk: AnnotationRiskCritical,
			annotations: map[string]string{
				"nginx.ingress.kubernetes.io/bla": "blo",
				"nginx.ingress.kubernetes.io/bli": "bl3",
			},
			config: AnnotationFields{
				"bla": {
					Risk: AnnotationRiskHigh,
				},
				"bli": {
					Risk: AnnotationRiskMedium,
				},
			},
			wantErr: false,
		},
		{
			name:    "low risk should  be accepted with maximum low",
			maxrisk: AnnotationRiskLow,
			annotations: map[string]string{
				"nginx.ingress.kubernetes.io/bla": "blo",
				"nginx.ingress.kubernetes.io/bli": "bl3",
			},
			config: AnnotationFields{
				"bla": {
					Risk: AnnotationRiskLow,
				},
				"bli": {
					Risk: AnnotationRiskLow,
				},
			},
			wantErr: false,
		},
		{
			name:    "critical risk should  be accepted with maximum critical",
			maxrisk: AnnotationRiskCritical,
			annotations: map[string]string{
				"nginx.ingress.kubernetes.io/bla": "blo",
				"nginx.ingress.kubernetes.io/bli": "bl3",
			},
			config: AnnotationFields{
				"bla": {
					Risk: AnnotationRiskCritical,
				},
				"bli": {
					Risk: AnnotationRiskCritical,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckAnnotationRisk(tt.annotations, tt.maxrisk, tt.config); (err != nil) != tt.wantErr {
				t.Errorf("CheckAnnotationRisk() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommonNameAnnotationValidator(t *testing.T) {
	tests := []struct {
		name       string
		annotation string
		wantErr    bool
	}{
		{
			name:       "correct example",
			annotation: `CN=(my\.common\.name)`,
			wantErr:    false,
		},
		{
			name:       "no CN= prefix",
			annotation: `(my\.common\.name)`,
			wantErr:    true,
		},
		{
			name:       "invalid prefix",
			annotation: `CN(my\.common\.name)`,
			wantErr:    true,
		},
		{
			name:       "invalid regex",
			annotation: `CN=(my\.common\.name]`,
			wantErr:    true,
		},
		{
			name:       "wildcard regex",
			annotation: `CN=(my\..*\.name)`,
			wantErr:    false,
		},
		{
			name:       "somewhat complex regex",
			annotation: "CN=(my\\.app\\.dev|.*\\.bbb\\.aaaa\\.tld)",
			wantErr:    false,
		},
		{
			name:       "another somewhat complex regex",
			annotation: `CN=(my-app.*\.c\.defg\.net|other.app.com)`,
			wantErr:    false,
		},
		{
			name:       "nested parenthesis regex",
			annotation: `CN=(api-one\.(asdf)?qwer\.webpage\.organization\.org)`,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CommonNameAnnotationValidator(tt.annotation); (err != nil) != tt.wantErr {
				t.Errorf("CommonNameAnnotationValidator() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
