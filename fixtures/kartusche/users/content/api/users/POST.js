const {email, username, password} = JSON.parse(requestBody())
const handler = () => {
    if (!email || !email.match(/^[^@]+@[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)+$/) ) {
        w.header().set("content-type","application/json")
        w.writeHeader(400)
        w.write(JSON.stringify({error: "invalid_email"}))
        return
    }  
    
    if (!username || !username.match(/^[a-zA-Z0-9_-]+$/)) {
        w.header().set("content-type","application/json")
        w.writeHeader(400)
        w.write(JSON.stringify({error: "invalid_username"}))
        return
    }

    if (!password || !password.toLowerCase().match(/^.{3,}$/)) {
        w.header().set("content-type","application/json")
        w.writeHeader(400)
        w.write(JSON.stringify({error: "invalid_password"}))
        return
    }

    const userId = uuidv4()

    write(tx => {
        tx.put(["users",userId], JSON.stringify({email, username, password}))
    })

    w.write(JSON.stringify({user_id: userId}))

}

handler()
