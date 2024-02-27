BEGIN;

    -- changes to exercise_users table
    alter table exercise_users
    add column if not exists first_attempted_at timestamp default now(),
    add column if not exists last_attempted_at timestamp,
    add column if not exists first_solved_at timestamp,
    add column if not exists last_solved_at timestamp,
    drop column if exists created_at,
    drop column if exists updated_at,
    drop column if exists times_attempted,
    drop column if exists solved;
    
COMMIT;
