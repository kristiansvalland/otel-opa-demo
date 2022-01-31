package main

blocked_users = {"mallory"}

default allow = false

# Allow all request not on user resources
allow {
	not blocked
	is_user_path == false
}

# Allow GET requests on self
allow {
	not blocked
	is_user_path
	is_self_path
	is_get
}

# Admin is allowed on all resources, provided not blocked
allow {
	not blocked
	is_admin
}

reason[r] {
	not is_get
	is_user_path
	is_self_path
	r := sprintf("Method %s not allowed on resource %s", [input.method, input.path])
}

reason["user is blocked"] {
	blocked
}

reason[r] {
	not is_self_path
	r := sprintf("Not self: requested %v, but user is %v", [path[1], payload.userid])
}

is_self_path {
	is_user_path
	to_number(path[1]) == payload.userid
}

is_get {
	input.method == "GET"
}

default blocked = true

blocked = false {
	not blocked_users[payload.username]
}

is_admin {
	payload.roles[_] == "admin"
}

is_user_path {
	"users" == path[0]
}

path := split(trim_prefix(input.path, "/"), "/")

header := decoded_token[0]

payload := decoded_token[1]

signature := decoded_token[2]

# TODO: verify as well
decoded_token = io.jwt.decode(input.jwt)
