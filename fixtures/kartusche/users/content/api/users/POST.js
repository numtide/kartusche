const {email, username, password} = JSON.parse(requestBody())
const handler = () => {
    if (!email || !email.match(/^[^@]+@[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)+$/) ) {
        jsonResponse(400, {error: "invalid_email"})
        return
    }  
    
    if (!username || !username.match(/^[a-zA-Z0-9_-]+$/)) {
        jsonResponse(400, {error: "invalid_username"})
        return
    }

    if (!password || !password.toLowerCase().match(/^.{3,}$/)) {
        jsonResponse(400, {error: "invalid_password"})
        return
    }

    const userId = uuidv4()

    write(tx => {
        tx.put(["users",userId], JSON.stringify({email, username, password}))
    })

    jsonResponse(200, {user_id: userId})
}

handler()
