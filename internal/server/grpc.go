package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/eviltomorrow/robber-core/pkg/grpclb"
	"github.com/eviltomorrow/robber-core/pkg/mysql"
	"github.com/eviltomorrow/robber-core/pkg/system"
	"github.com/eviltomorrow/robber-core/pkg/zlog"
	"github.com/eviltomorrow/robber-core/pkg/znet"
	"github.com/eviltomorrow/robber-repository/internal/middleware"
	"github.com/eviltomorrow/robber-repository/internal/model"
	"github.com/eviltomorrow/robber-repository/internal/service"
	"github.com/eviltomorrow/robber-repository/pkg/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	Host           = "0.0.0.0"
	Port           = 19092
	Endpoints      = []string{}
	RevokeEtcdConn func() error
	Key            = "grpclb/service/repository"

	server *grpc.Server
)

type GRPC struct {
	pb.UnimplementedServiceServer
}

func (g *GRPC) Version(ctx context.Context, _ *emptypb.Empty) (*wrapperspb.StringValue, error) {
	var buf bytes.Buffer
	buf.WriteString("Server: \r\n")
	buf.WriteString(fmt.Sprintf("   Robber-repository Version (Current): %s\r\n", system.MainVersion))
	buf.WriteString(fmt.Sprintf("   Go Version: %v\r\n", system.GoVersion))
	buf.WriteString(fmt.Sprintf("   Go OS/Arch: %v\r\n", system.GoOSArch))
	buf.WriteString(fmt.Sprintf("   Git Sha: %v\r\n", system.GitSha))
	buf.WriteString(fmt.Sprintf("   Git Tag: %v\r\n", system.GitTag))
	buf.WriteString(fmt.Sprintf("   Git Branch: %v\r\n", system.GitBranch))
	buf.WriteString(fmt.Sprintf("   Build Time: %v\r\n", system.BuildTime))
	buf.WriteString(fmt.Sprintf("   HostName: %v\r\n", system.HostName))
	buf.WriteString(fmt.Sprintf("   IP: %v\r\n", system.IP))
	buf.WriteString(fmt.Sprintf("   Running Time: %v\r\n", system.RunningTime()))
	return &wrapperspb.StringValue{Value: buf.String()}, nil
}

// Version(context.Context, *emptypb.Empty) (*wrapperspb.StringValue, error)
// PushQuoteWeek(context.Context, *wrapperspb.StringValue) (*Count, error)
// PushQuoteDay(Service_PushQuoteDayServer) error
// PushStock(Service_PushStockServer) error
// GetStockFull(*emptypb.Empty, Service_GetStockFullServer) error
// GetQuoteLatest(*QuoteRequest, Service_GetQuoteLatestServer) error

func (g *GRPC) PushQuoteDay(req pb.Service_PushQuoteDayServer) error {
	var (
		timeout = 10 * time.Second
		size    = 50
		total   int64
		success int64
		days    = make([]*model.Quote, 0, size)
	)
	for {
		quote, err := req.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		total++
		// 构建数据 day
		t, err := time.ParseInLocation("2006-01-02", quote.Date, time.Local)
		if err != nil {
			return err
		}
		day, err := service.BuildQuoteDay(quote, t)
		if err != nil {
			return err
		}
		days = append(days, day)
		if len(days) >= size {
			affected, err := service.SaveQuotes(days, model.Day, timeout)
			if err != nil {
				return err
			}
			days = days[:0]
			success += affected
		}
	}

	affected, err := service.SaveQuotes(days, model.Day, timeout)
	if err != nil {
		return err
	}
	success += affected

	return req.SendAndClose(&pb.Count{Total: total, Success: success})
}

func (g *GRPC) PushQuoteWeek(ctx context.Context, req *wrapperspb.StringValue) (*pb.Count, error) {
	date, err := time.ParseInLocation("2006-01-02", req.Value, time.Local)
	if err != nil {
		return nil, err
	}
	if date.Weekday() != time.Friday {
		return &pb.Count{}, nil
	}

	var (
		offset  int64 = 0
		limit   int64 = 50
		total   int64
		success int64
		timeout = 10 * time.Second
		weeks   = make([]*model.Quote, 0, limit)
	)
	for {
		stocks, err := model.StockWithSelectRange(mysql.DB, offset, limit, timeout)
		if err != nil {
			return nil, err
		}

		for _, stock := range stocks {
			total++
			week, err := service.BuildQuoteWeek(stock.Code, date)
			if err == service.ErrNoData {
				continue
			}
			if err != nil {
				return nil, err
			}
			weeks = append(weeks, week)
		}

		affected, err := service.SaveQuotes(weeks, model.Week, timeout)
		if err != nil {
			return nil, err
		}
		weeks = weeks[:0]
		success += affected

		if int64(len(stocks)) < limit {
			break
		}
	}

	return &pb.Count{Total: total, Success: success}, nil
}

