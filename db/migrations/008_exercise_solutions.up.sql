BEGIN;

    -- solutions table
    create table if not exists exercise_solutions (
        exercise_id uuid references exercises (id) primary key,
        solution_text text not null default '',
        created_at timestamp default now(),
        updated_at timestamp
    );

    insert into exercise_solutions select id from exercises;
    
COMMIT;
