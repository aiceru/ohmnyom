// +build test

package user

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
	"testing"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

const firestoreEmulatorHostStr = "FIRESTORE_EMULATOR_HOST"

func setupTestData(ctx context.Context, client *firestore.Client) {
	for _, user := range users {
		_, err := client.Collection(userCollection).Doc(user.Id).Set(ctx, user)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func teardownTestData(ctx context.Context, client *firestore.Client) {
	iter := client.Collection(userCollection).Limit(10).Documents(ctx)
	batch := client.Batch()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		batch.Delete(doc.Ref)
	}
	_, err := batch.Commit(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func killEmulator() {
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
		_ = syscall.Kill(-pids[i], syscall.SIGKILL)
	}
}

func TestMain(m *testing.M) {
	killEmulator()
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

	var result int
	defer func() {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		os.Exit(result)
	}()

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
	result = m.Run()
}

func newTestClient(ctx context.Context) *firestore.Client {
	client, err := firestore.NewClient(ctx, "test")
	if err != nil {
		log.Fatal(err)
	}
	return client
}
