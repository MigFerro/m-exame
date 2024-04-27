BEGIN;

    -- tests table
    create table if not exists tests (
        id uuid primary key not null,
        user_id uuid references users(id) not null,
        test_type varchar(50) not null,
        category_iid int references exercise_categories(iid),
        points_gained int,
        created_at timestamp default now(),
        finished_at timestamp,
        abandoned_at timestamp
    );

    -- test_exercises table
    create table if not exists test_exercises (
        test_id uuid references tests(id) not null,
        exercise_id uuid references exercises(id) not null,
        selected_choice_id uuid references exercise_choices(id),
        correct boolean,
        created_at timestamp default now(),
        finished_at timestamp,
        primary key(test_id, exercise_id)
    );

COMMIT;
