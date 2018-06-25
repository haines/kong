package kong_test

import (
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/alecthomas/kong"
)

func TestValueMapper(t *testing.T) {
	var cli struct {
		Flag string
	}
	k := mustNew(t, &cli, kong.ValueMapper(&cli.Flag, testMooMapper{}))
	_, err := k.Parse(nil)
	require.NoError(t, err)
	require.Equal(t, "", cli.Flag)
	_, err = k.Parse([]string{"--flag"})
	require.NoError(t, err)
	require.Equal(t, "MOO", cli.Flag)
}

func TestNamedMapper(t *testing.T) {
	var cli struct {
		Flag string `type:"moo"`
	}
	k := mustNew(t, &cli, kong.NamedMapper("moo", testMooMapper{}))
	_, err := k.Parse(nil)
	require.NoError(t, err)
	require.Equal(t, "", cli.Flag)
	_, err = k.Parse([]string{"--flag"})
	require.NoError(t, err)
	require.Equal(t, "MOO", cli.Flag)
}

type testMooMapper struct {
	text string
}

func (t testMooMapper) Decode(ctx *kong.DecodeContext, target reflect.Value) error {
	if t.text == "" {
		target.SetString("MOO")
	} else {
		target.SetString(t.text)
	}
	return nil
}
func (testMooMapper) IsBool() bool { return true }

func TestTimeMapper(t *testing.T) {
	var cli struct {
		Flag time.Time `format:"2006"`
	}
	k := mustNew(t, &cli)
	_, err := k.Parse([]string{"--flag=2008"})
	require.NoError(t, err)
	expected, err := time.Parse("2006", "2008")
	require.NoError(t, err)
	require.Equal(t, 2008, expected.Year())
	require.Equal(t, expected, cli.Flag)
}

func TestDurationMapper(t *testing.T) {
	var cli struct {
		Flag time.Duration
	}
	k := mustNew(t, &cli)
	_, err := k.Parse([]string{"--flag=5s"})
	require.NoError(t, err)
	require.Equal(t, time.Second*5, cli.Flag)
}

func TestSplitEscaped(t *testing.T) {
	require.Equal(t, []string{"a", "b"}, kong.SplitEscaped("a,b", ','))
	require.Equal(t, []string{"a,b", "c"}, kong.SplitEscaped(`a\,b,c`, ','))
}

func TestJoinEscaped(t *testing.T) {
	require.Equal(t, `a,b`, kong.JoinEscaped([]string{"a", "b"}, ','))
	require.Equal(t, `a\,b,c`, kong.JoinEscaped([]string{`a,b`, `c`}, ','))
	require.Equal(t, kong.JoinEscaped(kong.SplitEscaped(`a\,b,c`, ','), ','), `a\,b,c`)
}

func TestMapWithNamedTypes(t *testing.T) {
	var cli struct {
		TypedValue map[string]string `type:":moo"`
		TypedKey   map[string]string `type:"upper:"`
	}
	k := mustNew(t, &cli, kong.NamedMapper("moo", testMooMapper{}), kong.NamedMapper("upper", testUppercaseMapper{}))
	_, err := k.Parse([]string{"--typed-value", "first=5s", "--typed-value", "second=10s"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"first": "MOO", "second": "MOO"}, cli.TypedValue)
	_, err = k.Parse([]string{"--typed-key", "first=5s", "--typed-key", "second=10s"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"FIRST": "5s", "SECOND": "10s"}, cli.TypedKey)
}

func TestURLMapper(t *testing.T) {
	var cli struct {
		URL *url.URL `arg:""`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"http://w3.org"})
	require.NoError(t, err)
	require.Equal(t, "http://w3.org", cli.URL.String())
	_, err = p.Parse([]string{":foo"})
	require.Error(t, err)
}

func TestSliceConsumesRemainingPositionalArgs(t *testing.T) {
	var cli struct {
		Remainder []string `arg:""`
	}
	p := mustNew(t, &cli)
	_, err := p.Parse([]string{"--", "ls", "-lart"})
	require.NoError(t, err)
	require.Equal(t, []string{"ls", "-lart"}, cli.Remainder)
}