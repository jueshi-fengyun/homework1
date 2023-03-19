package webhomeworke1

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func Test_router_AddRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/login/:id",
		},
		{
			method: http.MethodGet,
			path:   "/user/:id(^[0-9]+$)",
		},
		//{
		//	method: http.MethodGet,
		//	path:   "/user/:id(^[0-9]+$)",
		//},
		{
			method: http.MethodGet,
			path:   "/host/*",
		},
		{
			method: http.MethodGet,
			path:   "/host/*/home",
		},
	}

	mockHandler := func(ctx *Context) {}
	r := newRouter()
	//能注册 /user/* /user/*/home ?
	//能重复注册/user/:id ?
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	// 非法用例
	r = newRouter()

	// 空字符串
	assert.PanicsWithValue(t, "web: 是空字符串", func() {
		r.addRoute(http.MethodGet, "", mockHandler)
	})

	// 前导没有 /
	assert.PanicsWithValue(t, "web: 路由 前导没有 /", func() {
		r.addRoute(http.MethodGet, "a/b/c", mockHandler)
	})

	// 后缀有 /
	assert.PanicsWithValue(t, "web: 路由以 / 结尾", func() {
		r.addRoute(http.MethodGet, "/a/b/c/", mockHandler)
	})

	// 根节点重复注册
	r.addRoute(http.MethodGet, "/", mockHandler)
	assert.PanicsWithValue(t, "web: 路由根节点重复注册", func() {
		r.addRoute(http.MethodGet, "/", mockHandler)
	})
	// 普通节点重复注册
	r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	assert.PanicsWithValue(t, "web: 路由普通节点重复注册", func() {
		r.addRoute(http.MethodGet, "/a/b/c", mockHandler)
	})

	// 多个 /
	assert.PanicsWithValue(t, "web: 非法路由。使用 //a/b, /a//b 之类的路由, [/a//b]", func() {
		r.addRoute(http.MethodGet, "/a//b", mockHandler)
	})
	assert.PanicsWithValue(t, "web: 非法路由。使用 //a/b, /a//b 之类的路由, [//a/b]", func() {
		r.addRoute(http.MethodGet, "//a/b", mockHandler)
	})

	//同一个位置只能注册路径参数，通配符路由和正则路由中的一个
	r.addRoute(http.MethodGet, "/user/*", mockHandler)
	assert.PanicsWithValue(t, "web: 非法路由,同一个位置只能注册路径参数，通配符路由和正则路由中的一个", func() {
		r.addRoute(http.MethodGet, "/user/:id", mockHandler)
	})
	assert.PanicsWithValue(t, "web: 非法路由,同一个位置只能注册路径参数，通配符路由和正则路由中的一个", func() {
		r.addRoute(http.MethodGet, "/user/:id(^[0-9]+$)", mockHandler)
	})

}

func (r router) equal(y router) (string, bool) {
	for k, v := range r.trees {
		yv, ok := y.trees[k]
		if !ok {
			return fmt.Sprintf("目标 router 里面没有方法 %s 的路由树", k), false
		}
		str, ok := v.equal(yv)
		if !ok {
			return k + "-" + str, ok
		}
	}
	return "", true
}

func (n *node) equal(y *node) (string, bool) {
	if y == nil {
		return "目标节点为 nil", false
	}
	if n.path != y.path {
		return fmt.Sprintf("%s 节点 path 不相等 x %s, y %s", n.path, n.path, y.path), false
	}

	nhv := reflect.ValueOf(n.handler)
	yhv := reflect.ValueOf(y.handler)
	if nhv != yhv {
		return fmt.Sprintf("%s 节点 handler 不相等 x %s, y %s", n.path, nhv.Type().String(), yhv.Type().String()), false
	}

	if len(n.children) != len(y.children) {
		return fmt.Sprintf("%s 子节点长度不等", n.path), false
	}
	if len(n.children) == 0 {
		return "", true
	}

	for k, v := range n.children {
		yv, ok := y.children[k]
		if !ok {
			return fmt.Sprintf("%s 目标节点缺少子节点 %s", n.path, k), false
		}
		str, ok := v.equal(yv)
		if !ok {
			return n.path + "-" + str, ok
		}
	}
	return "", true
}

