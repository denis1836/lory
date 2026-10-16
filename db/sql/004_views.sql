CREATE VIEW tag_usage AS
SELECT t.id, t.name, count(a.id) AS used_in_amount
FROM tags t
LEFT JOIN article_tags at ON at.tag_id = t.id
LEFT JOIN articles a ON a.id = at.article_id AND a.deleted_at IS NULL
GROUP BY t.id;
