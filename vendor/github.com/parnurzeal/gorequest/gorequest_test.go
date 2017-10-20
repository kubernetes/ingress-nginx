package gorequest

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"mime/multipart"

	"os"

	"github.com/elazarl/goproxy"
)

type (
	heyYou struct {
		Hey string `json:"hey"`
	}
	testStruct struct {
		String      string
		Int         int
		Btrue       bool
		Bfalse      bool
		Float       float64
		StringArray []string
		IntArray    []int
		BoolArray   []bool
		FloatArray  []float64
	}
)

// Test for changeMapToURLValues
func TestChangeMapToURLValues(t *testing.T) {

	data := map[string]interface{}{
		"s":  "a string",
		"i":  42,
		"bt": true,
		"bf": false,
		"f":  12.345,
		"sa": []string{"s1", "s2"},
		"ia": []int{47, 73},
		"fa": []float64{1.23, 4.56},
		"ba": []bool{true, false},
	}

	urlValues := changeMapToURLValues(data)

	var (
		s  string
		sd string
	)

	if s := urlValues.Get("s"); s != data["s"] {
		t.Errorf("Expected string %v, got %v", data["s"], s)
	}

	s = urlValues.Get("i")
	sd = strconv.Itoa(data["i"].(int))
	if s != sd {
		t.Errorf("Expected int %v, got %v", sd, s)
	}

	s = urlValues.Get("bt")
	sd = strconv.FormatBool(data["bt"].(bool))
	if s != sd {
		t.Errorf("Expected boolean %v, got %v", sd, s)
	}

	s = urlValues.Get("bf")
	sd = strconv.FormatBool(data["bf"].(bool))
	if s != sd {
		t.Errorf("Expected boolean %v, got %v", sd, s)
	}

	s = urlValues.Get("f")
	sd = strconv.FormatFloat(data["f"].(float64), 'f', -1, 64)
	if s != sd {
		t.Errorf("Expected float %v, got %v", data["f"], s)
	}

	// array cases
	// "To access multiple values, use the map directly."

	if size := len(urlValues["sa"]); size != 2 {
		t.Fatalf("Expected length %v, got %v", 2, size)
	}
	if urlValues["sa"][0] != "s1" {
		t.Errorf("Expected string %v, got %v", "s1", urlValues["sa"][0])
	}
	if urlValues["sa"][1] != "s2" {
		t.Errorf("Expected string %v, got %v", "s2", urlValues["sa"][1])
	}

	if size := len(urlValues["ia"]); size != 2 {
		t.Fatalf("Expected length %v, got %v", 2, size)
	}
	if urlValues["ia"][0] != "47" {
		t.Errorf("Expected string %v, got %v", "47", urlValues["ia"][0])
	}
	if urlValues["ia"][1] != "73" {
		t.Errorf("Expected string %v, got %v", "73", urlValues["ia"][1])
	}

	if size := len(urlValues["ba"]); size != 2 {
		t.Fatalf("Expected length %v, got %v", 2, size)
	}
	if urlValues["ba"][0] != "true" {
		t.Errorf("Expected string %v, got %v", "true", urlValues["ba"][0])
	}
	if urlValues["ba"][1] != "false" {
		t.Errorf("Expected string %v, got %v", "false", urlValues["ba"][1])
	}

	if size := len(urlValues["fa"]); size != 2 {
		t.Fatalf("Expected length %v, got %v", 2, size)
	}
	if urlValues["fa"][0] != "1.23" {
		t.Errorf("Expected string %v, got %v", "true", urlValues["fa"][0])
	}
	if urlValues["fa"][1] != "4.56" {
		t.Errorf("Expected string %v, got %v", "false", urlValues["fa"][1])
	}
}

// Test for Make request
func TestMakeRequest(t *testing.T) {
	var err error
	var cases = []struct {
		m string
		s *SuperAgent
	}{
		{POST, New().Post("/")},
		{GET, New().Get("/")},
		{HEAD, New().Head("/")},
		{PUT, New().Put("/")},
		{PATCH, New().Patch("/")},
		{DELETE, New().Delete("/")},
		{OPTIONS, New().Options("/")},
		{"TRACE", New().CustomMethod("TRACE", "/")}, // valid HTTP 1.1 method, see W3C RFC 2616
	}

	for _, c := range cases {
		_, err = c.s.MakeRequest()
		if err != nil {
			t.Errorf("Expected nil error for method %q; got %q", c.m, err.Error())
		}
	}

	// empty method should fail
	_, err = New().CustomMethod("", "/").MakeRequest()
	if err == nil {
		t.Errorf("Expected non-nil error for empty method; got %q", err.Error())
	}
}

// testing for Get method
func TestGet(t *testing.T) {
	const case1_empty = "/"
	const case2_set_header = "/set_header"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check method is GET before going to check other features
		if r.Method != GET {
			t.Errorf("Expected method %q; got %q", GET, r.Method)
		}
		if r.Header == nil {
			t.Error("Expected non-nil request Header")
		}
		switch r.URL.Path {
		default:
			t.Errorf("No testing for this case yet : %q", r.URL.Path)
		case case1_empty:
			t.Logf("case %v ", case1_empty)
		case case2_set_header:
			t.Logf("case %v ", case2_set_header)
			if r.Header.Get("API-Key") != "fookey" {
				t.Errorf("Expected 'API-Key' == %q; got %q", "fookey", r.Header.Get("API-Key"))
			}
		}
	}))

	defer ts.Close()

	New().Get(ts.URL + case1_empty).
		End()

	New().Get(ts.URL+case2_set_header).
		Set("API-Key", "fookey").
		End()
}

// testing for Get method with retry option
func TestRetryGet(t *testing.T) {
	const (
		case1_empty                         = "/"
		case24_after_3_attempt_return_valid = "/retry_3_attempt_then_valid"
		retry_count_expected                = "3"
	)

	var attempt int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check method is GET before going to check other features
		if r.Method != GET {
			t.Errorf("Expected method %q; got %q", GET, r.Method)
		}

		//set return status

		if r.Header == nil {
			t.Error("Expected non-nil request Header")
		}
		switch r.URL.Path {
		default:
			t.Errorf("No testing for this case yet : %q", r.URL.Path)
		case case1_empty:
			w.WriteHeader(400)
			t.Logf("case %v ", case1_empty)
		case case24_after_3_attempt_return_valid:
			if attempt == 3 {
				w.WriteHeader(200)
			} else {
				w.WriteHeader(400)
				t.Logf("case %v ", case24_after_3_attempt_return_valid)
			}
			attempt++
		}

	}))

	defer ts.Close()

	resp, _, errs := New().Get(ts.URL+case1_empty).
		Retry(3, 1*time.Nanosecond, http.StatusBadRequest).
		End()
	if errs != nil {
		t.Errorf("No testing for this case yet : %q", errs)
	}

	retryCountReturn := resp.Header.Get("Retry-Count")
	if retryCountReturn != retry_count_expected {
		t.Errorf("Expected [%s] retry but was [%s]", retry_count_expected, retryCountReturn)
	}

	resp, _, errs = New().Get(ts.URL+case24_after_3_attempt_return_valid).
		Retry(4, 1*time.Nanosecond, http.StatusBadRequest).
		End()
	if errs != nil {
		t.Errorf("No testing for this case yet : %q", errs)
	}

	retryCountReturn = resp.Header.Get("Retry-Count")
	if retryCountReturn != retry_count_expected {
		t.Errorf("Expected [%s] retry but was [%s]", retry_count_expected, retryCountReturn)
	}
}

// testing for Options method
func TestOptions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check method is OPTIONS before going to check other features
		if r.Method != OPTIONS {
			t.Errorf("Expected method %q; got %q", OPTIONS, r.Method)
		}
		t.Log("test Options")
		w.Header().Set("Allow", "HEAD, GET")
		w.WriteHeader(204)
	}))

	defer ts.Close()

	New().Options(ts.URL).
		End()
}

// testing that resp.Body is reusable
func TestResetBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Just some text"))
	}))

	defer ts.Close()

	resp, _, _ := New().Get(ts.URL).End()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	if string(bodyBytes) != "Just some text" {
		t.Error("Expected to be able to reuse the response body")
	}
}

