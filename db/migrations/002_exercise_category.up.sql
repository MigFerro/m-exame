BEGIN;

    -- category table
    create table if not exists exercise_categories (
        iid serial primary key,
        category varchar(50) not null,
        created_at timestamp default now(),
        updated_at timestamp
    );

    -- changes to exercises table
    alter table exercises
    add category_iid integer references exercise_categories (iid);

    alter table exercises
    add exame varchar(25);

    alter table exercises
    add fase varchar(25);

    
COMMIT;
