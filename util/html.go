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
