package main

import "time"

// // func main() {
// // 	conn := connect()
// // 	defer conn.Close()

// // 	// flags := int32(0)
// // 	acl := zk.WorldACL(zk.PermAll)

// // 	// path, err := conn.Create("/s01", []byte("data"), flags, acl)
// // 	path, err := conn.CreateProtectedEphemeralSequential("/01", []byte("data"), acl)
// // 	must(err)
// // 	fmt.Printf("create: %+v\n", path)
// // 	time.Sleep(60 * time.Second)

// // 	data, stat, err := conn.Get("/01")
// // 	must(err)
// // 	fmt.Printf("get:    %+v %+v\n", string(data), stat)
// // 	stat, err = conn.Set("/01", []byte("newdata"), stat.Version)
// // 	must(err)
// // 	fmt.Printf("set:    %+v\n", stat)

// // 	err = conn.Delete("/01", -1)
// // 	must(err)
// // 	fmt.Printf("delete: ok\n")

// // 	exists, stat, err := conn.Exists("/01")
// // 	must(err)
// // 	fmt.Printf("exists: %+v %+v\n", exists, stat)
// // }

func main() {
	// 服务器地址列表
	servers := []string{"localhost:32772", "localhost:32773", "localhost:32774"}
	client, err := NewClient(servers, "/websocket", 10)
	if err != nil {
		panic(err)
	}
	defer client.Close()
	node1 := &ServiceNode{"connect", "127.0.0.1", 4000, 1}
	node2 := &ServiceNode{"connect", "127.0.0.1", 4001, 2}
	node3 := &ServiceNode{"connect", "127.0.0.1", 4002, 3}
	// node1Path := ""
	if _, err = client.Register(node1); err != nil {
		panic(err)
	}
	// changeNode := &ServiceNode{}
	// client.SetNode(node1Path, node1, 2)
	// client.WatchNode(node1Path, changeNode)
	if _, err := client.Register(node2); err != nil {
		panic(err)
	}
	if _, err := client.Register(node3); err != nil {
		panic(err)
	}
	// nodes, err := client.GetNodes("user")
	// if err != nil {
	// 	panic(err)
	// }
	// for _, node := range nodes {
	// 	fmt.Println(node.Host, node.Port)
	// }
	time.Sleep(120 * time.Second)
}
