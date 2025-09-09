package syslog

import (
	"testing"

	"github.com/rickb777/expect"
)

func TestParseSeverities(t *testing.T) {
	expect.Slice(ParseSeverities("info")).ToBe(t, Info)
	expect.Slice(ParseSeverities("err,warning")).ToBe(t, Err, Warning)
	expect.Slice(ParseSeverities("error,warn")).ToBe(t, Err, Warning)
	expect.Error(ParseSeverities("foo,bar")).ToContain(t, "foo:")
}

func TestSeveritiesFilter(t *testing.T) {
	ss, _ := ParseSeverities("info")
	expect.Bool(ss.Filter()(&Message{Severity: Info})).ToBeTrue(t)
	expect.Bool(ss.Filter()(&Message{Severity: Warning})).ToBeFalse(t)
}

func TestParsePriorityFilter(t *testing.T) {
	expect.Error(ParsePriorityFilter("*")).ToContain(t, "*: invalid priority filter")
	expect.Error(ParsePriorityFilter("foo.bar")).ToContain(t, "foo: unknown facility")

	f, _ := ParsePriorityFilter("*.*")
	expect.Bool(f(&Message{Facility: User, Severity: Info})).ToBeTrue(t)

	f, _ = ParsePriorityFilter("user.*")
	expect.Bool(f(&Message{Facility: User, Severity: Info})).ToBeTrue(t)
	expect.Bool(f(&Message{Facility: Kern, Severity: Info})).ToBeFalse(t)

	f, _ = ParsePriorityFilter("*.info")
	expect.Bool(f(&Message{Facility: User, Severity: Info})).ToBeTrue(t)
	expect.Bool(f(&Message{Facility: User, Severity: Warning})).ToBeFalse(t)

	f, _ = ParsePriorityFilter("user.info")
	expect.Bool(f(&Message{Facility: User, Severity: Info})).ToBeTrue(t)
	expect.Bool(f(&Message{Facility: Kern, Severity: Info})).ToBeFalse(t)
	expect.Bool(f(&Message{Facility: User, Severity: Warning})).ToBeFalse(t)
}
