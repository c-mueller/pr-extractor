SELECT 'bot'                                 AS type,
       avg(a.additions)                      AS additions,
       avg(a.deletions)                      AS deletions,
       (avg(a.additions) + avg(a.deletions)) AS line_changes
FROM (
         SELECT DISTINCT pr2.pull_request_id, pr2.pr_author_login, pr2.additions, pr2.deletions
         FROM (
                  SELECT pull_request_id, max(pull_request_created_at) AS created_at
                  FROM pull_requests
                  GROUP BY pull_request_id
              ) AS prs
                  INNER JOIN pull_requests AS pr2 ON prs.pull_request_id = pr2.pull_request_id AND
                                                     prs.created_at = pr2.pull_request_created_at
         WHERE pr_author_login IN (SELECT username FROM bots)
     ) AS a
UNION
SELECT 'human'                               AS type,
       avg(a.additions)                      AS additions,
       avg(a.deletions)                      AS deletions,
       (avg(a.additions) + avg(a.deletions)) AS line_changes
FROM (
         SELECT DISTINCT pr2.pull_request_id, pr2.pr_author_login, pr2.additions, pr2.deletions
         FROM (
                  SELECT pull_request_id, max(pull_request_created_at) AS created_at
                  FROM pull_requests
                  GROUP BY pull_request_id
              ) AS prs
                  INNER JOIN pull_requests AS pr2 ON prs.pull_request_id = pr2.pull_request_id AND
                                                     prs.created_at = pr2.pull_request_created_at
         WHERE pr_author_login NOT IN (SELECT username FROM bots)
     ) AS a
UNION
SELECT 'total'                               AS type,
       avg(a.additions)                      AS additions,
       avg(a.deletions)                      AS deletions,
       (avg(a.additions) + avg(a.deletions)) AS line_changes
FROM (
         SELECT DISTINCT pr2.pull_request_id, pr2.pr_author_login, pr2.additions, pr2.deletions
         FROM (
                  SELECT pull_request_id, max(pull_request_created_at) AS created_at
                  FROM pull_requests
                  GROUP BY pull_request_id
              ) AS prs
                  INNER JOIN pull_requests AS pr2 ON prs.pull_request_id = pr2.pull_request_id AND
                                                     prs.created_at = pr2.pull_request_created_at
     ) AS a;