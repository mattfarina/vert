package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/Masterminds/semver"
	"github.com/codegangsta/cli"
)

const name = "vert"
const version = "0.1.0"
const description = `Compare versions.

Vert is a tool for comparing two version strings, or comparing a version string
to a version range.

	$ vert ">1.0.0" 1.1.0
	1.1.0

	$ vert "<1.0.0" 1.1.0
	   # No output because nothing matched.

	$ vert -f "<1.0.0" 1.1.0
	1.1.0   # -f returns all failures, rather than matches.

	$ vert ">1.0.0" 1.1.0 1.1.1 1.2.3 0.1.1
	1.1.0
	1.1.1
	1.2.3

See below for information about how to determine the number of failed tests.

EXIT CODES:

vert returns exit codes based on the number of failed matches. There are a few
reserved exit codes:

- 128: The command was not called correctly.
- 256: A version failed to parse, and comparisons could not continue. This will
  occur if the original constraint version cannot be parsed. If any
  subsequent version fails to parse, it will simply be counted as a failure.

Any other error codes indicate the number of failed tests. For example:

	$ vert 1.2.3 1.2.3 1.2.4 1.2.5
	1.2.3
	$ echo $?
	2   # <-- Two tests failed.

BASE VERSIONS:

The base version may be in any of the following formats:

- An exact semantic version number
	- 1.2.3
	- v1.2.3
	- 1.2.3-alpha.1+10212015
- A semantic version range
	- *
	- !=1.0.0
	- >=1.2.3
	- >1.2.3,<1.3.2
	- ~1.2.0
	- ^2.3

VERSIONS:

Other than the base version, all other supplied versions must follow the
SemVer 2 spec. Examples:

	- 1.2.3
	- v1.2.3
	- 1.2.3-alpha.1+10212015
	- v1.2.3-alpha.1+10212015
`

func main() {
	app := cli.NewApp()
	app.Name = name
	app.Usage = description
	app.Action = run
	app.Version = version
	app.ArgsUsage = "BASE VERSION [VERSION [VERSION [...]]"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "failed,f",
			Usage: "Show the versions that failed rather than the ones that passed.",
		},
		cli.BoolFlag{
			Name:  "sort,s",
			Usage: "Sort the versions before printing. Without this, versions are returned in the order they were tested.",
		},
	}
	app.Run(os.Args)
}

// run handles all of the flags and then runs the main action.
func run(c *cli.Context) {
	args := c.Args()
	if len(args) < 2 {
		perr("Not enough arguments")
		os.Exit(128)
	}

	pass, fail := compare(args[0], args[1:])

	out := pass
	if c.Bool("failed") {
		out = fail
	}

	if c.Bool("sort") {
		sort.Sort(semver.Collection(out))
	}

	pvers(out)
}

// compare compiles a base version comparator, and then compares all cases to it.
//
// It retuns an array of versions that passed, and an array of versions that failed.
func compare(base string, cases []string) ([]*semver.Version, []*semver.Version) {
	constraint, err := semver.NewConstraint(base)
	if err != nil {
		perr("Could not parse constraint %s", base)
		os.Exit(256)
	}

	passed, failed := []*semver.Version{}, []*semver.Version{}
	for _, t := range cases {
		ver, err := semver.NewVersion(t)
		if err != nil {
			failed = append(failed, ver)
			perr("Failed to parse %s", t)
			continue
		}
		if constraint.Check(ver) {
			passed = append(passed, ver)
			continue
		}
		failed = append(failed, ver)
	}

	return passed, failed
}

// pvers prints a list of versions to standard out.
func pvers(vers []*semver.Version) {
	for _, v := range vers {
		fmt.Fprintln(os.Stdout, v.String())
	}
}

// pout prints to stdout.
func pout(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, msg, args...)
	fmt.Fprintln(os.Stdout)
}

// perr prints to stderr.
func perr(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg, args...)
	fmt.Fprintln(os.Stderr)
}