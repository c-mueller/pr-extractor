SELECT 'bots' AS type, count(a.pull_request_id) AS total_cnt
FROM (
         SELECT DISTINCT pr2.pull_request_id
         FROM (
                  SELECT pull_request_id, max(pull_request_created_at) AS created_at
                  FROM pull_requests
                  WHERE pull_request_merged_at IS NOT NULL
                  GROUP BY pull_request_id
              ) AS prs
                  INNER JOIN pull_requests AS pr2 ON prs.pull_request_id = pr2.pull_request_id AND
                                                     prs.created_at = pr2.pull_request_created_at
         WHERE pr_author_login IN (SELECT username FROM bots)
     ) AS a
UNION
SELECT 'humans', count(a.pull_request_id) AS total_cnt
FROM (
         SELECT DISTINCT pr2.pull_request_id
         FROM (
                  SELECT pull_request_id, max(pull_request_created_at) AS created_at
                  FROM pull_requests
                  WHERE pull_request_merged_at IS NOT NULL
                  GROUP BY pull_request_id
              ) AS prs
                  INNER JOIN pull_requests AS pr2 ON prs.pull_request_id = pr2.pull_request_id AND
                                                     prs.created_at = pr2.pull_request_created_at
         WHERE pr_author_login NOT IN (SELECT username FROM bots)
     ) AS a
UNION
SELECT 'total', count(a.pull_request_id) AS total_cnt
FROM (
         SELECT DISTINCT pr2.pull_request_id
         FROM (
                  SELECT pull_request_id, max(pull_request_created_at) AS created_at
                  FROM pull_requests
                  WHERE pull_request_merged_at IS NOT NULL
                  GROUP BY pull_request_id
              ) AS prs
                  INNER JOIN pull_requests AS pr2 ON prs.pull_request_id = pr2.pull_request_id AND
                                                     prs.created_at = pr2.pull_request_created_at
     ) AS a;