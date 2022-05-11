const userJson = read(tx => tx.get(["users",vars.user_id]))
w.write(userJson)
