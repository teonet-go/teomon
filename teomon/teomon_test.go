package teomon

import (
	"fmt"
	"testing"
)

func TestParameter(t *testing.T) {

	var par Parameter

	// Value type bool
	par = Parameter{Name: "online", Value: true}

	data, err := par.MarshalBinary()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("MarshalBinary:", data)

	par = Parameter{}
	err = par.UnmarshalBinary(data)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("UnmarshalBinary:", par)

	// Value type string
	par = Parameter{Name: "online", Value: "string_value"}

	data, err = par.MarshalBinary()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("MarshalBinary:", data)

	par = Parameter{}
	err = par.UnmarshalBinary(data)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("UnmarshalBinary:", par)

	// Value type uint32
	par = Parameter{Name: "online", Value: uint32(121314)}

	data, err = par.MarshalBinary()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("MarshalBinary:", data)

	par = Parameter{}
	err = par.UnmarshalBinary(data)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("UnmarshalBinary:", par)

	// Value type int32
	par = Parameter{Name: "online", Value: int32(-121314)}

	data, err = par.MarshalBinary()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("MarshalBinary:", data)

	par = Parameter{}
	err = par.UnmarshalBinary(data)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("UnmarshalBinary:", par)

	// Value type int
	par = Parameter{Name: "online", Value: -12}

	data, err = par.MarshalBinary()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("MarshalBinary:", data)

	par = Parameter{}
	err = par.UnmarshalBinary(data)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("UnmarshalBinary:", par)

	// Value type []byte
	par = Parameter{Name: "online", Value: []byte("Hello!")}

	data, err = par.MarshalBinary()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("MarshalBinary:", data)

	par = Parameter{}
	err = par.UnmarshalBinary(data)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("UnmarshalBinary:", par)

	// Value type float
	par = Parameter{Name: "online", Value: 3.14}

	data, err = par.MarshalBinary()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("MarshalBinary:", data)

	par = Parameter{}
	err = par.UnmarshalBinary(data)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("UnmarshalBinary:", par)

	// Value type unknown
	par = Parameter{Name: "online", Value: struct{}{}}

	data, err = par.MarshalBinary()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("MarshalBinary:", data)

	par = Parameter{}
	err = par.UnmarshalBinary(data)
	if err == nil {
		err = fmt.Errorf("sucessfully unmarshal not supported type %T", par.Value)
		t.Error(err)
		return
	}
	fmt.Println("UnmarshalBinary:", err, par)

}

func TestMetric(t *testing.T) {

	m := Metric{
		Address:    "qUzILis",
		AppShort:   "test-metric",
		AppVersion: "0.0.1",
		Params: &Parameters{
			m: make(map[string]interface{}),
		},
	}
	m.Params.Add(OnlineParam, true)
	m.Params.Add("num_users", 234)

	data, err := m.MarshalBinary()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("marshalled data:", data)

	mout := Metric{
		Params: &Parameters{
			m: make(map[string]interface{}),
		},
	}
	err = mout.UnmarshalBinary(data)
	if err != nil {
		t.Error(err)
		return
	}
	if m.Address != mout.Address || m.AppName != mout.AppName || m.AppShort != mout.AppShort || m.AppVersion != mout.AppVersion {
		t.Error("wrong unmarshal metric")
		return
	}
	if val, ok := mout.Params.Get(OnlineParam); !ok || val != true {
		t.Error("wrong unmarshal param online")
		return
	}
	if val, ok := mout.Params.Get("num_users"); !ok || val != 234 {
		t.Error("wrong unmarshal param num_users")
		return
	}

	fmt.Println("unmarshalled metric:", mout)
}
