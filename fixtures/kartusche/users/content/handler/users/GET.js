w.header().set("content-type", "application/json")
const userList = read((tx) => {
    const users =[]
    for (const it = tx.iterator(["users"]); !it.isDone(); it.next()) {
        users.push(it.getKey())
    }
    return users
})
w.write(JSON.stringify(userList))