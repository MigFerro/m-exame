BEGIN;

    -- changes to exercise_users table
    alter table exercise_categories
    add column if not exists year varchar(10);
    
COMMIT;
