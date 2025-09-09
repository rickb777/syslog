package syslog

import (
	"testing"

	"github.com/rickb777/expect"
)

func TestParseFacilities(t *testing.T) {
	expect.Slice(ParseFacilities("user")).ToBe(t, User)
	expect.Slice(ParseFacilities("auth,daemon")).ToBe(t, Auth, Daemon)
	expect.Error(ParseFacilities("foo,bar")).ToContain(t, "foo:")
}

func TestFacilitiesFilter(t *testing.T) {
	fs, _ := ParseFacilities("user")
	expect.Bool(fs.Filter()(&Message{Facility: User})).ToBeTrue(t)
	expect.Bool(fs.Filter()(&Message{Facility: Auth})).ToBeFalse(t)
}
