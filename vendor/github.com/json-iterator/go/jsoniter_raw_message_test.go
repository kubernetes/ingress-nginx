package jsoniter

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_json_RawMessage(t *testing.T) {
	should := require.New(t)
	var data json.RawMessage
	should.Nil(Unmarshal([]byte(`[1,2,3]`), &data))
	should.Equal(`[1,2,3]`, string(data))
	str, err := MarshalToString(data)
	should.Nil(err)
	should.Equal(`[1,2,3]`, str)
}

func Test_jsoniter_RawMessage(t *testing.T) {
	should := require.New(t)
	var data RawMessage
	should.Nil(Unmarshal([]byte(`[1,2,3]`), &data))
	should.Equal(`[1,2,3]`, string(data))
	str, err := MarshalToString(data)
	should.Nil(err)
	should.Equal(`[1,2,3]`, str)
}

func Test_json_RawMessage_in_struct(t *testing.T) {
	type TestObject struct {
		Field1 string
		Field2 json.RawMessage
	}
	should := require.New(t)
	var data TestObject
	should.Nil(Unmarshal([]byte(`{"field1": "hello", "field2": [1,2,3]}`), &data))
	should.Equal(` [1,2,3]`, string(data.Field2))
	should.Equal(`hello`, data.Field1)
}

func Test_decode_map_of_raw_message(t *testing.T) {
	should := require.New(t)
	type RawMap map[string]*json.RawMessage
	b := []byte("{\"test\":[{\"key\":\"value\"}]}")
	var rawMap RawMap
	should.Nil(Unmarshal(b, &rawMap))
	should.Equal(`[{"key":"value"}]`, string(*rawMap["test"]))
	type Inner struct {
		Key string `json:"key"`
	}
	var inner []Inner
	Unmarshal(*rawMap["test"], &inner)
	should.Equal("value", inner[0].Key)
}

func Test_encode_map_of_raw_message(t *testing.T) {
	should := require.New(t)
	type RawMap map[string]*json.RawMessage
	value := json.RawMessage("[]")
	rawMap := RawMap{"hello": &value}
	output, err := MarshalToString(rawMap)
	should.Nil(err)
	should.Equal(`{"hello":[]}`, output)
}

func Test_encode_map_of_jsoniter_raw_message(t *testing.T) {
	should := require.New(t)
	type RawMap map[string]*RawMessage
	value := RawMessage("[]")
	rawMap := RawMap{"hello": &value}
	output, err := MarshalToString(rawMap)
	should.Nil(err)
	should.Equal(`{"hello":[]}`, output)
}

func Test_marshal_invalid_json_raw_message(t *testing.T) {
	type A struct {
		Raw json.RawMessage `json:"raw"`
	}
	message := []byte(`{}`)

	a := A{}
	should := require.New(t)
	should.Nil(ConfigCompatibleWithStandardLibrary.Unmarshal(message, &a))
	aout, aouterr := ConfigCompatibleWithStandardLibrary.Marshal(&a)
	should.Equal(`{"raw":null}`, string(aout))
	should.Nil(aouterr)
}