// testing for Param method
func TestParam(t *testing.T) {
	paramCode := "123456"
	paramFields := "f1;f2;f3"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Form.Get("code") != paramCode {
			t.Errorf("Expected 'code' == %s; got %v", paramCode, r.Form.Get("code"))
		}

		if r.Form.Get("fields") != paramFields {
			t.Errorf("Expected 'fields' == %s; got %v", paramFields, r.Form.Get("fields"))
		}
	}))

	defer ts.Close()

	New().Get(ts.URL).
		Param("code", paramCode).
		Param("fields", paramFields)
}

// testing for POST method
func TestPost(t *testing.T) {
	const case1_empty = "/"
	const case2_set_header = "/set_header"
	const case3_send_json = "/send_json"
	const case4_send_string = "/send_string"
	const case5_integration_send_json_string = "/integration_send_json_string"
	const case6_set_query = "/set_query"
	const case7_integration_send_json_struct = "/integration_send_json_struct"
	// Check that the number conversion should be converted as string not float64
	const case8_send_json_with_long_id_number = "/send_json_with_long_id_number"
	const case9_send_json_string_with_long_id_number_as_form_result = "/send_json_string_with_long_id_number_as_form_result"
	const case10_send_struct_pointer = "/send_struct_pointer"
	const case11_send_string_pointer = "/send_string_pointer"
	const case12_send_slice_string = "/send_slice_string"
	const case13_send_slice_string_pointer = "/send_slice_string_pointer"
	const case14_send_int_pointer = "/send_int_pointer"
	const case15_send_float_pointer = "/send_float_pointer"
	const case16_send_bool_pointer = "/send_bool_pointer"
	const case17_send_string_array = "/send_string_array"
	const case18_send_string_array_pointer = "/send_string_array_pointer"
	const case19_send_struct = "/send_struct"
	const case20_send_byte_char = "/send_byte_char"
	const case21_send_byte_char_pointer = "/send_byte_char_pointer"
	const case22_send_byte_int = "/send_byte_int"
	const case22_send_byte_int_pointer = "/send_byte_int_pointer"
	const case23_send_duplicate_query_params = "/send_duplicate_query_params"
	const case24_send_query_and_request_body = "/send_query_and_request_body"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check method is POST before going to check other features
		if r.Method != POST {
			t.Errorf("Expected method %q; got %q", POST, r.Method)
		}
		if r.Header == nil {
			t.Error("Expected non-nil request Header")
		}
		switch r.URL.Path {
		default:
			t.Errorf("No testing for this case yet : %q", r.URL.Path)
		case case1_empty:
			t.Logf("case %v ", case1_empty)
		case case2_set_header:
			t.Logf("case %v ", case2_set_header)
			if r.Header.Get("API-Key") != "fookey" {
				t.Errorf("Expected 'API-Key' == %q; got %q", "fookey", r.Header.Get("API-Key"))
			}
		case case3_send_json:
			t.Logf("case %v ", case3_send_json)
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			if string(body) != `{"query1":"test","query2":"test"}` {
				t.Error(`Expected Body with {"query1":"test","query2":"test"}`, "| but got", string(body))
			}
		case case4_send_string, case11_send_string_pointer:
			t.Logf("case %v ", r.URL.Path)
			if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
				t.Error("Expected Header Content-Type -> application/x-www-form-urlencoded", "| but got", r.Header.Get("Content-Type"))
			}
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			if string(body) != "query1=test&query2=test" {
				t.Error("Expected Body with \"query1=test&query2=test\"", "| but got", string(body))
			}
		case case5_integration_send_json_string:
			t.Logf("case %v ", case5_integration_send_json_string)
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			if string(body) != "query1=test&query2=test" {
				t.Error("Expected Body with \"query1=test&query2=test\"", "| but got", string(body))
			}
		case case6_set_query:
			t.Logf("case %v ", case6_set_query)
			v := r.URL.Query()
			if v["query1"][0] != "test" {
				t.Error("Expected query1:test", "| but got", v["query1"][0])
			}
			if v["query2"][0] != "test" {
				t.Error("Expected query2:test", "| but got", v["query2"][0])
			}
		case case7_integration_send_json_struct:
			t.Logf("case %v ", case7_integration_send_json_struct)
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			comparedBody := []byte(`{"Lower":{"Color":"green","Size":1.7},"Upper":{"Color":"red","Size":0},"a":"a","name":"Cindy"}`)
			if !bytes.Equal(body, comparedBody) {
				t.Errorf(`Expected correct json but got ` + string(body))
			}
		case case8_send_json_with_long_id_number:
			t.Logf("case %v ", case8_send_json_with_long_id_number)
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			if string(body) != `{"id":123456789,"name":"nemo"}` {
				t.Error(`Expected Body with {"id":123456789,"name":"nemo"}`, "| but got", string(body))
			}
		case case9_send_json_string_with_long_id_number_as_form_result:
			t.Logf("case %v ", case9_send_json_string_with_long_id_number_as_form_result)
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			if string(body) != `id=123456789&name=nemo` {
				t.Error(`Expected Body with "id=123456789&name=nemo"`, `| but got`, string(body))
			}
		case case19_send_struct, case10_send_struct_pointer:
			t.Logf("case %v ", r.URL.Path)
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			comparedBody := []byte(`{"Bfalse":false,"BoolArray":[true,false],"Btrue":true,"Float":12.345,"FloatArray":[1.23,4.56,7.89],"Int":42,"IntArray":[1,2],"String":"a string","StringArray":["string1","string2"]}`)
			if !bytes.Equal(body, comparedBody) {
				t.Errorf(`Expected correct json but got ` + string(body))
			}
		case case12_send_slice_string, case13_send_slice_string_pointer, case17_send_string_array, case18_send_string_array_pointer:
			t.Logf("case %v ", r.URL.Path)
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			comparedBody := []byte(`["string1","string2"]`)
			if !bytes.Equal(body, comparedBody) {
				t.Errorf(`Expected correct json but got ` + string(body))
			}
		case case14_send_int_pointer:
			t.Logf("case %v ", case14_send_int_pointer)
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			if string(body) != "42" {
				t.Error("Expected Body with \"42\"", "| but got", string(body))
			}
		case case15_send_float_pointer:
			t.Logf("case %v ", case15_send_float_pointer)
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			if string(body) != "12.345" {
				t.Error("Expected Body with \"12.345\"", "| but got", string(body))
			}
		case case16_send_bool_pointer:
			t.Logf("case %v ", case16_send_bool_pointer)
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			if string(body) != "true" {
				t.Error("Expected Body with \"true\"", "| but got", string(body))
			}
		case case20_send_byte_char, case21_send_byte_char_pointer, case22_send_byte_int, case22_send_byte_int_pointer:
			t.Logf("case %v ", r.URL.Path)
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			if string(body) != "71" {
				t.Error("Expected Body with \"71\"", "| but got", string(body))
			}
		case case23_send_duplicate_query_params:
			t.Logf("case %v ", case23_send_duplicate_query_params)
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			sbody := string(body)
			if sbody != "param=4&param=3&param=2&param=1" {
				t.Error("Expected Body \"param=4&param=3&param=2&param=1\"", "| but got", sbody)
			}
			values, _ := url.ParseQuery(sbody)
			if len(values["param"]) != 4 {
				t.Error("Expected Body with 4 params", "| but got", sbody)
			}
			if values["param"][0] != "4" || values["param"][1] != "3" || values["param"][2] != "2" || values["param"][3] != "1" {
				t.Error("Expected Body with 4 params and values", "| but got", sbody)
			}
		case case24_send_query_and_request_body:
			t.Logf("case %v ", case24_send_query_and_request_body)
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			sbody := string(body)
			if sbody != `{"name":"jkbbwr"}` {
				t.Error(`Expected Body "{"name":"jkbbwr"}"`, "| but got", sbody)
			}

			v := r.URL.Query()
			if v["test"][0] != "true" {
				t.Error("Expected test:true", "| but got", v["test"][0])
			}
		}
	}))
	defer ts.Close()

	New().Post(ts.URL + case1_empty).
		End()

	New().Post(ts.URL+case2_set_header).
		Set("API-Key", "fookey").
		End()

	New().Post(ts.URL + case3_send_json).
		Send(`{"query1":"test"}`).
		Send(`{"query2":"test"}`).
		End()

	New().Post(ts.URL + case4_send_string).
		Send("query1=test").
		Send("query2=test").
		End()

	New().Post(ts.URL + case5_integration_send_json_string).
		Send("query1=test").
		Send(`{"query2":"test"}`).
		End()

	/* TODO: More testing post for application/x-www-form-urlencoded
	   post.query(json), post.query(string), post.send(json), post.send(string), post.query(both).send(both)
	*/
	New().Post(ts.URL + case6_set_query).
		Query("query1=test").
		Query("query2=test").
		End()
	// TODO:
	// 1. test 2nd layer nested struct
	// 2. test lowercase won't be export to json
	// 3. test field tag change to json field name
	type Upper struct {
		Color string
		Size  int
		note  string
	}
	type Lower struct {
		Color string
		Size  float64
		note  string
	}
	type Style struct {
		Upper Upper
		Lower Lower
		Name  string `json:"name"`
	}
	myStyle := Style{Upper: Upper{Color: "red"}, Name: "Cindy", Lower: Lower{Color: "green", Size: 1.7}}
	New().Post(ts.URL + case7_integration_send_json_struct).
		Send(`{"a":"a"}`).
		Send(myStyle).
		End()

	New().Post(ts.URL + case8_send_json_with_long_id_number).
		Send(`{"id":123456789, "name":"nemo"}`).
		End()

	New().Post(ts.URL + case9_send_json_string_with_long_id_number_as_form_result).
		Type("form").
		Send(`{"id":123456789, "name":"nemo"}`).
		End()

	payload := testStruct{
		String:      "a string",
		Int:         42,
		Btrue:       true,
		Bfalse:      false,
		Float:       12.345,
		StringArray: []string{"string1", "string2"},
		IntArray:    []int{1, 2},
		BoolArray:   []bool{true, false},
		FloatArray:  []float64{1.23, 4.56, 7.89},
	}

	New().Post(ts.URL + case10_send_struct_pointer).
		Send(&payload).
		End()

	New().Post(ts.URL + case19_send_struct).
		Send(payload).
		End()

	s1 := "query1=test"
	s2 := "query2=test"
	New().Post(ts.URL + case11_send_string_pointer).
		Send(&s1).
		Send(&s2).
		End()

	New().Post(ts.URL + case12_send_slice_string).
		Send([]string{"string1", "string2"}).
		End()

	New().Post(ts.URL + case13_send_slice_string_pointer).
		Send(&[]string{"string1", "string2"}).
		End()

	i := 42
	New().Post(ts.URL + case14_send_int_pointer).
		Send(&i).
		End()

	f := 12.345
	New().Post(ts.URL + case15_send_float_pointer).
		Send(&f).
		End()

	b := true
	New().Post(ts.URL + case16_send_bool_pointer).
		Send(&b).
		End()

	var a [2]string
	a[0] = "string1"
	a[1] = "string2"
	New().Post(ts.URL + case17_send_string_array).
		Send(a).
		End()

	New().Post(ts.URL + case18_send_string_array_pointer).
		Send(&a).
		End()

	aByte := byte('G') // = 71 dec
	New().Post(ts.URL + case20_send_byte_char).
		Send(aByte).
		End()

	New().Post(ts.URL + case21_send_byte_char_pointer).
		Send(&aByte).
		End()

	iByte := byte(71) // = 'G'
	New().Post(ts.URL + case22_send_byte_int).
		Send(iByte).
		End()

	New().Post(ts.URL + case22_send_byte_int_pointer).
		Send(&iByte).
		End()

	New().Post(ts.URL + case23_send_duplicate_query_params).
		Send("param=1").
		Send("param=2").
		Send("param=3&param=4").
		End()

	data24 := struct {
		Name string `json:"name"`
	}{"jkbbwr"}
	New().Post(ts.URL + case24_send_query_and_request_body).
		Query("test=true").
		Send(data24).
		End()
}

