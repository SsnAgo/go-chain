package main

import (
	// 导入必要的包
	"fmt"
	"go-chain/network"
	"log"
	"os"
	"time"
)

func main() {
	// 创建三个服务器节点
	servers := make([]*network.Server, 3)
	addresses := []string{":9977", ":9978", ":9979"}

	for i := 0; i < 3; i++ {
		// 构建种子节点列表，不包括自身
		seedNodes := make([]string, 0, 2)
		for j, addr := range addresses {
			if j != i {
				seedNodes = append(seedNodes, "localhost"+addr)
			}
		}

		// 创建服务器选项
		opts := network.NewServerOpts(
			network.WithListenAddr(addresses[i]),
			network.WithSeedNodes(seedNodes),
			network.WithLogger(log.New(os.Stdout, fmt.Sprintf("服务器%d: ", i+1), log.LstdFlags)),
		)

		// 创建新服务器
		server, err := network.NewServer(*opts)
		if err != nil {
			fmt.Printf("创建服务器%d失败: %v\n", i+1, err)
			return
		}

		servers[i] = server
	}

	// 启动所有服务器
	for i, server := range servers {
		if err := server.Start(); err != nil {
			fmt.Printf("启动服务器%d失败: %v\n", i+1, err)
			return
		}
		fmt.Printf("服务器%d已启动\n", i+1)
	}

	// 让服务器运行一段时间
	fmt.Println("服务器正在运行...")
	time.Sleep(30 * time.Second)

	fmt.Println("程序结束")
}


