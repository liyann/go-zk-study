package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

type ServiceNode struct {
	Name      string `json:"name"` // 服务名称，这里是user
	Host      string `json:"host"`
	Port      int    `json:"port"`
	ConnCount int    `json:connCount`
}
type SdClient struct {
	zkServers []string // 多个节点地址
	zkRoot    string   // 服务根节点，这里是/api
	conn      *zk.Conn // zk的客户端连接
}

func NewClient(zkServers []string, zkRoot string, timeout int) (*SdClient, error) {
	client := new(SdClient)
	client.zkServers = zkServers
	client.zkRoot = zkRoot
	// 连接服务器
	conn, _, err := zk.Connect(zkServers, time.Duration(timeout)*time.Second)
	if err != nil {
		return nil, err
	}
	client.conn = conn
	// 创建服务根节点
	if err := client.ensureRoot(); err != nil {
		client.Close()
		return nil, err
	}
	return client, nil
}

// 关闭连接，释放临时节点
func (s *SdClient) Close() {
	s.conn.Close()
}

func (s *SdClient) ensureRoot() error {
	exists, _, err := s.conn.Exists(s.zkRoot)
	if err != nil {
		return err
	}
	if !exists {
		_, err := s.conn.Create(s.zkRoot, []byte(""), 0, zk.WorldACL(zk.PermAll))
		if err != nil && err != zk.ErrNodeExists {
			return err
		}
	}
	return nil
}

func (s *SdClient) Register(node *ServiceNode) (fullPath string, err error) {
	// fullPath := ""
	if err := s.ensureName(node.Name); err != nil {
		return "", err
	}
	path := s.zkRoot + "/" + node.Name + "/n"
	data, err := json.Marshal(node)
	if err != nil {
		return "", err
	}
	fullPath, err = s.conn.CreateProtectedEphemeralSequential(path, data, zk.WorldACL(zk.PermAll))
	fmt.Println(fullPath)
	if err != nil {
		return "", err
	}
	return fullPath, nil
}

func (s *SdClient) ensureName(name string) error {
	path := s.zkRoot + "/" + name
	exists, _, err := s.conn.Exists(path)
	if err != nil {
		return err
	}
	if !exists {
		_, err := s.conn.Create(path, []byte(""), 0, zk.WorldACL(zk.PermAll))
		if err != nil && err != zk.ErrNodeExists {
			return err
		}
	}
	return nil
}

// 先要创建/api/user节点作为服务列表的父节点。然后创建一个保护顺序临时(ProtectedEphemeralSequential)子节点，同时将地址信息存储在节点中。什么叫保护顺序临时节点，首先它是一个临时节点，会话关闭后节点自动消失。其它它是个顺序节点，zookeeper自动在名称后面增加自增后缀，确保节点名称的唯一性。同时还是个保护性节点，节点前缀增加了GUID字段，确保断开重连后临时节点可以和客户端状态对接上。接下来我们实现消费者获取服务列表方法
func (s *SdClient) GetNodes(name string) ([]*ServiceNode, error) {
	path := s.zkRoot + "/" + name
	// 获取字节点名称
	childs, _, err := s.conn.Children(path)
	if err != nil {
		if err == zk.ErrNoNode {
			return []*ServiceNode{}, nil
		}
		return nil, err
	}
	nodes := []*ServiceNode{}
	for _, child := range childs {
		fullPath := path + "/" + child
		data, _, err := s.conn.Get(fullPath)
		if err != nil {
			if err == zk.ErrNoNode {
				continue
			}
			return nil, err
		}
		node := new(ServiceNode)
		err = json.Unmarshal(data, node)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (s *SdClient) SetNode(nodePath string, node *ServiceNode, connCount int) {
	_, stat, err := s.conn.Get(nodePath)
	if err != nil {
		panic(err)
	}
	node.ConnCount = connCount
	data, err := json.Marshal(node)
	if err != nil {
		log.Println("json解析错误", err)
	}

	if _, err = s.conn.Set(nodePath, data, stat.Version); err != nil {
		panic(err)
	}

}

func (s *SdClient) WatchNode(nodePath string, changeNode *ServiceNode) {
	found, _, ech, err := s.conn.ExistsW(nodePath)
	must(err)
	if !found {
		log.Println("not found")
	}
	evt := <-ech
	fmt.Println("watch fired", evt)
	must(evt.Err)

	data, stat, err := s.conn.Get(nodePath)
	must(err)
	fmt.Printf("get:    %+v %+v\n", string(data), stat)
	json.Unmarshal(data, changeNode)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func connect() *zk.Conn {
	// zksStr := os.Getenv("ZOOKEEPER_SERVERS")
	// zks := strings.Split(zksStr, ",")
	zks := []string{"localhost:32769", "localhost:32770", "localhost:32771"}
	conn, _, err := zk.Connect(zks, time.Second)
	must(err)
	return conn
}

// func main() {
// 	conn := connect()
// 	defer conn.Close()

// 	// flags := int32(0)
// 	acl := zk.WorldACL(zk.PermAll)

// 	// path, err := conn.Create("/s01", []byte("data"), flags, acl)
// 	path, err := conn.CreateProtectedEphemeralSequential("/01", []byte("data"), acl)
// 	must(err)
// 	fmt.Printf("create: %+v\n", path)
// 	time.Sleep(60 * time.Second)

// 	data, stat, err := conn.Get("/01")
// 	must(err)
// 	fmt.Printf("get:    %+v %+v\n", string(data), stat)
// 	stat, err = conn.Set("/01", []byte("newdata"), stat.Version)
// 	must(err)
// 	fmt.Printf("set:    %+v\n", stat)

// 	err = conn.Delete("/01", -1)
// 	must(err)
// 	fmt.Printf("delete: ok\n")

// 	exists, stat, err := conn.Exists("/01")
// 	must(err)
// 	fmt.Printf("exists: %+v %+v\n", exists, stat)
// }
