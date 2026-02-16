package governance

import (
	"reflect"
	"testing"
)

func TestCompatibilityMetadataNormalized(t *testing.T) {
	in := CompatibilityMetadata{
		RequiredFlags:   []string{"cap.z", "cap.a"},
		AdvertisedFlags: []string{"cap.c", "cap.b"},
	}
	out := in.Normalized()
	if !reflect.DeepEqual(out.RequiredFlags, []string{"cap.a", "cap.z"}) {
		t.Fatalf("required=%v", out.RequiredFlags)
	}
	if !reflect.DeepEqual(out.AdvertisedFlags, []string{"cap.b", "cap.c"}) {
		t.Fatalf("advertised=%v", out.AdvertisedFlags)
	}
}
