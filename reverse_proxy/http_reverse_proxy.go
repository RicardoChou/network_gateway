package reverse_proxy

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zhj/go_gateway/middleware"
	"github.com/zhj/go_gateway/reverse_proxy/load_balance"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func NewLoadBalanceReverseProxy(c *gin.Context, lb load_balance.LoadBalance, trans *http.Transport) *httputil.ReverseProxy {
	//请求协调者
	director := func(req *http.Request) {
		nextAddr, err := lb.Get(req.URL.String())
		fmt.Println("nextAddr:", nextAddr)
		//todo 优化点3
		if err != nil || nextAddr == "" {
			panic("get next addr fail")
		}
		target, err := url.Parse(nextAddr)
		if err != nil {
			panic(err)
		}
		targetQuery := target.RawQuery
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host

		//target.Path : /base
		//req.URL.Path : /dir
		req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)
		//todo 当对域名(非内网)反向代理时需要设置此项, 当作后端反向代理时不需要
		req.Host = target.Host
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "user-agent")
		}
	}

	//更改内容
	modifyFunc := func(resp *http.Response) error {
		//todo 部分章节功能补充2
		//todo 兼容websocket
		if strings.Contains(resp.Header.Get("Connection"), "Upgrade") {
			return nil
		}
		//var payload []byte
		//var readErr error
		//
		////todo 部分章节功能补充3
		////todo 兼容gzip压缩
		//if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		//	gr, err := gzip.NewReader(resp.Body)
		//	if err != nil {
		//		return err
		//	}
		//	payload, readErr = ioutil.ReadAll(gr)
		//	resp.Header.Del("Content-Encoding")
		//} else {
		//	payload, readErr = ioutil.ReadAll(resp.Body)
		//}
		//if readErr != nil {
		//	return readErr
		//}
		//
		////异常请求时设置StatusCode
		//if resp.StatusCode != 200 {
		//	payload = []byte("StatusCode error:" + string(payload))
		//}
		//
		////todo 部分章节功能补充4
		////todo 因为预读了数据所以内容重新回写
		//c.Set("status_code", resp.StatusCode)
		//c.Set("payload", payload)
		//resp.Body = ioutil.NopCloser(bytes.NewBuffer(payload))
		//resp.ContentLength = int64(len(payload))
		//resp.Header.Set("Content-Length", strconv.FormatInt(int64(len(payload)), 10))
		return nil
	}

	//错误回调 ：关闭real_server时测试，错误回调
	//范围：transport.RoundTrip发生的错误、以及ModifyResponse发生的错误
	errFunc := func(w http.ResponseWriter, r *http.Request, err error) {
		middleware.ResponseError(c, 999, err)
	}

	return &httputil.ReverseProxy{Director: director, Transport: trans, ModifyResponse: modifyFunc, ErrorHandler: errFunc}
}
