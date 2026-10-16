CREATE VIEW tag_usage AS
SELECT t.tgID, t.name, count(a.artID) AS used_in_amount
FROM tags t
LEFT JOIN article_tags at ON at.tgID = t.tgID
LEFT JOIN articles a ON a.artID = at.artID AND a.deleted_at IS NULL
GROUP BY t.tgID;
