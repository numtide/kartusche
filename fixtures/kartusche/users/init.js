if (!tx.exists(["users"])) {
    tx.createMap(["users"])
}

if (!tx.exists(["userByEmail"])) {
    tx.createMap(["userByEmail"])
}