func checkFile(t *testing.T, fileheader *multipart.FileHeader) {
	infile, err := fileheader.Open()
	if err != nil {
		t.Error(err)
	}
	defer infile.Close()
	b, err := ioutil.ReadAll(infile)
	if err != nil {
		t.Error(err)
	}
	if len(b) == 0 {
		t.Error("Expected file-content > 0", "| but got", len(b), string(b))
	}
}

// testing for POST-Request of Type multipart
func TestMultipartRequest(t *testing.T) {

	const case0_send_not_supported_filetype = "/send_not_supported_filetype"
	const case1_send_string = "/send_string"
	const case2_send_json = "/send_json"
	const case3_integration_send_json_string = "/integration_send_json_string"
	const case4_set_query = "/set_query"
	const case5_send_struct = "/send_struct"
	const case6_send_slice_string = "/send_slice_string"
	const case6_send_slice_string_with_custom_fieldname = "/send_slice_string_with_custom_fieldname"
	const case7_send_array = "/send_array"
	const case8_integration_send_json_struct = "/integration_send_json_struct"
	const case9_send_duplicate_query_params = "/send_duplicate_query_params"

	const case10_send_file_by_path = "/send_file_by_path"
	const case10a_send_file_by_path_with_name = "/send_file_by_path_with_name"
	const case10b_send_file_by_path_pointer = "/send_file_by_path_pointer"
	const case11_send_file_by_path_without_name = "/send_file_by_path_without_name"
	const case12_send_file_by_path_without_name_but_with_fieldname = "/send_file_by_path_without_name_but_with_fieldname"

	const case13_send_file_by_content_without_name = "/send_file_by_content_without_name"
	const case13a_send_file_by_content_without_name_pointer = "/send_file_by_content_without_name_pointer"
	const case14_send_file_by_content_with_name = "/send_file_by_content_with_name"

	const case15_send_file_by_content_without_name_but_with_fieldname = "/send_file_by_content_without_name_but_with_fieldname"
	const case16_send_file_by_content_with_name_and_with_fieldname = "/send_file_by_content_with_name_and_with_fieldname"

	const case17_send_file_multiple_by_path_and_content_without_name = "/send_file_multiple_by_path_and_content_without_name"
	const case18_send_file_multiple_by_path_and_content_with_name = "/send_file_multiple_by_path_and_content_with_name"
	const case19_integration_send_file_and_data = "/integration_send_file_and_data"

	const case20_send_file_as_osfile = "/send_file_as_osfile"
	const case21_send_file_as_osfile_with_name = "/send_file_as_osfile_with_name"
	const case22_send_file_as_osfile_with_name_and_with_fieldname = "/send_file_as_osfile_with_name_and_with_fieldname"

	const case23_send_file_with_file_as_fieldname = "/send_file_with_file_as_fieldname"
	const case24_send_file_with_name_with_spaces = "/send_file_with_name_with_spaces"
	const case25_send_file_with_name_with_spaces_only = "/send_file_with_name_with_spaces_only"
	const case26_send_file_with_fieldname_with_spaces = "/send_file_with_fieldname_with_spaces"
	const case27_send_file_with_fieldname_with_spaces_only = "/send_file_with_fieldname_with_spaces_only"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check method is POST before going to check other features
		if r.Method != POST {
			t.Errorf("Expected method %q; got %q", POST, r.Method)
		}
		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Error("Expected Header Content-Type -> multipart/form-data", "| but got", r.Header.Get("Content-Type"))
		}
		const _24K = (1 << 20) * 24
		err := r.ParseMultipartForm(_24K)
		if err != nil {
			t.Errorf("Error: %v", err)
		}
		t.Logf("case %v ", r.URL.Path)
		switch r.URL.Path {
		default:
			t.Errorf("No testing for this case yet : %q", r.URL.Path)
		case case0_send_not_supported_filetype:
			// will be handled at place
		case case1_send_string, case2_send_json, case3_integration_send_json_string:
			if len(r.MultipartForm.Value["query1"]) != 1 {
				t.Error("Expected length of query1:test == 1", "| but got", len(r.MultipartForm.Value["query1"]))
			}
			if r.MultipartForm.Value["query1"][0] != "test" {
				t.Error("Expected query1:test", "| but got", r.MultipartForm.Value["query1"][0])
			}
			if len(r.MultipartForm.Value["query2"]) != 1 {
				t.Error("Expected length of query2:test == 1", "| but got", len(r.MultipartForm.Value["query2"]))
			}
			if r.MultipartForm.Value["query2"][0] != "test" {
				t.Error("Expected query2:test", "| but got", r.MultipartForm.Value["query2"][0])
			}
		case case4_set_query:
			v := r.URL.Query()
			if v["query1"][0] != "test" {
				t.Error("Expected query1:test", "| but got", v["query1"][0])
			}
			if v["query2"][0] != "test" {
				t.Error("Expected query2:test", "| but got", v["query2"][0])
			}
			if val, ok := r.MultipartForm.Value["query1"]; ok {
				t.Error("Expected no value", "| but got", val)
			}
			if val, ok := r.MultipartForm.Value["query2"]; ok {
				t.Error("Expected no value", "| but got", val)
			}
		case case5_send_struct:
			if r.MultipartForm.Value["String"][0] != "a string" {
				t.Error("Expected String:'a string'", "| but got", r.MultipartForm.Value["String"][0])
			}
			if r.MultipartForm.Value["Int"][0] != "42" {
				t.Error("Expected Int:42", "| but got", r.MultipartForm.Value["Int"][0])
			}
			if r.MultipartForm.Value["Btrue"][0] != "true" {
				t.Error("Expected Btrue:true", "| but got", r.MultipartForm.Value["Btrue"][0])
			}
			if r.MultipartForm.Value["Bfalse"][0] != "false" {
				t.Error("Expected Btrue:false", "| but got", r.MultipartForm.Value["Bfalse"][0])
			}
			if r.MultipartForm.Value["Float"][0] != "12.345" {
				t.Error("Expected Float:12.345", "| but got", r.MultipartForm.Value["Float"][0])
			}
			if len(r.MultipartForm.Value["StringArray"]) != 2 {
				t.Error("Expected length of StringArray:2", "| but got", len(r.MultipartForm.Value["StringArray"]))
			}
			if r.MultipartForm.Value["StringArray"][0] != "string1" {
				t.Error("Expected StringArray:string1", "| but got", r.MultipartForm.Value["StringArray"][0])
			}
			if r.MultipartForm.Value["StringArray"][1] != "string2" {
				t.Error("Expected StringArray:string2", "| but got", r.MultipartForm.Value["StringArray"][1])
			}
			if len(r.MultipartForm.Value["IntArray"]) != 2 {
				t.Error("Expected length of IntArray:2", "| but got", len(r.MultipartForm.Value["IntArray"]))
			}
			if r.MultipartForm.Value["IntArray"][0] != "1" {
				t.Error("Expected IntArray:1", "| but got", r.MultipartForm.Value["IntArray"][0])
			}
			if r.MultipartForm.Value["IntArray"][1] != "2" {
				t.Error("Expected IntArray:2", "| but got", r.MultipartForm.Value["IntArray"][1])
			}
			if len(r.MultipartForm.Value["BoolArray"]) != 2 {
				t.Error("Expected length of BoolArray:2", "| but got", len(r.MultipartForm.Value["BoolArray"]))
			}
			if r.MultipartForm.Value["BoolArray"][0] != "true" {
				t.Error("Expected BoolArray:true", "| but got", r.MultipartForm.Value["BoolArray"][0])
			}
			if r.MultipartForm.Value["BoolArray"][1] != "false" {
				t.Error("Expected BoolArray:false", "| but got", r.MultipartForm.Value["BoolArray"][1])
			}
			if len(r.MultipartForm.Value["FloatArray"]) != 3 {
				t.Error("Expected length of FloatArray:3", "| but got", len(r.MultipartForm.Value["FloatArray"]))
			}
			if r.MultipartForm.Value["FloatArray"][0] != "1.23" {
				t.Error("Expected FloatArray:1.23", "| but got", r.MultipartForm.Value["FloatArray"][0])
			}
			if r.MultipartForm.Value["FloatArray"][1] != "4.56" {
				t.Error("Expected FloatArray:4.56", "| but got", r.MultipartForm.Value["FloatArray"][1])
			}
			if r.MultipartForm.Value["FloatArray"][2] != "7.89" {
				t.Error("Expected FloatArray:7.89", "| but got", r.MultipartForm.Value["FloatArray"][2])
			}
		case case6_send_slice_string, case7_send_array:
			if len(r.MultipartForm.Value["data"]) != 1 {
				t.Error("Expected length of data:JSON == 1", "| but got", len(r.MultipartForm.Value["data"]))
			}
			if r.MultipartForm.Value["data"][0] != `["string1","string2"]` {
				t.Error(`Expected 'data' with ["string1","string2"]`, "| but got", r.MultipartForm.Value["data"][0])
			}
		case case6_send_slice_string_with_custom_fieldname:
			if len(r.MultipartForm.Value["my_custom_data"]) != 1 {
				t.Error("Expected length of my_custom_data:JSON == 1", "| but got", len(r.MultipartForm.Value["my_custom_data"]))
			}
			if r.MultipartForm.Value["my_custom_data"][0] != `["string1","string2"]` {
				t.Error(`Expected 'my_custom_data' with ["string1","string2"]`, "| but got", r.MultipartForm.Value["my_custom_data"][0])
			}
		case case8_integration_send_json_struct:
			if len(r.MultipartForm.Value["query1"]) != 1 {
				t.Error("Expected length of query1:test == 1", "| but got", len(r.MultipartForm.Value["query1"]))
			}
			if r.MultipartForm.Value["query1"][0] != "test" {
				t.Error("Expected query1:test", "| but got", r.MultipartForm.Value["query1"][0])
			}
			if r.MultipartForm.Value["hey"][0] != "hey" {
				t.Error("Expected hey:'hey'", "| but got", r.MultipartForm.Value["Hey"][0])
			}
		case case9_send_duplicate_query_params:
			if len(r.MultipartForm.Value["param"]) != 4 {
				t.Error("Expected length of param:[] == 4", "| but got", len(r.MultipartForm.Value["param"]))
			}
			if r.MultipartForm.Value["param"][0] != "4" {
				t.Error("Expected param:0:4", "| but got", r.MultipartForm.Value["param"][0])
			}
			if r.MultipartForm.Value["param"][1] != "3" {
				t.Error("Expected param:1:3", "| but got", r.MultipartForm.Value["param"][1])
			}
			if r.MultipartForm.Value["param"][2] != "2" {
				t.Error("Expected param:2:2", "| but got", r.MultipartForm.Value["param"][2])
			}
			if r.MultipartForm.Value["param"][3] != "1" {
				t.Error("Expected param:3:1", "| but got", r.MultipartForm.Value["param"][3])
			}
		case case10_send_file_by_path, case11_send_file_by_path_without_name, case14_send_file_by_content_with_name, case20_send_file_as_osfile:
			if len(r.MultipartForm.File) != 1 {
				t.Error("Expected length of files:[] == 1", "| but got", len(r.MultipartForm.File))
			}
			if r.MultipartForm.File["file1"][0].Filename != "LICENSE" {
				t.Error("Expected Filename:LICENSE", "| but got", r.MultipartForm.File["file1"][0].Filename)
			}
			if r.MultipartForm.File["file1"][0].Header["Content-Type"][0] != "application/octet-stream" {
				t.Error("Expected Header:Content-Type:application/octet-stream", "| but got", r.MultipartForm.File["file1"][0].Header["Content-Type"])
			}
			checkFile(t, r.MultipartForm.File["file1"][0])
		case case10a_send_file_by_path_with_name, case10b_send_file_by_path_pointer, case21_send_file_as_osfile_with_name:
			if len(r.MultipartForm.File) != 1 {
				t.Error("Expected length of files:[] == 1", "| but got", len(r.MultipartForm.File))
			}
			if r.MultipartForm.File["file1"][0].Filename != "MY_LICENSE" {
				t.Error("Expected Filename:MY_LICENSE", "| but got", r.MultipartForm.File["file1"][0].Filename)
			}
		case case12_send_file_by_path_without_name_but_with_fieldname:
			if len(r.MultipartForm.File) != 1 {
				t.Error("Expected length of files:[] == 1", "| but got", len(r.MultipartForm.File))
			}
			if _, ok := r.MultipartForm.File["my_fieldname"]; !ok {
				keys := reflect.ValueOf(r.MultipartForm.File).MapKeys()
				t.Error("Expected Fieldname:my_fieldname", "| but got", keys)
			}
			if r.MultipartForm.File["my_fieldname"][0].Filename != "LICENSE" {
				t.Error("Expected Filename:LICENSE", "| but got", r.MultipartForm.File["my_fieldname"][0].Filename)
			}
			if r.MultipartForm.File["my_fieldname"][0].Header["Content-Type"][0] != "application/octet-stream" {
				t.Error("Expected Header:Content-Type:application/octet-stream", "| but got", r.MultipartForm.File["my_fieldname"][0].Header["Content-Type"])
			}
			checkFile(t, r.MultipartForm.File["my_fieldname"][0])
		case case13_send_file_by_content_without_name, case13a_send_file_by_content_without_name_pointer:
			if len(r.MultipartForm.File) != 1 {
				t.Error("Expected length of files:[] == 1", "| but got", len(r.MultipartForm.File))
			}
			if r.MultipartForm.File["file1"][0].Filename != "filename" {
				t.Error("Expected Filename:filename", "| but got", r.MultipartForm.File["file1"][0].Filename)
			}
			if r.MultipartForm.File["file1"][0].Header["Content-Type"][0] != "application/octet-stream" {
				t.Error("Expected Header:Content-Type:application/octet-stream", "| but got", r.MultipartForm.File["file1"][0].Header["Content-Type"])
			}
			checkFile(t, r.MultipartForm.File["file1"][0])
		case case15_send_file_by_content_without_name_but_with_fieldname:
			if len(r.MultipartForm.File) != 1 {
				t.Error("Expected length of files:[] == 1", "| but got", len(r.MultipartForm.File))
			}
			if _, ok := r.MultipartForm.File["my_fieldname"]; !ok {
				keys := reflect.ValueOf(r.MultipartForm.File).MapKeys()
				t.Error("Expected Fieldname:my_fieldname", "| but got", keys)
			}
			if r.MultipartForm.File["my_fieldname"][0].Filename != "filename" {
				t.Error("Expected Filename:filename", "| but got", r.MultipartForm.File["my_fieldname"][0].Filename)
			}
			if r.MultipartForm.File["my_fieldname"][0].Header["Content-Type"][0] != "application/octet-stream" {
				t.Error("Expected Header:Content-Type:application/octet-stream", "| but got", r.MultipartForm.File["my_fieldname"][0].Header["Content-Type"])
			}
			checkFile(t, r.MultipartForm.File["my_fieldname"][0])
		case case16_send_file_by_content_with_name_and_with_fieldname, case22_send_file_as_osfile_with_name_and_with_fieldname:
			if len(r.MultipartForm.File) != 1 {
				t.Error("Expected length of files:[] == 1", "| but got", len(r.MultipartForm.File))
			}
			if _, ok := r.MultipartForm.File["my_fieldname"]; !ok {
				keys := reflect.ValueOf(r.MultipartForm.File).MapKeys()
				t.Error("Expected Fieldname:my_fieldname", "| but got", keys)
			}
			if r.MultipartForm.File["my_fieldname"][0].Filename != "MY_LICENSE" {
				t.Error("Expected Filename:MY_LICENSE", "| but got", r.MultipartForm.File["my_fieldname"][0].Filename)
			}
			if r.MultipartForm.File["my_fieldname"][0].Header["Content-Type"][0] != "application/octet-stream" {
				t.Error("Expected Header:Content-Type:application/octet-stream", "| but got", r.MultipartForm.File["my_fieldname"][0].Header["Content-Type"])
			}
			checkFile(t, r.MultipartForm.File["my_fieldname"][0])
		case case17_send_file_multiple_by_path_and_content_without_name:
			if len(r.MultipartForm.File) != 2 {
				t.Error("Expected length of files:[] == 2", "| but got", len(r.MultipartForm.File))
			}
			// depends on map iteration order
			if r.MultipartForm.File["file1"][0].Filename != "LICENSE" && r.MultipartForm.File["file1"][0].Filename != "filename" {
				t.Error("Expected Filename:LICENSE||filename", "| but got", r.MultipartForm.File["file1"][0].Filename)
			}
			if r.MultipartForm.File["file1"][0].Header["Content-Type"][0] != "application/octet-stream" {
				t.Error("Expected Header:Content-Type:application/octet-stream", "| but got", r.MultipartForm.File["file1"][0].Header["Content-Type"])
			}
			// depends on map iteration order
			if r.MultipartForm.File["file2"][0].Filename != "LICENSE" && r.MultipartForm.File["file2"][0].Filename != "filename" {
				t.Error("Expected Filename:LICENSE||filename", "| but got", r.MultipartForm.File["file2"][0].Filename)
			}
			if r.MultipartForm.File["file2"][0].Header["Content-Type"][0] != "application/octet-stream" {
				t.Error("Expected Header:Content-Type:application/octet-stream", "| but got", r.MultipartForm.File["file2"][0].Header["Content-Type"])
			}
			checkFile(t, r.MultipartForm.File["file1"][0])
			checkFile(t, r.MultipartForm.File["file2"][0])
		case case18_send_file_multiple_by_path_and_content_with_name:
			if len(r.MultipartForm.File) != 2 {
				t.Error("Expected length of files:[] == 2", "| but got", len(r.MultipartForm.File))
			}
			// depends on map iteration order
			if r.MultipartForm.File["file1"][0].Filename != "LICENSE" && r.MultipartForm.File["file1"][0].Filename != "MY_LICENSE" {
				t.Error("Expected Filename:LICENSE||MY_LICENSE", "| but got", r.MultipartForm.File["file1"][0].Filename)
			}
			if r.MultipartForm.File["file1"][0].Header["Content-Type"][0] != "application/octet-stream" {
				t.Error("Expected Header:Content-Type:application/octet-stream", "| but got", r.MultipartForm.File["file1"][0].Header["Content-Type"])
			}
			// depends on map iteration order
			if r.MultipartForm.File["file2"][0].Filename != "LICENSE" && r.MultipartForm.File["file2"][0].Filename != "MY_LICENSE" {
				t.Error("Expected Filename:LICENSE||MY_LICENSE", "| but got", r.MultipartForm.File["file2"][0].Filename)
			}
			if r.MultipartForm.File["file2"][0].Header["Content-Type"][0] != "application/octet-stream" {
				t.Error("Expected Header:Content-Type:application/octet-stream", "| but got", r.MultipartForm.File["file2"][0].Header["Content-Type"])
			}
			checkFile(t, r.MultipartForm.File["file1"][0])
			checkFile(t, r.MultipartForm.File["file2"][0])
		case case19_integration_send_file_and_data:
			if len(r.MultipartForm.File) != 1 {
				t.Error("Expected length of files:[] == 1", "| but got", len(r.MultipartForm.File))
			}
			if r.MultipartForm.File["file1"][0].Filename != "LICENSE" {
				t.Error("Expected Filename:LICENSE", "| but got", r.MultipartForm.File["file1"][0].Filename)
			}
			if r.MultipartForm.File["file1"][0].Header["Content-Type"][0] != "application/octet-stream" {
				t.Error("Expected Header:Content-Type:application/octet-stream", "| but got", r.MultipartForm.File["file1"][0].Header["Content-Type"])
			}
			checkFile(t, r.MultipartForm.File["file1"][0])
			if len(r.MultipartForm.Value["query1"]) != 1 {
				t.Error("Expected length of query1:test == 1", "| but got", len(r.MultipartForm.Value["query1"]))
			}
			if r.MultipartForm.Value["query1"][0] != "test" {
				t.Error("Expected query1:test", "| but got", r.MultipartForm.Value["query1"][0])
			}
		case case23_send_file_with_file_as_fieldname:
			if len(r.MultipartForm.File) != 2 {
				t.Error("Expected length of files:[] == 2", "| but got", len(r.MultipartForm.File))
			}
			if val, ok := r.MultipartForm.File["file1"]; !ok {
				t.Error("Expected file with key: file1", "| but got ", val)
			}
			if val, ok := r.MultipartForm.File["file2"]; !ok {
				t.Error("Expected file with key: file2", "| but got ", val)
			}
			if r.MultipartForm.File["file1"][0].Filename != "b.file" {
				t.Error("Expected Filename:b.file", "| but got", r.MultipartForm.File["file1"][0].Filename)
			}
			if r.MultipartForm.File["file2"][0].Filename != "LICENSE" {
				t.Error("Expected Filename:LICENSE", "| but got", r.MultipartForm.File["file2"][0].Filename)
			}
			checkFile(t, r.MultipartForm.File["file1"][0])
			checkFile(t, r.MultipartForm.File["file2"][0])
		case case24_send_file_with_name_with_spaces, case25_send_file_with_name_with_spaces_only, case27_send_file_with_fieldname_with_spaces_only:
			if len(r.MultipartForm.File) != 1 {
				t.Error("Expected length of files:[] == 1", "| but got", len(r.MultipartForm.File))
			}
			if val, ok := r.MultipartForm.File["file1"]; !ok {
				t.Error("Expected file with key: file1", "| but got ", val)
			}
			if r.MultipartForm.File["file1"][0].Filename != "LICENSE" {
				t.Error("Expected Filename:LICENSE", "| but got", r.MultipartForm.File["file1"][0].Filename)
			}
			checkFile(t, r.MultipartForm.File["file1"][0])
		case case26_send_file_with_fieldname_with_spaces:
			if len(r.MultipartForm.File) != 1 {
				t.Error("Expected length of files:[] == 1", "| but got", len(r.MultipartForm.File))
			}
			if val, ok := r.MultipartForm.File["my_fieldname"]; !ok {
				t.Error("Expected file with key: my_fieldname", "| but got ", val)
			}
			if r.MultipartForm.File["my_fieldname"][0].Filename != "LICENSE" {
				t.Error("Expected Filename:LICENSE", "| but got", r.MultipartForm.File["my_fieldname"][0].Filename)
			}
			checkFile(t, r.MultipartForm.File["my_fieldname"][0])
		}

	}))
	defer ts.Close()

	// "the zero case"
	t.Logf("case %v ", case0_send_not_supported_filetype)
	_, _, errs := New().Post(ts.URL + case0_send_not_supported_filetype).
		Type("multipart").
		SendFile(42).
		End()

	if len(errs) == 0 {
		t.Errorf("Expected error, but got nothing: %v", errs)
	}

	New().Post(ts.URL + case1_send_string).
		Type("multipart").
		Send("query1=test").
		Send("query2=test").
		End()

	New().Post(ts.URL + case2_send_json).
		Type("multipart").
		Send(`{"query1":"test"}`).
		Send(`{"query2":"test"}`).
		End()

	New().Post(ts.URL + case3_integration_send_json_string).
		Type("multipart").
		Send("query1=test").
		Send(`{"query2":"test"}`).
		End()

	New().Post(ts.URL + case4_set_query).
		Type("multipart").
		Query("query1=test").
		Query("query2=test").
		End()

	New().Post(ts.URL + case5_send_struct).
		Type("multipart").
		Send(testStruct{
			String:      "a string",
			Int:         42,
			Btrue:       true,
			Bfalse:      false,
			Float:       12.345,
			StringArray: []string{"string1", "string2"},
			IntArray:    []int{1, 2},
			BoolArray:   []bool{true, false},
			FloatArray:  []float64{1.23, 4.56, 7.89},
		}).
		End()

	New().Post(ts.URL + case6_send_slice_string).
		Type("multipart").
		Send([]string{"string1", "string2"}).
		End()

	New().Post(ts.URL+case6_send_slice_string_with_custom_fieldname).
		Type("multipart").
		Set("json_fieldname", "my_custom_data").
		Send([]string{"string1", "string2"}).
		End()

	New().Post(ts.URL + case7_send_array).
		Type("multipart").
		Send([2]string{"string1", "string2"}).
		End()

	New().Post(ts.URL + case8_integration_send_json_struct).
		Type("multipart").
		Send(`{"query1":"test"}`).
		Send(heyYou{
			Hey: "hey",
		}).
		End()

	New().Post(ts.URL + case9_send_duplicate_query_params).
		Type("multipart").
		Send("param=1").
		Send("param=2").
		Send("param=3&param=4").
		End()

	fileByPath := "./LICENSE"
	New().Post(ts.URL + case10_send_file_by_path).
		Type("multipart").
		SendFile(fileByPath).
		End()

	New().Post(ts.URL+case10a_send_file_by_path_with_name).
		Type("multipart").
		SendFile(fileByPath, "MY_LICENSE").
		End()

	New().Post(ts.URL+case10b_send_file_by_path_pointer).
		Type("multipart").
		SendFile(&fileByPath, "MY_LICENSE").
		End()

	New().Post(ts.URL+case11_send_file_by_path_without_name).
		Type("multipart").
		SendFile(fileByPath, "").
		End()

	New().Post(ts.URL+case12_send_file_by_path_without_name_but_with_fieldname).
		Type("multipart").
		SendFile(fileByPath, "", "my_fieldname").
		End()

	b, _ := ioutil.ReadFile("./LICENSE")
	New().Post(ts.URL + case13_send_file_by_content_without_name).
		Type("multipart").
		SendFile(b).
		End()

	New().Post(ts.URL + case13a_send_file_by_content_without_name_pointer).
		Type("multipart").
		SendFile(&b).
		End()

	New().Post(ts.URL+case14_send_file_by_content_with_name).
		Type("multipart").
		SendFile(b, "LICENSE").
		End()

	New().Post(ts.URL+case15_send_file_by_content_without_name_but_with_fieldname).
		Type("multipart").
		SendFile(b, "", "my_fieldname").
		End()

	New().Post(ts.URL+case16_send_file_by_content_with_name_and_with_fieldname).
		Type("multipart").
		SendFile(b, "MY_LICENSE", "my_fieldname").
		End()

	New().Post(ts.URL + case17_send_file_multiple_by_path_and_content_without_name).
		Type("multipart").
		SendFile("./LICENSE").
		SendFile(b).
		End()

	New().Post(ts.URL+case18_send_file_multiple_by_path_and_content_with_name).
		Type("multipart").
		SendFile("./LICENSE").
		SendFile(b, "MY_LICENSE").
		End()

	New().Post(ts.URL + case19_integration_send_file_and_data).
		Type("multipart").
		SendFile("./LICENSE").
		Send("query1=test").
		End()

	osFile, _ := os.Open("./LICENSE")
	New().Post(ts.URL + case20_send_file_as_osfile).
		Type("multipart").
		SendFile(osFile).
		End()

	New().Post(ts.URL+case21_send_file_as_osfile_with_name).
		Type("multipart").
		SendFile(osFile, "MY_LICENSE").
		End()

	New().Post(ts.URL+case22_send_file_as_osfile_with_name_and_with_fieldname).
		Type("multipart").
		SendFile(osFile, "MY_LICENSE", "my_fieldname").
		End()

	New().Post(ts.URL+case23_send_file_with_file_as_fieldname).
		Type("multipart").
		SendFile(b, "b.file").
		SendFile(osFile, "", "file").
		End()

	New().Post(ts.URL+case24_send_file_with_name_with_spaces).
		Type("multipart").
		SendFile(osFile, " LICENSE  ").
		End()

	New().Post(ts.URL+case25_send_file_with_name_with_spaces_only).
		Type("multipart").
		SendFile(osFile, "   ").
		End()

	New().Post(ts.URL+case26_send_file_with_fieldname_with_spaces).
		Type("multipart").
		SendFile(osFile, "", " my_fieldname  ").
		End()

	New().Post(ts.URL+case27_send_file_with_fieldname_with_spaces_only).
		Type("multipart").
		SendFile(osFile, "", "   ").
		End()
}

