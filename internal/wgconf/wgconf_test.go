package wgconf

import (
	"reflect"
	"testing"
	"time"
)

func TestWGConf_GetFriendlyNameMap(t *testing.T) {
	c, err := New("../../test/wg0.conf")
	if err != nil {
		t.Errorf("Cannot read test file: %s", err.Error())
	}

	// Test case 1
	// Normal

	wantNames := map[string]string{
		"i+VdaJmF7mSlQlDQnEuFbo1JFicB2X054uN0DF5MICA=": "1st person",
		"63clN7mNlJ7ckYH7VirX1VyAfXwR4t9DP9DRp2qMu0o=": "2nd person (comment in PublicKey)",
		"bws0GsCPM0IT8OSgVirk6lgiRcOw6Ga3X62plId+PBU=": "(no name)",
		"NyPEExViZP/KuPYkYPNAqd6jo3xrfy8yBGSKrEaKPyI=": "(no name)",
		"not described": "",
	}
	gotNames, err := c.GetFriendlyNameMap()
	if err != nil {
		t.Errorf("(Normal) WGConf.GetFriendlyNameMap() error = %v, wantErr %v", err, false)
		return
	}
	for k, v := range wantNames {
		if !reflect.DeepEqual(gotNames[k], v) {
			t.Errorf("(Normal) WGConf.GetFriendlyNameMap() gotNames[%s] = %v, want %v", k, gotNames[k], v)
		}
	}

	// Test case 2
	// return cached data when wgconf file modified time is not changed.

	// insert "cached" data
	c.friendryNameMap["i+VdaJmF7mSlQlDQnEuFbo1JFicB2X054uN0DF5MICA="] = "cached data"
	c.friendryNameMap["cached data"] = "(cached data)"
	wantNames["i+VdaJmF7mSlQlDQnEuFbo1JFicB2X054uN0DF5MICA="] = "cached data"
	wantNames["cached data"] = "(cached data)"
	gotNames, err = c.GetFriendlyNameMap()
	if err != nil {
		t.Errorf("(Cached) WGConf.GetFriendlyNameMap() error = %v, wantErr %v", err, false)
		return
	}
	for k, v := range wantNames {
		if !reflect.DeepEqual(gotNames[k], v) {
			t.Errorf("(Cached) WGConf.GetFriendlyNameMap() gotNames[%s] = %v, want %v", k, gotNames[k], v)
		}
	}

	// Test case 3
	// reload config file when modified time is changed.

	// change modified time
	c.ModTime = time.Unix(0, 0)
	c.friendryNameMap["i+VdaJmF7mSlQlDQnEuFbo1JFicB2X054uN0DF5MICA="] = "1st person"
	wantNames["i+VdaJmF7mSlQlDQnEuFbo1JFicB2X054uN0DF5MICA="] = "1st person"
	wantNames["cached data"] = ""
	gotNames, err = c.GetFriendlyNameMap()
	if err != nil {
		t.Errorf("(Reload) WGConf.GetFriendlyNameMap() error = %v, wantErr %v", err, false)
		return
	}
	for k, v := range wantNames {
		if !reflect.DeepEqual(gotNames[k], v) {
			t.Errorf("(Reload) WGConf.GetFriendlyNameMap() gotNames[%s] = %v, want %v", k, gotNames[k], v)
		}
	}

}