func Test_router_findRoute(t *testing.T) {
	testRoutes := []struct {
		method string
		path   string
	}{
		//{
		//	method: http.MethodGet,
		//	path:   "/",
		//},
		//{
		//	method: http.MethodGet,
		//	path:   "/user",
		//},
		{
			method: http.MethodGet,
			path:   "/login/:id",
		},
		//{
		//	method: http.MethodGet,
		//	path:   "/user/:id(^[0-9]+$)",
		//},
		//{
		//	method: http.MethodPost,
		//	path:   "/order/create",
		//},
		{
			method: http.MethodGet,
			path:   "/host/*",
		},
		{
			method: http.MethodGet,
			path:   "/host/*/home",
		},
	}

	mockHandler := func(ctx *Context) {}

	testCases := []struct {
		name     string
		method   string
		path     string
		found    bool
		wantNode *node
	}{
		//{
		//	name:   "method not found",
		//	method: http.MethodHead,
		//},
		//{
		//	name:   "path not found",
		//	method: http.MethodGet,
		//	path:   "/abc",
		//},
		//{
		//	name:   "root",
		//	method: http.MethodGet,
		//	path:   "/",
		//	found:  true,
		//	wantNode: &node{
		//		path:    "/",
		//		handler: mockHandler,
		//	},
		//},
		//{
		//	name:   "user",
		//	method: http.MethodGet,
		//	path:   "/user",
		//	found:  true,
		//	wantNode: &node{
		//		path:    "user",
		//		handler: mockHandler,
		//	},
		//},
		//{
		//	name:   "no handler",
		//	method: http.MethodPost,
		//	path:   "/order",
		//	found:  true,
		//	wantNode: &node{
		//		path: "order",
		//	},
		//},
		//{
		//	name:   "two layer",
		//	method: http.MethodPost,
		//	path:   "/order/create",
		//	found:  true,
		//	wantNode: &node{
		//		path:    "create",
		//		handler: mockHandler,
		//	},
		//},
		{
			name:   "*末尾通配符1",
			method: http.MethodGet,
			path:   "/host/create",
			found:  true,
			wantNode: &node{
				path:    "*",
				handler: mockHandler,
			},
		},
		{
			name:   "*末尾通配符2",
			method: http.MethodGet,
			path:   "/host/create/asdf",
			found:  true,
			wantNode: &node{
				path:    "*",
				handler: mockHandler,
			},
		},
		{
			name:   "*通配符中间",
			method: http.MethodGet,
			path:   "/host/create/home",
			found:  true,
			wantNode: &node{
				path:    "home",
				handler: mockHandler,
			},
		},
		{
			name:   "路径参数",
			method: http.MethodGet,
			path:   "/login/123",
			found:  true,
			wantNode: &node{
				path:    "123",
				handler: mockHandler,
			},
		},
		//{
		//	name:   "正则匹配",
		//	method: http.MethodPost,
		//	path:   "/user/346",
		//	found:  true,
		//	wantNode: &node{
		//		path:    "346",
		//		handler: mockHandler,
		//	},
		//},
	}

	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			n, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.found, found)
			if !found {
				return
			}
			wantVal := reflect.ValueOf(tc.wantNode.handler)
			nVal := reflect.ValueOf(n.handler)
			assert.Equal(t, wantVal, nVal)
		})
	}
}

func BenchmarkSprint(b *testing.B) {
	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/login/:id",
		},
		{
			method: http.MethodGet,
			path:   "/user/:id(^[0-9]+$)",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodGet,
			path:   "/host/*",
		},
		{
			method: http.MethodGet,
			path:   "/host/*/home",
		},
	}

	mockHandler := func(ctx *Context) {}

	testCases := []struct {
		name     string
		method   string
		path     string
		found    bool
		wantNode *node
	}{
		{
			name:   "method not found",
			method: http.MethodHead,
		},
		{
			name:   "path not found",
			method: http.MethodGet,
			path:   "/abc",
		},
		{
			name:   "root",
			method: http.MethodGet,
			path:   "/",
			found:  true,
			wantNode: &node{
				path:    "/",
				handler: mockHandler,
			},
		},
		{
			name:   "user",
			method: http.MethodGet,
			path:   "/user",
			found:  true,
			wantNode: &node{
				path:    "user",
				handler: mockHandler,
			},
		},
		{
			name:   "no handler",
			method: http.MethodPost,
			path:   "/order",
			found:  true,
			wantNode: &node{
				path: "order",
			},
		},
		{
			name:   "two layer",
			method: http.MethodPost,
			path:   "/order/create",
			found:  true,
			wantNode: &node{
				path:    "create",
				handler: mockHandler,
			},
		},
		{
			name:   "*末尾通配符",
			method: http.MethodPost,
			path:   "/host/create",
			found:  true,
			wantNode: &node{
				path:    "create",
				handler: mockHandler,
			},
		},
		{
			name:   "*通配符中间",
			method: http.MethodPost,
			path:   "/host/create/home",
			found:  true,
			wantNode: &node{
				path:    "home",
				handler: mockHandler,
			},
		},
		{
			name:   "路径参数",
			method: http.MethodPost,
			path:   "/login/123",
			found:  true,
			wantNode: &node{
				path:    "123",
				handler: mockHandler,
			},
		},
		{
			name:   "正则匹配",
			method: http.MethodPost,
			path:   "/user/346",
			found:  true,
			wantNode: &node{
				path:    "346",
				handler: mockHandler,
			},
		},
	}

	r := newRouter()
	for _, tr := range testRoutes {
		r.addRoute(tr.method, tr.path, mockHandler)
	}

	for _, tc := range testCases {
		n, found := r.findRoute(tc.method, tc.path)
		fmt.Println(n, found)
	}
}
