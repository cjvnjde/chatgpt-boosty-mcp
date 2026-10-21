# API coverage

83 MCP tools: 54 reads and 29 operations with side effects. This is the feature scope of the supplied MCP reference, with login kept in a local CLI command. It is not a claim to cover every internal Boosty endpoint.

The TypeScript references were read for endpoint contracts only. No source from those libraries is compiled, installed, or executed by this project. Go uses the official MCP SDK and its dependencies plus Go system calls for terminal input; Boosty requests use `net/http` directly.

`Live public` means an anonymous HTTP 200 was observed by this project on 2026-10-21. `Contract only` means implemented from the inspected contract and tested locally, without authenticated live verification. `Experimental` means the reference marks the path/body experimental, or this project's write operation has not been verified on a live account. These labels do not promise that private operations currently succeed.

Tool inputs use the exact query/form field names below. A `!` marks a required field. All `{path_parameters}` are also required string inputs. Lists default to limit=20 when they have a limit parameter; maximum 100. Output is `{"response": <original upstream JSON>}` with no flattening or removal of fields. In particular, financial and account reads can contain personal data.

| Tool | HTTP contract | Access / effect | Status | Query fields | Body fields |
| --- | --- | --- | --- | --- | --- |
| `get_blog` | `GET /v1/blog/{blog}` | Public / session · read | Live public | — | — |
| `list_posts` | `GET /v1/blog/{blog}/post/` | Public / session · read | Live public | limit offset comments_limit reply_limit from_ts to_ts level_id tags_ids | — |
| `get_post` | `GET /v1/blog/{blog}/post/{post_id}` | Public / session · read | Contract only | comments_limit reply_limit | — |
| `create_post` | `POST /v1/blog/{blog}/post/` | Login · WRITE | Experimental | — | Form: title data price teaser_data tags deny_comments wait_video publish_time advertiser_info has_chat |
| `update_post` | `PUT /v1/blog/{blog}/post/{post_id}` | Login · WRITE | Experimental | — | Form: title data price teaser_data tags deny_comments wait_video publish_time advertiser_info |
| `delete_post` | `DELETE /v1/blog/{blog}/post/{post_id}` | Login · WRITE | Experimental | — | — |
| `get_post_deferred_access` | `GET /v1/blog/{blog}/post/{post_id}/deferred_access` | Login · read | Contract only | — | — |
| `update_post_deferred_access` | `PUT /v1/blog/{blog}/post/{post_id}/deferred_access` | Login · WRITE | Experimental | — | JSON: isCommentsDenied! price! applyTime! |
| `list_subscription_levels` | `GET /v1/blog/{blog}/subscription_level/` | Public / session · read | Live public | show_free_level | — |
| `get_subscription_level` | `GET /v1/blog/{blog}/subscription/level/{level}` | Public / session · read | Contract only | with_content_counters | — |
| `list_subscribers` | `GET /v1/blog/{blog}/subscribers` | Login · read | Contract only | limit offset sort_by order with_follow | — |
| `list_blacklist` | `GET /v1/blacklist/` | Login · read | Contract only | blog_url! | — |
| `list_unsubscribe_reasons` | `GET /v1/blog/{blog}/unsubscribe_reasons/` | Login · read | Contract only | limit offset sort_by order level_ids answers from to | — |
| `list_media` | `GET /v1/blog/{blog}/media_album/` | Public / session · read | Live public | limit offset type | — |
| `get_feed` | `GET /v1/feed/post/` | Login · read | Contract only | limit offset only_allowed | — |
| `search_blogs` | `GET /v1/search/blog/` | Login · read | Contract only | search_query! limit offset  | — |
| `suggest_blogs` | `GET /v1/search/blog/suggest/` | Login · read | Contract only | search_query! limit offset  | — |
| `search_feed_tags` | `GET /v1/search/feed/tag/` | Login · read | Contract only | search_query! limit offset  | — |
| `search_feed_posts` | `GET /v1/search/feed/post/` | Login · read | Contract only | search_query! limit offset only_allowed only_bought from_ts to_ts tags_ids | — |
| `search_blog_posts` | `GET /v1/search/blog/post/` | Login · read | Contract only | search_query! limit offset blog_url! only_allowed subscription_level has_image has_audio_file has_audio has_ok_video has_video has_file from_ts to_ts | — |
| `list_comments` | `GET /v1/blog/{blog}/post/{post_id}/comment/` | Public / session · read | Contract only | limit offset reply_limit order | — |
| `list_comment_replies` | `GET /v1/blog/{blog}/post/{post_id}/comment/` | Public / session · read | Contract only | parent_id! limit offset order | — |
| `create_comment` | `POST /v1/blog/{blog}/post/{post_id}/comment/` | Login · WRITE | Experimental | — | Form: data! reply_to_id reply_id |
| `react_to_post` | `POST /v1/blog/{blog}/post/{post_id}/reaction` | Login · WRITE | Experimental | from_page | Form: reaction! |
| `remove_post_reaction` | `DELETE /v1/blog/{blog}/post/{post_id}/reaction` | Login · WRITE | Experimental | from_page | — |
| `like_post` | `POST /v1/blog/{blog}/post/{post_id}/like` | Login · WRITE | Experimental | from_page | — |
| `unlike_post` | `DELETE /v1/blog/{blog}/post/{post_id}/like` | Login · WRITE | Experimental | from_page | — |
| `react_to_comment` | `POST /v1/blog/{blog}/post/{post_id}/comment/{comment_id}/reaction` | Login · WRITE | Experimental | from_page | Form: reaction! |
| `remove_comment_reaction` | `DELETE /v1/blog/{blog}/post/{post_id}/comment/{comment_id}/reaction` | Login · WRITE | Experimental | from_page | — |
| `like_comment` | `POST /v1/blog/{blog}/post/{post_id}/comment/{comment_id}/reaction` | Login · WRITE | Experimental | from_page | — |
| `unlike_comment` | `DELETE /v1/blog/{blog}/post/{post_id}/comment/{comment_id}/reaction` | Login · WRITE | Experimental | from_page | — |
| `get_poll` | `GET /v1/blog/{blog}/poll/{poll_id}` | Public / session · read | Experimental | voters_limit | — |
| `list_poll_voters` | `GET /v1/blog/{blog}/poll/{poll_id}/vote/` | Login · read | Experimental | limit offset option | — |
| `vote_poll` | `POST /v1/poll/{poll_id}/vote` | Login · WRITE | Experimental | — | Form: answer! other is_anonymous |
| `unvote_poll` | `DELETE /v1/poll/{poll_id}/vote` | Login · WRITE | Experimental | — | — |
| `get_current_user` | `GET /v1/user/current` | Login · read | Contract only | — | — |
| `list_subscriptions` | `GET /v1/user/subscriptions` | Login · read | Contract only | limit offset with_follow | — |
| `list_sessions` | `GET /v1/user/session/` | Login · read | Contract only | — | — |
| `get_notification_settings` | `GET /v1/notification/settings/` | Login · read | Contract only | transports | — |
| `list_payment_cards` | `GET /v1/payment/bind/list` | Login · read | Contract only | currency deleted | — |
| `list_blog_currencies` | `GET /v1/user/current/blog_currency/available/` | Login · read | Contract only | — | — |
| `update_profile` | `PUT /v1/user/current` | Login · WRITE | Experimental | — | Form: name email can_view_adult_content send_msg_rule send_msg_level_id |
| `update_notification_setting` | `PUT /v1/notification/settings/{transport}/{switch}` | Login · WRITE | Experimental | — | Form: value! |
| `update_dialog_settings` | `POST /v1/user/dialog_settings` | Login · WRITE | Experimental | — | Form: can_send_all! can_send_donation! can_send_my_authors! can_send_paid_subscribers! can_send_subscribers! can_send_payers! send_msg_level_id |
| `set_locale` | `POST /v1/user/locale` | Login · WRITE | Experimental | — | Form: lang! |
| `end_sessions` | `DELETE /v1/user/session/` | Login · WRITE | Experimental | — | Form: including_current! |
| `list_dialogs` | `GET /v1/dialog/` | Login · read | Contract only | limit offset | — |
| `get_dialog_with_user` | `GET /v1/dialog` | Login · read | Contract only | user_id! | — |
| `create_dialog` | `POST /v1/dialog/` | Login · WRITE | Contract only | — | Form: user_id! |
| `list_messages` | `GET /v1/dialog/{dialog_id}/message/` | Login · read | Contract only | limit offset | — |
| `send_message` | `POST /v1/dialog/{dialog_id}/message` | Login · WRITE | Contract only | — | Form: data! |
| `list_notifications` | `GET /v1/notification/standalone/event/` | Login · read | Contract only | — | — |
| `mark_notifications_read` | `PUT /v1/notification/standalone/read/` | Login · WRITE | Contract only | — | Form: event_id! |
| `delete_notification` | `DELETE /v1/notification/standalone/event/{event}` | Login · WRITE | Experimental | — | — |
| `delete_notifications` | `DELETE /v1/notification/standalone/event/` | Login · WRITE | Experimental | — | Form: event_id! |
| `list_targets` | `GET /v1/target/{blog}/` | Public / session · read | Contract only | type show_deleted | — |
| `get_target` | `GET /v1/target/{target_id}` | Public / session · read | Contract only | currency | — |
| `create_target` | `POST /v1/target/{target_type}` | Login · WRITE | Experimental | — | Form: blog_url! description! target_sum! |
| `update_target` | `PUT /v1/target/{target_id}` | Login · WRITE | Experimental | — | Form: description! target_sum! |
| `delete_target` | `DELETE /v1/target/{target_id}` | Login · WRITE | Experimental | — | — |
| `list_post_sales` | `GET /v1/blog/{blog}/sales/post/` | Login · read | Contract only | limit offset sort_by order post_ids | — |
| `list_donations` | `GET /v1/blog/{blog}/sales/donation/` | Login · read | Contract only | limit offset sort_by order type target_id | — |
| `list_bundle_sales` | `GET /v1/blog/{blog}/sales/bundle/` | Login · read | Contract only | limit offset sort_by order bundle_ids | — |
| `list_holds` | `GET /v1/blog/{blog}/sales/hold/` | Login · read | Contract only | limit offset | — |
| `list_broadcast_sales` | `GET /v1/blog/{blog}/sales/dialog_broadcast_message/` | Login · read | Contract only | limit offset | — |
| `get_paid_sales_by_date` | `GET /v1/blog/{blog}/{sale_kind}/paid/by_date` | Login · read | Contract only | from! to! | — |
| `get_paid_sales_days` | `GET /v1/blog/{blog}/{sale_kind}/paid/days` | Login · read | Contract only | date! | — |
| `list_payout_methods` | `GET /v1/payout/methods` | Login · read | Contract only | — | — |
| `list_payouts` | `GET /v1/blog/{blog}/payout/` | Login · read | Contract only | — | — |
| `list_payout_history` | `GET /v1/blog/{blog}/payout/history/` | Login · read | Contract only | limit offset | — |
| `get_default_currency` | `GET /v1/payment/default_currency` | Anonymous only · read | Live public | — | — |
| `get_stats_summary` | `GET /v1/blog/stat/{blog}/current` | Login · read | Contract only | — | — |
| `get_stats_metrics` | `GET /v1/blog/stat/{blog}/metrics` | Login · read | Contract only | from! to! | — |
| `get_stats_charts` | `GET /v1/blog/{blog}/stat/data` | Login · read | Contract only | last_time limit | — |
| `list_stats_events` | `GET /v1/blog/stat/{blog}/events` | Login · read | Contract only | limit offset from to event_types user_ids | — |
| `list_visit_sources` | `GET /v1/blog/stat/{blog}/visits` | Login · read | Contract only | type! period! limit offset sort_by order timestamp | — |
| `list_payment_sources` | `GET /v1/blog/stat/{blog}/payments/analytics` | Login · read | Contract only | type! period! limit offset sort_by order timestamp | — |
| `get_referral_stats` | `GET /v1/blog/referral/{blog}/stat/` | Login · read | Contract only | — | — |
| `list_referral_users` | `GET /v1/blog/referral/{blog}/user/` | Login · read | Contract only | limit offset only_bloggers | — |
| `search_subscribers` | `GET /v1/blog/stat/{blog}/search` | Login · read | Contract only | chunk! limit | — |
| `get_post_stats` | `GET /v1/blog/stat/{blog}/post/{post_id}` | Login · read | Experimental | — | — |
| `get_report_info` | `GET /v1/blog/stat/{blog}/csv/{report_type}/info` | Login · read | Contract only | — | — |
| `build_report` | `GET /v1/blog/stat/{blog}/csv/{report_type}` | Login · WRITE | Contract only | separator from to period | — |

