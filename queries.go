package main

func buildGetAllClipsQuery() string {
	return `SELECT
        c.*,
        string_agg(DISTINCT t.tag_name, ', ') AS returned_clip_tags,
        string_agg(DISTINCT u.username, ', ') AS returned_featured_users,
        g.name AS game_name,
        u.username AS user_name
    FROM
        clips AS c
    LEFT JOIN
        clips_tags AS ct ON c.id = ct.clip_id
    LEFT JOIN
        tags AS t ON ct.tag_id = t.id
    LEFT JOIN
        clips_users AS cu ON c.id = cu.clip_id
    LEFT JOIN
        users AS u ON cu.user_id = u.id
    LEFT JOIN
        games AS g ON c.game_id = g.id
    GROUP BY
        c.id, c.description, g.name, u.username;`
}

func buildGetClipQuery() string {
	return `SELECT
	c.*,
	string_agg(DISTINCT t.tag_name, ', ') AS returned_clip_tags,
	string_agg(DISTINCT u.username, ', ') AS returned_featured_users,
	g.name AS game_name,
	u.username AS user_name
FROM
	clips AS c
LEFT JOIN
	clips_tags AS ct ON c.id = ct.clip_id
LEFT JOIN
	tags AS t ON ct.tag_id = t.id
LEFT JOIN
	clips_users AS cu ON c.id = cu.clip_id
LEFT JOIN
	users AS u ON cu.user_id = u.id
LEFT JOIN
	games AS g ON c.game_id = g.id
WHERE
	c.id = $1
GROUP BY
	c.id, c.description, g.name, u.username;`
}