// testing for Patch method
func TestPatch(t *testing.T) {
	const case1_empty = "/"
	const case2_set_header = "/set_header"
	const case3_send_json = "/send_json"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check method is PATCH before going to check other features
		if r.Method != PATCH {
			t.Errorf("Expected method %q; got %q", PATCH, r.Method)
		}
		if r.Header == nil {
			t.Error("Expected non-nil request Header")
		}
		switch r.URL.Path {
		default:
			t.Errorf("No testing for this case yet : %q", r.URL.Path)
		case case1_empty:
			t.Logf("case %v ", case1_empty)
		case case2_set_header:
			t.Logf("case %v ", case2_set_header)
			if r.Header.Get("API-Key") != "fookey" {
				t.Errorf("Expected 'API-Key' == %q; got %q", "fookey", r.Header.Get("API-Key"))
			}
		case case3_send_json:
			t.Logf("case %v ", case3_send_json)
			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			if string(body) != `{"query1":"test","query2":"test"}` {
				t.Error(`Expected Body with {"query1":"test","query2":"test"}`, "| but got", string(body))
			}
		}
	}))

	defer ts.Close()

	New().Patch(ts.URL + case1_empty).
		End()

	New().Patch(ts.URL+case2_set_header).
		Set("API-Key", "fookey").
		End()

	New().Patch(ts.URL + case3_send_json).
		Send(`{"query1":"test"}`).
		Send(`{"query2":"test"}`).
		End()
}

