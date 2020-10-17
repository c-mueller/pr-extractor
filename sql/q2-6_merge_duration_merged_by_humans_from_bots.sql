SELECT times.pull_request_id,
       merge_time_duration,
       prs.event_initiator_login AS merged_by,
       prs.pr_author_login       AS opened_by
FROM (
         SELECT pull_request_id                                                                  AS pull_request_id,
                EXTRACT(epoch FROM (max(pull_request_merged_at) - max(pull_request_created_at))) AS merge_time_duration
         FROM pull_requests AS prs
         WHERE pull_request_merged_at IS NOT NULL
           AND pr_author_login IN (SELECT username FROM bots)
           AND event_initiator_login NOT IN (SELECT username FROM bots)
         GROUP BY pull_request_id
     ) AS times,
     pull_requests AS prs
WHERE prs.pull_request_id = times.pull_request_id
  AND EXTRACT(epoch FROM (prs.pull_request_merged_at - prs.pull_request_created_at)) = times.merge_time_duration
ORDER BY merge_time_duration ASC;