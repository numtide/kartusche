const { email, username, password } = JSON.parse(requestBody())

const userId = uuidv4()

const response = write(tx => {
    if (tx.exists(["userByEmail", email])) {
        return `user with email ${email} already exist`
    }

    tx.put(["users", userId], JSON.stringify({ email, username, password }))
    tx.put(["userByEmail", email], userId)
    return `new user id: ${userId}`
})

render_template("create-user-result.mustache", { message: response })