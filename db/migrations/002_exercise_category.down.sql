BEGIN;

    -- changes to exercises table
    alter table exercises
    drop column category_iid,
    drop column exame,
    drop column fase;

    -- category table
    drop table if exists exercise_categories;
    
COMMIT;
