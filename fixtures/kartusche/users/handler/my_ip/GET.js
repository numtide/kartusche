const res = http_do("GET","https://api.ipify.org?format=json")
w.header().set("content-type","application/json")
w.write(JSON.stringify(res.body))
