const { inputError } = require("lib/responses")

const { email, username, password } = JSON.parse(requestBody())
const handler = () => {
    if (!email || !email.match(/^[^@]+@[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)+$/)) {
        return inputError("invalid_email")
    }

    if (!username || !username.match(/^[a-zA-Z0-9_-]+$/)) {
        return inputError("invalid_username")
    }

    if (!password || !password.toLowerCase().match(/^.{3,}$/)) {
        return inputError("invalid_password")
    }

    const userId = uuidv4()

    write(tx => {
        tx.put(["users", userId], JSON.stringify({ email, username, password }))
    })

    jsonResponse(200, { user_id: userId })
}

handler()
