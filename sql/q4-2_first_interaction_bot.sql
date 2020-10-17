SELECT r.pull_request_id,
       EXTRACT(epoch FROM min(r.comment_created_at - prs.pull_request_created_at)) AS min_comment_response_time
FROM (
         SELECT *
         FROM pull_requests
         WHERE pr_author_login IN (SELECT username FROM bots)
     ) AS prs,
     (
         SELECT pull_request_id, event_db_id, comment_author_name, comment_author_type, comment_created_at
         FROM pull_request_review_comments
         WHERE comment_created_at = comment_updated_at
           AND comment_author_name NOT IN (SELECT username FROM bots)
         UNION
         SELECT pull_request_id, event_db_id, comment_author_name, comment_author_type, comment_created_at
         FROM pull_request_comments
         WHERE comment_created_at = comment_updated_at
           AND comment_author_name NOT IN (SELECT username FROM bots)
         UNION
         SELECT pull_request_id, event_db_id, event_initiator_login, 'User', pull_request_closed_at
         FROM pull_requests
         WHERE pull_request_closed_at IS NOT NULL
           AND event_initiator_login NOT IN (SELECT username FROM bots)
     ) AS r
WHERE r.pull_request_id = prs.pull_request_id
GROUP BY r.pull_request_id
HAVING min(r.comment_created_at - prs.pull_request_created_at) > INTERVAL '0 secs'