func checkQuery(t *testing.T, q map[string][]string, key string, want string) {
	v, ok := q[key]
	if !ok {
		t.Error(key, "Not Found")
	} else if len(v) < 1 {
		t.Error("No values for", key)
	} else if v[0] != want {
		t.Errorf("Expected %v:%v | but got %v", key, want, v[0])
	}
	return
}

// TODO: more check on url query (all testcases)
func TestQueryFunc(t *testing.T) {
	const case1_send_string = "/send_string"
	const case2_send_struct = "/send_struct"
	const case3_send_string_with_duplicates = "/send_string_with_duplicates"
	const case4_send_map = "/send_map"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != POST {
			t.Errorf("Expected method %q; got %q", POST, r.Method)
		}
		if r.Header == nil {
			t.Error("Expected non-nil request Header")
		}
		v := r.URL.Query()

		switch r.URL.Path {
		default:
			t.Errorf("No testing for this case yet : %q", r.URL.Path)
		case case1_send_string, case2_send_struct:
			checkQuery(t, v, "query1", "test1")
			checkQuery(t, v, "query2", "test2")
		case case3_send_string_with_duplicates:
			checkQuery(t, v, "query1", "test1")
			checkQuery(t, v, "query2", "test2")

			if len(v["param"]) != 4 {
				t.Errorf("Expected Body with 4 params | but got %q", len(v["param"]))
			}
			if v["param"][0] != "1" || v["param"][1] != "2" || v["param"][2] != "3" || v["param"][3] != "4" {
				t.Error("Expected Body with 4 params and values", "| but got", r.URL.RawQuery)
			}
		case case4_send_map:
			checkQuery(t, v, "query1", "test1")
			checkQuery(t, v, "query2", "test2")
			checkQuery(t, v, "query3", "3.1415926")
			checkQuery(t, v, "query4", "true")
		}
	}))
	defer ts.Close()

	New().Post(ts.URL + case1_send_string).
		Query("query1=test1").
		Query("query2=test2").
		End()

	qq := struct {
		Query1 string
		Query2 string
	}{
		Query1: "test1",
		Query2: "test2",
	}
	New().Post(ts.URL + case2_send_struct).
		Query(qq).
		End()

	New().Post(ts.URL + case3_send_string_with_duplicates).
		Query("query1=test1").
		Query("query2=test2").
		Query("param=1").
		Query("param=2").
		Query("param=3&param=4").
		End()

	New().Post(ts.URL + case4_send_map).
		Query(map[string]interface{}{
			"query1": "test1",
			"query2": "test2",
			"query3": 3.1415926,
			"query4": true,
		}).
		End()
}

