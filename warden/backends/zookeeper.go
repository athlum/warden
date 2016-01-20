package backends

import (
	"crypto/md5"
	"fmt"
	"github.com/athlum/warden/warden/handlers"
	"github.com/samuel/go-zookeeper/zk"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

type ZKConn struct {
	*zk.Conn
	Auth string
}

var (
	zkClient         ZKConn
	ZKLock           = new(sync.RWMutex)
	watchedZnode     = make(map[string][16]byte)
	watchedZnodeLock = new(sync.RWMutex)
)

func ZKClient() ZKConn {
	return zkClient
}

func NewZKConn(zkHost []string, auth string, timeout time.Duration) error {
	ZKLock.Lock()
	defer ZKLock.Unlock()
	if conn, _, err := zk.Connect(zkHost, timeout); err != nil {
		return err
	} else {
		zkClient = ZKConn{conn, auth}
	}

	if err := zkClient.AddAuth("digest", []byte(auth)); err != nil {
		return err
	}
	return nil
}

func (conn *ZKConn) syncNodes(path string, appId string, children []string, handler handlers.NodeHandler) error {
	nodes := []AppNode{}
	nodeList := []string{}
	for _, value := range children {
		childPath := fmt.Sprintf("%s/%s", path, value)
		node, _, getErr := conn.Get(childPath)
		if getErr != nil {
			return getErr
		}
		nodes = append(nodes, AppNode{value, string(node)})
		nodeList = append(nodeList, string(node))
	}

	sort.Strings(nodeList)
	md5Code := md5.Sum([]byte(strings.Join(nodeList, ",")))

	watchedZnodeLock.Lock()
	defer watchedZnodeLock.Unlock()
	log.Println("sync nodes")
	log.Printf("last watchedZnode: %x", watchedZnode[appId])
	log.Printf("current md5string: %x", md5Code)
	if watchedZnode[appId] != md5Code {
		registError := handler.Register(appId, nodes)
		if registError != nil {
			return registError
		}
		watchedZnode[appId] = md5Code
	} else {
		log.Println("skipped.")
	}

	return nil
}

func (conn *ZKConn) Watch(appId string, handler handlers.NodeHandler, quit *chan string) {
	path := fmt.Sprintf("/app/%s", appId)
	for {
		children, _, eventChan, childErr := conn.ChildrenW(path)
		if childErr != nil {
			log.Println("children watch err:", childErr)
			continue
		}
		log.Println("new watcher on", path, "created.")
		if err := conn.syncNodes(path, appId, children, handler); err != nil {
			log.Println("sync err:", err)
			continue
		}

		select {
		case event := <-eventChan:
			log.Println(event.Type, "reached.")
			if event.Type == zk.EventNodeChildrenChanged {
				children, _, childErr := conn.Children(path)
				if childErr != nil {
					log.Println("read children:", childErr)
					continue
				}
				if err := conn.syncNodes(path, appId, children, handler); err != nil {
					log.Println("sync err:", err)
					continue
				}
			}
			log.Println("watcher on", path, "exited.")
		case <-*quit:
			return
		}
	}
}

func (conn *ZKConn) checkRoot(path string, acl []zk.ACL) error {
	tempPath := string(path[0])
	pathList := strings.Split(string(path[1:]), "/")
	for _, value := range pathList[:len(pathList)-1] {
		tempPath += value
		exists, _, err := conn.Exists(tempPath)
		if err != nil {
			log.Fatal(err)
		}
		if !exists {
			if _, newErr := conn.Create(tempPath, []byte{}, 0, acl); newErr != nil {
				return newErr
			}
		}
		tempPath += "/"
	}
	return nil
}

func (conn *ZKConn) NewNode(path string, value string) error {
	auth := strings.Split(conn.Auth, ":")
	readOnlyACL := append([]zk.ACL{}, zk.DigestACL(zk.PermAll, auth[0], auth[1])...)
	readOnlyACL = append(readOnlyACL, zk.WorldACL(zk.PermRead)...)
	if err := conn.checkRoot(path, readOnlyACL); err != nil {
		return err
	}

	retry := true
	var err error
	for retry {
		_, err := conn.Create(path, []byte(value), zk.FlagEphemeral, readOnlyACL)
		if err == nil {
			log.Printf("Node %s created", path)
		} else if err == zk.ErrNodeExists {
			log.Printf("Node %s existed, wait and retry", path)
			time.Sleep(time.Second * 3)
			continue
		}
		retry = false
	}
	return err
}

func (conn *ZKConn) RemoveNode(path string) error {
	err := conn.Delete(path, 0)
	if err == nil {
		log.Printf("Node %s deleted", path)
	}
	return err
}
