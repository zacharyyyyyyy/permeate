package util

import (
	"fmt"
	"net/http"
)

func ShowPasswordFormHtml(w http.ResponseWriter) {
	html := `  
 <!DOCTYPE html>  
 <html>  
 <head>  
 <title>token Form</title>  
 </head>  
 <body>  
 <h1>请输入token</h1>  
 <form method="post" action="/">  
 <label for="token">Token:</label>  
 <input type="token" id="token" name="token">  
 <button type="submit">Submit</button>  
 </form>  
 </body>  
 </html>  
 `
	fmt.Fprintf(w, html)
}

func GetHtmlDataPackage(r *http.Request) string {
	var scheme string
	if r.URL.Scheme == "" {
		scheme = "HTTP"
	} else {
		scheme = r.URL.Scheme
	}
	var str string
	str += r.Method + " " + r.URL.String() + " " + scheme + "/1.1\r\n"
	str += "Host: " + r.Host + "\r\n"
	for key, val := range r.Header {
		str += key + ": " + val[0] + "\r\n"
	}
	str += "\r\n"
	len := r.ContentLength // 获取请求实体长度
	if len > 0 {
		body := make([]byte, len) // 创建存放请求实体的字节切片
		r.Body.Read(body)
		str += string(body)
	}
	return str
}
