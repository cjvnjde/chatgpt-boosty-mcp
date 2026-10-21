package mcpserver

// Endpoint contracts are implemented here, without a third-party Boosty client.
func Endpoints() []Endpoint {
	return []Endpoint{
		{Name: "get_blog", Description: "Read a blog profile by its URL slug.", Method: "GET", Path: "/v1/blog/{blog}"},
		{Name: "list_posts", Description: "Read one page of blog posts, preserving access flags and pagination.", Method: "GET", Path: "/v1/blog/{blog}/post/", Query: "limit offset comments_limit reply_limit from_ts to_ts level_id tags_ids", Defaults: map[string]any{"limit": 20}},
		{Name: "get_post", Description: "Read one post and its content blocks.", Method: "GET", Path: "/v1/blog/{blog}/post/{post_id}", Query: "comments_limit reply_limit"},
		{Name: "get_current_user", Description: "Read the connected Boosty account.", Method: "GET", Path: "/v1/user/current", Auth: true},
	}
}
