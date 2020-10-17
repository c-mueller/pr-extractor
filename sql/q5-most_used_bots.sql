SELECT a.pr_author_login AS bot_login, count(*) AS cnt
FROM (
         SELECT DISTINCT pr2.pull_request_id, pr2.pr_author_login
         FROM (
                  SELECT pull_request_id, max(pull_request_created_at) AS created_at
                  FROM pull_requests

                  GROUP BY pull_request_id
              ) AS prs
                  INNER JOIN pull_requests AS pr2 ON prs.pull_request_id = pr2.pull_request_id AND
                                                     prs.created_at = pr2.pull_request_created_at
         WHERE pr_author_login IN (SELECT username FROM bots)
     ) AS a
GROUP BY a.pr_author_login
ORDER BY cnt DESC;