func (g *GRPC) PushStock(req pb.Service_PushStockServer) error {
	var (
		timeout        = 10 * time.Second
		size           = 50
		total, success int64
		stocks         = make([]*model.Stock, 0, size)
	)
	for {
		stock, err := req.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		total++

		stocks = append(stocks, &model.Stock{
			Code:            stock.Code,
			Name:            stock.Name,
			Suspend:         stock.Suspend,
			CreateTimestamp: time.Now(),
		})
		if len(stocks) >= size {
			affected, err := service.SaveStocks(stocks, timeout)
			if err != nil {
				return err
			}
			stocks = stocks[:0]
			success += affected
		}
	}

	affected, err := service.SaveStocks(stocks, timeout)
	if err != nil {
		return err
	}
	success += affected
	return req.SendAndClose(&pb.Count{Total: total, Success: success})
}

func (g *GRPC) GetStockFull(_ *emptypb.Empty, resp pb.Service_GetStockFullServer) error {
	var (
		offset  int64 = 0
		limit   int64 = 100
		timeout       = 10 * time.Second
	)

	for {
		stocks, err := model.StockWithSelectRange(mysql.DB, offset, limit, timeout)
		if err != nil {
			return err
		}

		for _, stock := range stocks {
			if err := resp.Send(&pb.Stock{Code: stock.Code, Name: stock.Name, Suspend: stock.Suspend}); err != nil {
				return err
			}
		}

		if int64(len(stocks)) < limit {
			break
		}
		offset += limit
	}
	return nil
}

func (g *GRPC) GetQuoteLatest(req *pb.QuoteRequest, resp pb.Service_GetQuoteLatestServer) error {
	var (
		offset  int64 = 0
		limit   int64 = req.Limit
		mode    string
		timeout = 10 * time.Second
	)
	switch req.Mode {
	case pb.QuoteRequest_Day:
		mode = model.Day
	case pb.QuoteRequest_Week:
		mode = model.Week
	default:
		mode = model.Day
	}

	for {
		quotes, err := model.QuoteWithSelectRangeByDate(mysql.DB, mode, req.Date, offset, limit, timeout)
		if err != nil {
			return err
		}

		for _, quote := range quotes {
			if err := resp.Send(&pb.Quote{
				Code:            quote.Code,
				Open:            quote.Open,
				Close:           quote.Close,
				High:            quote.High,
				Low:             quote.Low,
				YesterdayClosed: quote.YesterdayClosed,
				Volume:          quote.Volume,
				Account:         quote.Account,
				Date:            quote.Date.Format("2006-01-02"),
				NumOfYear:       int32(quote.NumOfYear),
			}); err != nil {
				return err
			}
		}

		if int64(len(quotes)) < limit {
			break
		}
		offset += limit
	}
	return nil
}

func StartupGRPC() error {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", Host, Port))
	if err != nil {
		return err
	}

	server = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middleware.UnaryServerRecoveryInterceptor,
			middleware.UnaryServerLogInterceptor,
		),
		grpc.ChainStreamInterceptor(
			middleware.StreamServerRecoveryInterceptor,
		),
	)

	reflection.Register(server)
	pb.RegisterServiceServer(server, &GRPC{})

	localIp, err := znet.GetLocalIP2()
	if err != nil {
		return fmt.Errorf("get local ip failure, nest error: %v", err)
	}

	close, err := grpclb.Register(Key, localIp, Port, Endpoints, 10)
	if err != nil {
		return fmt.Errorf("register service to etcd failure, nest error: %v", err)
	}
	RevokeEtcdConn = func() error {
		close()
		return nil
	}

	go func() {
		if err := server.Serve(listen); err != nil {
			zlog.Fatal("GRPC Server startup failure", zap.Error(err))
		}
	}()
	return nil
}

func ShutdownGRPC() error {
	if server == nil {
		return nil
	}
	server.Stop()
	return nil
}
