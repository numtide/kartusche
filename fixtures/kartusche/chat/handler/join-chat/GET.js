const username = r.uRL.query().get("username")
render_template('chat.mustache',{username})