package teomon

import (
	"fmt"
	"testing"
)

func TestParameter(t *testing.T) {

	var par Parameter

	// Value type bool
	par = Parameter{address: "aaadddrrr", name: "online", val: true}

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
	par = Parameter{address: "aaadddrrr", name: "online", val: "string_value"}

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
	par = Parameter{address: "aaadddrrr", name: "online", val: uint32(121314)}

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
	par = Parameter{address: "aaadddrrr", name: "online", val: int32(-121314)}

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
	par = Parameter{address: "aaadddrrr", name: "online", val: -12}

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
	par = Parameter{address: "aaadddrrr", name: "online", val: []byte("Hello!")}

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
	par = Parameter{address: "aaadddrrr", name: "online", val: 3.14}

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
	par = Parameter{address: "aaadddrrr", name: "online", val: struct{}{}}

	data, err = par.MarshalBinary()
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("MarshalBinary:", data)

	par = Parameter{}
	err = par.UnmarshalBinary(data)
	if err == nil {
		err = fmt.Errorf("sucessfully unmarshal not supported type %T", par.val)
		t.Error(err)
		return
	}
	fmt.Println("UnmarshalBinary:", err, par)

}
