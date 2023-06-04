package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

// Context 上下文，web 信息的封装
type Context struct {
	// origin objects
	Writer http.ResponseWriter // 返回
	Req    *http.Request       // 请求
	//Request info
	Path   string            // 请求路径
	Method string            // 请求方法
	Params map[string]string // 路由参数
	// response info
	StatusCode int // 状态码

	// middleware
	handlers []HandlerFunc
	index    int

	// Engine pointer
	engine *Engine
}

// Param 匹配状态码
func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    r,
		Path:   r.URL.Path,
		Method: r.Method,
		index:  -1,
	}
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Context-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		panic(err)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Context-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(http.StatusInternalServerError, err.Error())
	}
}

func (c *Context) Next() {
	c.index++
	//fmt.Println("Next: ", &c.StatusCode)
	s := len(c.handlers)
	//fmt.Println("Next: ", s)
	for ; c.index < s; c.index++ {
		fmt.Printf("Next: index: %d, statusCode: %d\n", c.index, c.StatusCode)
		c.handlers[c.index](c)
	}
}

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{"message": err})
}
