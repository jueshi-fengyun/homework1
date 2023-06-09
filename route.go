package webhomeworke1

import (
	"fmt"
	"regexp"
	"strings"
)

type node struct {
	path string
	// children 子节点
	// 子节点的 path => node
	children map[string]*node
	// handler 命中路由之后执行的逻辑
	handler HandleFunc
	// 通配符 * 表达的节点，任意匹配
	starChild *node
	// 参数路径
	paramChild *node
	paramName  string
	expr       string
	// 正则匹配
	regChild *node
}

type router struct {
	// trees 是按照 HTTP 方法来组织的
	// 如 GET => *node
	trees map[string]*node
}

func newRouter() router {
	return router{
		trees: map[string]*node{},
	}
}

// addRoute 注册路由。
// method 是 HTTP 方法
// path 必须以 / 开始并且结尾不能有 /，中间也不允许有连续的 /
func (r *router) addRoute(method string, path string, handler HandleFunc) {
	if path == "" {
		panic("web: 路由是空字符串")
	}
	if path[0] != '/' {
		panic("web: 路由必须以 / 开头")
	}

	if path != "/" && path[len(path)-1] == '/' {
		panic("web: 路由不能以 / 结尾")
	}

	root, ok := r.trees[method]
	// 这是一个全新的 HTTP 方法，创建根节点
	if !ok {
		// 创建根节点
		root = &node{path: "/"}
		r.trees[method] = root
	}
	if path == "/" {
		if root.handler != nil {
			panic("web: 路由冲突[/]")
		}
		root.handler = handler
		return
	}

	segs := strings.Split(path[1:], "/")
	// 开始一段段处理
	for _, s := range segs {
		if s == "" {
			panic(fmt.Sprintf("web: 非法路由。不允许使用 //a/b, /a//b 之类的路由, [%s]", path))
		}
		root = root.childOrCreate(s)
	}
	if root.handler != nil {
		panic(fmt.Sprintf("web: 路由冲突[%s]", path))
	}
	root.handler = handler
}

// findRoute 查找对应的节点
// 注意，返回的 node 内部 HandleFunc 不为 nil 才算是注册了路由
func (r *router) findRoute(method string, path string) (*node, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}

	if path == "/" {
		return root, true
	}

	segs := strings.Split(strings.Trim(path, "/"), "/")
	for _, s := range segs {
		var child *node
		child, ok = root.childOf(s)
		if !ok {
			//匹配*的通配符
			if root.path == "*" {
				return root, true
			}
			return nil, false
		}

		root = child

	}
	return root, true
}

func (n *node) childOf(path string) (*node, bool) {

	if n.children != nil {
		res, ok := n.children[path]
		return res, ok
	}
	if n.regChild != nil {
		found, err := regexp.MatchString(n.regChild.expr, path)
		if err != nil {
			return nil, false
		}
		return n.regChild, found
	}
	if n.paramChild != nil {
		return n.paramChild, true
	}

	if n.starChild != nil {
		return n.starChild, true
	}

	return nil, false
}

// childOrCreate 查找子节点，如果子节点不存在就创建一个
// 并且将子节点放回去了 children 中
func (n *node) childOrCreate(path string) *node {
	if path == "*" {
		if n.regChild != nil || n.paramChild != nil {
			panic(fmt.Sprintf("web: 路由冲突 同一个位置只能注册路径参数，通配符路由和正则路由中的一个 [%s]", path))
		}
		if n.starChild == nil {
			n.starChild = &node{path: path}
		}
		return n.starChild
	}
	if strings.HasPrefix(path, ":") && strings.Contains(path, "(") && strings.HasSuffix(path, ")") {
		if n.starChild != nil || n.paramChild != nil {
			panic(fmt.Sprintf("web: 路由冲突 同一个位置只能注册路径参数，通配符路由和正则路由中的一个 [%s]", path))
		}

		segs := strings.Split(path[1:], "(")
		if len(segs) != 2 {
			panic(fmt.Sprintf("web: 通配符路由[%s]格式不正确", path))
		}
		expr := segs[1][:len(segs[1])-1]
		if n.regChild != nil {
			if n.regChild.paramName != segs[0] || n.regChild.expr != expr {
				panic(fmt.Sprintf("web: 路由冲突，正则路由冲突，已有 %s，新注册 %s", n.regChild.path, path))
			}

		} else {
			n.regChild = &node{path: path, paramName: segs[0], expr: expr}
		}

		return n.regChild
	}

	if strings.HasPrefix(path, ":") && !strings.Contains(path, "(") && !strings.HasSuffix(path, ")") {
		if n.starChild != nil || n.regChild != nil {
			panic(fmt.Sprintf("web: 路由冲突 同一个位置只能注册路径参数，通配符路由和正则路由中的一个 [%s]", path))
		}
		if n.paramChild != nil {
			if n.paramChild.path != path {
				panic(fmt.Sprintf("web: 路由冲突，正则路由冲突，已有 %s，新注册 %s", n.regChild.path, path))
			}

		} else {
			n.paramChild = &node{path: path, paramName: path[1:]}
		}

		return n.paramChild

	}

	if n.children == nil {
		n.children = make(map[string]*node)
	}
	child, ok := n.children[path]
	if !ok {
		child = &node{path: path}
		n.children[path] = child
	}

	return child
}
