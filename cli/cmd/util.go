package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/criblio/scope/history"
	"github.com/criblio/scope/run"
	"github.com/criblio/scope/util"
	"github.com/spf13/cobra"
)

// sessionByID returns a session by ID, or if -1 (not set) returns last session
func sessionByID(id int) history.SessionList {
	var sessions history.SessionList
	if id == -1 {
		sessions = history.GetSessions().Last(1)
	} else {
		sessions = history.GetSessions().ID(id)
	}
	sessionCount := len(sessions)
	if sessionCount != 1 {
		util.ErrAndExit("error expected a single session, saw: %d", sessionCount)
	}
	return sessions
}

func promptClean(sl history.SessionList) {
	fmt.Print("Invalid session, likely an invalid command was scoped or a session file was modified. Would you like to delete this session? (default: yes) [y/n] ")
	var response string
	_, err := fmt.Scanf("%s", &response)
	util.CheckErrSprintf(err, "error reading response: %v", err)
	if !(response == "n" || response == "no") {
		sl.Remove()
	}
	os.Exit(0)
}

func helpErrAndExit(cmd *cobra.Command, errText string) {
	cmd.Help()
	fmt.Printf("\nerror: %s\n", errText)
	os.Exit(1)
}

func metricAndEventDestFlags(cmd *cobra.Command, rc *run.Config) {
	cmd.Flags().StringVarP(&rc.CriblDest, "cribldest", "c", "", "Set Cribl destination for metrics & events (host:port defaults to tls://)")
	cmd.Flags().StringVar(&rc.MetricsFormat, "metricformat", "ndjson", "Set format of metrics output (statsd|ndjson)")
	cmd.Flags().StringVarP(&rc.MetricsDest, "metricdest", "m", "", "Set destination for metrics (host:port defaults to tls://)")
	cmd.Flags().StringVarP(&rc.EventsDest, "eventdest", "e", "", "Set destination for events (host:port defaults to tls://)")
	cmd.Flags().BoolVarP(&rc.NoBreaker, "nobreaker", "n", false, "Set Cribl to not break streams into events.")
	cmd.Flags().StringVarP(&rc.AuthToken, "authtoken", "a", "", "Set AuthToken for Cribl")
}

func runCmdFlags(cmd *cobra.Command, rc *run.Config) {
	cmd.Flags().BoolVar(&rc.Passthrough, "passthrough", false, "Runs ldscope with current environment & no config.")
	cmd.Flags().IntVarP(&rc.Verbosity, "verbosity", "v", 4, "Set scope metric verbosity")
	cmd.Flags().BoolVarP(&rc.Payloads, "payloads", "p", false, "Capture payloads of network transactions")
	cmd.Flags().StringVar(&rc.Loglevel, "loglevel", "", "Set ldscope log level (debug, warning, info, error, none)")
	cmd.Flags().StringVarP(&rc.LibraryPath, "librarypath", "l", "", "Set path for dynamic libraries")
	cmd.Flags().StringVarP(&rc.UserConfig, "userconfig", "u", "", "Run ldscope with a user specified config file. Overrides all other settings.")
	metricAndEventDestFlags(cmd, rc)
}

/*
Incompatible flags list, key not present = no icompatibilities, key = nil exclusive flag, key = map incompatible flag list
--cribldest && --eventdest
--cribldest && --metricsdest
--userconfig && (--metricsdest || --eventsdest || --cribldest || --loglevel ....etc)
--help && [anything else]
--passthrough && [anything else]
*/
var IncompatibleFlags = map[string]map[string]int{
	"cribldest": {
		"eventdest":   1,
		"metricsdest": 1,
	},
	"metricsdest": {
		"cribldest": 1,
	},
	"eventdest": {
		"cribldest": 1,
	},
	"userconfig": {
		"metricsdest": 1,
		"eventsdest":  1,
		"cribldest":   1,
		"loglevel":    1,
	},
	"passthrough": nil,
	"help":        nil,
}

func checkIncompatibleFlags(flags []string) error {
	for _, fl := range flags {
		incompatible, exists := IncompatibleFlags[fl]
		if !exists {
			continue
		}
		if incompatible == nil {
			return fmt.Errorf("Flag \"%s\" can't be used with other flags", fl)
		}
		for _, nextFl := range flags {
			_, exist := incompatible[nextFl]
			if exist {
				return fmt.Errorf("Flag \"%s\" can't be used with \"%s\"", fl, nextFl)
			}
		}
	}
	return nil
}

func getFlags(args []string) []string {
	rv := make([]string, len(args))
	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			rv = append(rv, strings.Replace(arg, "--", "", 1))
		}
	}
	return rv
}
