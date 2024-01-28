BEGIN;

    alter table exercises
    add category varchar(50);

    alter table exercises
    add exame varchar(25);

    alter table exercises
    add fase varchar(25);
    
COMMIT;
