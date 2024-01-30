BEGIN;
    -- user role
    create type user_role as enum ('admin', 'prof', 'student');

    -- users table 
    create table if not exists users (
        id uuid primary key default gen_random_uuid(),
        auth_id varchar(50) not null,
        name varchar(30) not null,
        email varchar(30) not null,
        role user_role not null default 'student',
        created_at timestamp default now(),
        updated_at timestamp
    );

    -- exercises table
    create table if not exists exercises (
        id uuid primary key default gen_random_uuid(),
        problem_text text not null,
        created_at timestamp default now(),
        updated_at timestamp,
        created_by uuid not null,
        updated_by uuid
    );

    -- exercise_choices table
    create table if not exists exercise_choices (
        id uuid primary key default gen_random_uuid(),
        exercise_id uuid references exercises(id) not null,
        value varchar(10),
        is_solution boolean default false,
        created_at timestamp default now(),
        updated_at timestamp,
        created_by uuid not null,
        updated_by uuid
    );

    -- exercise_users table
    create table if not exists exercise_users (
        user_id uuid references users(id) not null,
        exercise_id uuid references exercises(id) not null,
        solved boolean default false not null,
        times_attempted int not null,
        created_at timestamp default now(),
        updated_at timestamp,
        primary key (user_id, exercise_id)
    );
    
COMMIT;
