package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
        "runtime"

	//"github.com/genuinetools/amicontained/version"
	"github.com/genuinetools/pkg/cli"
	"github.com/jessfraz/bpfd/proc"
	"github.com/sirupsen/logrus"
	"github.com/tv42/httpunix"
	//"golang.org/x/sys/unix"
)

var (
	debug bool
	docker_sock_hunt bool
)

func main() {
	// Create a new cli program.
	p := cli.NewProgram()
	p.Name = "amicontained"
	p.Description = "A container introspection tool"

	// Set the GitCommit and Version.
	//p.GitCommit = version.GITCOMMIT
	//p.Version = version.VERSION

	// Setup the global flags.
	p.FlagSet = flag.NewFlagSet("ship", flag.ExitOnError)
	p.FlagSet.BoolVar(&debug, "d", false, "enable debug logging")
	p.FlagSet.BoolVar(&docker_sock_hunt, "s", false, "enable docker sock hunting")

	// Set the before function.
	p.Before = func(ctx context.Context) error {
		// Set the log level.
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		return nil
	}

	// Set the main program action.
	p.Action = func(ctx context.Context, args []string) error {
		goarch := runtime.GOARCH
                fmt.Printf("\nGOARCH:%s", goarch)
		// Container Runtime
		runtime := proc.GetContainerRuntime(0, 0)
		fmt.Printf("Container Runtime: %s\n", runtime)

		// Namespaces
		namespaces := []string{"pid"}
		fmt.Println("Has Namespaces:")
		for _, namespace := range namespaces {
			ns, err := proc.HasNamespace(namespace)
			if err != nil {
				fmt.Printf("\t%s: error -> %v\n", namespace, err)
				continue
			}
			fmt.Printf("\t%s: %t\n", namespace, ns)
		}

		// User Namespaces
		userNS, userMappings := proc.GetUserNamespaceInfo(0)
		fmt.Printf("\tuser: %t\n", userNS)
		if len(userMappings) > 0 {
			fmt.Println("User Namespace Mappings:")
			for _, userMapping := range userMappings {
				fmt.Printf("\tContainer -> %d\tHost -> %d\tRange -> %d\n", userMapping.ContainerID, userMapping.HostID, userMapping.Range)
			}
		}

		// AppArmor Profile
		aaprof := proc.GetAppArmorProfile(0)
		fmt.Printf("AppArmor Profile: %s\n", aaprof)

		// Capabilities
		caps, err := proc.GetCapabilities(0)
		if err != nil {
			logrus.Warnf("getting capabilities failed: %v", err)
		}
		if len(caps) > 0 {
			fmt.Println("Capabilities:")
			for k, v := range caps {
				if len(v) > 0 {
					fmt.Printf("\t%s -> %s\n", k, strings.Join(v, " "))
				}
			}
		}

		// Seccomp
		seccompMode := proc.GetSeccompEnforcingMode(0)
		fmt.Printf("Seccomp: %s\n", seccompMode)

		// arm64 broken
		// https://github.com/genuinetools/amicontained/pull/15/commits
  		//  seccompIter()

		// Docker.sock
		if docker_sock_hunt {
			// Docker.sock
			fmt.Println("Looking for Docker.sock")
			getValidSockets("/")
		}

		return nil
	}

	// Run our program.
	p.Run()
}

func walkpath(path string, info os.FileInfo, err error) error {
	if err != nil {
		if debug {
			fmt.Println(err)
		}
	} else {
		switch mode := info.Mode(); {
		case mode&os.ModeSocket != 0:
			if debug {
				fmt.Println("Valid Socket: " + path)
			}
			resp, err := checkSock(path)
			if err == nil {
				if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
					fmt.Println("Valid Docker Socket: " + path)
				} else {
					if debug {
						fmt.Println("Invalid Docker Socket: " + path)
					}
				}
				defer resp.Body.Close()
			} else {
				if debug {
					fmt.Println("Invalid Docker Socket: " + path)
				}
			}
		default:
			if debug {
				fmt.Println("Invalid Socket: " + path)
			}
		}
	}
	return nil
}

func getValidSockets(startPath string) ([]string, error) {
	err := filepath.Walk(startPath, walkpath)
	if err != nil {
		if debug {
			fmt.Println(err)
		}
		return nil, err
	}
	return nil, nil
}

func checkSock(path string) (*http.Response, error) {

	if debug {
		fmt.Println("[-] Checking Sock for HTTP: " + path)
	}
	u := &httpunix.Transport{
		DialTimeout:           100 * time.Millisecond,
		RequestTimeout:        1 * time.Second,
		ResponseHeaderTimeout: 1 * time.Second,
	}
	u.RegisterLocation("dockerd", path)
	var client = http.Client{
		Transport: u,
	}
	resp, err := client.Get("http+unix://dockerd/info")

	if resp == nil {
		return nil, err
	}
	return resp, nil
}