// TODO: more tests on redirect
func TestRedirectPolicyFunc(t *testing.T) {
	redirectSuccess := false
	redirectFuncGetCalled := false
	tsRedirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectSuccess = true
	}))
	defer tsRedirect.Close()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, tsRedirect.URL, http.StatusMovedPermanently)
	}))
	defer ts.Close()

	New().
		Get(ts.URL).
		RedirectPolicy(func(req Request, via []Request) error {
			redirectFuncGetCalled = true
			return nil
		}).End()
	if !redirectSuccess {
		t.Error("Expected reaching another redirect url not original one")
	}
	if !redirectFuncGetCalled {
		t.Error("Expected redirect policy func to get called")
	}
}

func TestEndBytes(t *testing.T) {
	serverOutput := "hello world"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(serverOutput))
	}))
	defer ts.Close()

	// Callback.
	{
		resp, bodyBytes, errs := New().Get(ts.URL).EndBytes(func(resp Response, body []byte, errs []error) {
			if len(errs) > 0 {
				t.Fatalf("Unexpected errors: %s", errs)
			}
			if resp.StatusCode != 200 {
				t.Fatalf("Expected StatusCode=200, actual StatusCode=%v", resp.StatusCode)
			}
			if string(body) != serverOutput {
				t.Errorf("Expected bodyBytes=%s, actual bodyBytes=%s", serverOutput, string(body))
			}
		})
		if len(errs) > 0 {
			t.Fatalf("Unexpected errors: %s", errs)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("Expected StatusCode=200, actual StatusCode=%v", resp.StatusCode)
		}
		if string(bodyBytes) != serverOutput {
			t.Errorf("Expected bodyBytes=%s, actual bodyBytes=%s", serverOutput, string(bodyBytes))
		}
	}

	// No callback.
	{
		resp, bodyBytes, errs := New().Get(ts.URL).EndBytes()
		if len(errs) > 0 {
			t.Errorf("Unexpected errors: %s", errs)
		}
		if resp.StatusCode != 200 {
			t.Errorf("Expected StatusCode=200, actual StatusCode=%v", resp.StatusCode)
		}
		if string(bodyBytes) != serverOutput {
			t.Errorf("Expected bodyBytes=%s, actual bodyBytes=%s", serverOutput, string(bodyBytes))
		}
	}
}

