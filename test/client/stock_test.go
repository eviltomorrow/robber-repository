package client

import (
	"context"
	"fmt"
	"testing"

	"github.com/eviltomorrow/robber-repository/pkg/client"
	"github.com/eviltomorrow/robber-repository/pkg/pb"
)

func TestPushStock(t *testing.T) {
	stub, cancel, err := client.NewClientForRepository()
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	resp, err := stub.PushStock(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	var (
		i    = 0
		size = 5000
	)
	for {
		err := resp.Send(&pb.Stock{Code: fmt.Sprintf("%d", i)})
		if err != nil {
			t.Fatal(err)
		}
		i++
		if i >= size {
			break
		}
	}
	count, err := resp.CloseAndRecv()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("count: %v\r\n", count.String())
}
