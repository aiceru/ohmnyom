package firestore

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"cloud.google.com/go/firestore"
)

const firestoreEmulatorHostStr = "FIRESTORE_EMULATOR_HOST"

func KillEmulator() {
	cmd := exec.Command("bash", "-c", "ps aux | grep emulator | grep -v grep | awk '{print $2}'")
	out, _ := cmd.Output()
	var err error

	if len(out) < 1 {
		return
	}

	pidstrs := strings.Split(strings.Trim(string(out), "\n "), "\n")
	pids := make([]int, len(pidstrs))
	for i, s := range pidstrs {
		pids[i], err = strconv.Atoi(s)
		if err != nil {
			log.Fatal(err)
		}
	}
	for i := 0; i < len(pids); i++ {
		process, err := os.FindProcess(pids[i])
		if err == nil {
			// log.Printf("killing process %v", process.Pid)
			if err := process.Kill(); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func RunEmulator() {
	cmd := exec.Command("/Users/wooseok/google-cloud-sdk/bin/gcloud", "beta", "emulators", "firestore", "start",
		"--host-port=localhost")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	defer stderr.Close()

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		buf := make([]byte, 256, 256)
		for {
			n, err := stderr.Read(buf[:])
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatalf("reading stderr %v", err)
			}
			if n > 0 {
				d := string(buf[:n])
				// log.Print(d)
				if strings.Contains(d, "Dev App Server is now running.") {
					wg.Done()
					break
				}

				pos := strings.Index(d, firestoreEmulatorHostStr+"=")
				if pos > 0 {
					host := d[pos+len(firestoreEmulatorHostStr)+1 : len(d)-1]
					_ = os.Setenv(firestoreEmulatorHostStr, host)
				}
			}
		}
	}()

	wg.Wait()
}

func NewEmulatorClient(ctx context.Context) *firestore.Client {
	client, err := firestore.NewClient(ctx, "test")
	if err != nil {
		log.Fatal(err)
	}
	return client
}
