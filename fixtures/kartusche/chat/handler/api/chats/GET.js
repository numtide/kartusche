const chats = read(tx => {
    const chats = []
    for (const it=tx.iterator(["chats"]); !it.isDone(); it.next()) {
        chats.push(it.getKey())
    }
    return chats
})
w.header().set("content-type","application/json")
w.write(JSON.stringify(chats))