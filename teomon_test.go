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
		err = fmt.Errorf("successfully unmarshal not supported type %T", par.Value)
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
	m.Params.Add(ParamOnline, true)
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
	if val, ok := mout.Params.Get(ParamOnline); !ok || val != true {
		t.Error("wrong unmarshal param online")
		return
	}
	if val, ok := mout.Params.Get("num_users"); !ok || val != 234 {
		t.Error("wrong unmarshal param num_users")
		return
	}

	fmt.Println("unmarshalled metric:", mout)
}

func TestPeersDel(t *testing.T) {

	var peers *Peers
	var m *Metric

	peers = NewPeers()

	m = NewMetric()
	m.Address = "qUzILis-1"
	peers.Add(m)

	m = NewMetric()
	m.Address = "qUzILis-2"
	peers.Add(m)

	m = NewMetric()
	m.Address = "qUzILis-3"
	peers.Add(m)

	m = NewMetric()
	m.Address = "qUzILis-4"
	peers.Add(m)

	m = NewMetric()
	m.Address = "qUzILis-5"
	peers.Add(m)

	// Delete first record
	peers.Del("qUzILis-1")

	// Delete last record
	peers.Del("qUzILis-5")

	// Delete midle record
	peers.Del("qUzILis-3")

	// Check length
	l := len(peers.metrics)
	if l != 2 {
		t.Error("wrong peers length", l)
		return
	}
	fmt.Println(peers)

	// Delete last two peers
	peers.Del("qUzILis-2")
	peers.Del("qUzILis-4")

	// Check length
	l = len(peers.metrics)
	if l != 0 {
		t.Error("wrong peers length", l)
		return
	}
	fmt.Println(peers)
}

func TestSortMetric(t *testing.T) {

	for i := 0; i < 2; i++ {
		p := NewPeers()

		m1 := NewMetric()
		m1.AppShort = "app-01"
		m1.Address = "a1"
		p.Add(m1) // set online true

		m2 := NewMetric()
		m2.AppShort = "app-02"
		m2.Address = "a2"
		p.Add(m2)
		m2.Params.Add(ParamOnline, false)
		p.Add(m2)

		m3 := NewMetric()
		m3.AppShort = "app-03"
		m3.Address = "a3"
		p.Add(m3) // set online true

		m4 := NewMetric()
		m4.AppShort = "app-04"
		m4.Address = "a4"
		p.Add(m4)
		m4.Params.Add(ParamOnline, false)
		p.Add(m4)

		m5 := NewMetric()
		m5.AppShort = "_app-05"
		m5.Address = "a5"
		p.Add(m5) // set online true
		m5.Params.Add(ParamOnline, false)
		p.Add(m5)

		switch i {
		case 0:
			msg, _ := p.Json()
			fmt.Println(string(msg)) // metrics sort inside p.String() or p.Json
		default:
			msg := p.String()
			fmt.Println(msg) // metrics sort inside p.String() or p.Json
		}

		// Check sort
		if p.metrics[0].AppShort == "_app-05" &&
			p.metrics[1].AppShort == "app-02" &&
			p.metrics[2].AppShort == "app-04" &&
			p.metrics[3].AppShort == "app-01" &&
			p.metrics[4].AppShort == "app-03" {

		} else {
			t.Error("wrong metrics sort")
			return
		}
	}
}
