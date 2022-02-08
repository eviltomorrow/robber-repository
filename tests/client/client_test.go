package client

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/eviltomorrow/robber-repository/pkg/client"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestVersion(t *testing.T) {
	stub, close, err := client.NewClientForRepository()
	if err != nil {
		t.Fatal(err)
	}
	defer close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	repley, err := stub.Version(ctx, &emptypb.Empty{})
	if err != nil {
		log.Fatalf("Version error: %v", err)
	}
	fmt.Println(repley.Value)
}

func BenchmarkVersion(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if err := getVersion(); err != nil {
			b.Fatal(err)
		}
	}
}

func getVersion() error {
	stub, close, err := client.NewClientForRepository()
	if err != nil {
		return err
	}
	defer close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = stub.Version(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}
	return nil
}
