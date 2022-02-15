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

BLOCKED_API := "http://blocked-user-rs:8088/blocked"

request := {
	"url": BLOCKED_API,
	"method": "POST",
	"body": {"username": payload.username},
	"headers": {"Content-Type": "application/json"},
}

blocked_response = response {
	response = http.send(request)
}

default blocked = true

blocked = false {
	blocked_response.status_code == 200
	body = blocked_response.body

	body.result == false
}

reason[msg] {
	blocked_response.status_code != 200
	msg := sprintf("Failed to query for blocked user: POST %s failed with status_code=%v", [BLOCKED_API, blocked_response.status_code])
}

reason["Failed to query for blocked user: got no body"] {
	blocked_response.status_code == 200
	not blocked_response.body
}

reason["unable to query blocked api"] {
	not blocked_response
}

reason["user is blocked"] {
	blocked_response.body.result
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
