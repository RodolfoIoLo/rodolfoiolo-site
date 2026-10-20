UPDATE site_settings
SET value = '{"site_title":"Rodolfo Iolo","site_description":"Articles, projects, and engineering notes.","site_keywords":"computer science,go,web development","about_content":"# About\n\nComputer science student learning in public.","footer_text":"Built with care and documented end to end.","icp_number":null,"social_links":[]}'::jsonb,
    updated_at = now()
WHERE key = 'site';
