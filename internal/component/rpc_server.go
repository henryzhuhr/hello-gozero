package component

import (
	"context"
	"time"

	// "hello-gozero/demo/internal/server"
	"hello-gozero/internal/config"
	"hello-gozero/internal/svc"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// RPCServerComponent GoZero RPC 服务组件
type RPCServerComponent struct {
	config config.Config
	svcCtx *svc.ServiceContext
	server *rest.Server
	ready  chan struct{}
}

// NewRPCServerComponent 创建 RPC 服务组件
func NewRPCServerComponent(config config.Config, svcCtx *svc.ServiceContext) *RPCServerComponent {
	return &RPCServerComponent{
		config: config,
		svcCtx: svcCtx,
		ready:  make(chan struct{}),
	}
}

// Name Implements [Component.Name]
func (r *RPCServerComponent) Name() string {
	return "RPC Server"
}

// Start Implements [Component.Start]
func (r *RPCServerComponent) Start(ctx context.Context) error {
	// 创建 RPC 服务
	c := r.config
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// demo.RegisterDemoServer(grpcServer, server.NewDemoServer(ctx))

		if c.RpcServerConf.Mode == service.DevMode || c.RpcServerConf.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	// 在此注册 RPC 服务处理器
	// e.g., demo.RegisterDemoServer(r.server, server.NewDemoServer(r.svcCtx))

	// 启动 RPC 服务（非阻塞）
	go func() {
		s.Start()
	}()

	// 等待服务启动
	time.Sleep(200 * time.Millisecond)

	// 健康检查：确保服务真正可用
	// if err := r.healthCheck(); err != nil {
	// 	return fmt.Errorf("RPC server health check failed: %w", err)
	// }

	close(r.ready)
	return nil
}

// Stop Implements [Component.Stop]
func (r *RPCServerComponent) Stop(ctx context.Context) error {
	r.server.Stop()
	return nil
}
