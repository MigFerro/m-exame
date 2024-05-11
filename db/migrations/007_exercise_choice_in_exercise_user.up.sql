BEGIN;

    -- changes to exercise_users table
    alter table exercise_users
    add column if not exists choice_selected uuid references exercise_choices(id),
    add column if not exists points_gained int;
    
COMMIT;
