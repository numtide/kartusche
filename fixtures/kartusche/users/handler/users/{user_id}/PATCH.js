const { inputError } = require("lib/responses")

const { email } = JSON.parse(requestBody())
if (email === "") {
    inputError("invalid_email")
} else {
    write(tx => {
        const us = JSON.parse(tx.get(["users", vars.user_id]))
        us.email = email
        tx.put(["users", vars.user_id], JSON.stringify(us))
    })
    jsonResponse(204)
}


