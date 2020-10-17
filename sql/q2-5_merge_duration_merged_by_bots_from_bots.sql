SELECT pull_request_id,
       EXTRACT(epoch FROM (max(pull_request_merged_at) - max(pull_request_created_at))) AS merge_time_duration
FROM pull_requests AS prs
WHERE pull_request_merged_at IS NOT NULL
  AND pr_author_login IN (SELECT username FROM bots)
  AND event_initiator_login IN (SELECT username FROM bots)
GROUP BY pull_request_id;