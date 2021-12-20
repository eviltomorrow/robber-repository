package server

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/eviltomorrow/robber-core/pkg/grpclb"
	"github.com/eviltomorrow/robber-core/pkg/httpclient"
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
	Port           = 27321
	Endpoints      = []string{}
	RevokeEtcdConn func() error
	Key            = "grpclb/service/repository"
	timeout        = 10 * time.Second

	server *grpc.Server
)

type GRPC struct {
	pb.UnimplementedServiceServer
}

// PushData(Service_PushDataServer) error
// GetStockFull(*emptypb.Empty, Service_GetStockFullServer) error
// GetQuoteLatest(*QuoteRequest, Service_GetQuoteLatestServer) error

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

// CreateTask(context.Context, *Task) (*emptypb.Empty, error)
// Complete(context.Context, *Task) (*emptypb.Empty, error)
// PushData(Service_PushDataServer) error
// GetStockFull(*emptypb.Empty, Service_GetStockFullServer) error
// GetQuoteLatest(*QuoteRequest, Service_GetQuoteLatestServer) error

func (g *GRPC) CreateTask(ctx context.Context, req *pb.Task) (*emptypb.Empty, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid parameter, task is nil")
	}

	tx, err := mysql.DB.Begin()
	if err != nil {
		return nil, nil
	}
	_, err = model.TaskWithSelectOne(tx, req.Date, timeout)
	if err == nil {
		tx.Rollback()
		return nil, fmt.Errorf("exist same date[%v] task", req.Date)
	}
	if err != sql.ErrNoRows {
		tx.Rollback()
		return nil, err
	}

	if _, err := model.TaskWithInsertOne(tx, &model.Task{Date: req.Date, CallbackURL: req.CallbackUrl}, timeout); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (g *GRPC) Complete(ctx context.Context, req *pb.Task) (*emptypb.Empty, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid parameter, tak is nil")
	}

	task, err := model.TaskWithSelectOne(mysql.DB, req.Date, timeout)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("not found task with date[%s]", req.Date)
	}
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, fmt.Errorf("panic: invalid task is nil")
	}

	resp, err := httpclient.GetHTTP(task.CallbackURL, timeout, nil)
	if err != nil {
		return nil, err
	}
	zlog.Info("Callback success", zap.String("url", task.CallbackURL), zap.String("result", resp))

	tx, err := mysql.DB.Begin()
	if err != nil {
		return nil, err
	}
	_, err = model.TaskWithUpdateOne(tx, req.Date, &model.Task{
		Completed:     1,
		MetadataCount: req.MetadataCount,
		StockCount:    req.StockCount,
		DayCount:      req.DayCount,
		WeekCount:     req.WeekCount,
	}, timeout)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (g *GRPC) PushData(req pb.Service_PushDataServer) error {
	var (
		timeout = 20 * time.Second
		size    = 50
		stocks  = make([]*model.Stock, 0, size)
		days    = make([]*model.Quote, 0, size)
		weeks   = make([]*model.Quote, 0, size)

		stockCount, dayCount, weekCount int64
		cache                           = make([]*pb.Metadata, 0, size)
	)
	for {
		data, err := req.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		stocks = append(stocks, &model.Stock{
			Code:            data.Code,
			Name:            data.Name,
			Suspend:         data.Suspend,
			CreateTimestamp: time.Now(),
		})

		cache = append(cache, data)

		if len(cache) >= size {
			for _, c := range cache {
				t, err := time.ParseInLocation("2006-01-02", c.Date, time.Local)
				if err != nil {
					zlog.Error("ParseInLocation date failure", zap.String("data", c.String()), zap.Error(err))
					continue
				}
				day, err := service.BuildQuoteDay(c, t)
				if err != nil {
					zlog.Error("BuildQuoteDay failure", zap.String("data", c.String()), zap.Error(err))
				} else {
					days = append(days, day)
				}
			}

			affected, err := service.SaveStocks(stocks, timeout)
			if err != nil {
				zlog.Error("SaveStocks failure", zap.Any("stocks", stocks), zap.Error(err))
			}
			stocks = stocks[:0]
			stockCount += affected

			affected, err = service.SaveQuotes(days, model.Day, timeout)
			if err != nil {
				zlog.Error("SaveQuotes day failure", zap.Any("days", days), zap.Error(err))
			}
			days = days[:0]
			dayCount += affected

			for _, c := range cache {
				t, err := time.ParseInLocation("2006-01-02", c.Date, time.Local)
				if err != nil {
					zlog.Error("ParseInLocation date failure", zap.String("data", c.String()), zap.Error(err))
					continue
				}

				if t.Weekday() == time.Friday {
					week, err := service.BuildQuoteWeek(c.Code, t)
					if err != nil {
						zlog.Error("BuildQuoteWeek failure", zap.String("data", c.String()), zap.Error(err))
					} else {
						weeks = append(weeks, week)
					}
				}
			}

			affected, err = service.SaveQuotes(weeks, model.Week, timeout)
			if err != nil {
				zlog.Error("SaveQuotes week failure", zap.Any("weeks", weeks), zap.Error(err))
			}
			weeks = weeks[:0]
			weekCount += affected

			cache = cache[:0]
		}
	}

	if len(cache) != 0 {
		for _, c := range cache {
			t, err := time.ParseInLocation("2006-01-02", c.Date, time.Local)
			if err != nil {
				zlog.Error("ParseInLocation date failure", zap.String("data", c.String()), zap.Error(err))
				continue
			}
			day, err := service.BuildQuoteDay(c, t)
			if err != nil {
				zlog.Error("BuildQuoteDay failure", zap.String("data", c.String()), zap.Error(err))
			} else {
				days = append(days, day)
			}
		}

		affected, err := service.SaveStocks(stocks, timeout)
		if err != nil {
			zlog.Error("SaveStocks failure", zap.Any("stocks", stocks), zap.Error(err))
		}
		stockCount += affected

		affected, err = service.SaveQuotes(days, model.Day, timeout)
		if err != nil {
			zlog.Error("SaveQuotes day failure", zap.Any("days", days), zap.Error(err))
		}
		dayCount += affected

		for _, c := range cache {
			t, err := time.ParseInLocation("2006-01-02", c.Date, time.Local)
			if err != nil {
				zlog.Error("ParseInLocation date failure", zap.String("data", c.String()), zap.Error(err))
				continue
			}

			if t.Weekday() == time.Friday {
				week, err := service.BuildQuoteWeek(c.Code, t)
				if err != nil {
					zlog.Error("BuildQuoteWeek failure", zap.String("data", c.String()), zap.Error(err))
				} else {
					weeks = append(weeks, week)
				}
			}
		}
		affected, err = service.SaveQuotes(weeks, model.Week, timeout)
		if err != nil {
			zlog.Error("SaveQuotes week failure", zap.Any("weeks", weeks), zap.Error(err))
		}

		weekCount += affected
	}

	return req.SendAndClose(&pb.Count{Stock: stockCount, Day: dayCount, Week: weekCount})
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
			middleware.StreamServerLogInterceptor,
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