func TestEndStruct(t *testing.T) {
	var resStruct heyYou
	expStruct := heyYou{Hey: "you"}
	serverOutput, err := json.Marshal(expStruct)
	if err != nil {
		t.Errorf("Unexpected errors: %s", err)
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(serverOutput)
	}))
	defer ts.Close()

	// Callback.
	{
		resp, bodyBytes, errs := New().Get(ts.URL).EndStruct(func(resp Response, v interface{}, body []byte, errs []error) {
			if len(errs) > 0 {
				t.Fatalf("Unexpected errors: %s", errs)
			}
			if resp.StatusCode != 200 {
				t.Fatalf("Expected StatusCode=200, actual StatusCode=%v", resp.StatusCode)
			}
			if !reflect.DeepEqual(expStruct, resStruct) {
				resBytes, _ := json.Marshal(resStruct)
				t.Errorf("Expected body=%s, actual bodyBytes=%s", serverOutput, string(resBytes))
			}
			if !reflect.DeepEqual(body, serverOutput) {
				t.Errorf("Expected bodyBytes=%s, actual bodyBytes=%s", serverOutput, string(body))
			}
		})
		if len(errs) > 0 {
			t.Fatalf("Unexpected errors: %s", errs)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("Expected StatusCode=200, actual StatusCode=%v", resp.StatusCode)
		}
		if !reflect.DeepEqual(bodyBytes, serverOutput) {
			t.Errorf("Expected bodyBytes=%s, actual bodyBytes=%s", serverOutput, string(bodyBytes))
		}
	}

	// No callback.
	{
		resp, bodyBytes, errs := New().Get(ts.URL).EndStruct(&resStruct)
		if len(errs) > 0 {
			t.Errorf("Unexpected errors: %s", errs)
		}
		if resp.StatusCode != 200 {
			t.Errorf("Expected StatusCode=200, actual StatusCode=%v", resp.StatusCode)
		}
		if !reflect.DeepEqual(expStruct, resStruct) {
			resBytes, _ := json.Marshal(resStruct)
			t.Errorf("Expected body=%s, actual bodyBytes=%s", serverOutput, string(resBytes))
		}
		if !reflect.DeepEqual(bodyBytes, serverOutput) {
			t.Errorf("Expected bodyBytes=%s, actual bodyBytes=%s", serverOutput, string(bodyBytes))
		}
	}
}

func TestProxyFunc(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "proxy passed")
	}))
	defer ts.Close()
	// start proxy
	proxy := goproxy.NewProxyHttpServer()
	proxy.OnRequest().DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			return r, nil
		})
	ts2 := httptest.NewServer(proxy)
	// sending request via Proxy
	resp, body, _ := New().Proxy(ts2.URL).Get(ts.URL).End()
	if resp.StatusCode != 200 {
		t.Error("Expected 200 Status code")
	}
	if body != "proxy passed" {
		t.Error("Expected 'proxy passed' body string")
	}
}

