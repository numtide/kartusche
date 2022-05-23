if (!tx.exists(["chats"])) {
    tx.createMap(["chats"])
    tx.createMap(["chats","support"])
}
