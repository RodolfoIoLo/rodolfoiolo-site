UPDATE site_settings
SET value = '{"title":"Rodolfo Iolo","description":"Articles, projects, and engineering notes."}'::jsonb,
    updated_at = now()
WHERE key = 'site';