## Contract details

- `list_media` always sends `limit_by=media`, with default `type=all`.
- `list_payment_cards` adds `payment_partner=advpay`.
- `get_notification_settings` defaults to all four known transports.
- `list_visit_sources` and `list_payment_sources` default to `order=gt`; `type` is `referer` or `utm`, `period` is `day`, `week`, or `month`.
- `get_default_currency` never sends a Boosty bearer token, even when logged in.
- `send_message` uses a path without a trailing slash, while `list_messages` uses a trailing slash.
- `list_comment_replies` uses `parent_id` (numeric comment `intId`) on the ordinary comment endpoint.
- `data` and `teaser_data` arrays are JSON-encoded inside form fields. Omitted update fields are preserved. Only `create_post` defaults these arrays to empty arrays.
- `like_comment` sends `reaction=like`. Post likes use the separate `/like` route.
- `update_target` repeats its path ID as `target_id` in the form, matching the observed contract.
- `including_current=true` is a local validation requirement, not sent to Boosty. Session termination can require a new cookie import.
- `build_report` is a write despite using GET. `get_report_info` is read-only. Returned report links are not fetched by this server.
- Paid post and bundle statistics share tools through `sale_kind=post|bundle`; reports use `report_type=donations|subscriptions|holds|payouts`.
- No bulk pagination, media upload, payment initiation, scraping of inaccessible content, or arbitrary HTTP proxy tool is exposed.

See the README for authentication, deployment, reference commits and test limitations.
