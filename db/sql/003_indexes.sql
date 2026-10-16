-- tags
CREATE UNIQUE INDEX tags_name_lower_uniq ON Tags (lower(name));

-- articles 
CREATE UNIQUE INDEX articles_project_slug_uniq  ON articles (project_id, slug) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX articles_project_title_uniq ON articles (project_id, lower(title)) WHERE deleted_at IS NULL;

-- articles' tags
CREATE INDEX article_tags_tag_idx ON article_tags (tag_id);