func TestTimeoutFunc(t *testing.T) {
	// 1st case, dial timeout
	startTime := time.Now()
	_, _, errs := New().Timeout(1000 * time.Millisecond).Get("http://www.google.com:81").End()
	elapsedTime := time.Since(startTime)
	if errs == nil {
		t.Error("Expected dial timeout error but get nothing")
	}
	if elapsedTime < 1000*time.Millisecond || elapsedTime > 1500*time.Millisecond {
		t.Errorf("Expected timeout in between 1000 -> 1500 ms | but got %d", elapsedTime)
	}
	// 2st case, read/write timeout (Can dial to url but want to timeout because too long operation on the server)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1100 * time.Millisecond) // slightly longer than expected
		w.WriteHeader(200)
	}))
	request := New().Timeout(1000 * time.Millisecond)
	startTime = time.Now()
	_, _, errs = request.Get(ts.URL).End()
	elapsedTime = time.Since(startTime)
	if errs == nil {
		t.Error("Expected dial+read/write timeout | but get nothing")
	}
	if elapsedTime < 1000*time.Millisecond || elapsedTime > 1500*time.Millisecond {
		t.Errorf("Expected timeout in between 1000 -> 1500 ms | but got %d", elapsedTime)
	}
	// 3rd case, testing reuse of same request
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1100 * time.Millisecond) // slightly longer than expected
		w.WriteHeader(200)
	}))
	startTime = time.Now()
	_, _, errs = request.Get(ts.URL).End()
	elapsedTime = time.Since(startTime)
	if errs == nil {
		t.Error("Expected dial+read/write timeout | but get nothing")
	}
	if elapsedTime < 1000*time.Millisecond || elapsedTime > 1500*time.Millisecond {
		t.Errorf("Expected timeout in between 1000 -> 1500 ms | but got %d", elapsedTime)
	}

}

func TestCookies(t *testing.T) {
	request := New().Timeout(60 * time.Second)
	_, _, errs := request.Get("https://github.com").End()
	if errs != nil {
		t.Error("Cookies test request did not complete")
		return
	}
	domain, _ := url.Parse("https://github.com")
	if len(request.Client.Jar.Cookies(domain)) == 0 {
		t.Error("Expected cookies | but get nothing")
	}
}

func TestGetSetCookie(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != GET {
			t.Errorf("Expected method %q; got %q", GET, r.Method)
		}
		c, err := r.Cookie("API-Cookie-Name")
		if err != nil {
			t.Error(err)
		}
		if c == nil {
			t.Error("Expected non-nil request Cookie 'API-Cookie-Name'")
		} else if c.Value != "api-cookie-value" {
			t.Errorf("Expected 'API-Cookie-Name' == %q; got %q", "api-cookie-value", c.Value)
		}
	}))
	defer ts.Close()

	New().Get(ts.URL).
		AddCookie(&http.Cookie{Name: "API-Cookie-Name", Value: "api-cookie-value"}).
		End()
}

func TestGetSetCookies(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != GET {
			t.Errorf("Expected method %q; got %q", GET, r.Method)
		}
		c, err := r.Cookie("API-Cookie-Name1")
		if err != nil {
			t.Error(err)
		}
		if c == nil {
			t.Error("Expected non-nil request Cookie 'API-Cookie-Name1'")
		} else if c.Value != "api-cookie-value1" {
			t.Errorf("Expected 'API-Cookie-Name1' == %q; got %q", "api-cookie-value1", c.Value)
		}
		c, err = r.Cookie("API-Cookie-Name2")
		if err != nil {
			t.Error(err)
		}
		if c == nil {
			t.Error("Expected non-nil request Cookie 'API-Cookie-Name2'")
		} else if c.Value != "api-cookie-value2" {
			t.Errorf("Expected 'API-Cookie-Name2' == %q; got %q", "api-cookie-value2", c.Value)
		}
	}))
	defer ts.Close()

	New().Get(ts.URL).AddCookies([]*http.Cookie{
		{Name: "API-Cookie-Name1", Value: "api-cookie-value1"},
		{Name: "API-Cookie-Name2", Value: "api-cookie-value2"},
	}).End()
}

func TestErrorTypeWrongKey(t *testing.T) {
	//defer afterTest(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, checkTypeWrongKey")
	}))
	defer ts.Close()

	_, _, err := New().
		Get(ts.URL).
		Type("wrongtype").
		End()
	if len(err) != 0 {
		if err[0].Error() != "Type func: incorrect type \"wrongtype\"" {
			t.Errorf("Wrong error message: " + err[0].Error())
		}
	} else {
		t.Error("Should have error")
	}
}

func TestBasicAuth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := strings.SplitN(r.Header["Authorization"][0], " ", 2)
		if len(auth) != 2 || auth[0] != "Basic" {
			t.Error("bad syntax")
		}
		payload, _ := base64.StdEncoding.DecodeString(auth[1])
		pair := strings.SplitN(string(payload), ":", 2)
		if pair[0] != "myusername" || pair[1] != "mypassword" {
			t.Error("Wrong username/password")
		}
	}))
	defer ts.Close()
	New().Post(ts.URL).
		SetBasicAuth("myusername", "mypassword").
		End()
}

func TestXml(t *testing.T) {
	xml := `<note><to>Tove</to><from>Jani</from><heading>Reminder</heading><body>Don't forget me this weekend!</body></note>`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check method is PATCH before going to check other features
		if r.Method != POST {
			t.Errorf("Expected method %q; got %q", POST, r.Method)
		}
		if r.Header == nil {
			t.Error("Expected non-nil request Header")
		}

		if r.Header.Get("Content-Type") != "application/xml" {
			t.Error("Expected Header Content-Type -> application/xml", "| but got", r.Header.Get("Content-Type"))
		}

		defer r.Body.Close()
		body, _ := ioutil.ReadAll(r.Body)
		if string(body) != xml {
			t.Error(`Expected XML `, xml, "| but got", string(body))
		}
	}))

	defer ts.Close()

	New().Post(ts.URL).
		Type("xml").
		Send(xml).
		End()

	New().Post(ts.URL).
		Set("Content-Type", "application/xml").
		Send(xml).
		End()
}

func TestPlainText(t *testing.T) {
	text := `hello world \r\n I am GoRequest`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check method is PATCH before going to check other features
		if r.Method != POST {
			t.Errorf("Expected method %q; got %q", POST, r.Method)
		}
		if r.Header == nil {
			t.Error("Expected non-nil request Header")
		}
		if r.Header.Get("Content-Type") != "text/plain" {
			t.Error("Expected Header Content-Type -> text/plain", "| but got", r.Header.Get("Content-Type"))
		}

		defer r.Body.Close()
		body, _ := ioutil.ReadAll(r.Body)
		if string(body) != text {
			t.Error(`Expected text `, text, "| but got", string(body))
		}
	}))

	defer ts.Close()

	New().Post(ts.URL).
		Type("text").
		Send(text).
		End()

	New().Post(ts.URL).
		Set("Content-Type", "text/plain").
		Send(text).
		End()
}

func TestAsCurlCommand(t *testing.T) {
	var (
		endpoint = "http://github.com/parnurzeal/gorequest"
		jsonData = `{"here": "is", "some": {"json": ["data"]}}`
	)

	request := New().Timeout(10*time.Second).Put(endpoint).Set("Content-Type", "application/json").Send(jsonData)

	curlComand, err := request.AsCurlCommand()
	if err != nil {
		t.Fatal(err)
	}

	expected := fmt.Sprintf(`curl -X 'PUT' -d '%v' -H 'Content-Type: application/json' '%v'`, strings.Replace(jsonData, " ", "", -1), endpoint)
	if curlComand != expected {
		t.Fatalf("\nExpected curlCommand=%v\n   but actual result=%v", expected, curlComand)
	}
}

func TestSetDebugByEnvironmentVar(t *testing.T) {
	endpoint := "http://github.com/parnurzeal/gorequest"

	var buf bytes.Buffer
	logger := log.New(&buf, "[gorequest]", log.LstdFlags)

	os.Setenv("GOREQUEST_DEBUG", "1")
	New().SetLogger(logger).Get(endpoint).End()

	if len(buf.String()) == 0 {
		t.Fatalf("\nExpected gorequest to log request and response object if GOREQUEST_DEBUG=1")
	}

	os.Setenv("GOREQUEST_DEBUG", "")
	buf.Reset()

	New().SetLogger(logger).Get(endpoint).End()

	if len(buf.String()) > 0 {
		t.Fatalf("\nExpected gorequest not to log request and response object if GOREQUEST_DEBUG is not set.")
	}
}
