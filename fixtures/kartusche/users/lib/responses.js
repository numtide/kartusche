function jsonResponse(statusCode, data) {
    w.header().set("content-type","application/json")
    w.writeHeader(statusCode)
    w.write(JSON.stringify(data))
}

exports.inputError = (reason) => {
    jsonResponse(400, {reason: reason})
}
