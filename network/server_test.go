package network

import (
	"fmt"
	"testing"
)







func TestNewServerAndStart(t *testing.T) {
	tests := []struct {
		name    string
		opts    ServerOpts
		wantErr bool
	}{
		{
			name: "正常情况",
			opts: ServerOpts{
				listenAddr: ":9977",
				seedNodes:  []string{"127.0.0.1:8080"},
				privKey:    "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6a7b8c9d0e1f2a3b4c5d6a7b8c9d0e1f2a3b4c5d6a7b8c9d0e1f2",
			},
			wantErr: false,
		},
		// {
		// 	name: "无种子节点",
		// 	opts: ServerOpts{
		// 		listenAddr: ":9977",
		// 		seedNodes:  []string{},
		// 	},
		// 	wantErr: true,
		// },
		// {
		// 	name: "无监听地址",
		// 	opts: ServerOpts{
		// 		seedNodes: []string{"127.0.0.1:8080"},
		// 	},
		// 	wantErr: false,
		// },
		// {
		// 	name: "TCP传输层启动失败",
		// 	opts: ServerOpts{
		// 		listenAddr: "无效地址",
		// 		seedNodes:  []string{"127.0.0.1:8080"},
		// 	},
		// 	wantErr: true,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建新服务器
			s, err := NewServer(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return // 如果创建服务器失败，就不继续测试启动了
			}

			// 启动服务器
			err = s.Start()
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.Start() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 打印服务器信息（可选）
			fmt.Printf("服务器信息: %+v\n", s)
		})
	}
}
